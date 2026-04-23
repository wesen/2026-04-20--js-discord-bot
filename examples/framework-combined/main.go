package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	publicbotcli "github.com/manuel/wesen/2026-04-20--js-discord-bot/pkg/botcli"
	"github.com/manuel/wesen/2026-04-20--js-discord-bot/pkg/framework"
	"github.com/spf13/cobra"
)

func main() {
	root, err := newRootCommand(os.Args[1:]...)
	if err != nil {
		panic(err)
	}
	if err := root.Execute(); err != nil {
		panic(err)
	}
}

func newRootCommand(rawArgs ...string) (*cobra.Command, error) {
	root := &cobra.Command{
		Use:           "framework-combined",
		Short:         "Example app combining one built-in bot with repo-driven bots",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().StringArray(publicbotcli.BotRepositoryFlag, nil, "Bot repository root to scan for named JavaScript bots (repeatable)")

	root.AddCommand(newRunBuiltInCommand())

	bootstrap, err := publicbotcli.BuildBootstrap(
		rawArgs,
		publicbotcli.WithDefaultRepositories(filepath.Join("examples", "discord-bots")),
	)
	if err != nil {
		return nil, err
	}
	root.AddCommand(publicbotcli.NewCommand(bootstrap, publicbotcli.WithAppName("discord")))

	return root, nil
}

func newRunBuiltInCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "run-builtin",
		Short: "Run the built-in bot wired directly through pkg/framework",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			bot, err := framework.New(
				framework.WithCredentialsFromEnv(),
				framework.WithScript(filepath.Join("examples", "framework-combined", "builtin-bot", "index.js")),
				framework.WithRuntimeConfig(map[string]any{
					"mode":   "built-in",
					"source": "framework-combined",
				}),
				framework.WithSyncOnStart(true),
			)
			if err != nil {
				return fmt.Errorf("create built-in bot: %w", err)
			}
			return bot.Run(ctx)
		},
	}
}
