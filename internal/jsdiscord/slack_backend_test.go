package jsdiscord

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
)

func TestVerifySlackRequest(t *testing.T) {
	secret := "secret"
	body := []byte("token=x&team_id=T1")
	now := time.Unix(1700000000, 0)
	timestamp := "1700000000"
	base := "v0:" + timestamp + ":" + string(body)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(base))
	sig := "v0=" + hex.EncodeToString(mac.Sum(nil))
	if err := VerifySlackRequest(secret, timestamp, sig, body, now); err != nil {
		t.Fatalf("expected valid signature: %v", err)
	}
	if err := VerifySlackRequest(secret, timestamp, "v0=bad", body, now); err == nil {
		t.Fatalf("expected invalid signature to fail")
	}
}

func TestSlackManifestUsesCommands(t *testing.T) {
	desc := &BotDescriptor{Name: "adventure", Description: "Adventure bot", Commands: []CommandDescriptor{{Name: "adventure-start", Description: "Start", Spec: map[string]any{"options": map[string]any{"prompt": map[string]any{"type": "string"}}}}}}
	manifest := SlackManifest(desc, "https://bot.example")
	features := manifest["features"].(map[string]any)
	commands := features["slash_commands"].([]map[string]any)
	if len(commands) != 1 {
		t.Fatalf("expected one command, got %d", len(commands))
	}
	if commands[0]["command"] != "/adventure-start" {
		t.Fatalf("unexpected command: %#v", commands[0])
	}
	if commands[0]["url"] != "https://bot.example/slack/commands" {
		t.Fatalf("unexpected url: %#v", commands[0])
	}
	if commands[0]["usage_hint"] != "prompt text" {
		t.Fatalf("unexpected usage hint: %#v", commands[0])
	}
	scopes := manifest["oauth_config"].(map[string]any)["scopes"].(map[string]any)["bot"].([]string)
	if !containsString(scopes, "app_mentions:read") {
		t.Fatalf("expected app_mentions:read scope, got %#v", scopes)
	}
	if !containsString(scopes, "files:write") {
		t.Fatalf("expected files:write scope, got %#v", scopes)
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func TestSlackBaseRequestUsesUserMentionAsUsername(t *testing.T) {
	backend := &SlackBackend{host: &Host{scriptPath: "bot.js", runtimeConfig: map[string]any{}}}
	req := backend.slackBaseRequest("T1", "C1", "U05MP5JKZTP")
	if req.User.Username != "<@U05MP5JKZTP>" {
		t.Fatalf("expected Slack mention username, got %q", req.User.Username)
	}
	if req.Member == nil || req.Member.User == nil || req.Member.User.Username != "<@U05MP5JKZTP>" {
		t.Fatalf("expected member user mention, got %#v", req.Member)
	}
}

func TestSlackMessagePayloadMapsButtonsAndInlineFiles(t *testing.T) {
	payload := &normalizedResponse{
		Content: "Choose",
		Components: []discordgo.MessageComponent{discordgo.ActionsRow{Components: []discordgo.MessageComponent{
			discordgo.Button{CustomID: "adv:choice:0", Label: "Dive", Style: discordgo.PrimaryButton},
			discordgo.Button{CustomID: "adv:choice:1", Label: "Panic", Style: discordgo.DangerButton},
		}}},
		Files: []*discordgo.File{{Name: "export.json", ContentType: "application/json", Reader: strings.NewReader(`{"ok":true}`)}},
	}
	msg, err := slackMessagePayload(payload)
	if err != nil {
		t.Fatalf("payload failed: %v", err)
	}
	if !strings.Contains(msg["text"].(string), "export.json") {
		t.Fatalf("expected inline file in text: %#v", msg["text"])
	}
	blocks := msg["blocks"].([]map[string]any)
	if len(blocks) != 2 {
		t.Fatalf("expected section + actions, got %#v", blocks)
	}
	elements := blocks[1]["elements"].([]map[string]any)
	if elements[0]["action_id"] != "adv:choice:0" || elements[1]["style"] != "danger" {
		t.Fatalf("unexpected buttons: %#v", elements)
	}
}

func TestSlackResponderPublicCommandReplyCreatesEditableMessage(t *testing.T) {
	calls := []string{}
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.URL.Path)
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if r.URL.Path == "/chat.update" && payload["ts"] != "100.1" {
			t.Fatalf("expected update ts, got %#v", payload)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"channel":"C1","ts":"100.1"}`))
	}))
	defer api.Close()
	store, err := OpenSlackStore(t.TempDir() + "/slack.sqlite")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()
	client := &SlackClient{BotToken: "xoxb-test", APIBaseURL: api.URL, HTTPClient: api.Client()}
	responder := newSlackResponder(client, store, "T1", "https://response.example", "C1", "", "trigger", "script.js")
	if err := responder.Reply(t.Context(), map[string]any{"content": "first"}); err != nil {
		t.Fatalf("reply: %v", err)
	}
	if responder.messageTS != "100.1" {
		t.Fatalf("expected responder to remember ts, got %q", responder.messageTS)
	}
	if err := responder.Edit(t.Context(), map[string]any{"content": "second"}); err != nil {
		t.Fatalf("edit: %v", err)
	}
	if strings.Join(calls, ",") != "/chat.postMessage,/chat.update" {
		t.Fatalf("unexpected calls: %#v", calls)
	}
}

func TestSlackStorePersistsMessage(t *testing.T) {
	store, err := OpenSlackStore(t.TempDir() + "/slack.sqlite")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()
	if err := store.UpsertMessage(SlackMessage{TeamID: "T1", ChannelID: "C1", TS: "1.2", Content: "Turn 4"}, map[string]any{"ok": true}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	msg, err := store.GetMessage("T1", "C1", "1.2")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if msg.Content != "Turn 4" || msg.ID == "" {
		t.Fatalf("unexpected message: %#v", msg)
	}
}
