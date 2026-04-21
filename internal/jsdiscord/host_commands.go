package jsdiscord

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func applicationCommandFromSnapshot(snapshot map[string]any) (*discordgo.ApplicationCommand, error) {
	name := strings.TrimSpace(fmt.Sprint(snapshot["name"]))
	if name == "" {
		return nil, fmt.Errorf("discord command snapshot missing name")
	}
	spec, _ := snapshot["spec"].(map[string]any)
	cmdType := strings.ToLower(strings.TrimSpace(fmt.Sprint(snapshot["type"])))
	if cmdType == "" && spec != nil {
		cmdType = strings.ToLower(strings.TrimSpace(fmt.Sprint(spec["type"])))
	}

	description := "JavaScript command"
	if spec != nil {
		if raw, ok := spec["description"]; ok && strings.TrimSpace(fmt.Sprint(raw)) != "" {
			description = strings.TrimSpace(fmt.Sprint(raw))
		}
	}

	var options []*discordgo.ApplicationCommandOption
	var err error

	switch cmdType {
	case "user":
		return &discordgo.ApplicationCommand{Name: name, Type: discordgo.UserApplicationCommand}, nil
	case "message":
		return &discordgo.ApplicationCommand{Name: name, Type: discordgo.MessageApplicationCommand}, nil
	default:
		options, err = applicationCommandOptions(spec)
		if err != nil {
			return nil, fmt.Errorf("discord command %s: %w", name, err)
		}
		return &discordgo.ApplicationCommand{Name: name, Description: description, Options: options}, nil
	}
}

func applicationCommandOptions(spec map[string]any) ([]*discordgo.ApplicationCommandOption, error) {
	if len(spec) == 0 {
		return nil, nil
	}
	rawOptions, ok := spec["options"]
	if !ok || rawOptions == nil {
		return nil, nil
	}
	type optionDraft struct {
		name     string
		raw      any
		required bool
		order    int
	}
	drafts := make([]optionDraft, 0)
	appendDraft := func(name string, raw any, order int) error {
		name = strings.TrimSpace(name)
		if name == "" {
			return fmt.Errorf("option missing name")
		}
		drafts = append(drafts, optionDraft{
			name:     name,
			raw:      raw,
			required: optionRequired(raw),
			order:    order,
		})
		return nil
	}
	switch v := rawOptions.(type) {
	case map[string]any:
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for i, key := range keys {
			if err := appendDraft(key, v[key], i); err != nil {
				return nil, err
			}
		}
	case []any:
		for i, raw := range v {
			mapping, _ := raw.(map[string]any)
			if err := appendDraft(strings.TrimSpace(fmt.Sprint(mapping["name"])), mapping, i); err != nil {
				return nil, err
			}
		}
	default:
		return nil, fmt.Errorf("unsupported options payload type %T", rawOptions)
	}
	required := make([]optionDraft, 0, len(drafts))
	optional := make([]optionDraft, 0, len(drafts))
	for _, draft := range drafts {
		if draft.required {
			required = append(required, draft)
		} else {
			optional = append(optional, draft)
		}
	}
	sort.SliceStable(required, func(i, j int) bool {
		if required[i].name != required[j].name {
			return required[i].name < required[j].name
		}
		return required[i].order < required[j].order
	})
	sort.SliceStable(optional, func(i, j int) bool {
		if optional[i].name != optional[j].name {
			return optional[i].name < optional[j].name
		}
		return optional[i].order < optional[j].order
	})
	out := make([]*discordgo.ApplicationCommandOption, 0, len(drafts))
	for _, draft := range append(required, optional...) {
		child, err := optionSpecToDiscord(draft.name, draft.raw)
		if err != nil {
			return nil, err
		}
		out = append(out, child)
	}
	return out, nil
}

func optionRequired(raw any) bool {
	mapping, _ := raw.(map[string]any)
	if len(mapping) == 0 {
		return false
	}
	required, _ := mapping["required"].(bool)
	return required
}

