package jsdiscord

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func buildThreadOps(ops *DiscordOps, scriptPath string, session *discordgo.Session) {
	if ops == nil || session == nil {
		return
	}

	ops.ThreadFetch = func(ctx context.Context, threadID string) (map[string]any, error) {
		_ = ctx
		threadID = strings.TrimSpace(threadID)
		if threadID == "" {
			return nil, fmt.Errorf("thread fetch requires thread ID")
		}
		channel, err := session.Channel(threadID)
		if err != nil {
			return nil, err
		}
		if channel == nil || !channel.IsThread() {
			return nil, fmt.Errorf("channel %q is not a thread", threadID)
		}
		ret := channelSnapshotMap(channel)
		logLifecycleDebug("fetched discord thread from javascript", map[string]any{"script": scriptPath, "threadId": threadID, "action": "discord.threads.fetch"})
		return ret, nil
	}

	ops.ThreadJoin = func(ctx context.Context, threadID string) error {
		_ = ctx
		threadID = strings.TrimSpace(threadID)
		if threadID == "" {
			return fmt.Errorf("thread join requires thread ID")
		}
		err := session.ThreadJoin(threadID)
		if err == nil {
			logLifecycleDebug("joined discord thread from javascript", map[string]any{"script": scriptPath, "threadId": threadID, "action": "discord.threads.join"})
		}
		return err
	}

	ops.ThreadLeave = func(ctx context.Context, threadID string) error {
		_ = ctx
		threadID = strings.TrimSpace(threadID)
		if threadID == "" {
			return fmt.Errorf("thread leave requires thread ID")
		}
		err := session.ThreadLeave(threadID)
		if err == nil {
			logLifecycleDebug("left discord thread from javascript", map[string]any{"script": scriptPath, "threadId": threadID, "action": "discord.threads.leave"})
		}
		return err
	}

	ops.ThreadStart = func(ctx context.Context, channelID string, payload any) (map[string]any, error) {
		_ = ctx
		channelID = strings.TrimSpace(channelID)
		if channelID == "" {
			return nil, fmt.Errorf("thread start requires channel ID")
		}
		options, err := normalizeThreadStartOptions(payload)
		if err != nil {
			return nil, err
		}
		var thread *discordgo.Channel
		if options.MessageID != "" {
			thread, err = session.MessageThreadStartComplex(channelID, options.MessageID, options.Data)
		} else {
			thread, err = session.ThreadStartComplex(channelID, options.Data)
		}
		if err != nil {
			return nil, err
		}
		ret := channelSnapshotMap(thread)
		fields := mergeLogFields(map[string]any{"script": scriptPath, "channelId": channelID, "threadId": ret["id"], "action": "discord.threads.start"}, payloadLogFields(payload))
		if options.MessageID != "" {
			fields["messageId"] = options.MessageID
		}
		logLifecycleDebug("started discord thread from javascript", fields)
		return ret, nil
	}
}
