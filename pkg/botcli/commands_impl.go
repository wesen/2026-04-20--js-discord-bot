package botcli

import (
	"context"
	"fmt"
	"strings"

	glazed_cli "github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/spf13/cobra"

	"github.com/manuel/wesen/2026-04-20--js-discord-bot/internal/bot"
	appconfig "github.com/manuel/wesen/2026-04-20--js-discord-bot/internal/config"
	"github.com/manuel/wesen/2026-04-20--js-discord-bot/internal/jsdiscord"
)

func botCLIParserConfig(appName string) glazed_cli.CobraParserConfig {
	return glazed_cli.CobraParserConfig{
		AppName:           appName,
		ShortHelpSections: []string{schema.DefaultSlug, schema.GlobalDefaultSlug},
	}
}

type listBotsCommand struct {
	*cmds.CommandDescription
	bootstrap Bootstrap
	hostOpts  []jsdiscord.HostOption
}

func (c *listBotsCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	bots, err := DiscoverBots(ctx, c.bootstrap, c.hostOpts...)
	if err != nil {
		return err
	}
	for _, bot := range bots {
		row := types.NewRow(
			types.MRP("name", bot.Name()),
			types.MRP("source", bot.SourceLabel()),
			types.MRP("description", bot.Description()),
			types.MRP("commands", len(bot.Descriptor.Commands)),
			types.MRP("events", len(bot.Descriptor.Events)),
			types.MRP("components", len(bot.Descriptor.Components)),
			types.MRP("modals", len(bot.Descriptor.Modals)),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

type helpBotsCommand struct {
	*cmds.CommandDescription
	bootstrap Bootstrap
	hostOpts  []jsdiscord.HostOption
}

func (c *helpBotsCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	botName := ""
	if argSection, ok := vals.Get(schema.DefaultSlug); ok && argSection.Fields != nil {
		if fv, ok := argSection.Fields.Get("bot-name"); ok && fv != nil {
			botName = fmt.Sprint(fv.Value)
		}
	}
	botName = strings.TrimSpace(botName)
	if botName == "" {
		return fmt.Errorf("bot name is required")
	}

	bots, err := DiscoverBots(ctx, c.bootstrap, c.hostOpts...)
	if err != nil {
		return err
	}
	selected, err := ResolveBot(botName, bots)
	if err != nil {
		return err
	}

	for _, cmd := range selected.Descriptor.Commands {
		row := types.NewRow(
			types.MRP("kind", "command"),
			types.MRP("name", cmd.Name),
			types.MRP("description", cmd.Description),
			types.MRP("type", cmd.Type),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	for _, ev := range selected.Descriptor.Events {
		row := types.NewRow(types.MRP("kind", "event"), types.MRP("name", ev.Name), types.MRP("description", ""), types.MRP("type", ""))
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	for _, comp := range selected.Descriptor.Components {
		row := types.NewRow(types.MRP("kind", "component"), types.MRP("name", comp.CustomID), types.MRP("description", ""), types.MRP("type", ""))
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	for _, modal := range selected.Descriptor.Modals {
		row := types.NewRow(types.MRP("kind", "modal"), types.MRP("name", modal.CustomID), types.MRP("description", ""), types.MRP("type", ""))
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

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
	syncOnStart := false
	if fv, ok := parsedValues.GetField(schema.DefaultSlug, "sync-on-start"); ok && fv != nil && fv.Value != nil {
		if value, ok := fv.Value.(bool); ok {
			syncOnStart = value
		}
	}

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

// NewBotsCommand builds the public repo-driven bot command tree from a bootstrap.
func NewBotsCommand(bootstrap Bootstrap, opts ...CommandOption) (*cobra.Command, error) {
	cfg, err := applyCommandOptions(opts...)
	if err != nil {
		return nil, err
	}
	parserConfig := botCLIParserConfig(cfg.appName)
	hostOpts := cfg.hostOptions()

	root := &cobra.Command{
		Use:   "bots",
		Short: "List, inspect, and run named JavaScript bot implementations",
	}

	listDesc := cmds.NewCommandDescription(
		"list",
		cmds.WithShort("List discovered bot implementations"),
		cmds.WithLong("Emit all discovered bots as a structured table."),
	)
	listCmd := &listBotsCommand{CommandDescription: listDesc, bootstrap: bootstrap, hostOpts: hostOpts}
	listCobra, err := glazed_cli.BuildCobraCommandFromCommand(listCmd, glazed_cli.WithParserConfig(parserConfig))
	if err != nil {
		return nil, fmt.Errorf("build list command: %w", err)
	}
	root.AddCommand(listCobra)

	helpDesc := cmds.NewCommandDescription(
		"help",
		cmds.WithShort("Show metadata for one discovered bot implementation"),
		cmds.WithLong("Emit bot metadata (commands, events, components, modals) as structured rows."),
		cmds.WithArguments(
			fields.New("bot-name", fields.TypeString,
				fields.WithIsArgument(true),
				fields.WithRequired(true),
				fields.WithHelp("Bot name or source label"),
			),
		),
	)
	helpCmd := &helpBotsCommand{CommandDescription: helpDesc, bootstrap: bootstrap, hostOpts: hostOpts}
	helpCobra, err := glazed_cli.BuildCobraCommandFromCommand(helpCmd, glazed_cli.WithParserConfig(parserConfig))
	if err != nil {
		return nil, fmt.Errorf("build help command: %w", err)
	}
	root.AddCommand(helpCobra)

	if len(bootstrap.Repositories) == 0 {
		return root, nil
	}
	discoveredBots, err := DiscoverBots(context.Background(), bootstrap, hostOpts...)
	if err != nil {
		return nil, fmt.Errorf("discover bots: %w", err)
	}
	botByScriptPath := map[string]DiscoveredBot{}
	for _, discoveredBot := range discoveredBots {
		botByScriptPath[discoveredBot.ScriptPath()] = discoveredBot
	}

	registries, err := ScanBotRepositories(bootstrap.Repositories)
	if err != nil {
		return nil, fmt.Errorf("scan bot repositories: %w", err)
	}

	runCommandByBot := map[string]*cmds.CommandDescription{}
	discoveredCommands := make([]cmds.Command, 0)
	for _, registry := range registries {
		for _, verb := range registry.Verbs() {
			if verb.Name == "run" {
				baseDesc, err := registry.CommandDescriptionForVerb(verb)
				if err != nil {
					return nil, fmt.Errorf("build run command description for %s: %w", verb.SourceRef(), err)
				}
				desc := ensureRunCommandDefaults(baseDesc)
				discoveredCommands = append(discoveredCommands, &botRunCommand{CommandDescription: desc, scriptPath: verb.File.AbsPath, hostOpts: hostOpts})
				if bot, ok := botByScriptPath[verb.File.AbsPath]; ok {
					runCommandByBot[bot.Name()] = desc
				}
				continue
			}

			cmd, err := registry.CommandForVerbWithInvoker(verb, makeBotVerbInvoker(cfg))
			if err != nil {
				return nil, fmt.Errorf("build jsverb command for %s: %w", verb.SourceRef(), err)
			}
			discoveredCommands = append(discoveredCommands, cmd)
		}
	}

	for _, discoveredBot := range discoveredBots {
		if _, ok := runCommandByBot[discoveredBot.Name()]; !ok {
			baseDesc := buildSyntheticBotRunDescription(discoveredBot, discoveredBot.Name())
			discoveredCommands = append(discoveredCommands, &botRunCommand{CommandDescription: baseDesc, scriptPath: discoveredBot.ScriptPath(), hostOpts: hostOpts})
			runCommandByBot[discoveredBot.Name()] = baseDesc
		}
	}

	for _, discoveredBot := range discoveredBots {
		baseDesc, ok := runCommandByBot[discoveredBot.Name()]
		if !ok {
			continue
		}
		aliasDesc := buildCompatibilityRunAliasDescription(baseDesc, discoveredBot.Name())
		discoveredCommands = append(discoveredCommands, &botRunCommand{CommandDescription: aliasDesc, scriptPath: discoveredBot.ScriptPath(), hostOpts: hostOpts})
	}

	if len(discoveredCommands) > 0 {
		if err := glazed_cli.AddCommandsToRootCommand(root, discoveredCommands, nil, glazed_cli.WithParserConfig(parserConfig)); err != nil {
			return nil, fmt.Errorf("register discovered bot verbs: %w", err)
		}
	}

	return root, nil
}
