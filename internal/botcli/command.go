package botcli

import (
	"fmt"

	glazed_cli "github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/spf13/cobra"
)

// NewBotsCommand builds the "bots" subcommand tree from a Bootstrap.
// It discovers bot implementations, scans for jsverbs metadata, and
// registers glazed commands for list, help, and any __verb__("run") verbs.
func NewBotsCommand(bootstrap Bootstrap) (*cobra.Command, error) {
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
		glazed_cli.WithParserConfig(glazed_cli.CobraParserConfig{
			ShortHelpSections: []string{schema.DefaultSlug, schema.GlobalDefaultSlug},
		}),
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
		glazed_cli.WithParserConfig(glazed_cli.CobraParserConfig{
			ShortHelpSections: []string{schema.DefaultSlug, schema.GlobalDefaultSlug},
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("build help command: %w", err)
	}
	root.AddCommand(helpCobra)

	// --- scan for jsverbs and register discovered verbs ---
	if len(bootstrap.Repositories) == 0 {
		return root, nil
	}
	registries, err := ScanBotRepositories(bootstrap.Repositories)
	if err != nil {
		return nil, fmt.Errorf("scan bot repositories: %w", err)
	}

	discoveredCommands := make([]cmds.Command, 0)
	for _, registry := range registries {
		for _, verb := range registry.Verbs() {
			if verb.Name == "run" {
				desc, err := registry.CommandDescriptionForVerb(verb)
				if err != nil {
					return nil, fmt.Errorf("build run command description for %s: %w", verb.SourceRef(), err)
				}
				discoveredCommands = append(discoveredCommands, &botRunCommand{
					CommandDescription: desc,
					scriptPath:         verb.File.AbsPath,
				})
				continue
			}

			cmd, err := registry.CommandForVerbWithInvoker(verb, botVerbInvoker)
			if err != nil {
				return nil, fmt.Errorf("build jsverb command for %s: %w", verb.SourceRef(), err)
			}
			discoveredCommands = append(discoveredCommands, cmd)
		}
	}

	if len(discoveredCommands) > 0 {
		if err := glazed_cli.AddCommandsToRootCommand(
			root,
			discoveredCommands,
			nil,
			glazed_cli.WithParserConfig(glazed_cli.CobraParserConfig{
				ShortHelpSections: []string{schema.DefaultSlug, schema.GlobalDefaultSlug},
			}),
		); err != nil {
			return nil, fmt.Errorf("register discovered bot verbs: %w", err)
		}
	}

	return root, nil
}

// NewCommand creates a BotsCommand. When bootstrap is omitted, an empty bootstrap is used.
func NewCommand(bootstrap ...Bootstrap) *cobra.Command {
	b := Bootstrap{}
	if len(bootstrap) > 0 {
		b = bootstrap[0]
	}
	cmd, err := NewBotsCommand(b)
	if err != nil {
		panic(err)
	}
	return cmd
}
