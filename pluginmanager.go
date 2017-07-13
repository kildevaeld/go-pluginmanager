package pluginmanager

import (
	"errors"
	"fmt"
	"path/filepath"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/kildevaeld/go-args"
)

type PluginManager struct {
	providers map[string]PluginProvider
	plugins   []Plugin
	o         PluginManagerOptions
}

func (p *PluginManager) openProviders() error {
	if p.providers != nil {
		return errors.New("already open")
	}

	p.providers = make(map[string]PluginProvider)
	for name, pr := range p.o.Providers {
		if factory, ok := providers[name]; ok {
			provider, err := factory(pr)
			if err != nil {
				return fmt.Errorf("got error when create '%s': %s", name, err)
			}
			p.providers[name] = provider
		} else {
			return errors.New("provider not found: " + name)
		}
	}

	return nil
}

func (p *PluginManager) Open() error {

	if err := p.openProviders(); err != nil {
		return err
	}

	var result error
	var out []Plugin
	for _, plugin := range p.providers {

		if plugins, err := plugin.Open(p.o.Path); err != nil {
			result = multierror.Append(result, err)
		} else {
			out = append(out, plugins...)
		}
	}

	p.plugins = out

	return result
}

func (p *PluginManager) Close() error {
	var result error

	for _, plugin := range p.providers {
		if err := plugin.Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result
}

func (p *PluginManager) Initialize(fn PluginFactoryFunc, v interface{}) error {
	var result error
	for _, plugin := range p.plugins {
		if err := plugin.Initialize(fn, v); err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result
}

func (p *PluginManager) Shutdown(fn PluginFactoryFunc, v interface{}) error {
	var result error
	for _, plugin := range p.plugins {
		if err := plugin.Shutdown(fn, v); err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result
}

func (p *PluginManager) find(plugin string) Plugin {
	for _, pl := range p.plugins {
		if pl.Manifest().Name == plugin {
			return pl
		}
	}
	return nil
}

func (p *PluginManager) Call(plugin string, method string, args args.ArgumentList) (args.Argument, error) {
	pl := p.find(plugin)
	if pl == nil {
		return nil, errors.New("plugin not found")
	}
	return pl.Call(method, args)
}

func NewPluginManager(options PluginManagerOptions) (*PluginManager, error) {
	if options.Path == "" {
		return nil, errors.New("no path")
	}

	if !filepath.IsAbs(options.Path) {
		path, err := filepath.Abs(options.Path)
		if err != nil {
			return nil, err
		}
		options.Path = path
	}

	if options.Providers == nil {
		return nil, errors.New("no providers")
	}

	return &PluginManager{
		o: options,
	}, nil
}

var providers map[string]func(o interface{}) (PluginProvider, error)

func init() {
	providers = make(map[string]func(o interface{}) (PluginProvider, error))
}

func RegisterProvider(name string, fn func(o interface{}) (PluginProvider, error)) {
	providers[name] = fn
}
