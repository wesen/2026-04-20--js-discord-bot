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
			"artist":          "Test Set",
			"date":            "2026-05-22",
			"doors_time":      "7pm",
			"age_restriction": "All ages",
			"price":           "Free",
		},
		Member: &MemberSnapshot{ID: "user-1", Roles: []string{"member"}},
		Config: map[string]any{
			"upcomingShowsChannelId": "chan-1",
			"adminRoleId":            "admin-role",
			"bookerRoleId":           "booker-role",
			"timeZone":               "America/New_York",
		},
	})
	if err != nil {
		t.Fatalf("dispatch unauthorized command: %v", err)
	}
	if got := fmt.Sprint(unauthorized); !strings.Contains(got, "don't have permission") || !strings.Contains(got, "User ID") || !strings.Contains(got, "member role IDs") || !strings.Contains(got, "Exact matching role IDs") || !strings.Contains(got, "Why denied") || !strings.Contains(got, "debug-my-roles") || !strings.Contains(got, "/debug") {
		t.Fatalf("unauthorized result = %s", got)
	}

	var sends, pins, lists int
	var sentPayloads []any
	authorized, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name: "announce",
		Args: map[string]any{
			"artist":          "Test Set",
			"date":            "2026-05-22",
			"doors_time":      "7pm",
			"age_restriction": "All ages",
			"price":           "Free",
			"notes":           "Bring a jacket.",
		},
		Member: &MemberSnapshot{ID: "user-1", Roles: []string{"booker-role"}},
		Config: map[string]any{
			"upcomingShowsChannelId": "chan-1",
			"adminRoleId":            "admin-role",
			"bookerRoleId":           "booker-role",
			"timeZone":               "America/New_York",
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
					"id":     "msg-1",
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
		Name:   "unpin-old",
		Member: &MemberSnapshot{ID: "user-1", Roles: []string{"admin-role"}},
		Config: map[string]any{
			"upcomingShowsChannelId": "chan-1",
			"adminRoleId":            "admin-role",
			"bookerRoleId":           "booker-role",
			"timeZone":               "America/New_York",
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

func TestShowSpaceDatabaseCommandsCreateShowLookupsAndArchiveIt(t *testing.T) {
	handle := loadTestBot(t, filepath.Join(repoRootJSDiscord(t), "examples", "discord-bots", "show-space", "index.js"))
	dbPath := filepath.Join(t.TempDir(), "shows.sqlite")
	config := map[string]any{
		"dbPath":                 dbPath,
		"seedFromJson":           false,
		"upcomingShowsChannelId": "chan-1",
		"staffChannelId":         "staff-1",
		"adminRoleId":            "admin-role",
		"bookerRoleId":           "booker-role",
		"timeZone":               "America/New_York",
	}

	var sends, pins, unpins int
	var sentPayloads []any
	addResult, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name: "add-show",
		Args: map[string]any{
			"artist":     "Demo Night",
			"date":       "2026-06-03",
			"doors_time": "7pm",
			"age":        "All ages",
			"price":      "$12",
			"notes":      "Bring earplugs.",
		},
		Member: &MemberSnapshot{ID: "user-1", Roles: []string{"booker-role"}},
		Config: config,
		Discord: &DiscordOps{
			ChannelSend: func(_ context.Context, channelID string, payload any) error {
				sends++
				if channelID != "chan-1" {
					t.Fatalf("announcement channel = %s", channelID)
				}
				sentPayloads = append(sentPayloads, payload)
				return nil
			},
			MessageList: func(_ context.Context, channelID string, payload any) ([]map[string]any, error) {
				return []map[string]any{{
					"id":     "msg-1",
					"embeds": []any{map[string]any{"title": "🎵 Demo Night — Wed Jun 3, 2026"}},
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
		t.Fatalf("dispatch add-show: %v", err)
	}
	if got := fmt.Sprint(addResult); !strings.Contains(got, "Show added") || !strings.Contains(got, "#1") {
		t.Fatalf("add-show result = %s", got)
	}
	if sends != 1 || pins != 1 {
		t.Fatalf("add-show counts = sends:%d pins:%d", sends, pins)
	}
	if len(sentPayloads) != 1 {
		t.Fatalf("announcement payloads = %#v", sentPayloads)
	}

	showResult, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name:   "show",
		Args:   map[string]any{"id": 1},
		Config: config,
	})
	if err != nil {
		t.Fatalf("dispatch show: %v", err)
	}
	if got := fmt.Sprint(showResult); !strings.Contains(got, "Demo Night") || !strings.Contains(got, "ID") {
		t.Fatalf("show result = %s", got)
	}

	cancelResult, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name:   "cancel-show",
		Args:   map[string]any{"id": 1},
		Member: &MemberSnapshot{ID: "user-1", Roles: []string{"booker-role"}},
		Config: config,
		Discord: &DiscordOps{
			MessageUnpin: func(_ context.Context, channelID, messageID string) error {
				unpins++
				if channelID != "chan-1" || messageID != "msg-1" {
					t.Fatalf("cancel unpin target = %s/%s", channelID, messageID)
				}
				return nil
			},
			ChannelSend: func(_ context.Context, channelID string, payload any) error {
				sends++
				return nil
			},
		},
	})
	if err != nil {
		t.Fatalf("dispatch cancel-show: %v", err)
	}
	if got := fmt.Sprint(cancelResult); !strings.Contains(got, "cancelled") {
		t.Fatalf("cancel-show result = %s", got)
	}

	archiveResult, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name:   "archive-show",
		Args:   map[string]any{"id": 1},
		Member: &MemberSnapshot{ID: "user-1", Roles: []string{"admin-role"}},
		Config: config,
		Discord: &DiscordOps{
			MessageUnpin: func(_ context.Context, channelID, messageID string) error {
				unpins++
				return nil
			},
		},
	})
	if err != nil {
		t.Fatalf("dispatch archive-show: %v", err)
	}
	if got := fmt.Sprint(archiveResult); !strings.Contains(got, "archived") {
		t.Fatalf("archive-show result = %s", got)
	}

	pastResult, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name:   "past-shows",
		Config: config,
	})
	if err != nil {
		t.Fatalf("dispatch past-shows: %v", err)
	}
	if got := fmt.Sprint(pastResult); !strings.Contains(got, "Demo Night") || !strings.Contains(got, "Past Shows") {
		t.Fatalf("past-shows result = %s", got)
	}
	if unpins < 2 {
		t.Fatalf("unpins = %d", unpins)
	}
}

