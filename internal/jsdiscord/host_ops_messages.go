package jsdiscord

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func buildMessageOps(ops *DiscordOps, scriptPath string, session *discordgo.Session) {
	if ops == nil || session == nil {
		return
	}

	ops.MessageFetch = func(ctx context.Context, channelID, messageID string) (map[string]any, error) {
		_ = ctx
		channelID = strings.TrimSpace(channelID)
		messageID = strings.TrimSpace(messageID)
		if channelID == "" || messageID == "" {
			return nil, fmt.Errorf("message fetch requires channel ID and message ID")
		}
		message, err := session.ChannelMessage(channelID, messageID)
		if err != nil {
			return nil, err
		}
		ret := messageMap(message)
		logLifecycleDebug("fetched discord channel message from javascript", map[string]any{"script": scriptPath, "channelId": channelID, "messageId": messageID, "action": "discord.messages.fetch"})
		return ret, nil
	}

	ops.MessageList = func(ctx context.Context, channelID string, payload any) ([]map[string]any, error) {
		_ = ctx
		channelID = strings.TrimSpace(channelID)
		if channelID == "" {
			return nil, fmt.Errorf("message list requires channel ID")
		}
		options, err := normalizeMessageListOptions(payload)
		if err != nil {
			return nil, err
		}
		messages, err := session.ChannelMessages(channelID, options.Limit, options.Before, options.After, options.Around)
		if err != nil {
			return nil, err
		}
		ret := make([]map[string]any, 0, len(messages))
		for _, message := range messages {
			ret = append(ret, messageMap(message))
		}
		logLifecycleDebug("listed discord channel messages from javascript", map[string]any{"script": scriptPath, "channelId": channelID, "before": options.Before, "after": options.After, "around": options.Around, "limit": options.Limit, "count": len(ret), "action": "discord.messages.list"})
		return ret, nil
	}

	ops.MessageEdit = func(ctx context.Context, channelID, messageID string, payload any) error {
		_ = ctx
		channelID = strings.TrimSpace(channelID)
		messageID = strings.TrimSpace(messageID)
		if channelID == "" || messageID == "" {
			return fmt.Errorf("message edit requires channel ID and message ID")
		}
		edit, err := normalizeChannelMessageEdit(channelID, messageID, payload)
		if err != nil {
			return err
		}
		_, err = session.ChannelMessageEditComplex(edit)
		if err == nil {
			logLifecycleDebug("edited discord channel message from javascript", mergeLogFields(map[string]any{"script": scriptPath, "channelId": channelID, "messageId": messageID, "action": "discord.messages.edit"}, payloadLogFields(payload)))
		}
		return err
	}

	ops.MessageDelete = func(ctx context.Context, channelID, messageID string) error {
		_ = ctx
		channelID = strings.TrimSpace(channelID)
		messageID = strings.TrimSpace(messageID)
		if channelID == "" || messageID == "" {
			return fmt.Errorf("message delete requires channel ID and message ID")
		}
		err := session.ChannelMessageDelete(channelID, messageID)
		if err == nil {
			logLifecycleDebug("deleted discord channel message from javascript", map[string]any{"script": scriptPath, "channelId": channelID, "messageId": messageID, "action": "discord.messages.delete"})
		}
		return err
	}

	ops.MessageReact = func(ctx context.Context, channelID, messageID, emoji string) error {
		_ = ctx
		channelID = strings.TrimSpace(channelID)
		messageID = strings.TrimSpace(messageID)
		emoji = strings.TrimSpace(emoji)
		if channelID == "" || messageID == "" || emoji == "" {
			return fmt.Errorf("message react requires channel ID, message ID, and emoji")
		}
		err := session.MessageReactionAdd(channelID, messageID, emoji)
		if err == nil {
			logLifecycleDebug("reacted to discord channel message from javascript", map[string]any{"script": scriptPath, "channelId": channelID, "messageId": messageID, "emoji": emoji, "action": "discord.messages.react"})
		}
		return err
	}

	ops.MessagePin = func(ctx context.Context, channelID, messageID string) error {
		_ = ctx
		channelID = strings.TrimSpace(channelID)
		messageID = strings.TrimSpace(messageID)
		if channelID == "" || messageID == "" {
			return fmt.Errorf("message pin requires channel ID and message ID")
		}
		err := session.ChannelMessagePin(channelID, messageID)
		if err == nil {
			logLifecycleDebug("pinned discord channel message from javascript", map[string]any{"script": scriptPath, "channelId": channelID, "messageId": messageID, "action": "discord.messages.pin"})
		}
		return err
	}

	ops.MessageUnpin = func(ctx context.Context, channelID, messageID string) error {
		_ = ctx
		channelID = strings.TrimSpace(channelID)
		messageID = strings.TrimSpace(messageID)
		if channelID == "" || messageID == "" {
			return fmt.Errorf("message unpin requires channel ID and message ID")
		}
		err := session.ChannelMessageUnpin(channelID, messageID)
		if err == nil {
			logLifecycleDebug("unpinned discord channel message from javascript", map[string]any{"script": scriptPath, "channelId": channelID, "messageId": messageID, "action": "discord.messages.unpin"})
		}
		return err
	}

	ops.MessageListPinned = func(ctx context.Context, channelID string) ([]map[string]any, error) {
		_ = ctx
		channelID = strings.TrimSpace(channelID)
		if channelID == "" {
			return nil, fmt.Errorf("list pinned messages requires channel ID")
		}
		messages, err := session.ChannelMessagesPinned(channelID)
		if err != nil {
			return nil, err
		}
		ret := make([]map[string]any, 0, len(messages))
		for _, message := range messages {
			ret = append(ret, messageMap(message))
		}
		logLifecycleDebug("listed pinned discord channel messages from javascript", map[string]any{"script": scriptPath, "channelId": channelID, "count": len(ret), "action": "discord.messages.listPinned"})
		return ret, nil
	}

	ops.MessageBulkDelete = func(ctx context.Context, channelID string, payload any) error {
		_ = ctx
		channelID = strings.TrimSpace(channelID)
		if channelID == "" {
			return fmt.Errorf("message bulkDelete requires channel ID")
		}
		messageIDs, err := normalizeMessageIDList(payload)
		if err != nil {
			return err
		}
		err = session.ChannelMessagesBulkDelete(channelID, messageIDs)
		if err == nil {
			logLifecycleDebug("bulk deleted discord channel messages from javascript", mergeLogFields(map[string]any{"script": scriptPath, "channelId": channelID, "count": len(messageIDs), "action": "discord.messages.bulkDelete"}, payloadLogFields(map[string]any{"messageIds": messageIDs})))
		}
		return err
	}
}
