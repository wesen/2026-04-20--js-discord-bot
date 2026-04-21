package jsdiscord

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func normalizeTimeoutUntil(payload any) (*time.Time, error) {
	switch v := payload.(type) {
	case nil:
		return nil, nil
	case int:
		return timeoutFromDurationSeconds(int64(v))
	case int64:
		return timeoutFromDurationSeconds(v)
	case float64:
		return timeoutFromDurationSeconds(int64(v))
	case map[string]any:
		if clear, ok := v["clear"].(bool); ok && clear {
			return nil, nil
		}
		if untilRaw, ok := v["until"]; ok {
			until := strings.TrimSpace(fmt.Sprint(untilRaw))
			if until == "" {
				return nil, fmt.Errorf("member timeout until value is empty")
			}
			parsed, err := time.Parse(time.RFC3339, until)
			if err != nil {
				return nil, fmt.Errorf("parse member timeout until: %w", err)
			}
			return &parsed, nil
		}
		if seconds, ok := int64Value(v["durationSeconds"]); ok {
			return timeoutFromDurationSeconds(seconds)
		}
		return nil, fmt.Errorf("member timeout payload must include clear, until, or durationSeconds")
	default:
		return nil, fmt.Errorf("unsupported member timeout payload type %T", payload)
	}
}

func timeoutFromDurationSeconds(seconds int64) (*time.Time, error) {
	if seconds <= 0 {
		return nil, nil
	}
	until := time.Now().UTC().Add(time.Duration(seconds) * time.Second)
	return &until, nil
}

func normalizeModerationReason(payload any) string {
	switch v := payload.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	case map[string]any:
		return strings.TrimSpace(fmt.Sprint(v["reason"]))
	default:
		return strings.TrimSpace(fmt.Sprint(payload))
	}
}

func normalizeBanOptions(payload any) (string, int, error) {
	reason := ""
	deleteMessageDays := 0
	switch v := payload.(type) {
	case nil:
		return reason, deleteMessageDays, nil
	case string:
		return strings.TrimSpace(v), deleteMessageDays, nil
	case map[string]any:
		reason = strings.TrimSpace(fmt.Sprint(v["reason"]))
		if days, ok := int64Value(v["deleteMessageDays"]); ok {
			deleteMessageDays = int(days)
		}
		if deleteMessageDays < 0 {
			return "", 0, fmt.Errorf("ban deleteMessageDays must be >= 0")
		}
		return reason, deleteMessageDays, nil
	default:
		return "", 0, fmt.Errorf("unsupported member ban payload type %T", payload)
	}
}

type memberListOptions struct {
	After string
	Limit int
}

type messageListOptions struct {
	Before string
	After  string
	Around string
	Limit  int
}

func normalizeMemberListOptions(payload any) (memberListOptions, error) {
	ret := memberListOptions{Limit: 25}
	switch v := payload.(type) {
	case nil:
		return ret, nil
	case map[string]any:
		ret.After = optionalStringValue(v, "after")
		if limit, ok := int64Value(v["limit"]); ok {
			ret.Limit = int(limit)
		}
		if ret.Limit <= 0 {
			ret.Limit = 25
		}
		if ret.Limit > 1000 {
			ret.Limit = 1000
		}
		return ret, nil
	default:
		return memberListOptions{}, fmt.Errorf("unsupported member list payload type %T", payload)
	}
}

type threadStartOptions struct {
	MessageID string
	Data      *discordgo.ThreadStart
}

func optionalStringValue(mapping map[string]any, key string) string {
	raw, ok := mapping[key]
	if !ok || raw == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(raw))
}