func TestShowSpaceArchiveExpiredSummarizesAndUnpins(t *testing.T) {
	handle := loadTestBot(t, filepath.Join(repoRootJSDiscord(t), "examples", "discord-bots", "show-space", "index.js"))
	config := map[string]any{
		"dbPath":                 filepath.Join(t.TempDir(), "shows.sqlite"),
		"seedFromJson":           false,
		"upcomingShowsChannelId": "chan-1",
		"staffChannelId":         "staff-1",
		"adminRoleId":            "admin-role",
		"bookerRoleId":           "booker-role",
		"timeZone":               "America/New_York",
	}

	var sends, unpins int
	_, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name: "add-show",
		Args: map[string]any{
			"artist":     "Ancient Echo",
			"date":       "2021-01-01",
			"doors_time": "7pm",
			"age":        "All ages",
			"price":      "Free",
			"notes":      "This one should archive.",
		},
		Member: &MemberSnapshot{ID: "user-1", Roles: []string{"booker-role"}},
		Config: config,
		Discord: &DiscordOps{
			ChannelSend: func(_ context.Context, channelID string, payload any) error {
				sends++
				return nil
			},
			MessageList: func(_ context.Context, channelID string, payload any) ([]map[string]any, error) {
				return []map[string]any{{
					"id":     "msg-old",
					"embeds": []any{map[string]any{"title": "🎵 Ancient Echo — Fri Jan 1, 2021"}},
				}}, nil
			},
			MessagePin: func(_ context.Context, channelID, messageID string) error {
				return nil
			},
		},
	})
	if err != nil {
		t.Fatalf("dispatch add-show: %v", err)
	}

	result, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name:   "archive-expired",
		Member: &MemberSnapshot{ID: "user-1", Roles: []string{"admin-role"}},
		Config: config,
		Discord: &DiscordOps{
			MessageUnpin: func(_ context.Context, channelID, messageID string) error {
				unpins++
				if channelID != "chan-1" || messageID != "msg-old" {
					t.Fatalf("archive-expired unpin target = %s/%s", channelID, messageID)
				}
				return nil
			},
			ChannelSend: func(_ context.Context, channelID string, payload any) error {
				sends++
				return nil
			},
		},
	})
	if err != nil {
		t.Fatalf("dispatch archive-expired: %v", err)
	}
	if got := fmt.Sprint(result); !strings.Contains(got, "Archived 1") || !strings.Contains(got, "unpinned 1") {
		t.Fatalf("archive-expired result = %s", got)
	}
	if sends < 2 || unpins != 1 {
		t.Fatalf("counts = sends:%d unpins:%d", sends, unpins)
	}
}

