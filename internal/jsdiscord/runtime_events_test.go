package jsdiscord

import (
	"context"
	"fmt"
	"testing"
)

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
		Message: &MessageSnapshot{Content: "!pingjs"},
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
		Message: &MessageSnapshot{ID: "msg-1", Content: "edited text", ChannelID: "chan-1"},
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
		Message: &MessageSnapshot{ID: "msg-2", ChannelID: "chan-1", Deleted: true},
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
		Message:  &MessageSnapshot{ID: "msg-1", ChannelID: "chan-1"},
		User:     UserSnapshot{ID: "user-1"},
		Member:   &MemberSnapshot{ID: "user-1", Roles: []string{"mod"}},
		Reaction: ReactionSnapshot{MessageID: "msg-1", Emoji: EmojiSnapshot{Name: "🔥"}},
	})
	if err != nil {
		t.Fatalf("dispatch reactionAdd: %v", err)
	}
	if got := fmt.Sprint(added); got != "[add:🔥:user=user-1]" {
		t.Fatalf("reactionAdd result = %s", got)
	}

	removed, err := handle.DispatchEvent(context.Background(), DispatchRequest{
		Name:     "reactionRemove",
		Message:  &MessageSnapshot{ID: "msg-2", ChannelID: "chan-1"},
		User:     UserSnapshot{ID: "user-2"},
		Reaction: ReactionSnapshot{MessageID: "msg-2", Emoji: EmojiSnapshot{Name: "✅"}},
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
		User:   UserSnapshot{ID: "user-1"},
		Member: &MemberSnapshot{ID: "user-1", Roles: []string{"mod", "helper"}},
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
		User:   UserSnapshot{ID: "user-1"},
		Member: &MemberSnapshot{ID: "user-1", Roles: []string{"mod", "helper", "trusted"}},
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
		User:   UserSnapshot{ID: "user-2"},
		Member: &MemberSnapshot{ID: "user-2", Roles: []string{"member"}},
	})
	if err != nil {
		t.Fatalf("dispatch guildMemberRemove: %v", err)
	}
	if got := fmt.Sprint(removed); got != "[remove:user-2:guild=guild-1]" {
		t.Fatalf("guildMemberRemove result = %s", got)
	}
}
