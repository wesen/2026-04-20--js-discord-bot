package botcli

import (
	"context"
	"fmt"

	glazed_cli "github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/spf13/cobra"
)

func botCLIParserConfig(appName string) glazed_cli.CobraParserConfig {
	return glazed_cli.CobraParserConfig{
		AppName:           appName,
		ShortHelpSections: []string{schema.DefaultSlug, schema.GlobalDefaultSlug},
	}
}

// NewBotsCommand builds the "bots" subcommand tree from a Bootstrap.
// It discovers bot implementations, scans for jsverbs metadata, and
// registers glazed commands for list, help, and any __verb__("run") verbs.
func NewBotsCommand(bootstrap Bootstrap, opts ...CommandOption) (*cobra.Command, error) {
	cfg, err := applyCommandOptions(opts...)
	if err != nil {
		return nil, err
	}
	parserConfig := botCLIParserConfig(cfg.appName)

	root := &cobra.Command{
		Use:   "bots",
		Short: "List, inspect, and run named JavaScript bot implementations",
	}

	// --- list command (GlazeCommand) ---
	listDesc := cmds.NewCommandDescription(
		"list",
		cmds.WithShort("List discovered bot implementations"),
		cmds.WithLong("Emit all discovered bots as a structured table."),
	)
	listCmd := &listBotsCommand{
		CommandDescription: listDesc,
		bootstrap:          bootstrap,
	}
	listCobra, err := glazed_cli.BuildCobraCommandFromCommand(listCmd,
		glazed_cli.WithParserConfig(parserConfig),
	)
	if err != nil {
		return nil, fmt.Errorf("build list command: %w", err)
	}
	root.AddCommand(listCobra)

	// --- help command (GlazeCommand) ---
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
	helpCmd := &helpBotsCommand{
		CommandDescription: helpDesc,
		bootstrap:          bootstrap,
	}
	helpCobra, err := glazed_cli.BuildCobraCommandFromCommand(helpCmd,
		glazed_cli.WithParserConfig(parserConfig),
	)
	if err != nil {
		return nil, fmt.Errorf("build help command: %w", err)
	}
	root.AddCommand(helpCobra)

	// --- scan for jsverbs and register discovered verbs ---
	if len(bootstrap.Repositories) == 0 {
		return root, nil
	}
	discoveredBots, err := DiscoverBots(context.Background(), bootstrap)
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
				discoveredCommands = append(discoveredCommands, &botRunCommand{
					CommandDescription: desc,
					scriptPath:         verb.File.AbsPath,
				})
				if bot, ok := botByScriptPath[verb.File.AbsPath]; ok {
					runCommandByBot[bot.Name()] = desc
				}
				continue
			}

			cmd, err := registry.CommandForVerbWithInvoker(verb, botVerbInvoker)
			if err != nil {
				return nil, fmt.Errorf("build jsverb command for %s: %w", verb.SourceRef(), err)
			}
			discoveredCommands = append(discoveredCommands, cmd)
		}
	}

	for _, discoveredBot := range discoveredBots {
		if _, ok := runCommandByBot[discoveredBot.Name()]; !ok {
			baseDesc := buildSyntheticBotRunDescription(discoveredBot, discoveredBot.Name())
			discoveredCommands = append(discoveredCommands, &botRunCommand{
				CommandDescription: baseDesc,
				scriptPath:         discoveredBot.ScriptPath(),
			})
			runCommandByBot[discoveredBot.Name()] = baseDesc
		}
	}

	for _, discoveredBot := range discoveredBots {
		baseDesc, ok := runCommandByBot[discoveredBot.Name()]
		if !ok {
			continue
		}
		aliasDesc := buildCompatibilityRunAliasDescription(baseDesc, discoveredBot.Name())
		discoveredCommands = append(discoveredCommands, &botRunCommand{
			CommandDescription: aliasDesc,
			scriptPath:         discoveredBot.ScriptPath(),
		})
	}

	if len(discoveredCommands) > 0 {
		if err := glazed_cli.AddCommandsToRootCommand(
			root,
			discoveredCommands,
			nil,
			glazed_cli.WithParserConfig(parserConfig),
		); err != nil {
			return nil, fmt.Errorf("register discovered bot verbs: %w", err)
		}
	}

	return root, nil
}

func NewCommand(bootstrap Bootstrap, opts ...CommandOption) *cobra.Command {
	cmd, err := NewBotsCommand(bootstrap, opts...)
	if err != nil {
		panic(err)
	}
	return cmd
}