func TestShowSpaceDebugRolesRequiresDebugFlag(t *testing.T) {
	handle := loadTestBot(t, filepath.Join(repoRootJSDiscord(t), "examples", "discord-bots", "show-space", "index.js"))
	called := false
	result, err := handle.DispatchCommand(context.Background(), DispatchRequest{
		Name: "debug-roles",
		Config: map[string]any{
			"debug": false,
		},
		Guild: map[string]any{"id": "guild-1", "name": "The Venue"},
		Discord: &DiscordOps{
			RoleList: func(_ context.Context, guildID string) ([]map[string]any, error) {
				called = true
				return nil, nil
			},
		},
	})
	if err != nil {
		t.Fatalf("dispatch debug-roles: %v", err)
	}
	if called {
		t.Fatalf("role list should not be called when debug is disabled")
	}
	if got := fmt.Sprint(result); !strings.Contains(got, "Debug mode is disabled") {
		t.Fatalf("disabled result = %s", got)
	}
}

func TestShowSpaceDebugDashboardShowsUserAndRoleButtons(t *testing.T) {
	handle := loadTestBot(t, filepath.Join(repoRootJSDiscord(t), "examples", "discord-bots", "show-space", "index.js"))
	result, err := handle.DispatchCommandAsMap(context.Background(), DispatchRequest{
		Name: "debug",
		Config: map[string]any{
			"debug":        true,
			"adminRoleId":  "role-admin",
			"bookerRoleId": "role-booker",
		},
		User:   UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:  map[string]any{"id": "guild-1", "name": "The Venue"},
		Member: &MemberSnapshot{ID: "user-1", Roles: []string{"role-admin", "role-booker", "role-helper"}},
		Discord: &DiscordOps{
			RoleList: func(_ context.Context, guildID string) ([]map[string]any, error) {
				if guildID != "guild-1" {
					t.Fatalf("guildID = %s", guildID)
				}
				return []map[string]any{
					{"id": "role-admin", "name": "admin"},
					{"id": "role-booker", "name": "booker"},
					{"id": "role-helper", "name": "helper"},
				}, nil
			},
		},
	})
	if err != nil {
		t.Fatalf("dispatch debug dashboard: %v", err)
	}
	if got := fmt.Sprint(result["content"]); !strings.Contains(got, "User ID: user-1") || !strings.Contains(got, "Guild: The Venue") {
		t.Fatalf("dashboard content = %s", got)
	}
	components, ok := result["components"].([]any)
	if !ok || len(components) != 1 {
		t.Fatalf("dashboard components = %#v", result["components"])
	}
	if got := fmt.Sprint(components[0]); !strings.Contains(got, "show-space:debug:summary") || !strings.Contains(got, "show-space:debug:member") || !strings.Contains(got, "show-space:debug:guild") || !strings.Contains(got, "show-space:debug:config") || !strings.Contains(got, "show-space:debug:checks") {
		t.Fatalf("dashboard buttons = %s", got)
	}

	memberView, err := handle.DispatchComponentAsMap(context.Background(), DispatchRequest{
		Name:   "show-space:debug:member",
		User:   UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:  map[string]any{"id": "guild-1", "name": "The Venue"},
		Member: &MemberSnapshot{ID: "user-1", Roles: []string{"role-admin", "role-booker", "role-helper"}},
		Config: map[string]any{
			"debug":        true,
			"adminRoleId":  "role-admin",
			"bookerRoleId": "role-booker",
		},
		Discord: &DiscordOps{
			RoleList: func(_ context.Context, guildID string) ([]map[string]any, error) {
				return []map[string]any{
					{"id": "role-admin", "name": "admin"},
					{"id": "role-booker", "name": "booker"},
					{"id": "role-helper", "name": "helper"},
				}, nil
			},
		},
	})
	if err != nil {
		t.Fatalf("dispatch member view: %v", err)
	}
	if got := fmt.Sprint(memberView["content"]); !strings.Contains(got, "User ID: user-1") || !strings.Contains(strings.ToLower(got), "member roles") {
		t.Fatalf("member view content = %s", got)
	}
	if got := fmt.Sprint(memberView); !strings.Contains(got, "admin — role-admin") || !strings.Contains(got, "booker — role-booker") || !strings.Contains(got, "helper — role-helper") {
		t.Fatalf("member view = %s", got)
	}

	checksView, err := handle.DispatchComponentAsMap(context.Background(), DispatchRequest{
		Name:   "show-space:debug:checks",
		User:   UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:  map[string]any{"id": "guild-1", "name": "The Venue"},
		Member: &MemberSnapshot{ID: "user-1", Roles: []string{"role-admin", "role-booker", "role-helper"}},
		Config: map[string]any{
			"debug":        true,
			"adminRoleId":  "role-admin",
			"bookerRoleId": "role-booker",
		},
		Discord: &DiscordOps{
			RoleList: func(_ context.Context, guildID string) ([]map[string]any, error) {
				return []map[string]any{
					{"id": "role-admin", "name": "admin"},
					{"id": "role-booker", "name": "booker"},
					{"id": "role-helper", "name": "helper"},
				}, nil
			},
		},
	})
	if err != nil {
		t.Fatalf("dispatch checks view: %v", err)
	}
	if got := fmt.Sprint(checksView["content"]); !strings.Contains(got, "User ID: user-1") {
		t.Fatalf("checks content = %s", got)
	}
	if got := fmt.Sprint(checksView); !strings.Contains(got, "canManageShows: yes") || !strings.Contains(got, "canAdminOnly: yes") || !strings.Contains(got, "Exact matching role IDs") || !strings.Contains(got, "role-admin") || !strings.Contains(got, "role-booker") {
		t.Fatalf("checks view = %s", got)
	}
}

