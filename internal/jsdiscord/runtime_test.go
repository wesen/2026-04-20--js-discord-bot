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

func TestDiscordEventContextSupportsMessageUpdateAndDelete(t *testing.T) {
	scriptPath := writeBotScript(t, `
		const { defineBot } = require("discord")
		module.exports = defineBot(({ event }) => {
			event("messageUpdate", async (ctx) => {
				if ((ctx.message && ctx.message.content || "").trim() !== "edited text") {
					return null
				}
				return "update:" + String(ctx.before && ctx.before.content || "") + "->" + String(ctx.message && ctx.message.content || "")
			})
			event("messageDelete", async (ctx) => {
				return "delete:" + String(ctx.message && ctx.message.id || "") + ":before=" + String(ctx.before && ctx.before.content || "")
			})
		})
	`)

	handle := loadTestBot(t, scriptPath)

	updated, err := handle.DispatchEvent(context.Background(), DispatchRequest{
		Name:    "messageUpdate",
		Message: map[string]any{"id": "msg-1", "content": "edited text", "channelID": "chan-1"},
		Before:  map[string]any{"id": "msg-1", "content": "old text", "channelID": "chan-1"},
	})
	if err != nil {
		t.Fatalf("dispatch messageUpdate: %v", err)
	}
	if got := fmt.Sprint(updated); got != "[update:old text->edited text]" {
		t.Fatalf("messageUpdate result = %s", got)
	}

	deleted, err := handle.DispatchEvent(context.Background(), DispatchRequest{
		Name:    "messageDelete",
		Message: map[string]any{"id": "msg-2", "deleted": true, "channelID": "chan-1"},
		Before:  map[string]any{"id": "msg-2", "content": "removed text", "channelID": "chan-1"},
	})
	if err != nil {
		t.Fatalf("dispatch messageDelete: %v", err)
	}
	if got := fmt.Sprint(deleted); got != "[delete:msg-2:before=removed text]" {
		t.Fatalf("messageDelete result = %s", got)
	}
}

func TestDiscordEventContextSupportsReactionAddAndRemove(t *testing.T) {
	scriptPath := writeBotScript(t, `
		const { defineBot } = require("discord")
		module.exports = defineBot(({ event }) => {
			event("reactionAdd", async (ctx) => {
				return "add:" + String(ctx.reaction && ctx.reaction.emoji && ctx.reaction.emoji.name || "") + ":user=" + String(ctx.user && ctx.user.id || "")
			})
			event("reactionRemove", async (ctx) => {
				return "remove:" + String(ctx.reaction && ctx.reaction.emoji && ctx.reaction.emoji.name || "") + ":message=" + String(ctx.message && ctx.message.id || "")
			})
		})
	`)

	handle := loadTestBot(t, scriptPath)

	added, err := handle.DispatchEvent(context.Background(), DispatchRequest{
		Name:     "reactionAdd",
		Message:  map[string]any{"id": "msg-1", "channelID": "chan-1"},
		User:     map[string]any{"id": "user-1"},
		Member:   map[string]any{"id": "user-1", "roles": []string{"mod"}},
		Reaction: map[string]any{"messageId": "msg-1", "emoji": map[string]any{"name": "🔥"}},
	})
	if err != nil {
		t.Fatalf("dispatch reactionAdd: %v", err)
	}
	if got := fmt.Sprint(added); got != "[add:🔥:user=user-1]" {
		t.Fatalf("reactionAdd result = %s", got)
	}

	removed, err := handle.DispatchEvent(context.Background(), DispatchRequest{
		Name:     "reactionRemove",
		Message:  map[string]any{"id": "msg-2", "channelID": "chan-1"},
		User:     map[string]any{"id": "user-2"},
		Reaction: map[string]any{"messageId": "msg-2", "emoji": map[string]any{"name": "✅"}},
	})
	if err != nil {
		t.Fatalf("dispatch reactionRemove: %v", err)
	}
	if got := fmt.Sprint(removed); got != "[remove:✅:message=msg-2]" {
		t.Fatalf("reactionRemove result = %s", got)
	}
}

