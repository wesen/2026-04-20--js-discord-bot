package jsdiscord

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type normalizedResponse struct {
	Content         string
	Embeds          []*discordgo.MessageEmbed
	Components      []discordgo.MessageComponent
	AllowedMentions *discordgo.MessageAllowedMentions
	Files           []*discordgo.File
	Reference       *discordgo.MessageReference
	TTS             bool
	Ephemeral       bool
}

func normalizeResponsePayload(payload any) (*discordgo.InteractionResponseData, error) {
	normalized, err := normalizePayload(payload)
	if err != nil {
		return nil, err
	}
	data := &discordgo.InteractionResponseData{
		Content:         normalized.Content,
		Embeds:          normalized.Embeds,
		Components:      normalized.Components,
		AllowedMentions: normalized.AllowedMentions,
		Files:           normalized.Files,
		TTS:             normalized.TTS,
	}
	if normalized.Ephemeral {
		data.Flags = discordgo.MessageFlagsEphemeral
	}
	return data, nil
}

func normalizeModalPayload(payload any) (*discordgo.InteractionResponseData, error) {
	mapping, _ := payload.(map[string]any)
	if len(mapping) == 0 {
		return nil, fmt.Errorf("modal payload must be an object")
	}
	customID := strings.TrimSpace(fmt.Sprint(mapping["customId"]))
	if customID == "" {
		return nil, fmt.Errorf("modal payload missing customId")
	}
	title := strings.TrimSpace(fmt.Sprint(mapping["title"]))
	if title == "" {
		return nil, fmt.Errorf("modal payload missing title")
	}
	components, err := normalizeComponents(mapping["components"])
	if err != nil {
		return nil, err
	}
	if len(components) == 0 {
		return nil, fmt.Errorf("modal payload must include at least one component row")
	}
	return &discordgo.InteractionResponseData{CustomID: customID, Title: title, Components: components}, nil
}

func normalizeAutocompleteChoices(payload any) ([]*discordgo.ApplicationCommandOptionChoice, error) {
	switch v := payload.(type) {
	case nil:
		return nil, nil
	case []*discordgo.ApplicationCommandOptionChoice:
		return v, nil
	case []any:
		choices := make([]*discordgo.ApplicationCommandOptionChoice, 0, len(v))
		for _, item := range v {
			choice, err := normalizeAutocompleteChoice(item)
			if err != nil {
				return nil, err
			}
			if choice != nil {
				choices = append(choices, choice)
			}
		}
		if len(choices) > 25 {
			choices = choices[:25]
		}
		return choices, nil
	default:
		return nil, fmt.Errorf("unsupported autocomplete result type %T", payload)
	}
}

func normalizeAutocompleteChoice(raw any) (*discordgo.ApplicationCommandOptionChoice, error) {
	switch v := raw.(type) {
	case nil:
		return nil, nil
	case *discordgo.ApplicationCommandOptionChoice:
		return v, nil
	case map[string]any:
		name := strings.TrimSpace(fmt.Sprint(v["name"]))
		if name == "" {
			return nil, fmt.Errorf("autocomplete choice missing name")
		}
		value, ok := v["value"]
		if !ok {
			return nil, fmt.Errorf("autocomplete choice %q missing value", name)
		}
		return &discordgo.ApplicationCommandOptionChoice{Name: name, Value: value}, nil
	default:
		return nil, fmt.Errorf("unsupported autocomplete choice type %T", raw)
	}
}

func normalizeWebhookParams(payload any) (*discordgo.WebhookParams, error) {
	normalized, err := normalizePayload(payload)
	if err != nil {
		return nil, err
	}
	params := &discordgo.WebhookParams{
		Content:         normalized.Content,
		Embeds:          normalized.Embeds,
		Components:      normalized.Components,
		AllowedMentions: normalized.AllowedMentions,
		Files:           normalized.Files,
		TTS:             normalized.TTS,
	}
	if normalized.Ephemeral {
		params.Flags = discordgo.MessageFlagsEphemeral
	}
	return params, nil
}

func normalizeWebhookEdit(payload any) (*discordgo.WebhookEdit, error) {
	normalized, err := normalizePayload(payload)
	if err != nil {
		return nil, err
	}
	content := normalized.Content
	components := normalized.Components
	embeds := normalized.Embeds
	return &discordgo.WebhookEdit{
		Content:         &content,
		Components:      &components,
		Embeds:          &embeds,
		Files:           normalized.Files,
		AllowedMentions: normalized.AllowedMentions,
	}, nil
}

func normalizeMessageSend(payload any) (*discordgo.MessageSend, error) {
	normalized, err := normalizePayload(payload)
	if err != nil {
		return nil, err
	}
	message := &discordgo.MessageSend{
		Content:         normalized.Content,
		Embeds:          normalized.Embeds,
		Components:      normalized.Components,
		Files:           normalized.Files,
		AllowedMentions: normalized.AllowedMentions,
		Reference:       normalized.Reference,
		TTS:             normalized.TTS,
	}
	if normalized.Ephemeral {
		message.Flags = discordgo.MessageFlagsEphemeral
	}
	return message, nil
}

