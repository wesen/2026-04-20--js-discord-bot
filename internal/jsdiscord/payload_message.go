package jsdiscord

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func normalizeMessageReference(raw any) (*discordgo.MessageReference, error) {
	mapping, _ := raw.(map[string]any)
	if len(mapping) == 0 {
		return nil, nil
	}
	messageID := strings.TrimSpace(fmt.Sprint(mapping["messageId"]))
	if messageID == "" {
		return nil, fmt.Errorf("replyTo.messageId is required")
	}
	reference := &discordgo.MessageReference{MessageID: messageID}
	if channelID := strings.TrimSpace(fmt.Sprint(mapping["channelId"])); channelID != "" {
		reference.ChannelID = channelID
	}
	if guildID := strings.TrimSpace(fmt.Sprint(mapping["guildId"])); guildID != "" {
		reference.GuildID = guildID
	}
	return reference, nil
}