func TestDiscordEventContextSupportsGuildMemberEvents(t *testing.T) {
	scriptPath := writeBotScript(t, `
		const { defineBot } = require("discord")
		module.exports = defineBot(({ event }) => {
			event("guildMemberAdd", async (ctx) => {
				return "add:" + String(ctx.member && ctx.member.id || "") + ":roles=" + String((ctx.member && ctx.member.roles || []).length)
			})
			event("guildMemberUpdate", async (ctx) => {
				const before = Array.isArray(ctx.before && ctx.before.roles) ? ctx.before.roles.length : 0
				const after = Array.isArray(ctx.member && ctx.member.roles) ? ctx.member.roles.length : 0
				return "update:" + String(ctx.member && ctx.member.id || "") + ":" + String(before) + "->" + String(after)
			})
			event("guildMemberRemove", async (ctx) => {
				return "remove:" + String(ctx.user && ctx.user.id || "") + ":guild=" + String(ctx.guild && ctx.guild.id || "")
			})
		})
	`)

	handle := loadTestBot(t, scriptPath)

	added, err := handle.DispatchEvent(context.Background(), DispatchRequest{
		Name:   "guildMemberAdd",
		Guild:  map[string]any{"id": "guild-1"},
		User:   map[string]any{"id": "user-1"},
		Member: map[string]any{"id": "user-1", "roles": []string{"mod", "helper"}},
	})
	if err != nil {
		t.Fatalf("dispatch guildMemberAdd: %v", err)
	}
	if got := fmt.Sprint(added); got != "[add:user-1:roles=2]" {
		t.Fatalf("guildMemberAdd result = %s", got)
	}

	updated, err := handle.DispatchEvent(context.Background(), DispatchRequest{
		Name:   "guildMemberUpdate",
		Guild:  map[string]any{"id": "guild-1"},
		User:   map[string]any{"id": "user-1"},
		Member: map[string]any{"id": "user-1", "roles": []string{"mod", "helper", "trusted"}},
		Before: map[string]any{"id": "user-1", "roles": []string{"mod"}},
	})
	if err != nil {
		t.Fatalf("dispatch guildMemberUpdate: %v", err)
	}
	if got := fmt.Sprint(updated); got != "[update:user-1:1->3]" {
		t.Fatalf("guildMemberUpdate result = %s", got)
	}

	removed, err := handle.DispatchEvent(context.Background(), DispatchRequest{
		Name:   "guildMemberRemove",
		Guild:  map[string]any{"id": "guild-1"},
		User:   map[string]any{"id": "user-2"},
		Member: map[string]any{"id": "user-2", "roles": []string{"member"}},
	})
	if err != nil {
		t.Fatalf("dispatch guildMemberRemove: %v", err)
	}
	if got := fmt.Sprint(removed); got != "[remove:user-2:guild=guild-1]" {
		t.Fatalf("guildMemberRemove result = %s", got)
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

func TestDiscordContextSupportsMessageModerationOps(t *testing.T) {
	scriptPath := writeBotScript(t, `
		const { defineBot } = require("discord")
		module.exports = defineBot(({ command }) => {
			command("message-tools", async (ctx) => {
				const message = await ctx.discord.messages.fetch("chan-1", "msg-1")
				await ctx.discord.messages.pin("chan-1", "msg-1")
				await ctx.discord.messages.unpin("chan-1", "msg-1")
				const pinned = await ctx.discord.messages.listPinned("chan-1")
				return { content: String(message.id) + ":" + String(pinned.length) }
			})
		})
	`)

	handle := loadTestBot(t, scriptPath)
	var fetches, pins, unpins, listPinned int
	result, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name: "message-tools",
		Discord: &DiscordOps{
			MessageFetch: func(_ context.Context, channelID, messageID string) (map[string]any, error) {
				fetches++
				if channelID != "chan-1" || messageID != "msg-1" {
					t.Fatalf("fetch target = %s/%s", channelID, messageID)
				}
				return map[string]any{"id": "msg-1", "content": "hello"}, nil
			},
			MessagePin: func(_ context.Context, channelID, messageID string) error {
				pins++
				if channelID != "chan-1" || messageID != "msg-1" {
					t.Fatalf("pin target = %s/%s", channelID, messageID)
				}
				return nil
			},
			MessageUnpin: func(_ context.Context, channelID, messageID string) error {
				unpins++
				if channelID != "chan-1" || messageID != "msg-1" {
					t.Fatalf("unpin target = %s/%s", channelID, messageID)
				}
				return nil
			},
			MessageListPinned: func(_ context.Context, channelID string) ([]map[string]any, error) {
				listPinned++
				if channelID != "chan-1" {
					t.Fatalf("listPinned channel = %s", channelID)
				}
				return []map[string]any{{"id": "msg-1"}, {"id": "msg-2"}}, nil
			},
		},
	})
	if err != nil {
		t.Fatalf("dispatch command: %v", err)
	}
	if fmt.Sprint(result) != "map[content:msg-1:2]" {
		t.Fatalf("result = %#v", result)
	}
	if fetches != 1 || pins != 1 || unpins != 1 || listPinned != 1 {
		t.Fatalf("counts = fetch:%d pin:%d unpin:%d list:%d", fetches, pins, unpins, listPinned)
	}
}

