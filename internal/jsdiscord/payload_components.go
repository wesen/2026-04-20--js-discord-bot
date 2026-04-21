package jsdiscord

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func normalizeComponents(raw any) ([]discordgo.MessageComponent, error) {
	switch v := raw.(type) {
	case nil:
		return nil, nil
	case []discordgo.MessageComponent:
		return v, nil
	case []any:
		components := make([]discordgo.MessageComponent, 0, len(v))
		for _, item := range v {
			component, err := normalizeComponent(item)
			if err != nil {
				return nil, err
			}
			if component != nil {
				components = append(components, component)
			}
		}
		return components, nil
	default:
		return nil, fmt.Errorf("unsupported components payload type %T", raw)
	}
}

func normalizeComponent(raw any) (discordgo.MessageComponent, error) {
	mapping, _ := raw.(map[string]any)
	if len(mapping) == 0 {
		return nil, nil
	}
	switch strings.ToLower(strings.TrimSpace(fmt.Sprint(mapping["type"]))) {
	case "", "actionrow", "action-row", "row":
		rawChildren, _ := mapping["components"].([]any)
		children := make([]discordgo.MessageComponent, 0, len(rawChildren))
		for _, child := range rawChildren {
			component, err := normalizeLeafComponent(child)
			if err != nil {
				return nil, err
			}
			if component != nil {
				children = append(children, component)
			}
		}
		return discordgo.ActionsRow{Components: children}, nil
	default:
		return normalizeLeafComponent(mapping)
	}
}

func normalizeLeafComponent(raw any) (discordgo.MessageComponent, error) {
	mapping, _ := raw.(map[string]any)
	if len(mapping) == 0 {
		return nil, nil
	}
	switch strings.ToLower(strings.TrimSpace(fmt.Sprint(mapping["type"]))) {
	case "button":
		style, err := buttonStyleFromValue(mapping["style"])
		if err != nil {
			return nil, err
		}
		button := discordgo.Button{Style: style}
		if label, ok := mapping["label"]; ok {
			button.Label = fmt.Sprint(label)
		}
		if disabled, ok := mapping["disabled"].(bool); ok {
			button.Disabled = disabled
		}
		if customID, ok := mapping["customId"]; ok {
			button.CustomID = fmt.Sprint(customID)
		}
		if url, ok := mapping["url"]; ok {
			button.URL = fmt.Sprint(url)
		}
		return button, nil
	case "select", "stringselect", "string-select":
		return normalizeSelectMenu(mapping, discordgo.StringSelectMenu)
	case "userselect", "user-select":
		return normalizeSelectMenu(mapping, discordgo.UserSelectMenu)
	case "roleselect", "role-select":
		return normalizeSelectMenu(mapping, discordgo.RoleSelectMenu)
	case "mentionableselect", "mentionable-select":
		return normalizeSelectMenu(mapping, discordgo.MentionableSelectMenu)
	case "channelselect", "channel-select":
		return normalizeSelectMenu(mapping, discordgo.ChannelSelectMenu)
	case "textinput", "text-input":
		return normalizeTextInput(mapping)
	case "", "actionrow", "action-row", "row":
		return normalizeComponent(mapping)
	default:
		return nil, fmt.Errorf("unsupported component type %q", mapping["type"])
	}
}

func normalizeSelectMenu(mapping map[string]any, menuType discordgo.SelectMenuType) (discordgo.MessageComponent, error) {
	menu := discordgo.SelectMenu{MenuType: menuType}
	if customID, ok := mapping["customId"]; ok {
		menu.CustomID = fmt.Sprint(customID)
	}
	if placeholder, ok := mapping["placeholder"]; ok {
		menu.Placeholder = fmt.Sprint(placeholder)
	}
	if disabled, ok := mapping["disabled"].(bool); ok {
		menu.Disabled = disabled
	}
	if minValues, ok := intPointer(mapping["minValues"]); ok {
		menu.MinValues = minValues
	}
	if maxValues, ok := intValue(mapping["maxValues"]); ok {
		menu.MaxValues = maxValues
	}
	if options, err := normalizeSelectMenuOptions(mapping["options"]); err != nil {
		return nil, err
	} else if len(options) > 0 {
		menu.Options = options
	}
	return menu, nil
}

