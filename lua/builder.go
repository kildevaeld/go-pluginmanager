package lua

import (
	"errors"

	"github.com/aarzilli/golua/lua"
	"github.com/kildevaeld/go-args"
	"github.com/kildevaeld/goluaext"
)

var globals = []string{"print", "util", "tostring", "tonumber"}

func PushBuilder(state *lua.State, fn func(name string, args args.Argument) (args.Argument, error)) error {

	CreateSandbox(state, MetaMap{
		"__index": func(state *lua.State) int {
			name := state.ToString(2)

			state.PushGoFunction(func(state *lua.State) int {

				arguments, err := goluaext.LuaToArgument(state, 1, true)
				if err != nil {
					panic(err)
				}

				if arguments.Is(args.ArgumentSliceType) {
					arguments = args.Must(args.ArgumentList(arguments.Value().([]args.Argument)))
				}

				returnCount := 0
				if out, err := fn(name, arguments); err != nil {
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
		"__newindex": func(state *lua.State) int {

			name := state.ToString(2)

			args, err := goluaext.LuaToArgument(state, 3, false)
			if err != nil {
				panic(err)
			}

			returnCount := 0
			if out, err := fn(name, args); err != nil {
				args.Free()
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
		},
		"__tostring": func(state *lua.State) int {
			state.PushString("Builder")
			return 1
		},
	})

	return nil

	/*state.CreateTable(0, len(globals))

	for _, g := range globals {
		state.GetGlobal(g)
		state.SetField(-2, g)
	}

	state.CreateTable(0, 3)
	state.SetMetaMethod("__index", func(state *lua.State) int {
		name := state.ToString(2)

		state.PushGoFunction(func(state *lua.State) int {

			arguments, err := goluaext.LuaToArgument(state, 1)
			if err != nil {
				panic(err)
			}

			returnCount := 0
			if out, err := fn(name, arguments); err != nil {
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
	state.SetMetaMethod("__newindex", func(state *lua.State) int {

		name := state.ToString(2)

		args, err := goluaext.LuaToArgument(state, 3)
		if err != nil {
			panic(err)
		}

		returnCount := 0
		if out, err := fn(name, args); err != nil {
			args.Free()
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

	state.SetMetaMethod("__tostring", func(state *lua.State) int {
		state.PushString("Builder")
		return 1
	})

	state.SetMetaTable(-2)

	return nil*/
}

func createBuilder(state *lua.State) (args.Argument, error) {

	top := state.GetTop()

	if !state.IsFunction(top) {
		return nil, errors.New("invalid")
	}

	out := args.ArgumentMap{}

	PushBuilder(state, func(name string, a args.Argument) (args.Argument, error) {
		if a.Is(args.ArgumentListType) {
			i := a.Value().(args.ArgumentList)
			if i.Len() == 1 {
				a = i[0]
			}

		}

		out[name] = a
		return nil, nil
	})

	state.SetfEnv(top)

	if err := state.Call(0, 0); err != nil {
		return nil, err
	}

	return args.New(out)

}