func TestDiscordContextSupportsMessageBulkDelete(t *testing.T) {
	scriptPath := writeBotScript(t, `
		const { defineBot } = require("discord")
		module.exports = defineBot(({ command }) => {
			command("purge", async (ctx) => {
				await ctx.discord.messages.bulkDelete("chan-1", ["msg-1", "msg-2", "msg-2"])
				await ctx.discord.messages.bulkDelete("chan-1", { messageIds: ["msg-3"] })
				return { content: "purged" }
			})
		})
	`)

	handle := loadTestBot(t, scriptPath)
	var payloads [][]string
	result, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name: "purge",
		Discord: &DiscordOps{
			MessageBulkDelete: func(_ context.Context, channelID string, payload any) error {
				if channelID != "chan-1" {
					t.Fatalf("bulkDelete channel = %s", channelID)
				}
				ids, err := normalizeMessageIDList(payload)
				if err != nil {
					return err
				}
				payloads = append(payloads, ids)
				return nil
			},
		},
	})
	if err != nil {
		t.Fatalf("dispatch command: %v", err)
	}
	if fmt.Sprint(result) != "map[content:purged]" {
		t.Fatalf("result = %#v", result)
	}
	if len(payloads) != 2 {
		t.Fatalf("payloads = %#v", payloads)
	}
	if fmt.Sprint(payloads[0]) != "[msg-1 msg-2]" || fmt.Sprint(payloads[1]) != "[msg-3]" {
		t.Fatalf("payloads = %#v", payloads)
	}
}

func TestDiscordContextSupportsMessageListing(t *testing.T) {
	scriptPath := writeBotScript(t, `
		const { defineBot } = require("discord")
		module.exports = defineBot(({ command }) => {
			command("message-list", async (ctx) => {
				const messages = await ctx.discord.messages.list("chan-1", { around: "msg-2", limit: 2 })
				return { content: String(messages.length) + ":" + String(messages[0].id) }
			})
		})
	`)

	handle := loadTestBot(t, scriptPath)
	var lists int
	result, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name: "message-list",
		Discord: &DiscordOps{
			MessageList: func(_ context.Context, channelID string, payload any) ([]map[string]any, error) {
				lists++
				if channelID != "chan-1" {
					t.Fatalf("message list channel = %s", channelID)
				}
				mapping, _ := payload.(map[string]any)
				if fmt.Sprint(mapping["around"]) != "msg-2" {
					t.Fatalf("message list around = %#v", payload)
				}
				if mapping["limit"] != int64(2) && mapping["limit"] != 2 && mapping["limit"] != float64(2) {
					t.Fatalf("message list limit = %#v", payload)
				}
				return []map[string]any{{"id": "msg-2"}, {"id": "msg-3"}}, nil
			},
		},
	})
	if err != nil {
		t.Fatalf("dispatch command: %v", err)
	}
	if fmt.Sprint(result) != "map[content:2:msg-2]" {
		t.Fatalf("result = %#v", result)
	}
	if lists != 1 {
		t.Fatalf("lists = %d", lists)
	}
}

func TestDiscordContextSupportsMemberLookup(t *testing.T) {
	scriptPath := writeBotScript(t, `
		const { defineBot } = require("discord")
		module.exports = defineBot(({ command }) => {
			command("member-lookup", async (ctx) => {
				const member = await ctx.discord.members.fetch("guild-1", "user-1")
				const members = await ctx.discord.members.list("guild-1", { after: "user-0", limit: 2 })
				return { content: String(member.id) + ":" + String(members.length) }
			})
		})
	`)

	handle := loadTestBot(t, scriptPath)
	var fetches, lists int
	result, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name: "member-lookup",
		Discord: &DiscordOps{
			MemberFetch: func(_ context.Context, guildID, userID string) (map[string]any, error) {
				fetches++
				if guildID != "guild-1" || userID != "user-1" {
					t.Fatalf("member fetch = %s/%s", guildID, userID)
				}
				return map[string]any{"id": "user-1"}, nil
			},
			MemberList: func(_ context.Context, guildID string, payload any) ([]map[string]any, error) {
				lists++
				if guildID != "guild-1" {
					t.Fatalf("member list guild = %s", guildID)
				}
				mapping, _ := payload.(map[string]any)
				if fmt.Sprint(mapping["after"]) != "user-0" {
					t.Fatalf("member list after = %#v", payload)
				}
				if mapping["limit"] != int64(2) && mapping["limit"] != 2 && mapping["limit"] != float64(2) {
					t.Fatalf("member list limit = %#v", payload)
				}
				return []map[string]any{{"id": "user-1"}, {"id": "user-2"}}, nil
			},
		},
	})
	if err != nil {
		t.Fatalf("dispatch command: %v", err)
	}
	if fmt.Sprint(result) != "map[content:user-1:2]" {
		t.Fatalf("result = %#v", result)
	}
	if fetches != 1 || lists != 1 {
		t.Fatalf("counts = fetch:%d list:%d", fetches, lists)
	}
}

