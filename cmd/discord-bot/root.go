package main

import (
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	"github.com/spf13/cobra"

	publicbotcli "github.com/go-go-golems/discord-bot/pkg/botcli"
	appdoc "github.com/go-go-golems/discord-bot/pkg/doc"
)

func newRootCommand(rawArgs ...string) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:           "discord-bot",
		Short:         "A simple Go Discord bot",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return logging.InitLoggerFromCobra(cmd)
		},
	}

	if err := logging.AddLoggingSectionToRootCommand(rootCmd, "discord"); err != nil {
		return nil, fmt.Errorf("add logging section: %w", err)
	}
	rootCmd.PersistentFlags().StringArray(publicbotcli.BotRepositoryFlag, nil, "Bot repository root to scan for named JavaScript bots (repeatable)")

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

	bootstrap, err := publicbotcli.BuildBootstrap(rawArgs)
	if err != nil {
		return nil, err
	}
	botsCmd, err := publicbotcli.NewBotsCommand(bootstrap)
	if err != nil {
		return nil, err
	}
	rootCmd.AddCommand(botsCmd)

	return rootCmd, nil
}
