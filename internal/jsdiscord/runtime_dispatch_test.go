package jsdiscord

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestDiscordRegistrarSupportsRichCommandHelpers(t *testing.T) {
	scriptPath := writeBotScript(t, `
		const { defineBot } = require("discord")
		module.exports = defineBot(({ command, configure }) => {
			configure({ name: "js-bot" })
			command("ping", {
				description: "Ping from JS",
				options: {
					text: { type: "string", description: "Text", required: true }
				}
			}, async (ctx) => {
				const current = ctx.store.get("hits", 0)
				ctx.store.set("hits", current + 1)
				await ctx.defer({ ephemeral: true })
				await ctx.edit({
					content: ctx.args.text + ":" + current,
					embeds: [{ title: "Edited" }],
				})
				await ctx.followUp({ content: "after", ephemeral: true })
				return { content: "done", ephemeral: true }
			})
		})
	`)

	handle := loadTestBot(t, scriptPath)

	var (
		deferred  []any
		edits     []any
		followUps []any
		mu        sync.Mutex
	)
	result, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name: "ping",
		Args: map[string]any{"text": "hello"},
		Defer: func(_ context.Context, value any) error {
			mu.Lock()
			deferred = append(deferred, value)
			mu.Unlock()
			return nil
		},
		Edit: func(_ context.Context, value any) error {
			mu.Lock()
			edits = append(edits, value)
			mu.Unlock()
			return nil
		},
		FollowUp: func(_ context.Context, value any) error {
			mu.Lock()
			followUps = append(followUps, value)
			mu.Unlock()
			return nil
		},
	})
	if err != nil {
		t.Fatalf("dispatch command: %v", err)
	}
	if got := fmt.Sprint(result); got != "map[content:done ephemeral:true]" {
		t.Fatalf("command result = %s", got)
	}
	if len(deferred) != 1 || fmt.Sprint(deferred[0]) != "map[ephemeral:true]" {
		t.Fatalf("deferred = %#v", deferred)
	}
	if len(edits) != 1 {
		t.Fatalf("edits = %#v", edits)
	}
	if len(followUps) != 1 || fmt.Sprint(followUps[0]) != "map[content:after ephemeral:true]" {
		t.Fatalf("followUps = %#v", followUps)
	}
}

func TestDiscordCommandPromiseRejectionsIncludeJavaScriptErrorDetails(t *testing.T) {
	scriptPath := writeBotScript(t, `
		const { defineBot } = require("discord")
		module.exports = defineBot(({ command }) => {
			command("broken", async () => {
				throw new Error("boom from js")
			})
		})
	`)

	handle := loadTestBot(t, scriptPath)
	_, err := handle.DispatchCommand(context.Background(), DispatchRequest{Name: "broken"})
	if err == nil {
		t.Fatalf("expected dispatch error")
	}
	if got := err.Error(); !strings.Contains(got, "boom from js") {
		t.Fatalf("error = %q", got)
	}
}

func TestDiscordCommandContextSupportsTimerSleep(t *testing.T) {
	scriptPath := writeBotScript(t, `
		const { defineBot } = require("discord")
		const { sleep } = require("timer")
		module.exports = defineBot(({ command }) => {
			command("search", async (ctx) => {
				await ctx.defer({ ephemeral: true })
				await sleep(5)
				await ctx.edit({ content: "done" })
			})
		})
	`)

	handle := loadTestBot(t, scriptPath)
	var (
		deferred []any
		edits    []any
	)
	_, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name: "search",
		Defer: func(_ context.Context, value any) error {
			deferred = append(deferred, value)
			return nil
		},
		Edit: func(_ context.Context, value any) error {
			edits = append(edits, value)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("dispatch command: %v", err)
	}
	if len(deferred) != 1 || fmt.Sprint(deferred[0]) != "map[ephemeral:true]" {
		t.Fatalf("deferred = %#v", deferred)
	}
	if len(edits) != 1 || fmt.Sprint(edits[0]) != "map[content:done]" {
		t.Fatalf("edits = %#v", edits)
	}
}

