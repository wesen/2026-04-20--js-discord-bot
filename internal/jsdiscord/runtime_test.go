package jsdiscord

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
		Component: map[string]any{"customId": "support:queue", "type": "select"},
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
		Focused: map[string]any{"name": "query", "value": "arch"},
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

func TestApplicationCommandFromSnapshotSupportsAutocompleteAndConstraints(t *testing.T) {
	cmd, err := applicationCommandFromSnapshot(map[string]any{
		"name": "echo",
		"spec": map[string]any{
			"description": "Echo text",
			"options": map[string]any{
				"text": map[string]any{
					"type": "string", "description": "Text", "required": true,
					"autocomplete": true, "minLength": 2, "maxLength": 100,
				},
				"count": map[string]any{"type": "integer", "description": "Count", "minValue": 1, "maxValue": 10},
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
	if !cmd.Options[0].Autocomplete {
		t.Fatalf("expected autocomplete option: %#v", cmd.Options[0])
	}
	if cmd.Options[0].MinLength == nil || *cmd.Options[0].MinLength != 2 || cmd.Options[0].MaxLength != 100 {
		t.Fatalf("unexpected string constraints: %#v", cmd.Options[0])
	}
	if cmd.Options[1].MinValue == nil || *cmd.Options[1].MinValue != 1 || cmd.Options[1].MaxValue != 10 {
		t.Fatalf("unexpected number constraints: %#v", cmd.Options[1])
	}
}

func TestNormalizeResponsePayloadSupportsEmbedsAndSelectComponents(t *testing.T) {
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
					map[string]any{
						"type":     "select",
						"customId": "support:queue",
						"options": []any{
							map[string]any{"label": "Billing", "value": "billing"},
						},
					},
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

func TestNormalizeModalPayloadSupportsTextInputs(t *testing.T) {
	payload, err := normalizeModalPayload(map[string]any{
		"customId": "feedback:submit",
		"title":    "Feedback",
		"components": []any{
			map[string]any{
				"type": "actionRow",
				"components": []any{
					map[string]any{
						"type":     "textInput",
						"customId": "summary",
						"label":    "Summary",
						"style":    "short",
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("normalizeModalPayload: %v", err)
	}
	if payload.CustomID != "feedback:submit" || payload.Title != "Feedback" {
		t.Fatalf("payload = %#v", payload)
	}
	if len(payload.Components) != 1 {
		t.Fatalf("components = %#v", payload.Components)
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

func TestDiscordContextSupportsOutboundDiscordOps(t *testing.T) {
	scriptPath := writeBotScript(t, `
		const { defineBot } = require("discord")
		module.exports = defineBot(({ command }) => {
			command("announce", async (ctx) => {
				await ctx.discord.channels.send("chan-1", {
					content: "report",
					files: [{ name: "report.txt", content: "hello" }],
					replyTo: { messageId: "orig-1", channelId: "chan-1" }
				})
				await ctx.discord.messages.edit("chan-1", "msg-1", { content: "updated" })
				await ctx.discord.messages.react("chan-1", "msg-1", "✅")
				await ctx.discord.messages.delete("chan-1", "msg-1")
				return { content: "done" }
			})
		})
	`)

	handle := loadTestBot(t, scriptPath)
	var sends, edits, reacts, deletes int
	result, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name: "announce",
		Discord: &DiscordOps{
			ChannelSend: func(_ context.Context, channelID string, payload any) error {
				sends++
				if channelID != "chan-1" {
					t.Fatalf("channelID = %q", channelID)
				}
				msg, err := normalizeMessageSend(payload)
				if err != nil {
					return err
				}
				if len(msg.Files) != 1 || msg.Files[0].Name != "report.txt" {
					t.Fatalf("files = %#v", msg.Files)
				}
				if msg.Reference == nil || msg.Reference.MessageID != "orig-1" {
					t.Fatalf("reference = %#v", msg.Reference)
				}
				return nil
			},
			MessageEdit: func(_ context.Context, channelID, messageID string, payload any) error {
				edits++
				if channelID != "chan-1" || messageID != "msg-1" {
					t.Fatalf("edit target = %s/%s", channelID, messageID)
				}
				return nil
			},
			MessageReact: func(_ context.Context, channelID, messageID, emoji string) error {
				reacts++
				if emoji != "✅" {
					t.Fatalf("emoji = %q", emoji)
				}
				return nil
			},
			MessageDelete: func(_ context.Context, channelID, messageID string) error {
				deletes++
				return nil
			},
		},
	})
	if err != nil {
		t.Fatalf("dispatch command: %v", err)
	}
	if fmt.Sprint(result) != "map[content:done]" {
		t.Fatalf("result = %#v", result)
	}
	if sends != 1 || edits != 1 || reacts != 1 || deletes != 1 {
		t.Fatalf("counts = sends:%d edits:%d reacts:%d deletes:%d", sends, edits, reacts, deletes)
	}
}

func TestNormalizeMessageSendSupportsFilesAndReplyReference(t *testing.T) {
	message, err := normalizeMessageSend(map[string]any{
		"content": "report",
		"files": []any{
			map[string]any{"name": "report.txt", "content": "hello", "contentType": "text/plain"},
		},
		"replyTo": map[string]any{"messageId": "orig-1", "channelId": "chan-1"},
	})
	if err != nil {
		t.Fatalf("normalizeMessageSend: %v", err)
	}
	if len(message.Files) != 1 || message.Files[0].Name != "report.txt" {
		t.Fatalf("files = %#v", message.Files)
	}
	if message.Reference == nil || message.Reference.MessageID != "orig-1" || message.Reference.ChannelID != "chan-1" {
		t.Fatalf("reference = %#v", message.Reference)
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
