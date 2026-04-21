package botcli

import (
	"fmt"
	"io"
	"strings"

	glazed_cli "github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	glazed_schema "github.com/go-go-golems/glazed/pkg/cmds/schema"
	glazed_values "github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/spf13/cobra"

	"github.com/manuel/wesen/2026-04-20--js-discord-bot/internal/jsdiscord"
)

func buildRunSchema(bot DiscoveredBot) (*glazed_schema.Schema, map[string]string, error) {
	schema_ := glazed_schema.NewSchema()
	fieldNameToConfigKey := map[string]string{}
	runSchema := bot.Descriptor.RunSchema
	if runSchema == nil || len(runSchema.Sections) == 0 {
		return schema_, fieldNameToConfigKey, nil
	}
	for _, section := range runSchema.Sections {
		definitions := make([]*fields.Definition, 0, len(section.Fields))
		for _, field := range section.Fields {
			definition, err := runFieldToDefinition(field)
			if err != nil {
				return nil, nil, err
			}
			definitions = append(definitions, definition)
			fieldNameToConfigKey[field.InternalName] = field.Name
		}
		sec, err := glazed_schema.NewSection(
			section.Slug,
			section.Title,
			glazed_schema.WithFields(definitions...),
		)
		if err != nil {
			return nil, nil, err
		}
		schema_.Set(sec.GetSlug(), sec)
	}
	return schema_, fieldNameToConfigKey, nil
}

func runFieldToDefinition(field jsdiscord.RunFieldDescriptor) (*fields.Definition, error) {
	fieldType, err := glazedFieldType(field.Type)
	if err != nil {
		return nil, fmt.Errorf("run field %q: %w", field.Name, err)
	}
	options := []fields.Option{}
	if strings.TrimSpace(field.Help) != "" {
		options = append(options, fields.WithHelp(field.Help))
	}
	if field.Required {
		options = append(options, fields.WithRequired(true))
	}
	if field.Default != nil {
		options = append(options, fields.WithDefault(field.Default))
	}
	return fields.New(field.InternalName, fieldType, options...), nil
}

func glazedFieldType(raw string) (fields.Type, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "string":
		return fields.TypeString, nil
	case "bool", "boolean":
		return fields.TypeBool, nil
	case "int", "integer":
		return fields.TypeInteger, nil
	case "number", "float":
		return fields.TypeFloat, nil
	case "string_list", "string-list", "string[]":
		return fields.TypeStringList, nil
	default:
		return fields.TypeString, fmt.Errorf("unsupported run field type %q", raw)
	}
}

func parseRuntimeConfigArgs(bot DiscoveredBot, args []string) (map[string]any, error) {
	schema_, nameMap, err := buildRunSchema(bot)
	if err != nil {
		return nil, err
	}
	parser, err := glazed_cli.NewCobraParserFromSections(schema_.Clone(), &glazed_cli.CobraParserConfig{AppName: "discord"})
	if err != nil {
		return nil, err
	}
	cmd := &cobra.Command{Use: "run-runtime-config", SilenceErrors: true, SilenceUsage: true}
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	if err := parser.AddToCobraCommand(cmd); err != nil {
		return nil, err
	}
	if err := cmd.ParseFlags(args); err != nil {
		return nil, err
	}
	parsed, err := parser.Parse(cmd, cmd.Flags().Args())
	if err != nil {
		return nil, err
	}
	return runtimeConfigFromParsedValues(parsed, schema_, nameMap), nil
}

func runtimeConfigFromParsedValues(parsed *glazed_values.Values, schema_ *glazed_schema.Schema, nameMap map[string]string) map[string]any {
	ret := map[string]any{}
	if parsed == nil || schema_ == nil {
		return ret
	}
	schema_.ForEach(func(slug string, _ glazed_schema.Section) {
		sectionValues, ok := parsed.Get(slug)
		if !ok || sectionValues == nil || sectionValues.Fields == nil {
			return
		}
		sectionValues.Fields.ForEach(func(key string, value *fields.FieldValue) {
			if value == nil {
				return
			}
			configKey := nameMap[key]
			if configKey == "" {
				configKey = key
			}
			ret[configKey] = value.Value
		})
	})
	return ret
}
