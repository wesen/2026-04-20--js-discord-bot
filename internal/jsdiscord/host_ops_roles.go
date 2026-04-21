package jsdiscord

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func buildRoleOps(ops *DiscordOps, scriptPath string, session *discordgo.Session) {
	if ops == nil || session == nil {
		return
	}

	ops.RoleList = func(ctx context.Context, guildID string) ([]map[string]any, error) {
		_ = ctx
		guildID = strings.TrimSpace(guildID)
		if guildID == "" {
			return nil, fmt.Errorf("role list requires guild ID")
		}
		roles, err := session.GuildRoles(guildID)
		if err != nil {
			return nil, err
		}
		ret := make([]map[string]any, 0, len(roles))
		for _, role := range roles {
			ret = append(ret, roleMap(guildID, role))
		}
		logLifecycleDebug("listed discord guild roles from javascript", map[string]any{"script": scriptPath, "guildId": guildID, "count": len(ret), "action": "discord.roles.list"})
		return ret, nil
	}

	ops.RoleFetch = func(ctx context.Context, guildID, roleID string) (map[string]any, error) {
		_ = ctx
		guildID = strings.TrimSpace(guildID)
		roleID = strings.TrimSpace(roleID)
		if guildID == "" || roleID == "" {
			return nil, fmt.Errorf("role fetch requires guild ID and role ID")
		}
		roles, err := session.GuildRoles(guildID)
		if err != nil {
			return nil, err
		}
		for _, role := range roles {
			if role != nil && strings.TrimSpace(role.ID) == roleID {
				ret := roleMap(guildID, role)
				logLifecycleDebug("fetched discord guild role from javascript", map[string]any{"script": scriptPath, "guildId": guildID, "roleId": roleID, "action": "discord.roles.fetch"})
				return ret, nil
			}
		}
		return nil, fmt.Errorf("role %q not found in guild %q", roleID, guildID)
	}
}