func TestShowSpaceDebugRolesListsGuildRolesWhenEnabled(t *testing.T) {
	handle := loadTestBot(t, filepath.Join(repoRootJSDiscord(t), "examples", "discord-bots", "show-space", "index.js"))
	result, err := handle.DispatchCommandAsMap(context.Background(), DispatchRequest{
		Name: "debug-roles",
		Config: map[string]any{
			"debug": true,
		},
		User:  UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild: map[string]any{"id": "guild-1", "name": "The Venue"},
		Discord: &DiscordOps{
			RoleList: func(_ context.Context, guildID string) ([]map[string]any, error) {
				if guildID != "guild-1" {
					t.Fatalf("guildID = %s", guildID)
				}
				return []map[string]any{
					{"id": "role-admin", "name": "admin"},
					{"id": "role-booker", "name": "booker"},
				}, nil
			},
		},
	})
	if err != nil {
		t.Fatalf("dispatch debug-roles: %v", err)
	}
	if got := fmt.Sprint(result["content"]); !strings.Contains(got, "User ID: user-1") || !strings.Contains(got, "Guild: The Venue") {
		t.Fatalf("enabled content = %s", got)
	}
	if got := fmt.Sprint(result); !strings.Contains(got, "role-admin") || !strings.Contains(got, "role-booker") {
		t.Fatalf("enabled result = %s", got)
	}
}

func TestShowSpaceDebugMyRolesListsRolesVisibleOnMember(t *testing.T) {
	handle := loadTestBot(t, filepath.Join(repoRootJSDiscord(t), "examples", "discord-bots", "show-space", "index.js"))
	result, err := handle.DispatchCommandAsMap(context.Background(), DispatchRequest{
		Name: "debug-my-roles",
		Config: map[string]any{
			"debug": true,
		},
		User:   UserSnapshot{ID: "user-1", Username: "Ada", Bot: false},
		Guild:  map[string]any{"id": "guild-1", "name": "The Venue"},
		Member: &MemberSnapshot{ID: "user-1", Roles: []string{"role-admin", "role-booker", "role-helper"}},
		Discord: &DiscordOps{
			RoleList: func(_ context.Context, guildID string) ([]map[string]any, error) {
				if guildID != "guild-1" {
					t.Fatalf("guildID = %s", guildID)
				}
				return []map[string]any{
					{"id": "role-admin", "name": "admin"},
					{"id": "role-booker", "name": "booker"},
				}, nil
			},
		},
	})
	if err != nil {
		t.Fatalf("dispatch debug-my-roles: %v", err)
	}
	if got := fmt.Sprint(result["content"]); !strings.Contains(got, "User ID: user-1") || !strings.Contains(got, "Member ID: user-1") {
		t.Fatalf("debug-my-roles content = %s", got)
	}
	if got := fmt.Sprint(result); !strings.Contains(got, "admin — role-admin") || !strings.Contains(got, "booker — role-booker") || !strings.Contains(got, "(unknown role) — role-helper") {
		t.Fatalf("debug-my-roles result = %s", got)
	}
}