func TestDiscordContextSupportsGuildAndRoleLookup(t *testing.T) {
	scriptPath := writeBotScript(t, `
		const { defineBot } = require("discord")
		module.exports = defineBot(({ command }) => {
			command("lookup", async (ctx) => {
				const guild = await ctx.discord.guilds.fetch("guild-1")
				const roles = await ctx.discord.roles.list("guild-1")
				const role = await ctx.discord.roles.fetch("guild-1", "role-2")
				return { content: String(guild.id) + ":" + String(roles.length) + ":" + String(role.name) }
			})
		})
	`)

	handle := loadTestBot(t, scriptPath)
	var guildFetches, roleLists, roleFetches int
	result, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name: "lookup",
		Discord: &DiscordOps{
			GuildFetch: func(_ context.Context, guildID string) (map[string]any, error) {
				guildFetches++
				if guildID != "guild-1" {
					t.Fatalf("guild fetch = %s", guildID)
				}
				return map[string]any{"id": "guild-1"}, nil
			},
			RoleList: func(_ context.Context, guildID string) ([]map[string]any, error) {
				roleLists++
				if guildID != "guild-1" {
					t.Fatalf("role list = %s", guildID)
				}
				return []map[string]any{{"id": "role-1"}, {"id": "role-2"}}, nil
			},
			RoleFetch: func(_ context.Context, guildID, roleID string) (map[string]any, error) {
				roleFetches++
				if guildID != "guild-1" || roleID != "role-2" {
					t.Fatalf("role fetch = %s/%s", guildID, roleID)
				}
				return map[string]any{"id": "role-2", "name": "Moderator"}, nil
			},
		},
	})
	if err != nil {
		t.Fatalf("dispatch command: %v", err)
	}
	if fmt.Sprint(result) != "map[content:guild-1:2:Moderator]" {
		t.Fatalf("result = %#v", result)
	}
	if guildFetches != 1 || roleLists != 1 || roleFetches != 1 {
		t.Fatalf("counts = guild:%d roles:%d role:%d", guildFetches, roleLists, roleFetches)
	}
}

func TestDiscordContextSupportsChannelUtilities(t *testing.T) {
	scriptPath := writeBotScript(t, `
		const { defineBot } = require("discord")
		module.exports = defineBot(({ command }) => {
			command("channel-tools", async (ctx) => {
				const channel = await ctx.discord.channels.fetch("chan-1")
				await ctx.discord.channels.setTopic("chan-1", "Escalation queue")
				await ctx.discord.channels.setSlowmode("chan-1", 30)
				return { content: String(channel.id) + ":" + String(channel.rateLimitPerUser) }
			})
		})
	`)

	handle := loadTestBot(t, scriptPath)
	var fetches, topics, slowmodes int
	result, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name: "channel-tools",
		Discord: &DiscordOps{
			ChannelFetch: func(_ context.Context, channelID string) (map[string]any, error) {
				fetches++
				if channelID != "chan-1" {
					t.Fatalf("channel fetch = %s", channelID)
				}
				return map[string]any{"id": "chan-1", "rateLimitPerUser": 0}, nil
			},
			ChannelSetTopic: func(_ context.Context, channelID, topic string) error {
				topics++
				if channelID != "chan-1" || topic != "Escalation queue" {
					t.Fatalf("setTopic = %s %q", channelID, topic)
				}
				return nil
			},
			ChannelSetSlowmode: func(_ context.Context, channelID string, seconds int) error {
				slowmodes++
				if channelID != "chan-1" || seconds != 30 {
					t.Fatalf("setSlowmode = %s %d", channelID, seconds)
				}
				return nil
			},
		},
	})
	if err != nil {
		t.Fatalf("dispatch command: %v", err)
	}
	if fmt.Sprint(result) != "map[content:chan-1:0]" {
		t.Fatalf("result = %#v", result)
	}
	if fetches != 1 || topics != 1 || slowmodes != 1 {
		t.Fatalf("counts = fetch:%d topic:%d slowmode:%d", fetches, topics, slowmodes)
	}
}

