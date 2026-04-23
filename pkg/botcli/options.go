package botcli

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-go-golems/go-go-goja/engine"
	"github.com/go-go-golems/go-go-goja/pkg/jsverbs"
	"github.com/manuel/wesen/2026-04-20--js-discord-bot/internal/jsdiscord"
)

type RuntimeFactory interface {
	NewRuntimeForVerb(ctx context.Context, registry *jsverbs.Registry, verb *jsverbs.VerbSpec) (*engine.Runtime, error)
}

type RuntimeFactoryFunc func(ctx context.Context, registry *jsverbs.Registry, verb *jsverbs.VerbSpec) (*engine.Runtime, error)

func (f RuntimeFactoryFunc) NewRuntimeForVerb(ctx context.Context, registry *jsverbs.Registry, verb *jsverbs.VerbSpec) (*engine.Runtime, error) {
	return f(ctx, registry, verb)
}

type HostOptionsProvider interface {
	HostOptions() []jsdiscord.HostOption
}

type commandOptions struct {
	appName                 string
	runtimeModuleRegistrars []engine.RuntimeModuleRegistrar
	runtimeFactory          RuntimeFactory
}

type CommandOption func(*commandOptions) error

func defaultCommandOptions() commandOptions {
	return commandOptions{appName: "discord"}
}

func applyCommandOptions(opts ...CommandOption) (commandOptions, error) {
	cfg := defaultCommandOptions()
	for i, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&cfg); err != nil {
			return commandOptions{}, fmt.Errorf("apply command option %d: %w", i, err)
		}
	}
	return cfg, nil
}

// WithAppName controls the env prefix used by dynamic Glazed bot commands.
func WithAppName(name string) CommandOption {
	return func(cfg *commandOptions) error {
		name = strings.TrimSpace(name)
		if name == "" {
			return fmt.Errorf("app name is empty")
		}
		cfg.appName = name
		return nil
	}
}

// WithRuntimeModuleRegistrars appends custom runtime-scoped native module registrars.
func WithRuntimeModuleRegistrars(registrars ...engine.RuntimeModuleRegistrar) CommandOption {
	return func(cfg *commandOptions) error {
		for i, registrar := range registrars {
			if registrar == nil {
				return fmt.Errorf("runtime module registrar at index %d is nil", i)
			}
		}
		cfg.runtimeModuleRegistrars = append(cfg.runtimeModuleRegistrars, registrars...)
		return nil
	}
}

// WithRuntimeFactory overrides ordinary jsverbs runtime creation and may also
// contribute host options for discovery and host-managed bot runs when the
// factory implements HostOptionsProvider.
func WithRuntimeFactory(factory RuntimeFactory) CommandOption {
	return func(cfg *commandOptions) error {
		if factory == nil {
			return fmt.Errorf("runtime factory is nil")
		}
		cfg.runtimeFactory = factory
		return nil
	}
}

func (cfg commandOptions) hostOptions() []jsdiscord.HostOption {
	ret := []jsdiscord.HostOption{}
	if len(cfg.runtimeModuleRegistrars) > 0 {
		ret = append(ret, jsdiscord.WithRuntimeModuleRegistrars(cfg.runtimeModuleRegistrars...))
	}
	if provider, ok := cfg.runtimeFactory.(HostOptionsProvider); ok {
		ret = append(ret, provider.HostOptions()...)
	}
	return ret
}
