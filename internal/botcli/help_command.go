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
)

// helpBotsCommand implements cmds.GlazeCommand and emits bot metadata as rows.
type helpBotsCommand struct {
	*cmds.CommandDescription
	bootstrap Bootstrap
}

func (c *helpBotsCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	// The bot name is either a positional argument or from the "bot-name" flag.
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

	bots, err := DiscoverBots(ctx, c.bootstrap)
	if err != nil {
		return err
	}
	selected, err := ResolveBot(botName, bots)
	if err != nil {
		return err
	}

	// Emit one row per command
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

	// Emit one row per event
	for _, ev := range selected.Descriptor.Events {
		row := types.NewRow(
			types.MRP("kind", "event"),
			types.MRP("name", ev.Name),
			types.MRP("description", ""),
			types.MRP("type", ""),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	// Emit one row per component
	for _, comp := range selected.Descriptor.Components {
		row := types.NewRow(
			types.MRP("kind", "component"),
			types.MRP("name", comp.CustomID),
			types.MRP("description", ""),
			types.MRP("type", ""),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	// Emit one row per modal
	for _, modal := range selected.Descriptor.Modals {
		row := types.NewRow(
			types.MRP("kind", "modal"),
			types.MRP("name", modal.CustomID),
			types.MRP("description", ""),
			types.MRP("type", ""),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}
