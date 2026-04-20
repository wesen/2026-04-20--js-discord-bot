package botcli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "bots",
		Short: "List, inspect, and run named JavaScript bot implementations",
	}
	root.PersistentFlags().StringArray(BotRepositoryFlag, nil, "Directory containing named JavaScript bot implementations (repeatable)")
	root.AddCommand(newListCommand())
	root.AddCommand(newHelpCommand())
	root.AddCommand(newRunCommand())
	return root
}

func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List discovered bot implementations",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = args
			bootstrap, err := DiscoverBootstrapFromCommand(cmd)
			if err != nil {
				return err
			}
			bots, err := DiscoverBots(cmd.Context(), bootstrap)
			if err != nil {
				return err
			}
			for _, bot := range bots {
				line := bot.Name() + "\t" + bot.SourceLabel()
				if desc := bot.Description(); desc != "" {
					line += "\t" + desc
				}
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), line); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func newHelpCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "help <bot>",
		Short: "Show metadata for one discovered bot implementation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bootstrap, err := DiscoverBootstrapFromCommand(cmd)
			if err != nil {
				return err
			}
			bots, err := DiscoverBots(cmd.Context(), bootstrap)
			if err != nil {
				return err
			}
			bot, err := ResolveBot(args[0], bots)
			if err != nil {
				return err
			}
			return printBotHelp(cmd, bot)
		},
	}
}

func printBotHelp(cmd *cobra.Command, bot DiscoveredBot) error {
	out := cmd.OutOrStdout()
	if _, err := fmt.Fprintf(out, "Bot: %s\n", bot.Name()); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "Source: %s\n", bot.ScriptPath()); err != nil {
		return err
	}
	if desc := bot.Description(); desc != "" {
		if _, err := fmt.Fprintf(out, "Description: %s\n", desc); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(out, "Commands:"); err != nil {
		return err
	}
	if len(bot.Descriptor.Commands) == 0 {
		if _, err := fmt.Fprintln(out, "  (none)"); err != nil {
			return err
		}
	} else {
		for _, command := range bot.Descriptor.Commands {
			line := "  - " + command.Name
			if command.Description != "" {
				line += " — " + command.Description
			}
			if _, err := fmt.Fprintln(out, line); err != nil {
				return err
			}
		}
	}
	if _, err := fmt.Fprintln(out, "Events:"); err != nil {
		return err
	}
	if len(bot.Descriptor.Events) == 0 {
		if _, err := fmt.Fprintln(out, "  (none)"); err != nil {
			return err
		}
	} else {
		for _, event := range bot.Descriptor.Events {
			if _, err := fmt.Fprintf(out, "  - %s\n", event.Name); err != nil {
				return err
			}
		}
	}
	if err := printRunSchema(out, bot); err != nil {
		return err
	}
	return nil
}

func newRunCommand() *cobra.Command {
	var (
		botToken          string
		applicationID     string
		guildID           string
		publicKey         string
		clientID          string
		clientSecret      string
		syncOnStart       bool
		printParsedValues bool
	)
	cmd := &cobra.Command{
		Use:                "run <bot>",
		Short:              "Run one named bot implementation through the live Discord host",
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			parsed, err := preparseRunArgs(args, defaultPreParsedRunArgs())
			if err != nil {
				return err
			}
			if parsed.ShowHelp && strings.TrimSpace(parsed.Selector) == "" {
				return cmd.Help()
			}
			if parsed.ShowHelp {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("resolve cwd: %w", err)
				}
				repositories, err := repositoriesFromCLIPaths(parsed.BotRepositories, cwd)
				if err != nil {
					return err
				}
				bootstrap, err := bootstrapFromRepositories(repositories)
				if err != nil {
					return err
				}
				bots, err := DiscoverBots(cmd.Context(), bootstrap)
				if err != nil {
					return err
				}
				selected, err := ResolveBot(parsed.Selector, bots)
				if err != nil {
					return err
				}
				return printBotHelp(cmd, selected)
			}
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("resolve cwd: %w", err)
			}
			repositories, err := repositoriesFromCLIPaths(parsed.BotRepositories, cwd)
			if err != nil {
				return err
			}
			bootstrap, err := bootstrapFromRepositories(repositories)
			if err != nil {
				return err
			}
			bots, err := DiscoverBots(cmd.Context(), bootstrap)
			if err != nil {
				return err
			}
			selected, err := ResolveBot(parsed.Selector, bots)
			if err != nil {
				return err
			}
			runtimeConfig, err := parseRuntimeConfigArgs(selected, parsed.DynamicArgs)
			if err != nil {
				return err
			}
			cfg := settingsFromValues(parsed.BotToken, parsed.ApplicationID, parsed.GuildID, parsed.PublicKey, parsed.ClientID, parsed.ClientSecret)
			if err := cfg.Validate(); err != nil {
				return err
			}
			ctx, cancel := runContext(context.Background())
			defer cancel()
			return runSelectedBotsFn(ctx, RunRequest{Config: cfg, Bot: selected, RuntimeConfig: runtimeConfig, SyncOnStart: parsed.SyncOnStart, PrintParsedValues: parsed.PrintParsedValues, Out: cmd.OutOrStdout()})
		},
	}
	cmd.Flags().StringVar(&botToken, "bot-token", "", "Discord bot token")
	cmd.Flags().StringVar(&applicationID, "application-id", "", "Discord application/client ID")
	cmd.Flags().StringVar(&guildID, "guild-id", "", "Optional guild ID for development sync")
	cmd.Flags().StringVar(&publicKey, "public-key", "", "Discord public key")
	cmd.Flags().StringVar(&clientID, "client-id", "", "Discord client ID")
	cmd.Flags().StringVar(&clientSecret, "client-secret", "", "Discord client secret")
	cmd.Flags().BoolVar(&syncOnStart, "sync-on-start", false, "Sync slash commands for the selected bot before opening the gateway")
	cmd.Flags().BoolVar(&printParsedValues, "print-parsed-values", false, "Print the resolved config, selected bot, and runtime config, then exit")
	return cmd
}
