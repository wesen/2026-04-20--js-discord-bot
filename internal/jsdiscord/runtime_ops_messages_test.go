package jsdiscord

import (
	"context"
	"fmt"
	"testing"
)

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