func optionSpecToDiscord(name string, raw any) (*discordgo.ApplicationCommandOption, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("option missing name")
	}
	mapping, _ := raw.(map[string]any)
	description := "Option for JavaScript command"
	if mapping != nil {
		if rawDesc, ok := mapping["description"]; ok && strings.TrimSpace(fmt.Sprint(rawDesc)) != "" {
			description = strings.TrimSpace(fmt.Sprint(rawDesc))
		}
	}
	optionType, err := optionTypeFromSpec(mapping)
	if err != nil {
		return nil, fmt.Errorf("option %s: %w", name, err)
	}
	ret := &discordgo.ApplicationCommandOption{Name: name, Description: description, Type: optionType}
	if required, ok := mapping["required"].(bool); ok {
		ret.Required = required
	}
	if autocomplete, ok := mapping["autocomplete"].(bool); ok {
		ret.Autocomplete = autocomplete
	}
	choices, err := optionChoicesFromSpec(mapping)
	if err != nil {
		return nil, fmt.Errorf("option %s: %w", name, err)
	}
	ret.Choices = choices
	if ret.Autocomplete && len(ret.Choices) > 0 {
		return nil, fmt.Errorf("option %s cannot define both autocomplete and static choices", name)
	}
	if minLength, ok := intPointer(mapping["minLength"]); ok {
		ret.MinLength = minLength
	}
	if maxLength, ok := intValue(mapping["maxLength"]); ok {
		ret.MaxLength = maxLength
	}
	if minValue, ok := floatPointer(mapping["minValue"]); ok {
		ret.MinValue = minValue
	}
	if maxValue, ok := floatValue(mapping["maxValue"]); ok {
		ret.MaxValue = maxValue
	}
	if optionType == discordgo.ApplicationCommandOptionSubCommand || optionType == discordgo.ApplicationCommandOptionSubCommandGroup {
		nested, err := applicationCommandOptions(mapping)
		if err != nil {
			return nil, fmt.Errorf("option %s: %w", name, err)
		}
		ret.Options = nested
	}
	return ret, nil
}

func optionChoicesFromSpec(mapping map[string]any) ([]*discordgo.ApplicationCommandOptionChoice, error) {
	if len(mapping) == 0 {
		return nil, nil
	}
	raw, ok := mapping["choices"]
	if !ok || raw == nil {
		return nil, nil
	}
	values, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("choices must be an array")
	}
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0, len(values))
	for _, item := range values {
		choice, err := normalizeAutocompleteChoice(item)
		if err != nil {
			return nil, err
		}
		if choice != nil {
			choices = append(choices, choice)
		}
	}
	return choices, nil
}

func optionTypeFromSpec(mapping map[string]any) (discordgo.ApplicationCommandOptionType, error) {
	if mapping == nil {
		return discordgo.ApplicationCommandOptionString, nil
	}
	switch strings.ToLower(strings.TrimSpace(fmt.Sprint(mapping["type"]))) {
	case "", "string":
		return discordgo.ApplicationCommandOptionString, nil
	case "int", "integer":
		return discordgo.ApplicationCommandOptionInteger, nil
	case "bool", "boolean":
		return discordgo.ApplicationCommandOptionBoolean, nil
	case "number", "float":
		return discordgo.ApplicationCommandOptionNumber, nil
	case "user":
		return discordgo.ApplicationCommandOptionUser, nil
	case "channel":
		return discordgo.ApplicationCommandOptionChannel, nil
	case "role":
		return discordgo.ApplicationCommandOptionRole, nil
	case "mentionable":
		return discordgo.ApplicationCommandOptionMentionable, nil
	case "sub_command":
		return discordgo.ApplicationCommandOptionSubCommand, nil
	case "sub_command_group":
		return discordgo.ApplicationCommandOptionSubCommandGroup, nil
	default:
		return discordgo.ApplicationCommandOptionString, fmt.Errorf("unsupported option type %q", mapping["type"])
	}
}
