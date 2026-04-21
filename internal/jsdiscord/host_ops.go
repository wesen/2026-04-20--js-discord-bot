package jsdiscord

import "github.com/bwmarrin/discordgo"

func buildDiscordOps(scriptPath string, session *discordgo.Session) *DiscordOps {
	if session == nil {
		return nil
	}
	ops := &DiscordOps{}
	buildChannelOps(ops, scriptPath, session)
	buildMessageOps(ops, scriptPath, session)
	buildMemberOps(ops, scriptPath, session)
	return ops
}