func TestDiscordContextSupportsMemberAdminOps(t *testing.T) {
	scriptPath := writeBotScript(t, `
		const { defineBot } = require("discord")
		module.exports = defineBot(({ command }) => {
			command("moderate", async (ctx) => {
				await ctx.discord.members.addRole("guild-1", "user-1", "role-1")
				await ctx.discord.members.removeRole("guild-1", "user-1", "role-2")
				await ctx.discord.members.timeout("guild-1", "user-1", { durationSeconds: 600 })
				await ctx.discord.members.timeout("guild-1", "user-1", { clear: true })
				await ctx.discord.members.kick("guild-1", "user-2", { reason: "spam" })
				await ctx.discord.members.ban("guild-1", "user-3", { reason: "raid", deleteMessageDays: 2 })
				await ctx.discord.members.unban("guild-1", "user-3")
				return { content: "ok" }
			})
		})
	`)

	handle := loadTestBot(t, scriptPath)
	var adds, removes, timeouts, kicks, bans, unbans int
	result, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name: "moderate",
		Discord: &DiscordOps{
			MemberAddRole: func(_ context.Context, guildID, userID, roleID string) error {
				adds++
				if guildID != "guild-1" || userID != "user-1" || roleID != "role-1" {
					t.Fatalf("addRole target = %s/%s/%s", guildID, userID, roleID)
				}
				return nil
			},
			MemberRemoveRole: func(_ context.Context, guildID, userID, roleID string) error {
				removes++
				if roleID != "role-2" {
					t.Fatalf("removeRole role = %s", roleID)
				}
				return nil
			},
			MemberSetTimeout: func(_ context.Context, guildID, userID string, payload any) error {
				timeouts++
				mapping, _ := payload.(map[string]any)
				if timeouts == 1 {
					if mapping["durationSeconds"] != int64(600) && mapping["durationSeconds"] != 600 && mapping["durationSeconds"] != float64(600) {
						t.Fatalf("timeout payload = %#v", payload)
					}
				}
				if timeouts == 2 {
					if clear, _ := mapping["clear"].(bool); !clear {
						t.Fatalf("clear payload = %#v", payload)
					}
				}
				return nil
			},
			MemberKick: func(_ context.Context, guildID, userID string, payload any) error {
				kicks++
				mapping, _ := payload.(map[string]any)
				if guildID != "guild-1" || userID != "user-2" || fmt.Sprint(mapping["reason"]) != "spam" {
					t.Fatalf("kick = %s/%s %#v", guildID, userID, payload)
				}
				return nil
			},
			MemberBan: func(_ context.Context, guildID, userID string, payload any) error {
				bans++
				mapping, _ := payload.(map[string]any)
				if guildID != "guild-1" || userID != "user-3" {
					t.Fatalf("ban target = %s/%s %#v", guildID, userID, payload)
				}
				if fmt.Sprint(mapping["reason"]) != "raid" {
					t.Fatalf("ban reason = %#v", payload)
				}
				if mapping["deleteMessageDays"] != int64(2) && mapping["deleteMessageDays"] != 2 && mapping["deleteMessageDays"] != float64(2) {
					t.Fatalf("ban days = %#v", payload)
				}
				return nil
			},
			MemberUnban: func(_ context.Context, guildID, userID string) error {
				unbans++
				if guildID != "guild-1" || userID != "user-3" {
					t.Fatalf("unban target = %s/%s", guildID, userID)
				}
				return nil
			},
		},
	})
	if err != nil {
		t.Fatalf("dispatch command: %v", err)
	}
	if fmt.Sprint(result) != "map[content:ok]" {
		t.Fatalf("result = %#v", result)
	}
	if adds != 1 || removes != 1 || timeouts != 2 || kicks != 1 || bans != 1 || unbans != 1 {
		t.Fatalf("counts = add:%d remove:%d timeout:%d kick:%d ban:%d unban:%d", adds, removes, timeouts, kicks, bans, unbans)
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
