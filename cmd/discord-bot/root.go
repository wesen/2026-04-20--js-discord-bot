package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	"github.com/spf13/cobra"

	"github.com/manuel/wesen/2026-04-20--js-discord-bot/internal/botcli"
	appdoc "github.com/manuel/wesen/2026-04-20--js-discord-bot/pkg/doc"
)

func newRootCommand() (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:           "discord-bot",
		Short:         "A simple Go Discord bot",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return logging.InitLoggerFromCobra(cmd)
		},
	}

	if err := logging.AddLoggingSectionToRootCommand(rootCmd, "discord"); err != nil {
		return nil, fmt.Errorf("add logging section: %w", err)
	}

	helpSystem := help.NewHelpSystem()
	if err := appdoc.AddDocToHelpSystem(helpSystem); err != nil {
		return nil, fmt.Errorf("load embedded help docs: %w", err)
	}
	help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)

	runCmd, err := newRunCommand()
	if err != nil {
		return nil, err
	}
	validateCmd, err := newValidateConfigCommand()
	if err != nil {
		return nil, err
	}
	syncCmd, err := newSyncCommandsCommand()
	if err != nil {
		return nil, err
	}

	runCobraCmd, err := buildCobraCommand(runCmd)
	if err != nil {
		return nil, err
	}
	validateCobraCmd, err := buildCobraCommand(validateCmd)
	if err != nil {
		return nil, err
	}
	syncCobraCmd, err := buildCobraCommand(syncCmd)
	if err != nil {
		return nil, err
	}

	rootCmd.AddCommand(runCobraCmd, validateCobraCmd, syncCobraCmd)

	// Build bootstrap from default or discovered repositories.
	bootstrap := buildDefaultBootstrap()
	botsCmd := botcli.NewCommand(bootstrap)
	rootCmd.AddCommand(botsCmd)

	return rootCmd, nil
}

func buildDefaultBootstrap() botcli.Bootstrap {
	var repos []botcli.Repository

	appendRepo := func(path, source, sourceRef string) {
		abs, err := filepath.Abs(path)
		if err != nil {
			return
		}
		if info, err := os.Stat(abs); err != nil || !info.IsDir() {
			return
		}
		for _, existing := range repos {
			if existing.RootDir == abs {
				return
			}
		}
		repos = append(repos, botcli.Repository{
			Name:      filepath.Base(abs),
			Source:    source,
			SourceRef: sourceRef,
			RootDir:   abs,
		})
	}

	// If DISCORD_BOT_REPOSITORIES env var is set, use it.
	if envRepos := os.Getenv("DISCORD_BOT_REPOSITORIES"); envRepos != "" {
		for _, path := range strings.Split(envRepos, ":") {
			path = strings.TrimSpace(path)
			if path == "" {
				continue
			}
			appendRepo(path, "env", "DISCORD_BOT_REPOSITORIES")
		}
	}

	// Fall back to the repository's example bots directory for local/dev usage
	// only when no explicit repositories were configured.
	if len(repos) == 0 {
		appendRepo(filepath.Join("examples", "discord-bots"), "default", "examples/discord-bots")
	}

	return botcli.Bootstrap{Repositories: repos}
}
