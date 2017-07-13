package pluginmanager

import args "github.com/kildevaeld/go-args"

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
	Name string
	Type string
	Main string
	Path string
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
