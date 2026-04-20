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

func TestDiscordRegistrarCompilesBotAndSettlesAsyncHandlers(t *testing.T) {
	scriptPath := writeBotScript(t, `
		const { defineBot } = require("discord")
		module.exports = defineBot(({ command, event, configure }) => {
			configure({ name: "js-bot" })
			command("ping", {
				description: "Ping from JS",
				options: {
					text: { type: "string", description: "Text", required: true }
				}
			}, async (ctx) => {
				const current = ctx.store.get("hits", 0)
				ctx.store.set("hits", current + 1)
				await ctx.reply({ content: ctx.args.text + ":" + current })
				return { content: "done", ephemeral: true }
			})
			event("ready", async (ctx) => {
				ctx.store.set("ready", true)
				return "ready"
			})
		})
	`)

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
	defer func() { _ = rt.Close(context.Background()) }()

	value, err := rt.Require.Require(scriptPath)
	if err != nil {
		t.Fatalf("require bot script: %v", err)
	}
	handle, err := CompileBot(rt.VM, value)
	if err != nil {
		t.Fatalf("compile bot: %v", err)
	}
	desc, err := handle.Describe(context.Background())
	if err != nil {
		t.Fatalf("describe bot: %v", err)
	}
	if got := desc["kind"]; got != "discord.bot" {
		t.Fatalf("kind = %#v", got)
	}

	var (
		replies []string
		mu      sync.Mutex
	)
	result, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name: "ping",
		Args: map[string]any{"text": "hello"},
		Reply: func(_ context.Context, value any) error {
			mu.Lock()
			defer mu.Unlock()
			replies = append(replies, fmt.Sprint(value))
			return nil
		},
	})
	if err != nil {
		t.Fatalf("dispatch command: %v", err)
	}
	if got := fmt.Sprint(result); got != "map[content:done ephemeral:true]" {
		t.Fatalf("command result = %s", got)
	}
	if len(replies) != 1 || replies[0] != "map[content:hello:0]" {
		t.Fatalf("replies = %#v", replies)
	}

	eventResult, err := handle.DispatchEvent(context.Background(), DispatchRequest{Name: "ready"})
	if err != nil {
		t.Fatalf("dispatch event: %v", err)
	}
	if got := fmt.Sprint(eventResult); got != "[ready]" {
		t.Fatalf("event result = %s", got)
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
	if cmd.Options[1].Type != discordgo.ApplicationCommandOptionString {
		t.Fatalf("expected stable sorted options; first string option missing")
	}
}

func TestNormalizeResponsePayload(t *testing.T) {
	payload, err := normalizeResponsePayload(map[string]any{"content": "pong", "ephemeral": true})
	if err != nil {
		t.Fatalf("normalizeResponsePayload: %v", err)
	}
	if payload.Content != "pong" {
		t.Fatalf("content = %q", payload.Content)
	}
	if payload.Flags != discordgo.MessageFlagsEphemeral {
		t.Fatalf("flags = %v", payload.Flags)
	}
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
