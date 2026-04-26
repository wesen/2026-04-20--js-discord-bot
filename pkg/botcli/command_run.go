package botcli

import (
	"context"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"

	"github.com/go-go-golems/discord-bot/internal/bot"
	appconfig "github.com/go-go-golems/discord-bot/internal/config"
	"github.com/go-go-golems/discord-bot/internal/jsdiscord"
)

type botRunCommand struct {
	*cmds.CommandDescription
	scriptPath string
	hostOpts   []jsdiscord.HostOption
}

func (c *botRunCommand) Run(ctx context.Context, parsedValues *values.Values) error {
	cfg, err := appconfig.FromValues(parsedValues)
	if err != nil {
		return fmt.Errorf("decode discord settings: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return err
	}

	runtimeConfig := buildRuntimeConfig(parsedValues)
	syncOnStart := boolField(parsedValues, schema.DefaultSlug, "sync-on-start")

	b, err := bot.NewWithScript(cfg, c.scriptPath, runtimeConfig, c.hostOpts...)
	if err != nil {
		return err
	}
	defer func() { _ = b.Close() }()

	if syncOnStart {
		if _, err := b.SyncCommands(); err != nil {
			return err
		}
	}
	if err := b.Open(); err != nil {
		return err
	}
	<-ctx.Done()
	return nil
}