func normalizeChannelMessageEdit(channelID, messageID string, payload any) (*discordgo.MessageEdit, error) {
	normalized, err := normalizePayload(payload)
	if err != nil {
		return nil, err
	}
	edit := discordgo.NewMessageEdit(channelID, messageID)
	content := normalized.Content
	components := normalized.Components
	embeds := normalized.Embeds
	edit.Content = &content
	edit.Components = &components
	edit.Embeds = &embeds
	edit.Files = normalized.Files
	edit.AllowedMentions = normalized.AllowedMentions
	return edit, nil
}

func normalizePayload(payload any) (*normalizedResponse, error) {
	switch v := payload.(type) {
	case nil:
		return &normalizedResponse{}, nil
	case string:
		return &normalizedResponse{Content: v}, nil
	case map[string]any:
		ret := &normalizedResponse{}
		if content, ok := v["content"]; ok {
			ret.Content = fmt.Sprint(content)
		}
		if tts, ok := v["tts"].(bool); ok {
			ret.TTS = tts
		}
		if ephemeral, ok := v["ephemeral"].(bool); ok {
			ret.Ephemeral = ephemeral
		}
		embeds, err := normalizeEmbeds(v)
		if err != nil {
			return nil, err
		}
		ret.Embeds = embeds
		components, err := normalizeComponents(v["components"])
		if err != nil {
			return nil, err
		}
		ret.Components = components
		mentions, err := normalizeAllowedMentions(v["allowedMentions"])
		if err != nil {
			return nil, err
		}
		ret.AllowedMentions = mentions
		files, err := normalizeFiles(v["files"])
		if err != nil {
			return nil, err
		}
		ret.Files = files
		reference, err := normalizeMessageReference(v["replyTo"])
		if err != nil {
			return nil, err
		}
		ret.Reference = reference
		return ret, nil
	default:
		return &normalizedResponse{Content: fmt.Sprint(payload)}, nil
	}
}

func normalizeEmbeds(payload map[string]any) ([]*discordgo.MessageEmbed, error) {
	if payload == nil {
		return nil, nil
	}
	if raw, ok := payload["embeds"]; ok {
		return normalizeEmbedArray(raw)
	}
	if raw, ok := payload["embed"]; ok {
		embeds, err := normalizeEmbedArray([]any{raw})
		if err != nil {
			return nil, err
		}
		return embeds, nil
	}
	return nil, nil
}

func normalizeEmbedArray(raw any) ([]*discordgo.MessageEmbed, error) {
	switch v := raw.(type) {
	case nil:
		return nil, nil
	case []*discordgo.MessageEmbed:
		return v, nil
	case []any:
		embeds := make([]*discordgo.MessageEmbed, 0, len(v))
		for _, item := range v {
			embed, err := normalizeEmbed(item)
			if err != nil {
				return nil, err
			}
			if embed != nil {
				embeds = append(embeds, embed)
			}
		}
		return embeds, nil
	case map[string]any:
		embed, err := normalizeEmbed(v)
		if err != nil {
			return nil, err
		}
		if embed == nil {
			return nil, nil
		}
		return []*discordgo.MessageEmbed{embed}, nil
	default:
		return nil, fmt.Errorf("unsupported embeds payload type %T", raw)
	}
}

func normalizeEmbed(raw any) (*discordgo.MessageEmbed, error) {
	switch v := raw.(type) {
	case nil:
		return nil, nil
	case *discordgo.MessageEmbed:
		return v, nil
	case map[string]any:
		embed := &discordgo.MessageEmbed{}
		if title, ok := v["title"]; ok {
			embed.Title = fmt.Sprint(title)
		}
		if desc, ok := v["description"]; ok {
			embed.Description = fmt.Sprint(desc)
		}
		if url, ok := v["url"]; ok {
			embed.URL = fmt.Sprint(url)
		}
		if timestamp, ok := v["timestamp"]; ok {
			embed.Timestamp = fmt.Sprint(timestamp)
		}
		if color, ok := intValue(v["color"]); ok {
			embed.Color = color
		}
		if footer, ok := v["footer"].(map[string]any); ok {
			embed.Footer = &discordgo.MessageEmbedFooter{}
			if text, ok := footer["text"]; ok {
				embed.Footer.Text = fmt.Sprint(text)
			}
			if iconURL, ok := footer["iconURL"]; ok {
				embed.Footer.IconURL = fmt.Sprint(iconURL)
			}
		}
		if author, ok := v["author"].(map[string]any); ok {
			embed.Author = &discordgo.MessageEmbedAuthor{}
			if name, ok := author["name"]; ok {
				embed.Author.Name = fmt.Sprint(name)
			}
			if url, ok := author["url"]; ok {
				embed.Author.URL = fmt.Sprint(url)
			}
			if iconURL, ok := author["iconURL"]; ok {
				embed.Author.IconURL = fmt.Sprint(iconURL)
			}
		}
		if image, ok := v["image"].(map[string]any); ok {
			embed.Image = &discordgo.MessageEmbedImage{}
			if url, ok := image["url"]; ok {
				embed.Image.URL = fmt.Sprint(url)
			}
		}
		if thumbnail, ok := v["thumbnail"].(map[string]any); ok {
			embed.Thumbnail = &discordgo.MessageEmbedThumbnail{}
			if url, ok := thumbnail["url"]; ok {
				embed.Thumbnail.URL = fmt.Sprint(url)
			}
		}
		if fieldsRaw, ok := v["fields"].([]any); ok {
			fields := make([]*discordgo.MessageEmbedField, 0, len(fieldsRaw))
			for _, rawField := range fieldsRaw {
				fieldMap, _ := rawField.(map[string]any)
				if len(fieldMap) == 0 {
					continue
				}
				field := &discordgo.MessageEmbedField{}
				if name, ok := fieldMap["name"]; ok {
					field.Name = fmt.Sprint(name)
				}
				if value, ok := fieldMap["value"]; ok {
					field.Value = fmt.Sprint(value)
				}
				if inline, ok := fieldMap["inline"].(bool); ok {
					field.Inline = inline
				}
				fields = append(fields, field)
			}
			embed.Fields = fields
		}
		return embed, nil
	default:
		return nil, fmt.Errorf("unsupported embed payload type %T", raw)
	}
}

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

