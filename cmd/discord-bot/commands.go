package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/spf13/cobra"

	botapp "github.com/manuel/wesen/2026-04-20--js-discord-bot/internal/bot"
	appconfig "github.com/manuel/wesen/2026-04-20--js-discord-bot/internal/config"
)

type runCommand struct {
	*cmds.CommandDescription
}

type validateConfigCommand struct {
	*cmds.CommandDescription
}

type syncCommandsCommand struct {
	*cmds.CommandDescription
}

func newDiscordCommandDescription(name, short, long string) (*cmds.CommandDescription, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}

	commandSettingsSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}

	return cmds.NewCommandDescription(
		name,
		cmds.WithShort(short),
		cmds.WithLong(long),
		cmds.WithFlags(
			fields.New("bot-token", fields.TypeString, fields.WithHelp("Discord bot token from the Developer Portal")),
			fields.New("application-id", fields.TypeString, fields.WithHelp("Discord application/client ID")),
			fields.New("guild-id", fields.TypeString, fields.WithHelp("Optional guild ID for fast local slash-command sync")),
			fields.New("public-key", fields.TypeString, fields.WithHelp("Discord application public key (only needed for HTTP interactions)")),
			fields.New("client-id", fields.TypeString, fields.WithHelp("Discord client ID")),
			fields.New("client-secret", fields.TypeString, fields.WithHelp("Discord client secret")),
			fields.New("bot-script", fields.TypeString, fields.WithHelp("Optional path to a JavaScript Discord bot script loaded through goja")),
		),
		cmds.WithSections(glazedSection, commandSettingsSection),
	), nil
}

func newRunCommand() (*runCommand, error) {
	desc, err := newDiscordCommandDescription(
		"run",
		"Start the Discord bot",
		`Start the Discord bot, connect to the gateway, and wait for shutdown.

Examples:
  discord-bot run
  discord-bot run --log-level debug
`,
	)
	if err != nil {
		return nil, err
	}
	return &runCommand{CommandDescription: desc}, nil
}

func newValidateConfigCommand() (*validateConfigCommand, error) {
	desc, err := newDiscordCommandDescription(
		"validate-config",
		"Validate Discord configuration",
		`Validate the Discord-related settings before connecting to the gateway.

Examples:
  discord-bot validate-config
`,
	)
	if err != nil {
		return nil, err
	}
	return &validateConfigCommand{CommandDescription: desc}, nil
}

func newSyncCommandsCommand() (*syncCommandsCommand, error) {
	desc, err := newDiscordCommandDescription(
		"sync-commands",
		"Register slash commands",
		`Register or replace the bot's slash commands.

Examples:
  discord-bot sync-commands
  discord-bot sync-commands --guild-id 123456789012345678
`,
	)
	if err != nil {
		return nil, err
	}
	return &syncCommandsCommand{CommandDescription: desc}, nil
}

func (c *runCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	cfg, err := appconfig.FromValues(vals)
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}

	bot, err := botapp.New(cfg)
	if err != nil {
		return err
	}
	defer func() { _ = bot.Close() }()

	if err := bot.Open(); err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("status", "running"),
		types.MRP("application_id", cfg.ApplicationID),
		types.MRP("guild_id", strings.TrimSpace(cfg.GuildID)),
		types.MRP("bot_script", strings.TrimSpace(cfg.BotScript)),
		types.MRP("token", cfg.RedactedToken()),
	)
	if err := gp.AddRow(ctx, row); err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}

func (c *validateConfigCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	_ = ctx

	cfg, err := appconfig.FromValues(vals)
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("configured", true),
		types.MRP("application_id", cfg.ApplicationID),
		types.MRP("guild_id", strings.TrimSpace(cfg.GuildID)),
		types.MRP("bot_script", strings.TrimSpace(cfg.BotScript)),
		types.MRP("bot_token_present", strings.TrimSpace(cfg.BotToken) != ""),
		// Keep secrets out of output; show only the redacted token marker.
		types.MRP("token", cfg.RedactedToken()),
	)
	return gp.AddRow(ctx, row)
}

func (c *syncCommandsCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	_ = ctx

	cfg, err := appconfig.FromValues(vals)
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}

	bot, err := botapp.New(cfg)
	if err != nil {
		return err
	}
	defer func() { _ = bot.Close() }()

	commands, err := bot.SyncCommands()
	if err != nil {
		return err
	}

	for _, command := range commands {
		row := types.NewRow(
			types.MRP("command", command.Name),
			types.MRP("description", command.Description),
			types.MRP("scope", syncScope(cfg)),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func syncScope(cfg appconfig.Settings) string {
	if strings.TrimSpace(cfg.GuildID) != "" {
		return fmt.Sprintf("guild:%s", strings.TrimSpace(cfg.GuildID))
	}
	return "global"
}

func buildCobraCommand(cmd cmds.GlazeCommand) (*cobra.Command, error) {
	return cli.BuildCobraCommandFromCommand(cmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			AppName:           "discord",
			ShortHelpSections: []string{schema.DefaultSlug},
		}),
	)
}
