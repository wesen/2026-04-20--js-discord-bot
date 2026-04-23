package jsdiscord

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestShowSpaceUpcomingListsSeedShows(t *testing.T) {
	handle := loadTestBot(t, filepath.Join(repoRootJSDiscord(t), "examples", "discord-bots", "show-space", "index.js"))
	result, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name: "upcoming",
		Config: map[string]any{
			"timeZone": "America/New_York",
		},
	})
	if err != nil {
		t.Fatalf("dispatch command: %v", err)
	}
	payload, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("result type = %T", result)
	}
	if got := fmt.Sprint(payload["content"]); got == "" || got == "map[]" {
		t.Fatalf("content = %#v", payload["content"])
	}
	if !strings.Contains(fmt.Sprint(payload["content"]), "Upcoming Shows") {
		t.Fatalf("content = %#v", payload["content"])
	}
}

func TestShowSpaceAnnouncePinsAndRejectsUnauthorizedUsers(t *testing.T) {
	handle := loadTestBot(t, filepath.Join(repoRootJSDiscord(t), "examples", "discord-bots", "show-space", "index.js"))

	unauthorized, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name: "announce",
		Args: map[string]any{
			"artist": "Test Set",
			"date": "2026-05-22",
			"doors_time": "7pm",
			"age_restriction": "All ages",
			"price": "Free",
		},
		Member: &MemberSnapshot{ID: "user-1", Roles: []string{"member"}},
		Config: map[string]any{
			"upcomingShowsChannelId": "chan-1",
			"adminRoleId":             "admin-role",
			"bookerRoleId":            "booker-role",
			"timeZone":                "America/New_York",
		},
	})
	if err != nil {
		t.Fatalf("dispatch unauthorized command: %v", err)
	}
	if got := fmt.Sprint(unauthorized); !strings.Contains(got, "don't have permission") {
		t.Fatalf("unauthorized result = %s", got)
	}

	var sends, pins, lists int
	var sentPayloads []any
	authorized, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name: "announce",
		Args: map[string]any{
			"artist":           "Test Set",
			"date":             "2026-05-22",
			"doors_time":       "7pm",
			"age_restriction":  "All ages",
			"price":            "Free",
			"notes":            "Bring a jacket.",
		},
		Member: &MemberSnapshot{ID: "user-1", Roles: []string{"booker-role"}},
		Config: map[string]any{
			"upcomingShowsChannelId": "chan-1",
			"adminRoleId":             "admin-role",
			"bookerRoleId":            "booker-role",
			"timeZone":                "America/New_York",
		},
		Discord: &DiscordOps{
			ChannelSend: func(_ context.Context, channelID string, payload any) error {
				sends++
				if channelID != "chan-1" {
					t.Fatalf("channelID = %s", channelID)
				}
				sentPayloads = append(sentPayloads, payload)
				return nil
			},
			MessageList: func(_ context.Context, channelID string, payload any) ([]map[string]any, error) {
				lists++
				return []map[string]any{{
					"id": "msg-1",
					"embeds": []any{map[string]any{"title": "🎵 Test Set — Fri May 22, 2026"}},
				}}, nil
			},
			MessagePin: func(_ context.Context, channelID, messageID string) error {
				pins++
				if channelID != "chan-1" || messageID != "msg-1" {
					t.Fatalf("pin target = %s/%s", channelID, messageID)
				}
				return nil
			},
		},
	})
	if err != nil {
		t.Fatalf("dispatch command: %v", err)
	}
	if got := fmt.Sprint(authorized); !strings.Contains(got, "Posted and pinned") {
		t.Fatalf("authorized result = %s", got)
	}
	if sends != 1 || lists != 1 || pins != 1 {
		t.Fatalf("counts = sends:%d lists:%d pins:%d", sends, lists, pins)
	}
	if len(sentPayloads) != 1 {
		t.Fatalf("sent payloads = %#v", sentPayloads)
	}
	if payload, ok := sentPayloads[0].(map[string]any); !ok || fmt.Sprint(payload["embeds"]) == "" {
		t.Fatalf("announcement payload = %#v", sentPayloads[0])
	}
}

func TestShowSpaceUnpinOldRemovesExpiredPins(t *testing.T) {
	handle := loadTestBot(t, filepath.Join(repoRootJSDiscord(t), "examples", "discord-bots", "show-space", "index.js"))
	var unpins int
	result, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name: "unpin-old",
		Member: &MemberSnapshot{ID: "user-1", Roles: []string{"admin-role"}},
		Config: map[string]any{
			"upcomingShowsChannelId": "chan-1",
			"adminRoleId":             "admin-role",
			"bookerRoleId":            "booker-role",
			"timeZone":                "America/New_York",
		},
		Discord: &DiscordOps{
			MessageListPinned: func(_ context.Context, channelID string) ([]map[string]any, error) {
				return []map[string]any{
					{"id": "old-1", "embeds": []any{map[string]any{"title": "🎵 Old Timer — Fri Jan 1, 2021"}}},
					{"id": "future-1", "embeds": []any{map[string]any{"title": "🎵 Future Wave — Fri Jan 1, 2099"}}},
				}, nil
			},
			MessageUnpin: func(_ context.Context, channelID, messageID string) error {
				unpins++
				if messageID != "old-1" {
					t.Fatalf("unexpected unpin target = %s", messageID)
				}
				return nil
			},
		},
	})
	if err != nil {
		t.Fatalf("dispatch command: %v", err)
	}
	if got := fmt.Sprint(result); !strings.Contains(got, "Removed 1") {
		t.Fatalf("result = %s", got)
	}
	if unpins != 1 {
		t.Fatalf("unpins = %d", unpins)
	}
}
