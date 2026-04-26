package jsdiscord

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUIShowcaseBotMessageBuilders(t *testing.T) {
	scriptPath := filepath.Join(repoRootJSDiscord(t), "examples", "discord-bots", "ui-showcase", "index.js")
	handle := loadTestBot(t, scriptPath)

	result, err := handle.DispatchCommandAsMap(context.Background(), DispatchRequest{
		Name:    "demo-message",
		User:    UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
	})
	require.NoError(t, err)

	msg := result
	require.Contains(t, fmt.Sprint(msg["content"]), "UI DSL builder")
	require.True(t, msg["ephemeral"].(bool))

	embeds := msg["embeds"].([]any)
	require.Len(t, embeds, 1)
	embed := embeds[0].(map[string]any)
	require.Contains(t, fmt.Sprint(embed["title"]), "Builder Showcase")
	require.NotEmpty(t, embed["fields"])

	components := msg["components"].([]any)
	require.Len(t, components, 2)
	require.Contains(t, fmt.Sprint(components[0]), "showcase:msg:primary")
	require.Contains(t, fmt.Sprint(components[1]), "showcase:msg:select")
}

func TestUIShowcaseBotSearchFlow(t *testing.T) {
	scriptPath := filepath.Join(repoRootJSDiscord(t), "examples", "discord-bots", "ui-showcase", "index.js")
	handle := loadTestBot(t, scriptPath)

	config := map[string]any{}

	// Search for "discord" — should find multiple articles
	result, err := handle.DispatchCommandAsMap(context.Background(), DispatchRequest{
		Name:    "demo-search",
		Args:    map[string]any{"query": "discord"},
		User:    UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
		Config:  config,
	})
	require.NoError(t, err)

	msg := result
	require.Contains(t, fmt.Sprint(msg["content"]), "discord")
	require.Contains(t, fmt.Sprint(msg["content"]), "Found")

	embeds := msg["embeds"].([]any)
	require.NotEmpty(t, embeds)
	embed := embeds[0].(map[string]any)
	require.NotEmpty(t, embed["title"])

	components := msg["components"].([]any)
	require.Len(t, components, 3) // select + pager + action buttons

	// Verify select, pager, and action buttons are present
	require.Contains(t, fmt.Sprint(components[0]), "showcase.search:select")
	require.Contains(t, fmt.Sprint(components[1]), "showcase.search:previous")
	require.Contains(t, fmt.Sprint(components[2]), "showcase.search:open")

	// Click "next" to paginate
	nextResult, err := handle.DispatchComponentAsMap(context.Background(), DispatchRequest{
		Name:    "showcase.search:next",
		User:    UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
		Config:  config,
	})
	require.NoError(t, err)
	nextMsg := nextResult
	require.Contains(t, fmt.Sprint(nextMsg["content"]), "discord")

	// Click "previous" to go back
	prevResult, err := handle.DispatchComponentAsMap(context.Background(), DispatchRequest{
		Name:    "showcase.search:previous",
		User:    UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
		Config:  config,
	})
	require.NoError(t, err)
	prevMsg := prevResult
	require.Contains(t, fmt.Sprint(prevMsg["content"]), "discord")

	// Open the selected article
	openResult, err := handle.DispatchComponentAsMap(context.Background(), DispatchRequest{
		Name:    "showcase.search:open",
		User:    UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
		Config:  config,
	})
	require.NoError(t, err)
	openMsg := openResult
	require.Contains(t, fmt.Sprint(openMsg["content"]), "Opened")

	// Search alias (/find) should return the same shape
	aliasResult, err := handle.DispatchCommandAsMap(context.Background(), DispatchRequest{
		Name:    "find",
		Args:    map[string]any{"query": "pagination"},
		User:    UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
		Config:  config,
	})
	require.NoError(t, err)
	aliasMsg := aliasResult
	require.Contains(t, fmt.Sprint(aliasMsg["content"]), "pagination")

	// Autocomplete
	acResult, err := handle.DispatchAutocomplete(context.Background(), DispatchRequest{
		Name:    "demo-search",
		Args:    map[string]any{"query": "ui"},
		Focused: FocusedOptionSnapshot{Name: "query", Value: "ui"},
		Command: map[string]any{"name": "demo-search"},
		Config:  config,
	})
	require.NoError(t, err)
	acItems, ok := acResult.([]any)
	require.True(t, ok)
	require.NotEmpty(t, acItems)
}