func normalizeAllowedMentions(raw any) (*discordgo.MessageAllowedMentions, error) {
	mapping, _ := raw.(map[string]any)
	if len(mapping) == 0 {
		return nil, nil
	}
	mentions := &discordgo.MessageAllowedMentions{}
	if parseRaw, ok := mapping["parse"].([]any); ok {
		for _, item := range parseRaw {
			switch strings.ToLower(strings.TrimSpace(fmt.Sprint(item))) {
			case "users":
				mentions.Parse = append(mentions.Parse, discordgo.AllowedMentionTypeUsers)
			case "roles":
				mentions.Parse = append(mentions.Parse, discordgo.AllowedMentionTypeRoles)
			case "everyone":
				mentions.Parse = append(mentions.Parse, discordgo.AllowedMentionTypeEveryone)
			}
		}
	}
	if repliedUser, ok := mapping["repliedUser"].(bool); ok {
		mentions.RepliedUser = repliedUser
	}
	if usersRaw, ok := mapping["users"].([]any); ok {
		mentions.Users = stringSlice(usersRaw)
	}
	if rolesRaw, ok := mapping["roles"].([]any); ok {
		mentions.Roles = stringSlice(rolesRaw)
	}
	return mentions, nil
}

func normalizeFiles(raw any) ([]*discordgo.File, error) {
	switch v := raw.(type) {
	case nil:
		return nil, nil
	case []any:
		files := make([]*discordgo.File, 0, len(v))
		for _, item := range v {
			mapping, _ := item.(map[string]any)
			if len(mapping) == 0 {
				continue
			}
			name := strings.TrimSpace(fmt.Sprint(mapping["name"]))
			if name == "" {
				return nil, fmt.Errorf("file payload requires name")
			}
			content, ok := mapping["content"]
			if !ok {
				return nil, fmt.Errorf("file payload %q requires content", name)
			}
			file := &discordgo.File{Name: name, Reader: bytes.NewReader([]byte(fmt.Sprint(content)))}
			if contentType, ok := mapping["contentType"]; ok {
				file.ContentType = fmt.Sprint(contentType)
			}
			files = append(files, file)
		}
		return files, nil
	default:
		return nil, fmt.Errorf("unsupported files payload type %T", raw)
	}
}

func normalizeMessageReference(raw any) (*discordgo.MessageReference, error) {
	mapping, _ := raw.(map[string]any)
	if len(mapping) == 0 {
		return nil, nil
	}
	messageID := strings.TrimSpace(fmt.Sprint(mapping["messageId"]))
	if messageID == "" {
		return nil, fmt.Errorf("replyTo.messageId is required")
	}
	reference := &discordgo.MessageReference{MessageID: messageID}
	if channelID := strings.TrimSpace(fmt.Sprint(mapping["channelId"])); channelID != "" {
		reference.ChannelID = channelID
	}
	if guildID := strings.TrimSpace(fmt.Sprint(mapping["guildId"])); guildID != "" {
		reference.GuildID = guildID
	}
	return reference, nil
}

func intValue(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case float32:
		return int(v), true
	default:
		return 0, false
	}
}

func boolValue(value any) bool {
	ret, _ := value.(bool)
	return ret
}

func intPointer(value any) (*int, bool) {
	if v, ok := intValue(value); ok {
		ret := v
		return &ret, true
	}
	return nil, false
}

func floatValue(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

func floatPointer(value any) (*float64, bool) {
	if v, ok := floatValue(value); ok {
		ret := v
		return &ret, true
	}
	return nil, false
}

func stringSlice(values []any) []string {
	ret := make([]string, 0, len(values))
	for _, value := range values {
		ret = append(ret, fmt.Sprint(value))
	}
	return ret
}
