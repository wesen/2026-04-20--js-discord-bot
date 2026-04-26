package botcli

import (
	"context"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/discord-bot/internal/jsdiscord"
)

type listBotsCommand struct {
	*cmds.CommandDescription
	bootstrap Bootstrap
	hostOpts  []jsdiscord.HostOption
}

func (c *listBotsCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	bots, err := DiscoverBots(ctx, c.bootstrap, c.hostOpts...)
	if err != nil {
		return err
	}
	for _, bot := range bots {
		row := types.NewRow(
			types.MRP("name", bot.Name()),
			types.MRP("source", bot.SourceLabel()),
			types.MRP("description", bot.Description()),
			types.MRP("commands", len(bot.Descriptor.Commands)),
			types.MRP("events", len(bot.Descriptor.Events)),
			types.MRP("components", len(bot.Descriptor.Components)),
			types.MRP("modals", len(bot.Descriptor.Modals)),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}
