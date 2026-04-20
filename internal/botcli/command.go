package botcli

import (
	"context"
	"fmt"
	"os"

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
		Use:   "run <bot...>",
		Short: "Run one or more named bot implementations through the live Discord host",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bootstrap, err := DiscoverBootstrapFromCommand(cmd)
			if err != nil {
				return err
			}
			bots, err := DiscoverBots(cmd.Context(), bootstrap)
			if err != nil {
				return err
			}
			selected, err := ResolveBots(args, bots)
			if err != nil {
				return err
			}
			cfg := settingsFromValues(botToken, applicationID, guildID, publicKey, clientID, clientSecret)
			if err := cfg.Validate(); err != nil {
				return err
			}
			ctx, cancel := runContext(context.Background())
			defer cancel()
			return runSelectedBotsFn(ctx, RunRequest{Config: cfg, Bots: selected, SyncOnStart: syncOnStart, PrintParsedValues: printParsedValues, Out: cmd.OutOrStdout()})
		},
	}
	cmd.Flags().StringVar(&botToken, "bot-token", os.Getenv("DISCORD_BOT_TOKEN"), "Discord bot token")
	cmd.Flags().StringVar(&applicationID, "application-id", os.Getenv("DISCORD_APPLICATION_ID"), "Discord application/client ID")
	cmd.Flags().StringVar(&guildID, "guild-id", os.Getenv("DISCORD_GUILD_ID"), "Optional guild ID for development sync")
	cmd.Flags().StringVar(&publicKey, "public-key", os.Getenv("DISCORD_PUBLIC_KEY"), "Discord public key")
	cmd.Flags().StringVar(&clientID, "client-id", os.Getenv("DISCORD_CLIENT_ID"), "Discord client ID")
	cmd.Flags().StringVar(&clientSecret, "client-secret", os.Getenv("DISCORD_CLIENT_SECRET"), "Discord client secret")
	cmd.Flags().BoolVar(&syncOnStart, "sync-on-start", false, "Sync slash commands for the selected bots before opening the gateway")
	cmd.Flags().BoolVar(&printParsedValues, "print-parsed-values", false, "Print the resolved config and selected bots, then exit")
	return cmd
}
