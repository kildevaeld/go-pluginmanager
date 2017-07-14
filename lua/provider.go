package lua

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/aarzilli/golua/lua"
	"github.com/kildevaeld/go-pluginmanager"
	"github.com/mitchellh/mapstructure"
)

type LuaProviderOptions struct {
	Path    string
	Prelude func(state *lua.State) int
}

type LuaProvider struct {
	p []*LuaPlugin
	o LuaProviderOptions
}

func (l *LuaProvider) Open(path string) ([]pluginmanager.Plugin, error) {

	logger := zap.L().With(zap.String("prefix", "plugins:lua")).Sugar()

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() {
			logger.Debugf("skipping %s", file.Name())
			continue
		}

		pkgjson := filepath.Join(path, file.Name(), "plugin.json")

		bs, err := ioutil.ReadFile(pkgjson)
		if err != nil {
			logger.Debugf("skipping %s", file.Name())
			continue
		}
		var manifest pluginmanager.PluginManifest
		if err := json.Unmarshal(bs, &manifest); err != nil {
			logger.Warnf("skipping %s because of invalid manifest", file.Name())
			continue
		}

		manifest.Path = filepath.Join(path, file.Name())

		main := manifest.Main
		if main == "" {
			main = "main.lua"
		}

		mainlua := filepath.Join(path, file.Name(), main)

		if _, err := os.Stat(mainlua); err != nil {
			logger.Warnf("skipping %s: has now main file", file.Name())
		}

		plugin := &LuaPlugin{
			manifest: manifest,
			path:     mainlua,
			log:      logger.With(zap.String("plugin", manifest.Name)),
		}

		if err := plugin.Open(); err != nil {
			return nil, err
		}

		if l.o.Prelude != nil {
			s := plugin.l
			s.PushGoFunction(l.o.Prelude)
			if err := s.Call(0, 0); err != nil {
				return nil, err
			}
		}

		l.p = append(l.p, plugin)

	}

	var out []pluginmanager.Plugin
	for _, pl := range l.p {
		out = append(out, pl)
	}

	return out, nil
}

func (l *LuaProvider) Close() error {

	return nil
}

func NewProvider(options LuaProviderOptions) *LuaProvider {
	return &LuaProvider{nil, options}
}

func init() {

	pluginmanager.RegisterProvider("lua", func(o interface{}) (pluginmanager.PluginProvider, error) {

		var options LuaProviderOptions
		switch t := o.(type) {
		case LuaProviderOptions:
			options = t
		case *LuaProviderOptions:
			options = *t
		case map[string]interface{}:
			if err := mapstructure.Decode(&options, t); err != nil {
				return nil, err
			}
		}
		return NewProvider(options), nil
	})
}
