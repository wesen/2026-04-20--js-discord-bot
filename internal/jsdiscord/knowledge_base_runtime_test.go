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
		"dbPath":                filepath.Join(t.TempDir(), "knowledge.sqlite"),
		"captureEnabled":        true,
		"captureThreshold":      0.2,
		"reviewLimit":           5,
		"seedEntries":           false,
		"reactionPromoteEmojis": "🧠",
		"trustedReviewerIds":    "user-2",
	}

	var (
		replies         []any
		searchDefers    []any
		searchEdits     []any
		searchFollowUps []any
	)
	_, err := handle.DispatchEvent(context.Background(), DispatchRequest{
		Name: "messageCreate",
		Message: &MessageSnapshot{
			ID:        "msg-1",
			Content:   "Here is the fix:\n```js\ndb.configure(\"sqlite3\", \":memory:\")\n```\nUse /teach to save this knowledge.",
			GuildID:   "guild-1",
			ChannelID: "channel-1",
			Author:    &UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		},
		User:    UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
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

	_, err = handle.DispatchEvent(context.Background(), DispatchRequest{
		Name: "messageCreate",
		Message: &MessageSnapshot{
			ID:        "msg-2",
			Content:   "SQLite also stores the project notes for this bot so search can find them later.",
			GuildID:   "guild-1",
			ChannelID: "channel-1",
			Author:    &UserSnapshot{ID: "user-3", Username: "Lin", Bot: false},
		},
		User:    UserSnapshot{ID: "user-3", Username: "Lin", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
		Config:  config,
		Reply: func(_ context.Context, value any) error {
			replies = append(replies, value)
			return nil
		},
	})
	require.NoError(t, err)
	require.Len(t, replies, 2)

	_, err = handle.DispatchEvent(context.Background(), DispatchRequest{
		Name: "reactionAdd",
		Message: &MessageSnapshot{
			ID:        "msg-1",
			GuildID:   "guild-1",
			ChannelID: "channel-1",
		},
		User:    UserSnapshot{ID: "user-2", Username: "Reviewer", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
		Reaction: ReactionSnapshot{
			Emoji: EmojiSnapshot{Name: "🧠", Animated: false},
		},
		Config: config,
		Reply: func(_ context.Context, value any) error {
			replies = append(replies, value)
			return nil
		},
	})
	require.NoError(t, err)

	searchResult, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name:    "kb-search",
		Args:    map[string]any{"query": "sqlite"},
		User:    UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
		Config:  config,
	})
	require.NoError(t, err)
	searchMap := searchResult.(map[string]any)
	require.Contains(t, fmt.Sprint(searchMap["content"]), "sqlite")
	searchEmbeds := searchMap["embeds"].([]any)
	require.NotEmpty(t, searchEmbeds)
	searchEmbed := searchEmbeds[0].(map[string]any)
	searchFields := searchEmbed["fields"].([]any)
	require.NotEmpty(t, searchFields)
	searchEntryID := extractFieldValue(searchFields, "Entry ID")
	require.NotEmpty(t, searchEntryID)
	require.Contains(t, extractFieldValue(searchFields, "Source citation"), "channel-1")
	require.Contains(t, extractFieldValue(searchFields, "Source details"), "msg-")
	require.NotEmpty(t, extractFieldValue(searchFields, "Canonical source"))
	require.NotEqual(t, "(unknown)", extractFieldValue(searchFields, "Canonical source"))
	require.Contains(t, fmt.Sprint(searchEmbed["footer"]), "Search: sqlite")
	searchComponents := searchMap["components"].([]any)
	require.Len(t, searchComponents, 3)
	require.Contains(t, fmt.Sprint(searchComponents[0]), "knowledge:search:select")
	require.Contains(t, fmt.Sprint(searchComponents[1]), "knowledge:search:previous")
	require.Contains(t, fmt.Sprint(searchComponents[2]), "knowledge:search:export")

	searchSelectResult, err := handle.DispatchComponent(context.Background(), DispatchRequest{
		Name:        "knowledge:search:select",
		Values:      []string{searchEntryID},
		Config:      config,
		User:        UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:       map[string]any{"id": "guild-1"},
		Channel:     map[string]any{"id": "channel-1"},
		Interaction: InteractionSnapshot{ID: "interaction-search-select"},
	})
	require.NoError(t, err)
	require.Contains(t, fmt.Sprint(searchSelectResult), "Found")
	require.Contains(t, fmt.Sprint(searchSelectResult), "knowledge:search:previous")

	searchSourceResult, err := handle.DispatchComponent(context.Background(), DispatchRequest{
		Name:        "knowledge:search:source",
		Config:      config,
		User:        UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:       map[string]any{"id": "guild-1"},
		Channel:     map[string]any{"id": "channel-1"},
		Interaction: InteractionSnapshot{ID: "interaction-search-source"},
	})
	require.NoError(t, err)
	require.Contains(t, fmt.Sprint(searchSourceResult), "source citation")
	require.Contains(t, fmt.Sprint(searchSourceResult), "msg-")

	searchOpenResult, err := handle.DispatchComponent(context.Background(), DispatchRequest{
		Name:        "knowledge:search:open",
		Config:      config,
		User:        UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:       map[string]any{"id": "guild-1"},
		Channel:     map[string]any{"id": "channel-1"},
		Interaction: InteractionSnapshot{ID: "interaction-search-open"},
	})
	require.NoError(t, err)
	require.Contains(t, fmt.Sprint(searchOpenResult), "Opened knowledge entry")

	searchNextResult, err := handle.DispatchComponent(context.Background(), DispatchRequest{
		Name:        "knowledge:search:next",
		Config:      config,
		User:        UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:       map[string]any{"id": "guild-1"},
		Channel:     map[string]any{"id": "channel-1"},
		Interaction: InteractionSnapshot{ID: "interaction-search-next"},
	})
	require.NoError(t, err)
	require.Contains(t, fmt.Sprint(searchNextResult), "Found")

	_, err = handle.DispatchComponent(context.Background(), DispatchRequest{
		Name:        "knowledge:search:export",
		Config:      config,
		User:        UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:       map[string]any{"id": "guild-1"},
		Channel:     map[string]any{"id": "channel-1"},
		Interaction: InteractionSnapshot{ID: "interaction-search-export"},
		Defer: func(_ context.Context, value any) error {
			searchDefers = append(searchDefers, value)
			return nil
		},
		Edit: func(_ context.Context, value any) error {
			searchEdits = append(searchEdits, value)
			return nil
		},
		FollowUp: func(_ context.Context, value any) error {
			searchFollowUps = append(searchFollowUps, value)
			return nil
		},
	})
	require.NoError(t, err)
	require.Len(t, searchDefers, 1)
	require.Len(t, searchFollowUps, 1)
	require.Len(t, searchEdits, 1)
	require.Contains(t, fmt.Sprint(searchFollowUps[0]), "Shared from /ask for sqlite")
	require.Contains(t, fmt.Sprint(searchFollowUps[0]), "Source citation")
	require.Contains(t, fmt.Sprint(searchFollowUps[0]), "msg-")
	require.Contains(t, fmt.Sprint(searchEdits[0]), "Exported")

	askAutocomplete, err := handle.DispatchAutocomplete(context.Background(), DispatchRequest{
		Name:    "ask",
		Args:    map[string]any{"query": "sqlite"},
		Focused: FocusedOptionSnapshot{Name: "query", Value: "sqlite"},
		Command: map[string]any{"name": "ask"},
		Config:  config,
	})
	require.NoError(t, err)
	askItems, ok := askAutocomplete.([]any)
	require.True(t, ok)
	require.NotEmpty(t, askItems)

	articleAutocomplete, err := handle.DispatchAutocomplete(context.Background(), DispatchRequest{
		Name:    "article",
		Args:    map[string]any{"name": "kb_"},
		Focused: FocusedOptionSnapshot{Name: "name", Value: "kb_"},
		Command: map[string]any{"name": "article"},
		Config:  config,
	})
	require.NoError(t, err)
	articleItems, ok := articleAutocomplete.([]any)
	require.True(t, ok)
	require.NotEmpty(t, articleItems)

	reviewResult, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name:   "kb-review",
		Args:   map[string]any{"status": "review"},
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
	require.Equal(t, "review", statusBefore)
	components := reviewMap["components"].([]any)
	require.Len(t, components, 2)
	require.Contains(t, fmt.Sprint(components[0]), "knowledge:review:select")
	require.Contains(t, fmt.Sprint(components[1]), "knowledge:review:verify")

	selectResult, err := handle.DispatchComponent(context.Background(), DispatchRequest{
		Name:        "knowledge:review:select",
		Values:      []string{entryID},
		Config:      config,
		User:        UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:       map[string]any{"id": "guild-1"},
		Channel:     map[string]any{"id": "channel-1"},
		Interaction: InteractionSnapshot{ID: "interaction-select"},
	})
	require.NoError(t, err)
	require.Contains(t, fmt.Sprint(selectResult), "Selected")

	verifyResult, err := handle.DispatchComponent(context.Background(), DispatchRequest{
		Name:        "knowledge:review:verify",
		Config:      config,
		User:        UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:       map[string]any{"id": "guild-1"},
		Channel:     map[string]any{"id": "channel-1"},
		Interaction: InteractionSnapshot{ID: "interaction-verify"},
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
