package jsdiscord

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

func payloadLogFields(payload any) map[string]any {
	fields := map[string]any{
		"payloadType": fmt.Sprintf("%T", payload),
	}
	switch v := payload.(type) {
	case nil:
		return fields
	case string:
		fields["contentPreview"] = truncateLogText(v, 120)
		return fields
	case *normalizedResponse:
		if v.Content != "" {
			fields["contentPreview"] = truncateLogText(v.Content, 120)
		}
		fields["ephemeral"] = v.Ephemeral
		fields["embedCount"] = len(v.Embeds)
		fields["componentCount"] = len(v.Components)
		fields["fileCount"] = len(v.Files)
		return fields
	case map[string]any:
		if content, ok := v["content"]; ok {
			fields["contentPreview"] = truncateLogText(fmt.Sprint(content), 120)
		}
		if ephemeral, ok := v["ephemeral"].(bool); ok {
			fields["ephemeral"] = ephemeral
		}
		if customID := strings.TrimSpace(fmt.Sprint(v["customId"])); customID != "" {
			fields["customId"] = customID
		}
		if title := strings.TrimSpace(fmt.Sprint(v["title"])); title != "" {
			fields["title"] = title
		}
		if embeds, ok := sliceLen(v["embeds"]); ok {
			fields["embedCount"] = embeds
		}
		if components, ok := sliceLen(v["components"]); ok {
			fields["componentCount"] = components
		}
		if files, ok := sliceLen(v["files"]); ok {
			fields["fileCount"] = files
		}
	}
	return fields
}

func sliceLen(value any) (int, bool) {
	switch v := value.(type) {
	case []any:
		return len(v), true
	case []map[string]any:
		return len(v), true
	case []string:
		return len(v), true
	default:
		return 0, false
	}
}

func truncateLogText(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func mergeLogFields(base map[string]any, extras ...map[string]any) map[string]any {
	ret := map[string]any{}
	for key, value := range base {
		ret[key] = value
	}
	for _, extra := range extras {
		for key, value := range extra {
			ret[key] = value
		}
	}
	return ret
}

func logLifecycleDebug(message string, fields map[string]any) {
	e := log.Debug()
	applyFields(e, fields)
	e.Msg(message)
}
