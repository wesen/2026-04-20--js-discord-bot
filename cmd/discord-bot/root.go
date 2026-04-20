package main

import (
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	"github.com/spf13/cobra"
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
	return rootCmd, nil
}
