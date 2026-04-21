package jsdiscord

import (
	"fmt"
	"strings"
	"time"
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
