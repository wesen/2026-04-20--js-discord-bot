package jsdiscord

import (
	"context"
	"fmt"
	"testing"
)

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
					if shouldClear, _ := mapping["clear"].(bool); !shouldClear {
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
