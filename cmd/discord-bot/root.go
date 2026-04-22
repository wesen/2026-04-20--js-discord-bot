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

func newRootCommand(rawArgs ...string) (*cobra.Command, error) {
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
	rootCmd.PersistentFlags().StringArray(botcli.BotRepositoryFlag, nil, "Bot repository root to scan for named JavaScript bots (repeatable)")

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

	bootstrap, err := buildBootstrap(rawArgs)
	if err != nil {
		return nil, err
	}
	botsCmd := botcli.NewCommand(bootstrap)
	rootCmd.AddCommand(botsCmd)

	return rootCmd, nil
}

func buildBootstrap(rawArgs []string) (botcli.Bootstrap, error) {
	repos := []botcli.Repository{}
	seen := map[string]struct{}{}

	appendRepo := func(path, source, sourceRef string) error {
		abs, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("resolve %s repository %q: %w", source, path, err)
		}
		info, err := os.Stat(abs)
		if err != nil {
			return fmt.Errorf("resolve %s repository %q: %w", source, path, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("resolve %s repository %q: not a directory", source, path)
		}
		if _, ok := seen[abs]; ok {
			return nil
		}
		seen[abs] = struct{}{}
		repos = append(repos, botcli.Repository{
			Name:      filepath.Base(abs),
			Source:    source,
			SourceRef: sourceRef,
			RootDir:   abs,
		})
		return nil
	}

	cliRepoPaths, err := botRepositoryPathsFromArgs(rawArgs)
	if err != nil {
		return botcli.Bootstrap{}, err
	}
	for _, path := range cliRepoPaths {
		if err := appendRepo(path, "cli", "--"+botcli.BotRepositoryFlag); err != nil {
			return botcli.Bootstrap{}, err
		}
	}
	if len(repos) > 0 {
		return botcli.Bootstrap{Repositories: repos}, nil
	}

	if envRepos := os.Getenv("DISCORD_BOT_REPOSITORIES"); envRepos != "" {
		for _, path := range strings.Split(envRepos, string(os.PathListSeparator)) {
			path = strings.TrimSpace(path)
			if path == "" {
				continue
			}
			if err := appendRepo(path, "env", "DISCORD_BOT_REPOSITORIES"); err != nil {
				return botcli.Bootstrap{}, err
			}
		}
	}
	if len(repos) > 0 {
		return botcli.Bootstrap{Repositories: repos}, nil
	}

	defaultRepo := filepath.Join("examples", "discord-bots")
	if info, err := os.Stat(defaultRepo); err == nil && info.IsDir() {
		if err := appendRepo(defaultRepo, "default", "examples/discord-bots"); err != nil {
			return botcli.Bootstrap{}, err
		}
	}

	return botcli.Bootstrap{Repositories: repos}, nil
}

func botRepositoryPathsFromArgs(args []string) ([]string, error) {
	ret := []string{}
	flagName := "--" + botcli.BotRepositoryFlag
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			break
		}
		if arg == flagName {
			if i+1 >= len(args) {
				return nil, fmt.Errorf("flag needs an argument: %s", flagName)
			}
			i++
			ret = append(ret, args[i])
			continue
		}
		if strings.HasPrefix(arg, flagName+"=") {
			ret = append(ret, strings.TrimPrefix(arg, flagName+"="))
		}
	}
	return ret, nil
}
