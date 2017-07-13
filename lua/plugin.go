package lua

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	"github.com/aarzilli/golua/lua"
	"github.com/kildevaeld/go-args"
	"github.com/kildevaeld/goluaext"
	"github.com/kildevaeld/pluginmanager"
	"github.com/stevedonovan/luar"
)

type LuaPlugin struct {
	manifest    pluginmanager.PluginManifest
	l           *lua.State
	path        string
	log         *zap.SugaredLogger
	initializer *luar.LuaObject
	finalizer   *luar.LuaObject
	fn          pluginmanager.PluginFactoryFunc
}

func argument(t args.Type, v interface{}) args.Argument {
	return args.NewOrNil(v)
}

func closeArguments(args []args.Argument) {
	//for _, a := range args {
	/*if i, o := a.(*call_argument); o {
		i.v.Close()
	}*/
	//}
}

func createSandbox(hook pluginmanager.Hook, plugin *LuaPlugin, state *lua.State) {

	initIndex := state.GetTop()

	CreateSandbox(state, MetaMap{
		"__newindex": func(state *lua.State) int {

			if plugin.fn == nil {
				panic(errors.New("called outside factory"))
			}

			name := state.ToString(2)

			args, err := goluaext.LuaToArgument(state, 3, false)
			if err != nil {
				panic(err)
			}

			returnCount := 0
			if out, err := plugin.fn(plugin, hook, name, args); err != nil {
				//closeArguments(args)
				args.Free()
				panic(err)
			} else if out != nil {
				goluaext.PushArgument(state, out)
				returnCount = 1
			}
			//closeArguments(args)
			return returnCount
		},
		"__index": func(state *lua.State) int {
			name := state.ToString(2)

			state.PushGoFunction(func(state *lua.State) int {

				if plugin.fn == nil {
					panic(errors.New("called outside factory"))
				}

				var arguments args.Argument
				if state.IsFunction(1) {
					a, err := createBuilder(state)
					if err != nil {
						panic(err)
					}
					arguments = a

				} else {
					a, err := goluaext.LuaToArgument(state, 1, true)
					if err != nil {
						panic(err)
					}
					arguments = a
				}

				returnCount := 0
				if out, err := plugin.fn(plugin, hook, name, arguments); err != nil {
					arguments.Free()
					panic(err)
				} else if out != nil {
					if err := goluaext.PushArgument(state, out); err != nil {
						out.Free()
						panic(err)
					}
					out.Free()
					returnCount = 1
				}
				return returnCount
			})

			return 1
		},
	})

	state.SetfEnv(initIndex)
	/*initIndex := state.GetTop()

	state.CreateTable(0, 2)
	state.GetGlobal("print")
	state.SetField(-2, "print")
	state.GetGlobal("util")
	state.SetField(-2, "util")

	state.CreateTable(0, 2)
	state.SetMetaMethod("__newindex", func(state *lua.State) int {

		if plugin.fn == nil {
			panic(errors.New("called outside factory"))
		}

		name := state.ToString(2)

		args, err := goluaext.LuaToArgument(state, 3)
		if err != nil {
			panic(err)
		}

		returnCount := 0
		if out, err := plugin.fn(plugin, hook, name, args); err != nil {
			//closeArguments(args)
			args.Free()
			panic(err)
		} else if out != nil {
			goluaext.PushArgument(state, out)
			returnCount = 1
		}
		//closeArguments(args)
		return returnCount
	})

	state.SetMetaMethod("__index", func(state *lua.State) int {
		name := state.ToString(2)

		state.PushGoFunction(func(state *lua.State) int {

			if plugin.fn == nil {
				panic(errors.New("called outside factory"))
			}

			var arguments args.Argument
			if state.IsFunction(1) {
				a, err := createBuilder(state)
				if err != nil {
					panic(err)
				}
				arguments = a

			} else {
				a, err := goluaext.LuaToArgument(state, 1)
				if err != nil {
					panic(err)
				}
				arguments = a
			}

			returnCount := 0
			if out, err := plugin.fn(plugin, hook, name, arguments); err != nil {
				arguments.Free()
				panic(err)
			} else if out != nil {
				if err := goluaext.PushArgument(state, out); err != nil {
					out.Free()
					panic(err)
				}
				out.Free()
				returnCount = 1
			}
			return returnCount
		})

		return 1
	})
	state.SetMetaTable(-2)
	state.SetfEnv(initIndex)*/
}

func (l *LuaPlugin) init() error {
	top := l.l.GetTop()

	l.l.GetField(top, "initialize")
	if !l.l.IsFunction(-1) {
		return errors.New("no initializer")
	}

	createSandbox(pluginmanager.InitializeHook, l, l.l)
	l.initializer = luar.NewLuaObject(l.l, -1)

	l.l.GetField(top, "shutdown")
	if l.l.IsFunction(-1) {
		createSandbox(pluginmanager.FinalizerHook, l, l.l)
		l.finalizer = luar.NewLuaObject(l.l, -1)
	}

	return nil
}

func (l *LuaPlugin) Open() error {

	if l.l != nil {
		return errors.New("already open")
	}

	oldLuaPath := os.Getenv("LUA_PATH")
	luaPath := oldLuaPath
	if luaPath != "" {
		luaPath += ";"
	}

	luaPath += filepath.Dir(l.path) + "/?.lua"

	os.Setenv("LUA_PATH", luaPath)
	l.l = goluaext.Init()
	os.Setenv("LUA_PATH", oldLuaPath)

	then := time.Now()
	l.log.Debugw("opening file", zap.String("path", l.path), zap.String("lua_path", luaPath))
	if err := l.l.DoFile(l.path); err != nil {
		return err
	}

	if !l.l.IsTable(-1) {
		return errors.New("invalid return type")
	}

	if err := l.init(); err != nil {
		return err
	}

	l.log.Debugw("file opened", zap.Duration("elapsed", time.Since(then)))
	return nil
}

func (l *LuaPlugin) Close() error {
	if l.l == nil {
		return errors.New("already closed")
	}

	if l.initializer != nil {
		l.initializer.Close()
	}

	l.l.Close()

	return nil
}

func (l *LuaPlugin) Manifest() pluginmanager.PluginManifest {
	return l.manifest
}
func (l *LuaPlugin) Initialize(fn pluginmanager.PluginFactoryFunc, v interface{}) error {

	l.fn = fn
	top := l.l.GetTop()
	l.initializer.Push()
	callCount := 0

	a, e := args.New(v)
	if e != nil {
		l.l.SetTop(top)
		return fmt.Errorf("could not converter to argument: %s", e)
	}
	if a != nil {
		callCount = 1
		if err := goluaext.PushArgument(l.l, a); err != nil {
			l.l.SetTop(top)
			l.fn = nil
			return err
		}

	}

	if err := l.l.Call(callCount, 0); err != nil {
		l.fn = nil
		return err
	}

	l.fn = nil

	return nil
}
func (l *LuaPlugin) Shutdown(fn pluginmanager.PluginFactoryFunc, v interface{}) error {
	if l.finalizer != nil {
		l.fn = fn
		err := l.finalizer.Call(nil)
		l.fn = nil
		return err
	}
	return nil
}
func (l *LuaPlugin) Call(method string, a args.ArgumentList) (args.Argument, error) {
	return nil, nil
}