func TestUIShowcaseBotReviewFlow(t *testing.T) {
	scriptPath := filepath.Join(repoRootJSDiscord(t), "examples", "discord-bots", "ui-showcase", "index.js")
	handle := loadTestBot(t, scriptPath)

	config := map[string]any{}

	// Open review queue for "review" status
	result, err := handle.DispatchCommandAsMap(context.Background(), DispatchRequest{
		Name:    "demo-review",
		Args:    map[string]any{"status": "review"},
		User:    UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
		Config:  config,
	})
	require.NoError(t, err)

	msg := result
	require.Contains(t, fmt.Sprint(msg["content"]), "review")
	require.Contains(t, fmt.Sprint(msg["content"]), "Review queue")

	embeds := msg["embeds"].([]any)
	require.NotEmpty(t, embeds)

	components := msg["components"].([]any)
	require.Len(t, components, 2) // select + action buttons
	require.Contains(t, fmt.Sprint(components[0]), "showcase.review:select")
	require.Contains(t, fmt.Sprint(components[1]), "showcase.review:verify")
	require.Contains(t, fmt.Sprint(components[1]), "showcase.review:reject")

	// Get the selected entry ID from the embed fields
	reviewEmbed := embeds[0].(map[string]any)
	reviewFields := reviewEmbed["fields"].([]any)
	entryID := extractFieldValue(reviewFields, "Status")
	require.NotEmpty(t, entryID)

	// Verify the article
	verifyResult, err := handle.DispatchComponentAsMap(context.Background(), DispatchRequest{
		Name:    "showcase.review:verify",
		User:    UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
		Config:  config,
	})
	require.NoError(t, err)
	verifyMsg := verifyResult
	require.Contains(t, fmt.Sprint(verifyMsg["content"]), "Verified")
}

func TestUIShowcaseBotConfirmDialog(t *testing.T) {
	scriptPath := filepath.Join(repoRootJSDiscord(t), "examples", "discord-bots", "ui-showcase", "index.js")
	handle := loadTestBot(t, scriptPath)

	// Trigger confirmation
	result, err := handle.DispatchCommandAsMap(context.Background(), DispatchRequest{
		Name:    "demo-confirm",
		Args:    map[string]any{"action": "delete everything"},
		User:    UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
	})
	require.NoError(t, err)

	msg := result
	require.True(t, msg["ephemeral"].(bool))
	require.Contains(t, fmt.Sprint(msg["content"]), "delete everything")

	embeds := msg["embeds"].([]any)
	require.NotEmpty(t, embeds)
	require.Contains(t, fmt.Sprint(embeds[0]), "Confirm")

	components := msg["components"].([]any)
	require.Len(t, components, 1)
	require.Contains(t, fmt.Sprint(components[0]), "showcase:confirm:yes")
	require.Contains(t, fmt.Sprint(components[0]), "showcase:confirm:no")

	// Confirm the action
	confirmResult, err := handle.DispatchComponentAsMap(context.Background(), DispatchRequest{
		Name:    "showcase:confirm:yes",
		User:    UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
	})
	require.NoError(t, err)
	confirmMsg := confirmResult
	require.Contains(t, fmt.Sprint(confirmMsg["content"]), "Confirmed")

	// Cancel the action
	cancelResult, err := handle.DispatchComponentAsMap(context.Background(), DispatchRequest{
		Name:    "showcase:confirm:no",
		User:    UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
	})
	require.NoError(t, err)
	cancelMsg := cancelResult
	require.Contains(t, fmt.Sprint(cancelMsg["content"]), "Cancelled")
}

func TestUIShowcaseBotPager(t *testing.T) {
	scriptPath := filepath.Join(repoRootJSDiscord(t), "examples", "discord-bots", "ui-showcase", "index.js")
	handle := loadTestBot(t, scriptPath)

	result, err := handle.DispatchCommandAsMap(context.Background(), DispatchRequest{
		Name:    "demo-pager",
		User:    UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
	})
	require.NoError(t, err)

	msg := result
	require.Contains(t, fmt.Sprint(msg["content"]), "page 1/")

	components := msg["components"].([]any)
	require.Len(t, components, 1)
	require.Contains(t, fmt.Sprint(components[0]), "showcase.pager:previous")

	// Next page
	nextResult, err := handle.DispatchComponentAsMap(context.Background(), DispatchRequest{
		Name:    "showcase.pager:next",
		User:    UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
	})
	require.NoError(t, err)
	nextMsg := nextResult
	require.Contains(t, fmt.Sprint(nextMsg["content"]), "page 2/")

	// Previous page
	prevResult, err := handle.DispatchComponentAsMap(context.Background(), DispatchRequest{
		Name:    "showcase.pager:previous",
		User:    UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
	})
	require.NoError(t, err)
	prevMsg := prevResult
	require.Contains(t, fmt.Sprint(prevMsg["content"]), "page 1/")
}

