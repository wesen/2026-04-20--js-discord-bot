package botcli

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"

	"github.com/manuel/wesen/2026-04-20--js-discord-bot/internal/jsdiscord"
)

func buildSyntheticBotRunDescription(bot DiscoveredBot, parents ...string) *cmds.CommandDescription {
	desc := cmds.NewCommandDescription(
		"run",
		cmds.WithShort(fmt.Sprintf("Run the %s Discord bot", bot.Name())),
		cmds.WithLong("Start the selected JavaScript Discord bot with host-managed lifecycle and runtime config."),
		cmds.WithParents(parents...),
	)

	addCoreRunFields(desc)
	addRunSchemaFields(desc, bot.Descriptor.RunSchema)
	relaxEnvBackedRequiredFlags(desc)
	return desc
}

func buildCompatibilityRunAliasDescription(base *cmds.CommandDescription, botName string) *cmds.CommandDescription {
	short := strings.TrimSpace(base.Short)
	if short == "" {
		short = fmt.Sprintf("Run the %s Discord bot", botName)
	}
	return base.Clone(true,
		cmds.WithName(botName),
		cmds.WithParents("run"),
		cmds.WithShort(short),
	)
}

func ensureRunCommandDefaults(desc *cmds.CommandDescription) *cmds.CommandDescription {
	if desc == nil {
		return nil
	}
	ret := desc.Clone(true)
	addCoreRunFields(ret)
	relaxEnvBackedRequiredFlags(ret)
	return ret
}

func addCoreRunFields(desc *cmds.CommandDescription) {
	if desc == nil {
		return
	}
	section, ok := desc.GetDefaultSection()
	if !ok {
		var err error
		section, err = schema.NewSection(schema.DefaultSlug, "Arguments")
		if err != nil {
			panic(err)
		}
		desc.Schema.Set(section.GetSlug(), section)
		if err := desc.Schema.MoveToFront(section.GetSlug()); err != nil {
			panic(err)
		}
	}

	addFieldIfMissing(section, fields.New("bot-token", fields.TypeString,
		fields.WithHelp("Discord bot token"),
	))
	addFieldIfMissing(section, fields.New("application-id", fields.TypeString,
		fields.WithHelp("Discord application/client ID"),
	))
	addFieldIfMissing(section, fields.New("guild-id", fields.TypeString,
		fields.WithHelp("Optional guild ID for development sync"),
	))
	addFieldIfMissing(section, fields.New("sync-on-start", fields.TypeBool,
		fields.WithHelp("Sync application commands before opening the gateway session"),
	))
}

func addRunSchemaFields(desc *cmds.CommandDescription, runSchema *jsdiscord.RunSchemaDescriptor) {
	if desc == nil || runSchema == nil {
		return
	}
	section, ok := desc.GetDefaultSection()
	if !ok {
		return
	}
	for _, runSection := range runSchema.Sections {
		for _, field := range runSection.Fields {
			addFieldIfMissing(section, fields.New(field.Name, glazedFieldType(field.Type),
				fields.WithHelp(field.Help),
				fields.WithRequired(field.Required),
				fields.WithDefault(field.Default),
			))
		}
	}
}

func addFieldIfMissing(section schema.Section, def *fields.Definition) {
	if section == nil || def == nil {
		return
	}
	if _, ok := section.GetDefinitions().Get(def.Name); ok {
		return
	}
	section.AddFields(def)
}

func relaxEnvBackedRequiredFlags(desc *cmds.CommandDescription) {
	if desc == nil {
		return
	}
	section, ok := desc.GetDefaultSection()
	if !ok {
		return
	}
	for _, name := range []string{"bot-token", "application-id"} {
		def, ok := section.GetDefinitions().Get(name)
		if !ok || def == nil {
			continue
		}
		def.Required = false
	}
}

func glazedFieldType(fieldType string) fields.Type {
	switch strings.ToLower(strings.TrimSpace(fieldType)) {
	case "bool", "boolean":
		return fields.TypeBool
	case "int", "integer":
		return fields.TypeInteger
	case "float", "number":
		return fields.TypeFloat
	default:
		return fields.TypeString
	}
}
