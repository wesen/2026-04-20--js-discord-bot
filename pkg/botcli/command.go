package botcli

import (
	internalbotcli "github.com/manuel/wesen/2026-04-20--js-discord-bot/internal/botcli"
	"github.com/spf13/cobra"
)

// NewBotsCommand builds the public repo-driven bot command tree from a bootstrap.
func NewBotsCommand(bootstrap Bootstrap, opts ...CommandOption) (*cobra.Command, error) {
	cfg, err := applyCommandOptions(opts...)
	if err != nil {
		return nil, err
	}
	return internalbotcli.NewBotsCommand(bootstrap, toInternalCommandOptions(cfg)...)
}

// NewCommand creates a public embeddable Cobra command for repo-driven bots.
func NewCommand(bootstrap Bootstrap, opts ...CommandOption) *cobra.Command {
	cfg, err := applyCommandOptions(opts...)
	if err != nil {
		panic(err)
	}
	return internalbotcli.NewCommand(bootstrap, toInternalCommandOptions(cfg)...)
}
