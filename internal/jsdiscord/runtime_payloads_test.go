package jsdiscord

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestNormalizeResponsePayloadSupportsEmbedsAndSelectComponents(t *testing.T) {
	payload, err := normalizeResponsePayload(map[string]any{
		"content":   "pong",
		"ephemeral": true,
		"embeds": []any{
			map[string]any{"title": "Pong", "description": "hello", "color": 123},
		},
		"components": []any{
			map[string]any{
				"type": "actionRow",
				"components": []any{
					map[string]any{
						"type":     "select",
						"customId": "support:queue",
						"options": []any{
							map[string]any{"label": "Billing", "value": "billing"},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("normalizeResponsePayload: %v", err)
	}
	if payload.Content != "pong" {
		t.Fatalf("content = %q", payload.Content)
	}
	if payload.Flags != discordgo.MessageFlagsEphemeral {
		t.Fatalf("flags = %v", payload.Flags)
	}
	if len(payload.Embeds) != 1 || payload.Embeds[0].Title != "Pong" {
		t.Fatalf("embeds = %#v", payload.Embeds)
	}
	if len(payload.Components) != 1 {
		t.Fatalf("components = %#v", payload.Components)
	}
}

func TestNormalizeModalPayloadSupportsTextInputs(t *testing.T) {
	payload, err := normalizeModalPayload(map[string]any{
		"customId": "feedback:submit",
		"title":    "Feedback",
		"components": []any{
			map[string]any{
				"type": "actionRow",
				"components": []any{
					map[string]any{
						"type":     "textInput",
						"customId": "summary",
						"label":    "Summary",
						"style":    "short",
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("normalizeModalPayload: %v", err)
	}
	if payload.CustomID != "feedback:submit" || payload.Title != "Feedback" {
		t.Fatalf("payload = %#v", payload)
	}
	if len(payload.Components) != 1 {
		t.Fatalf("components = %#v", payload.Components)
	}
}

func TestNormalizeThreadStartOptions(t *testing.T) {
	options, err := normalizeThreadStartOptions(map[string]any{
		"name":                "Support follow-up",
		"type":                "private",
		"autoArchiveDuration": 60,
		"invitable":           true,
		"rateLimitPerUser":    5,
	})
	if err != nil {
		t.Fatalf("normalizeThreadStartOptions: %v", err)
	}
	if options.MessageID != "" {
		t.Fatalf("messageID = %q", options.MessageID)
	}
	if options.Data == nil || options.Data.Name != "Support follow-up" {
		t.Fatalf("data = %#v", options.Data)
	}
	if options.Data.Type != discordgo.ChannelTypeGuildPrivateThread {
		t.Fatalf("type = %v", options.Data.Type)
	}
	if options.Data.AutoArchiveDuration != 60 || !options.Data.Invitable || options.Data.RateLimitPerUser != 5 {
		t.Fatalf("data = %#v", options.Data)
	}
	messageOptions, err := normalizeThreadStartOptions(map[string]any{"name": "Follow-up", "messageId": "msg-1"})
	if err != nil {
		t.Fatalf("normalizeThreadStartOptions message thread: %v", err)
	}
	if messageOptions.MessageID != "msg-1" {
		t.Fatalf("message thread options = %#v", messageOptions)
	}
	if _, err := normalizeThreadStartOptions(map[string]any{"type": "public"}); err == nil {
		t.Fatalf("expected missing name error")
	}
}

func TestNormalizeMessageSendSupportsFilesAndReplyReference(t *testing.T) {
	message, err := normalizeMessageSend(map[string]any{
		"content": "report",
		"files": []any{
			map[string]any{"name": "report.txt", "content": "hello", "contentType": "text/plain"},
		},
		"replyTo": map[string]any{"messageId": "orig-1", "channelId": "chan-1"},
	})
	if err != nil {
		t.Fatalf("normalizeMessageSend: %v", err)
	}
	if len(message.Files) != 1 || message.Files[0].Name != "report.txt" {
		t.Fatalf("files = %#v", message.Files)
	}
	if message.Reference == nil || message.Reference.MessageID != "orig-1" || message.Reference.ChannelID != "chan-1" {
		t.Fatalf("reference = %#v", message.Reference)
	}
}
