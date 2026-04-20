package botcli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	glazed_cli "github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	glazed_schema "github.com/go-go-golems/glazed/pkg/cmds/schema"
	glazed_values "github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/spf13/cobra"

	"github.com/manuel/wesen/2026-04-20--js-discord-bot/internal/jsdiscord"
)

type preParsedRunArgs struct {
	Selector          string
	DynamicArgs       []string
	BotRepositories   []string
	BotToken          string
	ApplicationID     string
	GuildID           string
	PublicKey         string
	ClientID          string
	ClientSecret      string
	SyncOnStart       bool
	PrintParsedValues bool
	ShowHelp          bool
}

func defaultPreParsedRunArgs() preParsedRunArgs {
	return preParsedRunArgs{
		BotToken:      os.Getenv("DISCORD_BOT_TOKEN"),
		ApplicationID: os.Getenv("DISCORD_APPLICATION_ID"),
		GuildID:       os.Getenv("DISCORD_GUILD_ID"),
		PublicKey:     os.Getenv("DISCORD_PUBLIC_KEY"),
		ClientID:      os.Getenv("DISCORD_CLIENT_ID"),
		ClientSecret:  os.Getenv("DISCORD_CLIENT_SECRET"),
	}
}

func preparseRunArgs(args []string, defaults preParsedRunArgs) (preParsedRunArgs, error) {
	ret := defaults
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			ret.DynamicArgs = append(ret.DynamicArgs, args[i+1:]...)
			break
		}
		if arg == "--help" || arg == "-h" {
			ret.ShowHelp = true
			continue
		}
		if strings.HasPrefix(arg, "--") {
			name, inlineValue, hasInline := splitLongFlag(arg)
			switch name {
			case "bot-repository":
				value, next, err := consumeStringFlagValue(name, inlineValue, hasInline, args, i)
				if err != nil {
					return preParsedRunArgs{}, err
				}
				ret.BotRepositories = append(ret.BotRepositories, value)
				i = next
			case "bot-token":
				value, next, err := consumeStringFlagValue(name, inlineValue, hasInline, args, i)
				if err != nil {
					return preParsedRunArgs{}, err
				}
				ret.BotToken = value
				i = next
			case "application-id":
				value, next, err := consumeStringFlagValue(name, inlineValue, hasInline, args, i)
				if err != nil {
					return preParsedRunArgs{}, err
				}
				ret.ApplicationID = value
				i = next
			case "guild-id":
				value, next, err := consumeStringFlagValue(name, inlineValue, hasInline, args, i)
				if err != nil {
					return preParsedRunArgs{}, err
				}
				ret.GuildID = value
				i = next
			case "public-key":
				value, next, err := consumeStringFlagValue(name, inlineValue, hasInline, args, i)
				if err != nil {
					return preParsedRunArgs{}, err
				}
				ret.PublicKey = value
				i = next
			case "client-id":
				value, next, err := consumeStringFlagValue(name, inlineValue, hasInline, args, i)
				if err != nil {
					return preParsedRunArgs{}, err
				}
				ret.ClientID = value
				i = next
			case "client-secret":
				value, next, err := consumeStringFlagValue(name, inlineValue, hasInline, args, i)
				if err != nil {
					return preParsedRunArgs{}, err
				}
				ret.ClientSecret = value
				i = next
			case "sync-on-start":
				value, next, err := consumeBoolFlagValue(name, inlineValue, hasInline, args, i)
				if err != nil {
					return preParsedRunArgs{}, err
				}
				ret.SyncOnStart = value
				i = next
			case "print-parsed-values":
				value, next, err := consumeBoolFlagValue(name, inlineValue, hasInline, args, i)
				if err != nil {
					return preParsedRunArgs{}, err
				}
				ret.PrintParsedValues = value
				i = next
			default:
				ret.DynamicArgs = appendUnknownDynamicArg(ret.DynamicArgs, args, &i)
			}
			continue
		}
		if strings.HasPrefix(arg, "-") {
			ret.DynamicArgs = appendUnknownDynamicArg(ret.DynamicArgs, args, &i)
			continue
		}
		if strings.TrimSpace(ret.Selector) == "" {
			ret.Selector = strings.TrimSpace(arg)
			continue
		}
		return preParsedRunArgs{}, fmt.Errorf("bots run accepts exactly one bot selector; unexpected argument %q", arg)
	}
	if strings.TrimSpace(ret.Selector) == "" && !ret.ShowHelp {
		return preParsedRunArgs{}, fmt.Errorf("bot selector is required")
	}
	return ret, nil
}

func splitLongFlag(arg string) (string, string, bool) {
	trimmed := strings.TrimPrefix(arg, "--")
	parts := strings.SplitN(trimmed, "=", 2)
	if len(parts) == 2 {
		return parts[0], parts[1], true
	}
	return trimmed, "", false
}

func consumeStringFlagValue(name, inlineValue string, hasInline bool, args []string, index int) (string, int, error) {
	if hasInline {
		return inlineValue, index, nil
	}
	if index+1 >= len(args) {
		return "", index, fmt.Errorf("missing value for flag --%s", name)
	}
	return args[index+1], index + 1, nil
}

func consumeBoolFlagValue(name, inlineValue string, hasInline bool, args []string, index int) (bool, int, error) {
	if hasInline {
		switch strings.ToLower(strings.TrimSpace(inlineValue)) {
		case "true", "1", "yes", "on":
			return true, index, nil
		case "false", "0", "no", "off":
			return false, index, nil
		default:
			return false, index, fmt.Errorf("invalid boolean value %q for flag --%s", inlineValue, name)
		}
	}
	if index+1 < len(args) && isBoolLiteral(args[index+1]) {
		return consumeBoolFlagValue(name, args[index+1], true, args, index+1)
	}
	return true, index, nil
}

func isBoolLiteral(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1", "yes", "on", "false", "0", "no", "off":
		return true
	default:
		return false
	}
}

func appendUnknownDynamicArg(dynamic []string, args []string, index *int) []string {
	dynamic = append(dynamic, args[*index])
	if *index+1 < len(args) && !strings.HasPrefix(args[*index+1], "-") {
		dynamic = append(dynamic, args[*index+1])
		*index = *index + 1
	}
	return dynamic
}

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

func printRunSchema(w io.Writer, bot DiscoveredBot) error {
	if w == nil || bot.Descriptor == nil || bot.Descriptor.RunSchema == nil || len(bot.Descriptor.RunSchema.Sections) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(w, "Run config:"); err != nil {
		return err
	}
	for _, section := range bot.Descriptor.RunSchema.Sections {
		heading := section.Title
		if strings.TrimSpace(heading) == "" {
			heading = section.Slug
		}
		if _, err := fmt.Fprintf(w, "  [%s]\n", heading); err != nil {
			return err
		}
		fields_ := append([]jsdiscord.RunFieldDescriptor(nil), section.Fields...)
		sort.Slice(fields_, func(i, j int) bool { return fields_[i].Name < fields_[j].Name })
		for _, field := range fields_ {
			line := fmt.Sprintf("    - %s (--%s) <%s>", field.Name, strings.ReplaceAll(field.InternalName, "_", "-"), field.Type)
			if field.Required {
				line += " required"
			}
			if field.Help != "" {
				line += " — " + field.Help
			}
			if field.Default != nil {
				line += fmt.Sprintf(" (default: %v)", field.Default)
			}
			if _, err := fmt.Fprintln(w, line); err != nil {
				return err
			}
		}
	}
	return nil
}
