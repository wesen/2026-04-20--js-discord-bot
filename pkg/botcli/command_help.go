package botcli

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/discord-bot/internal/jsdiscord"
)

type helpBotsCommand struct {
	*cmds.CommandDescription
	bootstrap Bootstrap
	hostOpts  []jsdiscord.HostOption
}

func (c *helpBotsCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	botName := ""
	if argSection, ok := vals.Get(schema.DefaultSlug); ok && argSection.Fields != nil {
		if fv, ok := argSection.Fields.Get("bot-name"); ok && fv != nil {
			botName = fmt.Sprint(fv.Value)
		}
	}
	botName = strings.TrimSpace(botName)
	if botName == "" {
		return fmt.Errorf("bot name is required")
	}

	bots, err := DiscoverBots(ctx, c.bootstrap, c.hostOpts...)
	if err != nil {
		return err
	}
	selected, err := ResolveBot(botName, bots)
	if err != nil {
		return err
	}

	for _, cmd := range selected.Descriptor.Commands {
		row := types.NewRow(
			types.MRP("kind", "command"),
			types.MRP("name", cmd.Name),
			types.MRP("description", cmd.Description),
			types.MRP("type", cmd.Type),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	for _, ev := range selected.Descriptor.Events {
		row := types.NewRow(types.MRP("kind", "event"), types.MRP("name", ev.Name), types.MRP("description", ""), types.MRP("type", ""))
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	for _, comp := range selected.Descriptor.Components {
		row := types.NewRow(types.MRP("kind", "component"), types.MRP("name", comp.CustomID), types.MRP("description", ""), types.MRP("type", ""))
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	for _, modal := range selected.Descriptor.Modals {
		row := types.NewRow(types.MRP("kind", "modal"), types.MRP("name", modal.CustomID), types.MRP("description", ""), types.MRP("type", ""))
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}
