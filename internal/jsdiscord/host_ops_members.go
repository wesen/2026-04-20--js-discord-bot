package jsdiscord

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func buildMemberOps(ops *DiscordOps, scriptPath string, session *discordgo.Session) {
	if ops == nil || session == nil {
		return
	}

	ops.MemberFetch = func(ctx context.Context, guildID, userID string) (map[string]any, error) {
		_ = ctx
		guildID = strings.TrimSpace(guildID)
		userID = strings.TrimSpace(userID)
		if guildID == "" || userID == "" {
			return nil, fmt.Errorf("member fetch requires guild ID and user ID")
		}
		member, err := session.GuildMember(guildID, userID)
		if err != nil {
			return nil, err
		}
		ret := memberMap(member)
		logLifecycleDebug("fetched discord guild member from javascript", map[string]any{"script": scriptPath, "guildId": guildID, "userId": userID, "action": "discord.members.fetch"})
		return ret, nil
	}

	ops.MemberList = func(ctx context.Context, guildID string, payload any) ([]map[string]any, error) {
		_ = ctx
		guildID = strings.TrimSpace(guildID)
		if guildID == "" {
			return nil, fmt.Errorf("member list requires guild ID")
		}
		options, err := normalizeMemberListOptions(payload)
		if err != nil {
			return nil, err
		}
		members, err := session.GuildMembers(guildID, options.After, options.Limit)
		if err != nil {
			return nil, err
		}
		ret := make([]map[string]any, 0, len(members))
		for _, member := range members {
			ret = append(ret, memberMap(member))
		}
		logLifecycleDebug("listed discord guild members from javascript", map[string]any{"script": scriptPath, "guildId": guildID, "after": options.After, "limit": options.Limit, "count": len(ret), "action": "discord.members.list"})
		return ret, nil
	}

	ops.MemberAddRole = func(ctx context.Context, guildID, userID, roleID string) error {
		_ = ctx
		guildID = strings.TrimSpace(guildID)
		userID = strings.TrimSpace(userID)
		roleID = strings.TrimSpace(roleID)
		if guildID == "" || userID == "" || roleID == "" {
			return fmt.Errorf("member addRole requires guild ID, user ID, and role ID")
		}
		err := session.GuildMemberRoleAdd(guildID, userID, roleID)
		if err == nil {
			logLifecycleDebug("added discord guild member role from javascript", map[string]any{"script": scriptPath, "guildId": guildID, "userId": userID, "roleId": roleID, "action": "discord.members.addRole"})
		}
		return err
	}

	ops.MemberRemoveRole = func(ctx context.Context, guildID, userID, roleID string) error {
		_ = ctx
		guildID = strings.TrimSpace(guildID)
		userID = strings.TrimSpace(userID)
		roleID = strings.TrimSpace(roleID)
		if guildID == "" || userID == "" || roleID == "" {
			return fmt.Errorf("member removeRole requires guild ID, user ID, and role ID")
		}
		err := session.GuildMemberRoleRemove(guildID, userID, roleID)
		if err == nil {
			logLifecycleDebug("removed discord guild member role from javascript", map[string]any{"script": scriptPath, "guildId": guildID, "userId": userID, "roleId": roleID, "action": "discord.members.removeRole"})
		}
		return err
	}

	ops.MemberSetTimeout = func(ctx context.Context, guildID, userID string, payload any) error {
		_ = ctx
		guildID = strings.TrimSpace(guildID)
		userID = strings.TrimSpace(userID)
		if guildID == "" || userID == "" {
			return fmt.Errorf("member timeout requires guild ID and user ID")
		}
		until, err := normalizeTimeoutUntil(payload)
		if err != nil {
			return err
		}
		err = session.GuildMemberTimeout(guildID, userID, until)
		if err == nil {
			fields := mergeLogFields(map[string]any{"script": scriptPath, "guildId": guildID, "userId": userID, "action": "discord.members.timeout"}, payloadLogFields(payload))
			if until == nil {
				fields["cleared"] = true
			} else {
				fields["until"] = until.Format(time.RFC3339)
			}
			logLifecycleDebug("updated discord guild member timeout from javascript", fields)
		}
		return err
	}

	ops.MemberKick = func(ctx context.Context, guildID, userID string, payload any) error {
		_ = ctx
		guildID = strings.TrimSpace(guildID)
		userID = strings.TrimSpace(userID)
		if guildID == "" || userID == "" {
			return fmt.Errorf("member kick requires guild ID and user ID")
		}
		reason := normalizeModerationReason(payload)
		err := session.GuildMemberDeleteWithReason(guildID, userID, reason)
		if err == nil {
			fields := mergeLogFields(map[string]any{"script": scriptPath, "guildId": guildID, "userId": userID, "action": "discord.members.kick"}, payloadLogFields(payload))
			if reason != "" {
				fields["reason"] = reason
			}
			logLifecycleDebug("kicked discord guild member from javascript", fields)
		}
		return err
	}

	ops.MemberBan = func(ctx context.Context, guildID, userID string, payload any) error {
		_ = ctx
		guildID = strings.TrimSpace(guildID)
		userID = strings.TrimSpace(userID)
		if guildID == "" || userID == "" {
			return fmt.Errorf("member ban requires guild ID and user ID")
		}
		reason, deleteMessageDays, err := normalizeBanOptions(payload)
		if err != nil {
			return err
		}
		err = session.GuildBanCreateWithReason(guildID, userID, reason, deleteMessageDays)
		if err == nil {
			fields := mergeLogFields(map[string]any{"script": scriptPath, "guildId": guildID, "userId": userID, "action": "discord.members.ban", "deleteMessageDays": deleteMessageDays}, payloadLogFields(payload))
			if reason != "" {
				fields["reason"] = reason
			}
			logLifecycleDebug("banned discord guild member from javascript", fields)
		}
		return err
	}

	ops.MemberUnban = func(ctx context.Context, guildID, userID string) error {
		_ = ctx
		guildID = strings.TrimSpace(guildID)
		userID = strings.TrimSpace(userID)
		if guildID == "" || userID == "" {
			return fmt.Errorf("member unban requires guild ID and user ID")
		}
		err := session.GuildBanDelete(guildID, userID)
		if err == nil {
			logLifecycleDebug("unbanned discord guild member from javascript", map[string]any{"script": scriptPath, "guildId": guildID, "userId": userID, "action": "discord.members.unban"})
		}
		return err
	}
}
