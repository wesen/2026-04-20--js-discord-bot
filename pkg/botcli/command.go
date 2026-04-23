package botcli

import (
	internalbotcli "github.com/manuel/wesen/2026-04-20--js-discord-bot/internal/botcli"
	"github.com/spf13/cobra"
)

// NewBotsCommand builds the public repo-driven bot command tree from a bootstrap.
func NewBotsCommand(bootstrap Bootstrap) (*cobra.Command, error) {
	return internalbotcli.NewBotsCommand(bootstrap)
}

// NewCommand creates a public embeddable Cobra command for repo-driven bots.
func NewCommand(bootstrap ...Bootstrap) *cobra.Command {
	return internalbotcli.NewCommand(bootstrap...)
}
