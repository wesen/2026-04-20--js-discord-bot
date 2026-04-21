package jsdiscord

import (
	"context"
	"fmt"
	"testing"
)

func TestDiscordContextSupportsThreadUtilities(t *testing.T) {
	scriptPath := writeBotScript(t, `
		const { defineBot } = require("discord")
		module.exports = defineBot(({ command }) => {
			command("thread-tools", async (ctx) => {
				const thread = await ctx.discord.threads.fetch("thread-1")
				await ctx.discord.threads.join("thread-1")
				await ctx.discord.threads.leave("thread-1")
				const started = await ctx.discord.threads.start("chan-1", {
					name: "Support follow-up",
					type: "public",
					autoArchiveDuration: 60,
				})
				return { content: String(thread.id) + ":" + String(started.id) }
			})
		})
	`)

	handle := loadTestBot(t, scriptPath)
	var fetches, joins, leaves, starts int
	result, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name: "thread-tools",
		Discord: &DiscordOps{
			ThreadFetch: func(_ context.Context, threadID string) (map[string]any, error) {
				fetches++
				if threadID != "thread-1" {
					t.Fatalf("thread fetch = %s", threadID)
				}
				return map[string]any{"id": "thread-1"}, nil
			},
			ThreadJoin: func(_ context.Context, threadID string) error {
				joins++
				if threadID != "thread-1" {
					t.Fatalf("thread join = %s", threadID)
				}
				return nil
			},
			ThreadLeave: func(_ context.Context, threadID string) error {
				leaves++
				if threadID != "thread-1" {
					t.Fatalf("thread leave = %s", threadID)
				}
				return nil
			},
			ThreadStart: func(_ context.Context, channelID string, payload any) (map[string]any, error) {
				starts++
				if channelID != "chan-1" {
					t.Fatalf("thread start channel = %s", channelID)
				}
				mapping, _ := payload.(map[string]any)
				if fmt.Sprint(mapping["name"]) != "Support follow-up" {
					t.Fatalf("thread start payload = %#v", payload)
				}
				return map[string]any{"id": "thread-2"}, nil
			},
		},
	})
	if err != nil {
		t.Fatalf("dispatch command: %v", err)
	}
	if fmt.Sprint(result) != "map[content:thread-1:thread-2]" {
		t.Fatalf("result = %#v", result)
	}
	if fetches != 1 || joins != 1 || leaves != 1 || starts != 1 {
		t.Fatalf("counts = fetch:%d join:%d leave:%d start:%d", fetches, joins, leaves, starts)
	}
}