func TestDiscordCommandContextSupportsShowModal(t *testing.T) {
	scriptPath := writeBotScript(t, `
		const { defineBot } = require("discord")
		module.exports = defineBot(({ command }) => {
			command("feedback", async (ctx) => {
				await ctx.showModal({
					customId: "feedback:submit",
					title: "Feedback",
					components: [{
						type: "actionRow",
						components: [{
							type: "textInput",
							customId: "summary",
							label: "Summary",
							style: "short",
							required: true,
						}]
					}]
				})
			})
		})
	`)

	handle := loadTestBot(t, scriptPath)
	var shown []any
	_, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name: "feedback",
		ShowModal: func(_ context.Context, value any) error {
			shown = append(shown, value)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("dispatch command: %v", err)
	}
	if len(shown) != 1 {
		t.Fatalf("shown = %#v", shown)
	}
	payload, ok := shown[0].(map[string]any)
	if !ok {
		t.Fatalf("modal payload type = %T", shown[0])
	}
	if fmt.Sprint(payload["customId"]) != "feedback:submit" {
		t.Fatalf("customId = %#v", payload["customId"])
	}
}

func TestDiscordRuntimeSupportsComponentsModalsAndAutocomplete(t *testing.T) {
	scriptPath := writeBotScript(t, `
		const { defineBot } = require("discord")
		module.exports = defineBot(({ component, modal, autocomplete }) => {
			component("support:queue", async (ctx) => {
				return { content: "selected:" + ((ctx.values && ctx.values[0]) || "") }
			})
			modal("feedback:submit", async (ctx) => {
				return { content: "feedback:" + (ctx.values.summary || "") }
			})
			autocomplete("kb-search", "query", async (ctx) => {
				const current = String((ctx.focused && ctx.focused.value) || "")
				return [
					{ name: "Architecture", value: "architecture" },
					{ name: current, value: current }
				]
			})
		})
	`)

	handle := loadTestBot(t, scriptPath)

	componentResult, err := handle.DispatchComponent(context.Background(), DispatchRequest{
		Name:      "support:queue",
		Values:    []string{"billing"},
		Component: ComponentSnapshot{CustomID: "support:queue", Type: "select"},
	})
	if err != nil {
		t.Fatalf("dispatch component: %v", err)
	}
	if got := fmt.Sprint(componentResult); got != "map[content:selected:billing]" {
		t.Fatalf("component result = %s", got)
	}

	modalResult, err := handle.DispatchModal(context.Background(), DispatchRequest{
		Name:   "feedback:submit",
		Values: map[string]any{"summary": "looks good"},
		Modal:  map[string]any{"customId": "feedback:submit"},
	})
	if err != nil {
		t.Fatalf("dispatch modal: %v", err)
	}
	if got := fmt.Sprint(modalResult); got != "map[content:feedback:looks good]" {
		t.Fatalf("modal result = %s", got)
	}

	autocompleteResult, err := handle.DispatchAutocomplete(context.Background(), DispatchRequest{
		Name:    "kb-search",
		Args:    map[string]any{"query": "arch"},
		Focused: FocusedOptionSnapshot{Name: "query", Value: "arch"},
		Command: map[string]any{"name": "kb-search"},
	})
	if err != nil {
		t.Fatalf("dispatch autocomplete: %v", err)
	}
	items, ok := autocompleteResult.([]any)
	if !ok || len(items) != 2 {
		t.Fatalf("autocomplete result = %#v", autocompleteResult)
	}
}

func TestDiscordContextSupportsRuntimeConfig(t *testing.T) {
	scriptPath := writeBotScript(t, `
		const { defineBot } = require("discord")
		module.exports = defineBot(({ command }) => {
			command("show-config", async (ctx) => {
				return { content: String(ctx.config.index_path) + ":" + String(ctx.config.read_only) }
			})
		})
	`)

	handle := loadTestBot(t, scriptPath)
	result, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name:   "show-config",
		Config: map[string]any{"index_path": "./docs", "read_only": true},
	})
	if err != nil {
		t.Fatalf("dispatch command: %v", err)
	}
	if fmt.Sprint(result) != "map[content:./docs:true]" {
		t.Fatalf("result = %#v", result)
	}
}
