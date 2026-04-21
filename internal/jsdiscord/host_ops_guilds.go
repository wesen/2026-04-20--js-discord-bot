package jsdiscord

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func buildGuildOps(ops *DiscordOps, scriptPath string, session *discordgo.Session) {
	if ops == nil || session == nil {
		return
	}

	ops.GuildFetch = func(ctx context.Context, guildID string) (map[string]any, error) {
		_ = ctx
		guildID = strings.TrimSpace(guildID)
		if guildID == "" {
			return nil, fmt.Errorf("guild fetch requires guild ID")
		}
		guild, err := session.Guild(guildID)
		if err != nil {
			return nil, err
		}
		ret := guildSnapshotMap(guild)
		logLifecycleDebug("fetched discord guild from javascript", map[string]any{"script": scriptPath, "guildId": guildID, "action": "discord.guilds.fetch"})
		return ret, nil
	}
}
