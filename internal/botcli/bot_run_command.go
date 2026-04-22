package botcli

import (
	"context"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/values"

	"github.com/manuel/wesen/2026-04-20--js-discord-bot/internal/bot"
	appconfig "github.com/manuel/wesen/2026-04-20--js-discord-bot/internal/config"
)

// botRunCommand implements cmds.BareCommand for __verb__("run") verbs.
// It orchestrates the Discord bot lifecycle: parse credentials and config,
// load the JS script, inject runtime config, connect to gateway, block.
type botRunCommand struct {
	*cmds.CommandDescription
	scriptPath string
}

// Run extracts Discord credentials and runtime config from parsed CLI values,
// creates a bot.Bot for the script, injects config, connects, and blocks.
func (c *botRunCommand) Run(ctx context.Context, parsedValues *values.Values) error {
	cfg, err := appconfig.FromValues(parsedValues)
	if err != nil {
		return fmt.Errorf("decode discord settings: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return err
	}

	runtimeConfig := buildRuntimeConfig(parsedValues)

	b, err := bot.NewWithScript(cfg, c.scriptPath, runtimeConfig)
	if err != nil {
		return err
	}
	defer func() { _ = b.Close() }()

	if err := b.Open(); err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}

// buildRuntimeConfig walks all parsed sections/fields and builds a map
// keyed by snake_case names suitable for ctx.config in JS.
func buildRuntimeConfig(parsedValues *values.Values) map[string]any {
	ret := map[string]any{}
	if parsedValues == nil {
		return ret
	}
	parsedValues.ForEach(func(slug string, sectionVals *values.SectionValues) {
		if sectionVals == nil || sectionVals.Fields == nil {
			return
		}
		sectionVals.Fields.ForEach(func(fieldName string, fv *fields.FieldValue) {
			if fv == nil || fv.Definition == nil {
				return
			}
			configKey := runtimeFieldInternalName(fieldName)
			ret[configKey] = fv.Value
		})
	})
	return ret
}
