package jsdiscord

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/go-go-golems/go-go-goja/engine"
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

func TestDiscordEventContextSupportsMessageCreate(t *testing.T) {
	scriptPath := writeBotScript(t, `
		const { defineBot } = require("discord")
		module.exports = defineBot(({ event }) => {
			event("messageCreate", async (ctx) => {
				if ((ctx.message && ctx.message.content || "").trim() !== "!pingjs") {
					return null
				}
				await ctx.reply({
					content: "pong",
					embeds: [{ title: "From message" }],
					components: [{
						type: "actionRow",
						components: [{ type: "button", style: "primary", label: "OK", customId: "ok" }]
					}]
				})
				return "handled"
			})
		})
	`)

	handle := loadTestBot(t, scriptPath)

	var replies []any
	result, err := handle.DispatchEvent(context.Background(), DispatchRequest{
		Name:    "messageCreate",
		Message: map[string]any{"content": "!pingjs"},
		Reply: func(_ context.Context, value any) error {
			replies = append(replies, value)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("dispatch event: %v", err)
	}
	if got := fmt.Sprint(result); got != "[handled]" {
		t.Fatalf("event result = %s", got)
	}
	if len(replies) != 1 {
		t.Fatalf("replies = %#v", replies)
	}
	replyMap, ok := replies[0].(map[string]any)
	if !ok {
		t.Fatalf("reply payload = %T", replies[0])
	}
	if fmt.Sprint(replyMap["content"]) != "pong" {
		t.Fatalf("reply content = %#v", replyMap["content"])
	}
	if _, ok := replyMap["embeds"]; !ok {
		t.Fatalf("reply embeds missing: %#v", replyMap)
	}
	if _, ok := replyMap["components"]; !ok {
		t.Fatalf("reply components missing: %#v", replyMap)
	}
}

func TestApplicationCommandFromSnapshot(t *testing.T) {
	cmd, err := applicationCommandFromSnapshot(map[string]any{
		"name": "echo",
		"spec": map[string]any{
			"description": "Echo text",
			"options": map[string]any{
				"text":  map[string]any{"type": "string", "description": "Text", "required": true},
				"count": map[string]any{"type": "integer", "description": "Count"},
			},
		},
	})
	if err != nil {
		t.Fatalf("applicationCommandFromSnapshot: %v", err)
	}
	if cmd.Name != "echo" || cmd.Description != "Echo text" {
		t.Fatalf("unexpected command: %#v", cmd)
	}
	if len(cmd.Options) != 2 {
		t.Fatalf("options = %d", len(cmd.Options))
	}
	if cmd.Options[0].Type != discordgo.ApplicationCommandOptionInteger {
		t.Fatalf("expected sorted options by key, got first type %v", cmd.Options[0].Type)
	}
	if cmd.Options[1].Type != discordgo.ApplicationCommandOptionString {
		t.Fatalf("expected string option second, got %v", cmd.Options[1].Type)
	}
}

func TestNormalizeResponsePayloadSupportsEmbedsAndComponents(t *testing.T) {
	payload, err := normalizeResponsePayload(map[string]any{
		"content":   "pong",
		"ephemeral": true,
		"embeds": []any{
			map[string]any{"title": "Pong", "description": "hello", "color": 123},
		},
		"components": []any{
			map[string]any{
				"type": "actionRow",
				"components": []any{
					map[string]any{"type": "button", "style": "link", "label": "Docs", "url": "https://example.com"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("normalizeResponsePayload: %v", err)
	}
	if payload.Content != "pong" {
		t.Fatalf("content = %q", payload.Content)
	}
	if payload.Flags != discordgo.MessageFlagsEphemeral {
		t.Fatalf("flags = %v", payload.Flags)
	}
	if len(payload.Embeds) != 1 || payload.Embeds[0].Title != "Pong" {
		t.Fatalf("embeds = %#v", payload.Embeds)
	}
	if len(payload.Components) != 1 {
		t.Fatalf("components = %#v", payload.Components)
	}
}

func loadTestBot(t *testing.T, scriptPath string) *BotHandle {
	t.Helper()
	factory, err := engine.NewBuilder(
		engine.WithModuleRootsFromScript(scriptPath, engine.DefaultModuleRootsOptions()),
	).WithModules(engine.DefaultRegistryModules()).
		WithRuntimeModuleRegistrars(NewRegistrar(Config{})).
		Build()
	if err != nil {
		t.Fatalf("build factory: %v", err)
	}
	rt, err := factory.NewRuntime(context.Background())
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	t.Cleanup(func() { _ = rt.Close(context.Background()) })
	value, err := rt.Require.Require(scriptPath)
	if err != nil {
		t.Fatalf("require bot script: %v", err)
	}
	handle, err := CompileBot(rt.VM, value)
	if err != nil {
		t.Fatalf("compile bot: %v", err)
	}
	return handle
}

func writeBotScript(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "bot.js")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write bot script: %v", err)
	}
	return path
}