func TestUIShowcaseBotCardGallery(t *testing.T) {
	scriptPath := filepath.Join(repoRootJSDiscord(t), "examples", "discord-bots", "ui-showcase", "index.js")
	handle := loadTestBot(t, scriptPath)

	result, err := handle.DispatchCommandAsMap(context.Background(), DispatchRequest{
		Name:    "demo-cards",
		User:    UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
	})
	require.NoError(t, err)

	msg := result
	require.Contains(t, fmt.Sprint(msg["content"]), "catalog")

	embeds := msg["embeds"].([]any)
	require.NotEmpty(t, embeds)
	embed := embeds[0].(map[string]any)
	require.Contains(t, fmt.Sprint(embed["title"]), "Widget Pro")

	components := msg["components"].([]any)
	require.Len(t, components, 2) // select + action buttons

	// Select a different product
	selectResult, err := handle.DispatchComponentAsMap(context.Background(), DispatchRequest{
		Name:    "showcase.cards:select",
		Values:  []string{"prod-3"},
		User:    UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
	})
	require.NoError(t, err)
	selectMsg := selectResult
	selectEmbeds := selectMsg["embeds"].([]any)
	selectEmbed := selectEmbeds[0].(map[string]any)
	require.Contains(t, fmt.Sprint(selectEmbed["title"]), "Doodad Max")

	// Get info for the selected product
	infoResult, err := handle.DispatchComponentAsMap(context.Background(), DispatchRequest{
		Name:    "showcase.cards:info",
		User:    UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
	})
	require.NoError(t, err)
	infoMsg := infoResult
	require.Contains(t, fmt.Sprint(infoMsg["content"]), "Product info")

	// Alias: /browse should work the same
	browseResult, err := handle.DispatchCommandAsMap(context.Background(), DispatchRequest{
		Name:    "browse",
		User:    UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
	})
	require.NoError(t, err)
	browseMsg := browseResult
	require.Contains(t, fmt.Sprint(browseMsg["content"]), "catalog")
}

func TestUIShowcaseBotSelects(t *testing.T) {
	scriptPath := filepath.Join(repoRootJSDiscord(t), "examples", "discord-bots", "ui-showcase", "index.js")
	handle := loadTestBot(t, scriptPath)

	result, err := handle.DispatchCommandAsMap(context.Background(), DispatchRequest{
		Name:    "demo-selects",
		User:    UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
	})
	require.NoError(t, err)

	msg := result
	components := msg["components"].([]any)
	// Should have 5 rows: string, user, role, channel, mentionable selects
	require.Len(t, components, 5)
	require.Contains(t, fmt.Sprint(components[0]), "showcase:select:string")
	require.Contains(t, fmt.Sprint(components[1]), "showcase:select:user")
	require.Contains(t, fmt.Sprint(components[2]), "showcase:select:role")
	require.Contains(t, fmt.Sprint(components[3]), "showcase:select:channel")
	require.Contains(t, fmt.Sprint(components[4]), "showcase:select:mentionable")

	// Click a string select option
	selectResult, err := handle.DispatchComponentAsMap(context.Background(), DispatchRequest{
		Name:    "showcase:select:string",
		Values:  []string{"banana"},
		User:    UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
	})
	require.NoError(t, err)
	selectMsg := selectResult
	require.Contains(t, fmt.Sprint(selectMsg["content"]), "banana")
}

func TestUIShowcaseBotModalForm(t *testing.T) {
	scriptPath := filepath.Join(repoRootJSDiscord(t), "examples", "discord-bots", "ui-showcase", "index.js")
	handle := loadTestBot(t, scriptPath)

	// Submit the feedback form modal
	result, err := handle.DispatchModalAsMap(context.Background(), DispatchRequest{
		Name: "showcase:form:submit",
		Values: map[string]any{
			"title":    "Great DSL!",
			"feedback": "The builder pattern is very readable and easy to use.",
			"rating":   "5",
			"tags":     "ui, dsl, builders",
		},
		User:    UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
	})
	require.NoError(t, err)

	msg := result
	require.Contains(t, fmt.Sprint(msg["content"]), "Thanks")
	embeds := msg["embeds"].([]any)
	require.NotEmpty(t, embeds)
	embed := embeds[0].(map[string]any)
	require.Contains(t, fmt.Sprint(embed["title"]), "Great DSL!")
	require.Contains(t, fmt.Sprint(embed["description"]), "builder pattern")
}

func TestUIShowcaseBotAliasRegistration(t *testing.T) {
	scriptPath := filepath.Join(repoRootJSDiscord(t), "examples", "discord-bots", "ui-showcase", "index.js")
	handle := loadTestBot(t, scriptPath)

	// Both aliases should return the same content
	result1, err := handle.DispatchCommandAsMap(context.Background(), DispatchRequest{
		Name:    "demo-alias",
		User:    UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
	})
	require.NoError(t, err)
	msg1 := result1
	require.Contains(t, fmt.Sprint(msg1["content"]), "alias demo")

	result2, err := handle.DispatchCommandAsMap(context.Background(), DispatchRequest{
		Name:    "demo-alias-alt",
		User:    UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:   map[string]any{"id": "guild-1"},
		Channel: map[string]any{"id": "channel-1"},
	})
	require.NoError(t, err)
	msg2 := result2
	require.Contains(t, fmt.Sprint(msg2["content"]), "alias demo")
	// Same content from both aliases
	require.Equal(t, msg1["content"], msg2["content"])
}