func normalizeSelectMenuOptions(raw any) ([]discordgo.SelectMenuOption, error) {
	switch v := raw.(type) {
	case nil:
		return nil, nil
	case []any:
		ret := make([]discordgo.SelectMenuOption, 0, len(v))
		for _, item := range v {
			mapping, _ := item.(map[string]any)
			if len(mapping) == 0 {
				continue
			}
			label := strings.TrimSpace(fmt.Sprint(mapping["label"]))
			value := strings.TrimSpace(fmt.Sprint(mapping["value"]))
			if label == "" || value == "" {
				return nil, fmt.Errorf("select option requires label and value")
			}
			option := discordgo.SelectMenuOption{Label: label, Value: value}
			if description, ok := mapping["description"]; ok {
				option.Description = fmt.Sprint(description)
			}
			if emoji, ok := mapping["emoji"].(map[string]any); ok {
				normalized := normalizeComponentEmoji(emoji)
				option.Emoji = &normalized
			}
			if defaultValue, ok := mapping["default"].(bool); ok {
				option.Default = defaultValue
			}
			ret = append(ret, option)
		}
		return ret, nil
	default:
		return nil, fmt.Errorf("unsupported select menu options payload type %T", raw)
	}
}

func normalizeComponentEmoji(mapping map[string]any) discordgo.ComponentEmoji {
	return discordgo.ComponentEmoji{
		ID:       strings.TrimSpace(fmt.Sprint(mapping["id"])),
		Name:     strings.TrimSpace(fmt.Sprint(mapping["name"])),
		Animated: boolValue(mapping["animated"]),
	}
}

func normalizeTextInput(mapping map[string]any) (discordgo.MessageComponent, error) {
	customID := strings.TrimSpace(fmt.Sprint(mapping["customId"]))
	if customID == "" {
		return nil, fmt.Errorf("text input requires customId")
	}
	label := strings.TrimSpace(fmt.Sprint(mapping["label"]))
	if label == "" {
		return nil, fmt.Errorf("text input %q requires label", customID)
	}
	style, err := textInputStyleFromValue(mapping["style"])
	if err != nil {
		return nil, err
	}
	input := discordgo.TextInput{CustomID: customID, Label: label, Style: style}
	if placeholder, ok := mapping["placeholder"]; ok {
		input.Placeholder = fmt.Sprint(placeholder)
	}
	if value, ok := mapping["value"]; ok {
		input.Value = fmt.Sprint(value)
	}
	if required, ok := mapping["required"].(bool); ok {
		input.Required = required
	}
	if minLength, ok := intValue(mapping["minLength"]); ok {
		input.MinLength = minLength
	}
	if maxLength, ok := intValue(mapping["maxLength"]); ok {
		input.MaxLength = maxLength
	}
	return input, nil
}

func buttonStyleFromValue(raw any) (discordgo.ButtonStyle, error) {
	if raw == nil {
		return discordgo.PrimaryButton, nil
	}
	if value, ok := intValue(raw); ok {
		return discordgo.ButtonStyle(value), nil
	}
	switch strings.ToLower(strings.TrimSpace(fmt.Sprint(raw))) {
	case "", "primary":
		return discordgo.PrimaryButton, nil
	case "secondary":
		return discordgo.SecondaryButton, nil
	case "success", "green":
		return discordgo.SuccessButton, nil
	case "danger", "red":
		return discordgo.DangerButton, nil
	case "link":
		return discordgo.LinkButton, nil
	default:
		return discordgo.PrimaryButton, fmt.Errorf("unsupported button style %q", raw)
	}
}

func textInputStyleFromValue(raw any) (discordgo.TextInputStyle, error) {
	if raw == nil {
		return discordgo.TextInputShort, nil
	}
	if value, ok := intValue(raw); ok {
		return discordgo.TextInputStyle(value), nil
	}
	switch strings.ToLower(strings.TrimSpace(fmt.Sprint(raw))) {
	case "", "short", "single-line":
		return discordgo.TextInputShort, nil
	case "paragraph", "long", "multi-line":
		return discordgo.TextInputParagraph, nil
	default:
		return discordgo.TextInputShort, fmt.Errorf("unsupported text input style %q", raw)
	}
}
