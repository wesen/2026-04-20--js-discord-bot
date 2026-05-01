package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/go-go-golems/discord-bot/internal/jsdiscord"
)

func newSlackManifestCommand() *cobra.Command {
	var botScript string
	var baseURL string
	cmd := &cobra.Command{
		Use:   "slack-manifest",
		Short: "Generate a Slack app manifest for a JavaScript bot",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(botScript) == "" {
				return fmt.Errorf("--bot-script is required")
			}
			desc, err := jsdiscord.InspectScript(cmd.Context(), botScript)
			if err != nil {
				return err
			}
			manifest, err := jsdiscord.SlackManifestJSON(desc, baseURL)
			if err != nil {
				return err
			}
			_, err = cmd.OutOrStdout().Write(append(manifest, '\n'))
			return err
		},
	}
	cmd.Flags().StringVar(&botScript, "bot-script", "", "Path to JavaScript bot script")
	cmd.Flags().StringVar(&baseURL, "base-url", "https://example.com", "Public base URL for Slack request endpoints")
	return cmd
}

func newSlackServeCommand() *cobra.Command {
	var botScript string
	var baseURL string
	var listenAddr string
	var botToken string
	var signingSecret string
	var stateDB string
	cmd := &cobra.Command{
		Use:   "slack-serve",
		Short: "Serve Slack HTTP slash command, interactivity, and events endpoints",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(botScript) == "" {
				return fmt.Errorf("--bot-script is required")
			}
			if strings.TrimSpace(botToken) == "" {
				botToken = os.Getenv("SLACK_BOT_TOKEN")
			}
			if strings.TrimSpace(signingSecret) == "" {
				signingSecret = os.Getenv("SLACK_SIGNING_SECRET")
			}
			backend, err := jsdiscord.NewSlackBackend(cmd.Context(), botScript, jsdiscord.SlackConfig{
				BotToken:      botToken,
				SigningSecret: signingSecret,
				BaseURL:       baseURL,
				ListenAddr:    listenAddr,
				StateDBPath:   stateDB,
			})
			if err != nil {
				return err
			}
			defer func() { _ = backend.Close(context.Background()) }()
			return backend.Serve(cmd.Context())
		},
	}
	cmd.Flags().StringVar(&botScript, "bot-script", "", "Path to JavaScript bot script")
	cmd.Flags().StringVar(&baseURL, "base-url", "", "Public base URL for generated Slack URLs/logging")
	cmd.Flags().StringVar(&listenAddr, "listen-addr", ":8080", "Address for Slack HTTP server")
	cmd.Flags().StringVar(&botToken, "slack-bot-token", "", "Slack bot token; defaults to SLACK_BOT_TOKEN")
	cmd.Flags().StringVar(&signingSecret, "slack-signing-secret", "", "Slack signing secret; defaults to SLACK_SIGNING_SECRET")
	cmd.Flags().StringVar(&stateDB, "slack-state-db", "./slack-state.sqlite", "SQLite path for Slack adapter state")
	return cmd
}
