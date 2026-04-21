package jsdiscord

import (
	"context"
	"fmt"
	"testing"
)

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
