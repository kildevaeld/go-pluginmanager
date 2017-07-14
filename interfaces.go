package pluginmanager

import (
	"errors"

	args "github.com/kildevaeld/go-args"
	"github.com/mitchellh/mapstructure"
)

type Hook int

const (
	InitializeHook Hook = iota + 1
	FinalizerHook
)

//type Map map[string]interface{}

type PluginFactoryFunc func(plugin Plugin, hook Hook, method string, args args.Argument) (args.Argument, error)

type PluginManifest struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Type    string `json:"type"`
	Main    string `json:"main"`
	Path    string `json:"-"`

	Dependencies []string `json:"dependencies"`
	Features     map[string]map[string]interface{}
}

func (p PluginManifest) GetFeature(name string, out interface{}) error {
	if p.Features[name] == nil {
		return errors.New("invalid feature")
	}

	return mapstructure.Decode(p.Features[name], out)
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
