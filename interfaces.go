package pluginmanager

import (
	"encoding/json"

	args "github.com/kildevaeld/go-args"
)

type Hook int

const (
	InitializeHook Hook = iota + 1
	FinalizerHook
)

//type Map map[string]interface{}

type PluginManagerOptions struct {
	Path      string
	Providers map[string]interface{}
}

type PluginFactoryFunc func(plugin Plugin, hook Hook, method string, args args.Argument) (args.Argument, error)

type PluginManifest struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Type    string `json:"type"`
	Main    string `json:"main"`
	Path    string `json:"-"`

	Dependencies []string `json:"dependencies"`
	Features     map[string]json.RawMessage
}

type Plugin interface {
	Manifest() PluginManifest
	Initialize(PluginFactoryFunc, interface{}) error
	Shutdown(PluginFactoryFunc, interface{}) error
	Call(method string, args args.ArgumentList) (args.Argument, error)
}

type PluginProvider interface {
	Open(path string) ([]Plugin, error)
	Close() error
}
