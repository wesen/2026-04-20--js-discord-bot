package jsdiscord

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func buildChannelOps(ops *DiscordOps, scriptPath string, session *discordgo.Session) {
	if ops == nil || session == nil {
		return
	}

	ops.ChannelSend = func(ctx context.Context, channelID string, payload any) error {
		_ = ctx
		channelID = strings.TrimSpace(channelID)
		if channelID == "" {
			return fmt.Errorf("channel send requires channel ID")
		}
		message, err := normalizeMessageSend(payload)
		if err != nil {
			return err
		}
		_, err = session.ChannelMessageSendComplex(channelID, message)
		if err == nil {
			logLifecycleDebug("sent discord channel message from javascript", mergeLogFields(map[string]any{"script": scriptPath, "channelId": channelID, "action": "discord.channels.send"}, payloadLogFields(payload)))
		}
		return err
	}

	ops.ChannelFetch = func(ctx context.Context, channelID string) (map[string]any, error) {
		_ = ctx
		channelID = strings.TrimSpace(channelID)
		if channelID == "" {
			return nil, fmt.Errorf("channel fetch requires channel ID")
		}
		channel, err := session.Channel(channelID)
		if err != nil {
			return nil, err
		}
		ret := channelSnapshotMap(channel)
		logLifecycleDebug("fetched discord channel from javascript", map[string]any{"script": scriptPath, "channelId": channelID, "action": "discord.channels.fetch"})
		return ret, nil
	}

	ops.ChannelSetTopic = func(ctx context.Context, channelID, topic string) error {
		_ = ctx
		channelID = strings.TrimSpace(channelID)
		if channelID == "" {
			return fmt.Errorf("channel setTopic requires channel ID")
		}
		_, err := session.ChannelEdit(channelID, &discordgo.ChannelEdit{Topic: topic})
		if err == nil {
			logLifecycleDebug("updated discord channel topic from javascript", map[string]any{"script": scriptPath, "channelId": channelID, "topic": truncateLogText(topic, 120), "action": "discord.channels.setTopic"})
		}
		return err
	}

	ops.ChannelSetSlowmode = func(ctx context.Context, channelID string, seconds int) error {
		_ = ctx
		channelID = strings.TrimSpace(channelID)
		if channelID == "" {
			return fmt.Errorf("channel setSlowmode requires channel ID")
		}
		rate := seconds
		_, err := session.ChannelEdit(channelID, &discordgo.ChannelEdit{RateLimitPerUser: &rate})
		if err == nil {
			logLifecycleDebug("updated discord channel slowmode from javascript", map[string]any{"script": scriptPath, "channelId": channelID, "seconds": seconds, "action": "discord.channels.setSlowmode"})
		}
		return err
	}
}
