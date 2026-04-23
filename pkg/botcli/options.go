package botcli

import (
	"fmt"
	"strings"

	internalbotcli "github.com/manuel/wesen/2026-04-20--js-discord-bot/internal/botcli"
)

type commandOptions struct {
	appName string
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

func toInternalCommandOptions(cfg commandOptions) []internalbotcli.CommandOption {
	return []internalbotcli.CommandOption{
		internalbotcli.WithAppName(cfg.appName),
	}
}
