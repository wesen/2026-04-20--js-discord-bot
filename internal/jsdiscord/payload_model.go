package jsdiscord

import (
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
	case *normalizedResponse:
		return v, nil
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
