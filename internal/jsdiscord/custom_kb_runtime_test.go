package jsdiscord

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCustomKBBotStoresAndSearchesLinksWithSQLiteAndUIDSL(t *testing.T) {
	scriptPath := filepath.Join(repoRootJSDiscord(t), "examples", "discord-bots", "custom-kb", "index.js")
	handle := loadTestBot(t, scriptPath)
	config := map[string]any{"dbPath": filepath.Join(t.TempDir(), "custom-kb.sqlite"), "searchLimit": 5}

	addResult, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name:    "kb-link",
		Args:    map[string]any{"url": "https://example.com/sqlite-ui-dsl", "title": "SQLite UI DSL Notes", "summary": "Links about the Discord UI DSL and SQLite persistence.", "tags": "sqlite,ui,dsl"},
		User:    UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
		Config:  config,
	})
	require.NoError(t, err)
	require.Contains(t, fmt.Sprint(addResult), "https://example.com/sqlite-ui-dsl")

	searchResult, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name:    "kb-search",
		Args:    map[string]any{"query": "sqlite"},
		User:    UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
		Config:  config,
	})
	require.NoError(t, err)
	searchResponse := searchResult.(*normalizedResponse)
	require.Contains(t, searchResponse.Content, "Found 1 link")
	require.NotEmpty(t, searchResponse.Embeds)
	require.Contains(t, searchResponse.Embeds[0].Description, "SQLite UI DSL Notes")
	require.Contains(t, fmt.Sprint(searchResponse.Components), "custom-kb:search:select")
	require.Contains(t, fmt.Sprint(searchResponse.Components), "custom-kb:search:refresh")
}
