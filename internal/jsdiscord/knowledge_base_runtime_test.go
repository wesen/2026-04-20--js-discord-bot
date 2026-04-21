package jsdiscord

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKnowledgeBaseBotUsesSQLiteStoreForCaptureSearchAndReview(t *testing.T) {
	scriptPath := filepath.Join(repoRootJSDiscord(t), "examples", "discord-bots", "knowledge-base", "index.js")
	handle := loadTestBot(t, scriptPath)

	config := map[string]any{
		"dbPath":           filepath.Join(t.TempDir(), "knowledge.sqlite"),
		"captureEnabled":   true,
		"captureThreshold": 0.2,
		"reviewLimit":      5,
		"seedEntries":      false,
	}

	var replies []any
	_, err := handle.DispatchEvent(context.Background(), DispatchRequest{
		Name: "messageCreate",
		Message: map[string]any{
			"id":        "msg-1",
			"content":   "Here is the fix:\n```js\ndb.configure(\"sqlite3\", \":memory:\")\n```\nUse /teach to save this knowledge.",
			"guildID":   "guild-1",
			"channelID": "channel-1",
			"author":    map[string]any{"id": "user-1", "username": "Ada", "bot": false},
		},
		User:    map[string]any{"id": "user-1", "username": "Ada", "bot": false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
		Config:  config,
		Reply: func(_ context.Context, value any) error {
			replies = append(replies, value)
			return nil
		},
	})
	require.NoError(t, err)
	require.Len(t, replies, 1)
	require.Contains(t, fmt.Sprint(replies[0]), "Captured knowledge entry")

	searchResult, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name:   "kb-search",
		Args:   map[string]any{"query": "sqlite"},
		Config: config,
	})
	require.NoError(t, err)
	searchMap := searchResult.(map[string]any)
	require.Contains(t, fmt.Sprint(searchMap["content"]), "sqlite")
	searchEmbeds := searchMap["embeds"].([]any)
	require.NotEmpty(t, searchEmbeds)
	require.Contains(t, fmt.Sprint(searchEmbeds[0]), "Here is the fix")

	reviewResult, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name:   "kb-review",
		Config: config,
	})
	require.NoError(t, err)
	reviewMap := reviewResult.(map[string]any)
	reviewEmbeds := reviewMap["embeds"].([]any)
	require.NotEmpty(t, reviewEmbeds)
	reviewEmbed := reviewEmbeds[0].(map[string]any)
	reviewFields := reviewEmbed["fields"].([]any)
	require.NotEmpty(t, reviewFields)
	entryID := extractFieldValue(reviewFields, "Entry ID")
	require.NotEmpty(t, entryID)
	statusBefore := extractFieldValue(reviewFields, "Status")
	require.Equal(t, "draft", statusBefore)
	components := reviewMap["components"].([]any)
	require.Len(t, components, 2)
	require.Contains(t, fmt.Sprint(components[0]), "knowledge:review:select")
	require.Contains(t, fmt.Sprint(components[1]), "knowledge:review:verify")

	selectResult, err := handle.DispatchComponent(context.Background(), DispatchRequest{
		Name:        "knowledge:review:select",
		Values:      []string{entryID},
		Config:      config,
		User:        map[string]any{"id": "user-1", "username": "Ada", "bot": false},
		Guild:       map[string]any{"id": "guild-1"},
		Channel:     map[string]any{"id": "channel-1"},
		Interaction: map[string]any{"id": "interaction-select"},
	})
	require.NoError(t, err)
	require.Contains(t, fmt.Sprint(selectResult), "Selected")

	verifyResult, err := handle.DispatchComponent(context.Background(), DispatchRequest{
		Name:        "knowledge:review:verify",
		Config:      config,
		User:        map[string]any{"id": "user-1", "username": "Ada", "bot": false},
		Guild:       map[string]any{"id": "guild-1"},
		Channel:     map[string]any{"id": "channel-1"},
		Interaction: map[string]any{"id": "interaction-verify"},
	})
	require.NoError(t, err)
	require.Contains(t, fmt.Sprint(verifyResult), "Verified")

	articleResult, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name:   "article",
		Args:   map[string]any{"name": entryID},
		Config: config,
	})
	require.NoError(t, err)
	articleMap := articleResult.(map[string]any)
	require.Contains(t, fmt.Sprint(articleMap["content"]), "Opened knowledge entry")
	require.Contains(t, fmt.Sprint(articleMap["embeds"]), "verified")
}

func extractFieldValue(fields []any, name string) string {
	for _, item := range fields {
		field, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if fmt.Sprint(field["name"]) == name {
			return fmt.Sprint(field["value"])
		}
	}
	return ""
}

func repoRootJSDiscord(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	return root
}