func normalizeThreadStartOptions(payload any) (threadStartOptions, error) {
	switch v := payload.(type) {
	case string:
		name := strings.TrimSpace(v)
		if name == "" {
			return threadStartOptions{}, fmt.Errorf("thread start name is empty")
		}
		return threadStartOptions{Data: &discordgo.ThreadStart{Name: name, AutoArchiveDuration: 1440, Type: discordgo.ChannelTypeGuildPublicThread}}, nil
	case map[string]any:
		name := optionalStringValue(v, "name")
		if name == "" {
			return threadStartOptions{}, fmt.Errorf("thread start payload missing name")
		}
		messageID := optionalStringValue(v, "messageId")
		autoArchive := 1440
		if value, ok := int64Value(v["autoArchiveDuration"]); ok {
			autoArchive = int(value)
		}
		threadType, err := threadTypeValue(v["type"], messageID != "")
		if err != nil {
			return threadStartOptions{}, err
		}
		ret := threadStartOptions{
			MessageID: messageID,
			Data: &discordgo.ThreadStart{
				Name:                name,
				AutoArchiveDuration: autoArchive,
				Type:                threadType,
				Invitable:           boolValue(v["invitable"]),
				RateLimitPerUser:    intValueOrZero(v["rateLimitPerUser"]),
			},
		}
		return ret, nil
	default:
		return threadStartOptions{}, fmt.Errorf("unsupported thread start payload type %T", payload)
	}
}

func threadTypeValue(raw any, fromMessage bool) (discordgo.ChannelType, error) {
	if fromMessage && raw == nil {
		return 0, nil
	}
	if raw == nil {
		return discordgo.ChannelTypeGuildPublicThread, nil
	}
	switch v := raw.(type) {
	case string:
		switch strings.TrimSpace(strings.ToLower(v)) {
		case "", "public", "public-thread", "guild-public-thread":
			return discordgo.ChannelTypeGuildPublicThread, nil
		case "private", "private-thread", "guild-private-thread":
			return discordgo.ChannelTypeGuildPrivateThread, nil
		case "news", "news-thread", "guild-news-thread":
			return discordgo.ChannelTypeGuildNewsThread, nil
		default:
			return 0, fmt.Errorf("unsupported thread type %q", v)
		}
	case int:
		return discordgo.ChannelType(v), nil
	case int64:
		return discordgo.ChannelType(v), nil
	case float64:
		return discordgo.ChannelType(int(v)), nil
	default:
		return 0, fmt.Errorf("unsupported thread type value %T", raw)
	}
}

func intValueOrZero(value any) int {
	if n, ok := int64Value(value); ok {
		return int(n)
	}
	return 0
}

func normalizeMessageListOptions(payload any) (messageListOptions, error) {
	ret := messageListOptions{Limit: 25}
	switch v := payload.(type) {
	case nil:
		return ret, nil
	case map[string]any:
		ret.Before = optionalStringValue(v, "before")
		ret.After = optionalStringValue(v, "after")
		ret.Around = optionalStringValue(v, "around")
		anchors := 0
		if ret.Before != "" {
			anchors++
		}
		if ret.After != "" {
			anchors++
		}
		if ret.Around != "" {
			anchors++
		}
		if anchors > 1 {
			return messageListOptions{}, fmt.Errorf("message list payload may include only one of before, after, or around")
		}
		if limit, ok := int64Value(v["limit"]); ok {
			ret.Limit = int(limit)
		}
		if ret.Limit <= 0 {
			ret.Limit = 25
		}
		if ret.Limit > 100 {
			ret.Limit = 100
		}
		return ret, nil
	default:
		return messageListOptions{}, fmt.Errorf("unsupported message list payload type %T", payload)
	}
}

func normalizeMessageIDList(payload any) ([]string, error) {
	switch v := payload.(type) {
	case nil:
		return nil, fmt.Errorf("message bulkDelete requires at least one message ID")
	case []string:
		return cleanedMessageIDs(v)
	case []any:
		ids := make([]string, 0, len(v))
		for _, item := range v {
			id := strings.TrimSpace(fmt.Sprint(item))
			if id != "" {
				ids = append(ids, id)
			}
		}
		return cleanedMessageIDs(ids)
	case map[string]any:
		if ids, ok := v["messageIds"]; ok {
			return normalizeMessageIDList(ids)
		}
		return nil, fmt.Errorf("message bulkDelete payload must include messageIds")
	default:
		return nil, fmt.Errorf("unsupported message bulkDelete payload type %T", payload)
	}
}

func cleanedMessageIDs(ids []string) ([]string, error) {
	ret := make([]string, 0, len(ids))
	seen := map[string]struct{}{}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ret = append(ret, id)
	}
	if len(ret) == 0 {
		return nil, fmt.Errorf("message bulkDelete requires at least one message ID")
	}
	return ret, nil
}

func int64Value(value any) (int64, bool) {
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int64:
		return v, true
	case float64:
		return int64(v), true
	default:
		return 0, false
	}
}
