package botcli

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/go-go-goja/engine"
	"github.com/manuel/wesen/2026-04-20--js-discord-bot/internal/jsdiscord"
)

type commandOptions struct {
	appName                 string
	runtimeModuleRegistrars []engine.RuntimeModuleRegistrar
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

func (cfg commandOptions) hostOptions() []jsdiscord.HostOption {
	if len(cfg.runtimeModuleRegistrars) == 0 {
		return nil
	}
	return []jsdiscord.HostOption{jsdiscord.WithRuntimeModuleRegistrars(cfg.runtimeModuleRegistrars...)}
}
