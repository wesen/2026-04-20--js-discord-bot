package jsdiscord

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

// SlackConfig configures the HTTP Slack backend. The JavaScript layer remains
// Discord-shaped; this adapter maps Slack payloads to the existing DispatchRequest.
type SlackConfig struct {
	BotToken      string
	SigningSecret string
	BaseURL       string
	ListenAddr    string
	StateDBPath   string
}

// SlackBackend serves Slack slash command, interactivity, and events endpoints.
type SlackBackend struct {
	host   *Host
	config SlackConfig
	store  *SlackStore
	client *SlackClient
	server *http.Server
}

func NewSlackBackend(ctx context.Context, scriptPath string, config SlackConfig, opts ...HostOption) (*SlackBackend, error) {
	if strings.TrimSpace(scriptPath) == "" {
		return nil, fmt.Errorf("slack backend bot script path is empty")
	}
	host, err := NewHost(ctx, scriptPath, opts...)
	if err != nil {
		return nil, err
	}
	statePath := strings.TrimSpace(config.StateDBPath)
	if statePath == "" {
		statePath = "./slack-state.sqlite"
	}
	absState, err := filepath.Abs(statePath)
	if err != nil {
		_ = host.Close(context.Background())
		return nil, fmt.Errorf("resolve slack state path: %w", err)
	}
	store, err := OpenSlackStore(absState)
	if err != nil {
		_ = host.Close(context.Background())
		return nil, err
	}
	config.StateDBPath = absState
	return &SlackBackend{host: host, config: config, store: store, client: NewSlackClient(config.BotToken)}, nil
}

func (b *SlackBackend) Close(ctx context.Context) error {
	if b == nil {
		return nil
	}
	if b.server != nil {
		_ = b.server.Shutdown(ctx)
	}
	if b.store != nil {
		_ = b.store.Close()
	}
	if b.host != nil {
		return b.host.Close(ctx)
	}
	return nil
}

func (b *SlackBackend) Serve(ctx context.Context) error {
	if b == nil {
		return fmt.Errorf("slack backend is nil")
	}
	addr := strings.TrimSpace(b.config.ListenAddr)
	if addr == "" {
		addr = ":8080"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/slack/commands", b.handleCommand)
	mux.HandleFunc("/slack/interactivity", b.handleInteractivity)
	mux.HandleFunc("/slack/events", b.handleEvents)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok")) })
	b.server = &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = b.server.Shutdown(shutdownCtx)
	}()
	log.Info().Str("addr", addr).Str("script", b.host.scriptPath).Msg("starting slack backend")
	err := b.server.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (b *SlackBackend) Descriptor(ctx context.Context) (*BotDescriptor, error) {
	desc, err := b.host.Describe(ctx)
	if err != nil {
		return nil, err
	}
	return descriptorFromDescribe(b.host.scriptPath, desc)
}

// SlackManifest returns a Slack app manifest as JSON-compatible data.
func SlackManifest(desc *BotDescriptor, baseURL string) map[string]any {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		base = "https://example.com"
	}
	commands := make([]map[string]any, 0, len(desc.Commands))
	for _, command := range desc.Commands {
		if strings.TrimSpace(command.Name) == "" || command.Type == "user" || command.Type == "message" {
			continue
		}
		entry := map[string]any{
			"command":       "/" + strings.TrimPrefix(command.Name, "/"),
			"url":           base + "/slack/commands",
			"description":   firstNonEmpty(command.Description, desc.Description, "Bot command"),
			"should_escape": false,
		}
		if usageHint := slackUsageHint(command); usageHint != "" {
			entry["usage_hint"] = usageHint
		}
		commands = append(commands, entry)
	}
	sort.Slice(commands, func(i, j int) bool { return fmt.Sprint(commands[i]["command"]) < fmt.Sprint(commands[j]["command"]) })
	name := firstNonEmpty(desc.Name, "discord-bot")
	return map[string]any{
		"display_information": map[string]any{"name": name},
		"features": map[string]any{
			"bot_user":       map[string]any{"display_name": name, "always_online": false},
			"slash_commands": commands,
		},
		"oauth_config": map[string]any{"scopes": map[string]any{"bot": []string{"commands", "chat:write", "app_mentions:read"}}},
		"settings": map[string]any{
			"interactivity":       map[string]any{"is_enabled": true, "request_url": base + "/slack/interactivity"},
			"event_subscriptions": map[string]any{"request_url": base + "/slack/events", "bot_events": []string{"app_mention"}},
			"socket_mode_enabled": false,
		},
	}
}

func SlackManifestJSON(desc *BotDescriptor, baseURL string) ([]byte, error) {
	return json.MarshalIndent(SlackManifest(desc, baseURL), "", "  ")
}

func slackUsageHint(command CommandDescriptor) string {
	options := commandOptionNames(command.Spec)
	if len(options) == 0 {
		return ""
	}
	return options[0] + " text"
}

func commandOptionNames(spec map[string]any) []string {
	options, _ := spec["options"].(map[string]any)
	keys := make([]string, 0, len(options))
	for key := range options {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, preferred := range []string{"prompt", "text", "query", "input", "message"} {
		for i, key := range keys {
			if key == preferred {
				return append([]string{key}, append(keys[:i], keys[i+1:]...)...)
			}
		}
	}
	return keys
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func VerifySlackRequest(signingSecret, timestamp, signature string, body []byte, now time.Time) error {
	if strings.TrimSpace(signingSecret) == "" {
		return fmt.Errorf("slack signing secret is required")
	}
	if strings.TrimSpace(timestamp) == "" || strings.TrimSpace(signature) == "" {
		return fmt.Errorf("missing slack signature headers")
	}
	secs, err := parseInt64(timestamp)
	if err != nil {
		return fmt.Errorf("invalid slack timestamp: %w", err)
	}
	requestTime := time.Unix(secs, 0)
	if now.IsZero() {
		now = time.Now()
	}
	if d := now.Sub(requestTime); d > 5*time.Minute || d < -5*time.Minute {
		return fmt.Errorf("slack request timestamp outside tolerance")
	}
	base := "v0:" + timestamp + ":" + string(body)
	mac := hmac.New(sha256.New, []byte(signingSecret))
	_, _ = mac.Write([]byte(base))
	expected := "v0=" + hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return fmt.Errorf("invalid slack signature")
	}
	return nil
}

func parseInt64(text string) (int64, error) {
	var ret int64
	_, err := fmt.Sscanf(strings.TrimSpace(text), "%d", &ret)
	return ret, err
}

func (b *SlackBackend) readVerifiedBody(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body", http.StatusBadRequest)
		return nil, false
	}
	if err := VerifySlackRequest(b.config.SigningSecret, r.Header.Get("X-Slack-Request-Timestamp"), r.Header.Get("X-Slack-Signature"), body, time.Now()); err != nil {
		log.Warn().Err(err).Msg("rejected slack request")
		http.Error(w, "invalid slack signature", http.StatusUnauthorized)
		return nil, false
	}
	return body, true
}

func (b *SlackBackend) handleCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, ok := b.readVerifiedBody(w, r)
	if !ok {
		return
	}
	values, err := url.ParseQuery(string(body))
	if err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(""))
	go b.dispatchSlashCommand(context.Background(), values, body)
}

func (b *SlackBackend) handleInteractivity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, ok := b.readVerifiedBody(w, r)
	if !ok {
		return
	}
	values, err := url.ParseQuery(string(body))
	if err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	payloadText := values.Get("payload")
	var payload map[string]any
	if err := json.Unmarshal([]byte(payloadText), &payload); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(""))
	go b.dispatchInteractivePayload(context.Background(), payload, []byte(payloadText))
}

func (b *SlackBackend) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, ok := b.readVerifiedBody(w, r)
	if !ok {
		return
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if payload["type"] == "url_verification" {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(fmt.Sprint(payload["challenge"])))
		return
	}
	w.WriteHeader(http.StatusOK)
	go b.dispatchEventPayload(context.Background(), payload, body)
}

func (b *SlackBackend) dispatchSlashCommand(ctx context.Context, values url.Values, raw []byte) {
	commandName := strings.TrimPrefix(values.Get("command"), "/")
	teamID := values.Get("team_id")
	channelID := values.Get("channel_id")
	userID := values.Get("user_id")
	responseURL := values.Get("response_url")
	triggerID := values.Get("trigger_id")
	args := b.slackArgsForCommand(ctx, commandName, values.Get("text"))
	responder := newSlackResponder(b.client, b.store, teamID, responseURL, channelID, "", triggerID, b.host.scriptPath)
	_ = b.store.AddInteraction(SlackInteraction{Kind: "slash_command", TeamID: teamID, ChannelID: channelID, UserID: userID, TriggerID: triggerID, ResponseURL: responseURL, RawPayload: string(raw)})
	req := b.slackBaseRequest(teamID, channelID, userID)
	req.Name = commandName
	req.Args = args
	req.Command = map[string]any{"name": commandName, "type": "slash_command"}
	req.Interaction = InteractionSnapshot{ID: triggerID, Type: "applicationCommand", GuildID: teamID, ChannelID: channelID}
	req = withSlackResponder(req, responder)
	result, err := b.host.handle.DispatchCommand(ctx, req)
	if err != nil {
		_ = responder.Reply(ctx, map[string]any{"content": "command failed: " + err.Error(), "ephemeral": true})
		return
	}
	if !responder.Acknowledged() {
		_ = emitEventResult(ctx, responder.Reply, result)
	}
}

func (b *SlackBackend) slackArgsForCommand(ctx context.Context, commandName, text string) map[string]any {
	desc, err := b.Descriptor(ctx)
	if err != nil {
		return map[string]any{"text": text}
	}
	for _, command := range desc.Commands {
		if command.Name != commandName {
			continue
		}
		options := commandOptionNames(command.Spec)
		if len(options) == 1 {
			return map[string]any{options[0]: strings.TrimSpace(text)}
		}
		if len(options) == 0 {
			return map[string]any{}
		}
		log.Warn().Str("command", commandName).Strs("options", options).Msg("slack backend supports only one option; using first option")
		return map[string]any{options[0]: strings.TrimSpace(text)}
	}
	return map[string]any{"text": strings.TrimSpace(text)}
}

func (b *SlackBackend) dispatchInteractivePayload(ctx context.Context, payload map[string]any, raw []byte) {
	switch fmt.Sprint(payload["type"]) {
	case "block_actions":
		b.dispatchBlockAction(ctx, payload, raw)
	case "view_submission":
		b.dispatchViewSubmission(ctx, payload, raw)
	}
}

func (b *SlackBackend) dispatchBlockAction(ctx context.Context, payload map[string]any, raw []byte) {
	teamID := nestedString(payload, "team", "id")
	channelID := nestedString(payload, "channel", "id")
	userID := nestedString(payload, "user", "id")
	responseURL := fmt.Sprint(payload["response_url"])
	triggerID := fmt.Sprint(payload["trigger_id"])
	message := mapValue(payload["message"])
	messageTS := fmt.Sprint(message["ts"])
	action := firstMap(payload["actions"])
	actionID := fmt.Sprint(action["action_id"])
	values := []string{}
	if v := strings.TrimSpace(fmt.Sprint(action["value"])); v != "" {
		values = append(values, v)
	}
	stored, _ := b.store.GetMessage(teamID, channelID, messageTS)
	content := stored.Content
	if content == "" {
		content = slackTextFromBlocks(message)
	}
	responder := newSlackResponder(b.client, b.store, teamID, responseURL, channelID, messageTS, triggerID, b.host.scriptPath)
	_ = b.store.AddInteraction(SlackInteraction{Kind: "block_actions", TeamID: teamID, ChannelID: channelID, UserID: userID, CallbackID: actionID, ActionID: actionID, TriggerID: triggerID, ResponseURL: responseURL, MessageTS: messageTS, RawPayload: string(raw)})
	req := b.slackBaseRequest(teamID, channelID, userID)
	req.Name = actionID
	req.Values = values
	req.Command = map[string]any{"event": "component"}
	req.Interaction = InteractionSnapshot{ID: triggerID, Type: "component", GuildID: teamID, ChannelID: channelID}
	req.Message = &MessageSnapshot{ID: slackMessageID(teamID, channelID, messageTS), Content: content, GuildID: teamID, ChannelID: channelID}
	req.Component = ComponentSnapshot{CustomID: actionID, Type: "button"}
	req = withSlackResponder(req, responder)
	result, err := b.host.handle.DispatchComponent(ctx, req)
	if err != nil {
		_ = responder.Reply(ctx, map[string]any{"content": "component failed: " + err.Error(), "ephemeral": true})
		return
	}
	if !responder.Acknowledged() {
		_ = emitEventResult(ctx, responder.Reply, result)
	}
}

func (b *SlackBackend) dispatchViewSubmission(ctx context.Context, payload map[string]any, raw []byte) {
	teamID := nestedString(payload, "team", "id")
	userID := nestedString(payload, "user", "id")
	view := mapValue(payload["view"])
	callbackID := fmt.Sprint(view["callback_id"])
	triggerID := fmt.Sprint(payload["trigger_id"])
	metadata := parseJSONMap(fmt.Sprint(view["private_metadata"]))
	channelID := fmt.Sprint(metadata["channel_id"])
	messageTS := fmt.Sprint(metadata["message_ts"])
	responseURL := fmt.Sprint(metadata["response_url"])
	values := modalValuesFromSlack(view)
	stored, _ := b.store.GetMessage(teamID, channelID, messageTS)
	responder := newSlackResponder(b.client, b.store, teamID, responseURL, channelID, messageTS, triggerID, b.host.scriptPath)
	_ = b.store.AddInteraction(SlackInteraction{Kind: "view_submission", TeamID: teamID, ChannelID: channelID, UserID: userID, CallbackID: callbackID, TriggerID: triggerID, ResponseURL: responseURL, MessageTS: messageTS, RawPayload: string(raw)})
	req := b.slackBaseRequest(teamID, channelID, userID)
	req.Name = callbackID
	req.Values = values
	req.Command = map[string]any{"event": "modal"}
	req.Interaction = InteractionSnapshot{ID: triggerID, Type: "modal", GuildID: teamID, ChannelID: channelID}
	req.Message = &MessageSnapshot{ID: slackMessageID(teamID, channelID, messageTS), Content: stored.Content, GuildID: teamID, ChannelID: channelID}
	req.Modal = map[string]any{"customId": callbackID}
	req = withSlackResponder(req, responder)
	result, err := b.host.handle.DispatchModal(ctx, req)
	if err != nil {
		_ = responder.Reply(ctx, map[string]any{"content": "modal failed: " + err.Error(), "ephemeral": true})
		return
	}
	if !responder.Acknowledged() {
		_ = emitEventResult(ctx, responder.Reply, result)
	}
}

func (b *SlackBackend) dispatchEventPayload(ctx context.Context, payload map[string]any, raw []byte) {
	event := mapValue(payload["event"])
	typeName := fmt.Sprint(event["type"])
	if typeName == "app_mention" {
		typeName = "messageCreate"
	}
	teamID := fmt.Sprint(payload["team_id"])
	channelID := fmt.Sprint(event["channel"])
	userID := fmt.Sprint(event["user"])
	_ = b.store.AddInteraction(SlackInteraction{Kind: "event", TeamID: teamID, ChannelID: channelID, UserID: userID, CallbackID: typeName, RawPayload: string(raw)})
	req := b.slackBaseRequest(teamID, channelID, userID)
	req.Name = typeName
	req.Command = map[string]any{"event": typeName}
	req.Interaction = InteractionSnapshot{ID: fmt.Sprint(event["event_ts"]), Type: "event", GuildID: teamID, ChannelID: channelID}
	req.Message = &MessageSnapshot{ID: slackMessageID(teamID, channelID, fmt.Sprint(event["ts"])), Content: fmt.Sprint(event["text"]), GuildID: teamID, ChannelID: channelID}
	responder := newSlackResponder(b.client, b.store, teamID, "", channelID, "", "", b.host.scriptPath)
	req = withSlackResponder(req, responder)
	result, err := b.host.handle.DispatchEvent(ctx, req)
	if err != nil {
		log.Error().Err(err).Str("event", typeName).Msg("slack event dispatch failed")
		return
	}
	if !responder.Acknowledged() {
		_ = emitEventResult(ctx, responder.Reply, result)
	}
}

func (b *SlackBackend) slackBaseRequest(teamID, channelID, userID string) DispatchRequest {
	return DispatchRequest{
		Metadata: map[string]any{"scriptPath": b.host.scriptPath, "platform": "slack"},
		Config:   cloneMap(b.host.runtimeConfig),
		Guild:    map[string]any{"id": teamID},
		Channel:  map[string]any{"id": channelID},
		User:     UserSnapshot{ID: userID, Username: userID},
		Member:   &MemberSnapshot{GuildID: teamID, ID: userID, User: &UserSnapshot{ID: userID, Username: userID}},
		Me:       UserSnapshot{ID: "slack-bot", Username: "slack-bot", Bot: true},
	}
}

func withSlackResponder(req DispatchRequest, r *slackResponder) DispatchRequest {
	req.Reply = r.Reply
	req.FollowUp = r.FollowUp
	req.Edit = r.Edit
	req.Defer = r.Defer
	req.ShowModal = r.ShowModal
	return req
}

// SlackClient wraps the Slack Web API and response_url calls.
type SlackClient struct {
	BotToken   string
	APIBaseURL string
	HTTPClient *http.Client
}

func NewSlackClient(token string) *SlackClient {
	return &SlackClient{BotToken: strings.TrimSpace(token), HTTPClient: http.DefaultClient}
}

func (c *SlackClient) PostMessage(ctx context.Context, payload map[string]any) (map[string]any, error) {
	return c.api(ctx, "chat.postMessage", payload)
}

func (c *SlackClient) UpdateMessage(ctx context.Context, payload map[string]any) (map[string]any, error) {
	return c.api(ctx, "chat.update", payload)
}

func (c *SlackClient) OpenView(ctx context.Context, payload map[string]any) (map[string]any, error) {
	return c.api(ctx, "views.open", payload)
}

func (c *SlackClient) ResponseURL(ctx context.Context, responseURL string, payload map[string]any) (map[string]any, error) {
	if strings.TrimSpace(responseURL) == "" {
		return nil, fmt.Errorf("slack response_url is empty")
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, responseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("slack response_url status %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	return map[string]any{"ok": true}, nil
}

func (c *SlackClient) api(ctx context.Context, method string, payload map[string]any) (map[string]any, error) {
	if strings.TrimSpace(c.BotToken) == "" {
		return nil, fmt.Errorf("slack bot token is required for %s", method)
	}
	body, _ := json.Marshal(payload)
	baseURL := strings.TrimRight(c.APIBaseURL, "/")
	if baseURL == "" {
		baseURL = "https://slack.com/api"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/"+method, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.BotToken)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("slack rate limited; retry after %s", resp.Header.Get("Retry-After"))
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("slack api %s status %d: %s", method, resp.StatusCode, strings.TrimSpace(string(data)))
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	if ok, _ := out["ok"].(bool); !ok {
		return out, fmt.Errorf("slack api %s failed: %s", method, out["error"])
	}
	return out, nil
}

func (c *SlackClient) httpClient() *http.Client {
	if c != nil && c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

// SlackStore keeps adapter state durable across restarts.
type SlackStore struct {
	db *sql.DB
}

type SlackMessage struct {
	ID        string
	TeamID    string
	ChannelID string
	TS        string
	Content   string
}

type SlackInteraction struct {
	Kind        string
	TeamID      string
	ChannelID   string
	UserID      string
	CallbackID  string
	ActionID    string
	TriggerID   string
	ResponseURL string
	MessageTS   string
	RawPayload  string
}

func OpenSlackStore(path string) (*SlackStore, error) {
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create slack state directory: %w", err)
		}
	}
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	store := &SlackStore{db: db}
	if err := store.ensure(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *SlackStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *SlackStore) ensure() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS slack_messages (id TEXT PRIMARY KEY, team_id TEXT NOT NULL, channel_id TEXT NOT NULL, message_ts TEXT NOT NULL, thread_ts TEXT, content TEXT NOT NULL DEFAULT '', response_json TEXT NOT NULL DEFAULT '{}', context_json TEXT NOT NULL DEFAULT '{}', created_at TEXT NOT NULL, updated_at TEXT NOT NULL, UNIQUE(team_id, channel_id, message_ts))`,
		`CREATE TABLE IF NOT EXISTS slack_interactions (id TEXT PRIMARY KEY, team_id TEXT NOT NULL, channel_id TEXT, user_id TEXT, kind TEXT NOT NULL, callback_id TEXT, action_id TEXT, trigger_id TEXT, response_url TEXT, message_ts TEXT, raw_payload_json TEXT NOT NULL, acked_at TEXT, created_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS slack_modal_contexts (id TEXT PRIMARY KEY, team_id TEXT NOT NULL, channel_id TEXT, message_ts TEXT, callback_id TEXT NOT NULL, metadata_json TEXT NOT NULL DEFAULT '{}', created_at TEXT NOT NULL, expires_at TEXT)`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *SlackStore) UpsertMessage(msg SlackMessage, response map[string]any) error {
	if s == nil || s.db == nil || msg.ChannelID == "" || msg.TS == "" {
		return nil
	}
	if msg.ID == "" {
		msg.ID = slackMessageID(msg.TeamID, msg.ChannelID, msg.TS)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	responseJSON, _ := json.Marshal(response)
	_, err := s.db.Exec(`INSERT INTO slack_messages (id, team_id, channel_id, message_ts, content, response_json, created_at, updated_at) VALUES (?,?,?,?,?,?,?,?) ON CONFLICT(team_id, channel_id, message_ts) DO UPDATE SET content=excluded.content, response_json=excluded.response_json, updated_at=excluded.updated_at`, msg.ID, msg.TeamID, msg.ChannelID, msg.TS, msg.Content, string(responseJSON), now, now)
	return err
}

func (s *SlackStore) GetMessage(teamID, channelID, ts string) (SlackMessage, error) {
	if s == nil || s.db == nil || channelID == "" || ts == "" {
		return SlackMessage{}, nil
	}
	row := s.db.QueryRow(`SELECT id, team_id, channel_id, message_ts, content FROM slack_messages WHERE team_id = ? AND channel_id = ? AND message_ts = ? LIMIT 1`, teamID, channelID, ts)
	var msg SlackMessage
	err := row.Scan(&msg.ID, &msg.TeamID, &msg.ChannelID, &msg.TS, &msg.Content)
	if err == sql.ErrNoRows {
		return SlackMessage{}, nil
	}
	return msg, err
}

func (s *SlackStore) AddInteraction(in SlackInteraction) error {
	if s == nil || s.db == nil {
		return nil
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	id := fmt.Sprintf("slackint:%d", time.Now().UnixNano())
	_, err := s.db.Exec(`INSERT INTO slack_interactions (id, team_id, channel_id, user_id, kind, callback_id, action_id, trigger_id, response_url, message_ts, raw_payload_json, created_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`, id, in.TeamID, in.ChannelID, in.UserID, in.Kind, in.CallbackID, in.ActionID, in.TriggerID, in.ResponseURL, in.MessageTS, in.RawPayload, now)
	return err
}

type slackResponder struct {
	client      *SlackClient
	store       *SlackStore
	responseURL string
	channelID   string
	messageTS   string
	triggerID   string
	scriptPath  string
	mu          sync.Mutex
	acked       bool
	teamID      string
}

func newSlackResponder(client *SlackClient, store *SlackStore, teamID, responseURL, channelID, messageTS, triggerID, scriptPath string) *slackResponder {
	return &slackResponder{client: client, store: store, teamID: teamID, responseURL: responseURL, channelID: channelID, messageTS: messageTS, triggerID: triggerID, scriptPath: scriptPath}
}

func (r *slackResponder) Acknowledged() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.acked
}

func (r *slackResponder) markAcked() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.acked {
		return false
	}
	r.acked = true
	return true
}

func (r *slackResponder) Defer(ctx context.Context, payload any) error {
	_ = payload
	if !r.markAcked() {
		return nil
	}
	// HTTP handlers ACK Slack before dispatch. Defer only records JS intent.
	return nil
}

func (r *slackResponder) Reply(ctx context.Context, payload any) error {
	if !r.Acknowledged() {
		r.markAcked()
	}
	message, err := slackMessagePayload(payload)
	if err != nil {
		return err
	}
	if r.channelID != "" {
		message["channel"] = r.channelID
	}
	var resp map[string]any
	if r.messageTS != "" && !boolMap(message, "replace_original") {
		message["ts"] = r.messageTS
		resp, err = r.client.UpdateMessage(ctx, message)
	} else if r.channelID != "" && fmt.Sprint(message["response_type"]) != "ephemeral" {
		// Prefer chat.postMessage for public command responses so Slack returns
		// channel+ts and later ctx.edit calls can update the same message.
		resp, err = r.client.PostMessage(ctx, message)
	} else if r.responseURL != "" {
		if r.messageTS != "" {
			message["replace_original"] = true
		}
		resp, err = r.client.ResponseURL(ctx, r.responseURL, message)
	} else {
		resp, err = r.client.PostMessage(ctx, message)
	}
	if err != nil {
		return err
	}
	r.recordMessage(message, resp)
	return nil
}

func (r *slackResponder) FollowUp(ctx context.Context, payload any) error {
	message, err := slackMessagePayload(payload)
	if err != nil {
		return err
	}
	if r.channelID != "" {
		message["channel"] = r.channelID
	}
	delete(message, "replace_original")
	var resp map[string]any
	if r.responseURL != "" {
		resp, err = r.client.ResponseURL(ctx, r.responseURL, message)
	} else {
		resp, err = r.client.PostMessage(ctx, message)
	}
	if err != nil {
		return err
	}
	r.recordMessage(message, resp)
	return nil
}

func (r *slackResponder) Edit(ctx context.Context, payload any) error {
	if strings.TrimSpace(r.messageTS) == "" {
		return r.Reply(ctx, payload)
	}
	message, err := slackMessagePayload(payload)
	if err != nil {
		return err
	}
	message["channel"] = r.channelID
	message["ts"] = r.messageTS
	resp, err := r.client.UpdateMessage(ctx, message)
	if err != nil {
		return err
	}
	r.recordMessage(message, resp)
	return nil
}

func (r *slackResponder) ShowModal(ctx context.Context, payload any) error {
	mapping, _ := payload.(map[string]any)
	view, err := slackViewPayload(mapping, map[string]any{"channel_id": r.channelID, "message_ts": r.messageTS, "response_url": r.responseURL})
	if err != nil {
		return err
	}
	_, err = r.client.OpenView(ctx, map[string]any{"trigger_id": r.triggerID, "view": view})
	if err == nil {
		r.markAcked()
	}
	return err
}

func (r *slackResponder) recordMessage(message map[string]any, resp map[string]any) {
	if r == nil || r.store == nil {
		return
	}
	channel := firstNonEmpty(fmt.Sprint(resp["channel"]), fmt.Sprint(message["channel"]), r.channelID)
	ts := firstNonEmpty(fmt.Sprint(resp["ts"]), r.messageTS)
	if channel == "" || ts == "" {
		return
	}
	r.channelID = channel
	r.messageTS = ts
	_ = r.store.UpsertMessage(SlackMessage{TeamID: r.teamID, ChannelID: channel, TS: ts, Content: fmt.Sprint(message["text"])}, resp)
}

func slackMessageID(teamID, channelID, ts string) string {
	return "slack:" + teamID + ":" + channelID + ":" + ts
}

func slackMessagePayload(payload any) (map[string]any, error) {
	normalized, err := normalizePayload(payload)
	if err != nil {
		return nil, err
	}
	content := inlineFiles(normalized.Content, normalized.Files)
	blocks := []map[string]any{}
	if strings.TrimSpace(content) != "" {
		blocks = append(blocks, map[string]any{"type": "section", "text": map[string]any{"type": "mrkdwn", "text": truncateSlackText(content, 2900)}})
	}
	for i, component := range normalized.Components {
		if row, ok := component.(discordgo.ActionsRow); ok {
			elements := []map[string]any{}
			for _, child := range row.Components {
				if button, ok := child.(discordgo.Button); ok {
					elements = append(elements, slackButton(button))
				}
			}
			if len(elements) > 0 {
				blocks = append(blocks, map[string]any{"type": "actions", "block_id": fmt.Sprintf("row_%d", i), "elements": elements})
			}
		}
	}
	ret := map[string]any{"text": content}
	if normalized.Ephemeral {
		ret["response_type"] = "ephemeral"
	} else {
		ret["response_type"] = "in_channel"
	}
	if len(blocks) > 0 {
		ret["blocks"] = blocks
	}
	return ret, nil
}

func slackButton(button discordgo.Button) map[string]any {
	ret := map[string]any{"type": "button", "action_id": button.CustomID, "value": firstNonEmpty(button.CustomID, button.URL), "text": map[string]any{"type": "plain_text", "text": truncateSlackPlain(button.Label, 75), "emoji": true}}
	switch button.Style {
	case discordgo.PrimaryButton, discordgo.SuccessButton:
		ret["style"] = "primary"
	case discordgo.DangerButton:
		ret["style"] = "danger"
	}
	if button.Disabled {
		ret["disabled"] = true
	}
	if button.URL != "" {
		ret["url"] = button.URL
		delete(ret, "value")
	}
	return ret
}

func slackViewPayload(mapping map[string]any, metadata map[string]any) (map[string]any, error) {
	if len(mapping) == 0 {
		return nil, fmt.Errorf("modal payload must be an object")
	}
	customID := strings.TrimSpace(fmt.Sprint(mapping["customId"]))
	title := truncateSlackPlain(fmt.Sprint(mapping["title"]), 24)
	if customID == "" || title == "" {
		return nil, fmt.Errorf("modal payload missing customId/title")
	}
	components, err := normalizeComponents(mapping["components"])
	if err != nil {
		return nil, err
	}
	blocks := []map[string]any{}
	for i, component := range components {
		row, ok := component.(discordgo.ActionsRow)
		if !ok {
			continue
		}
		for _, child := range row.Components {
			if input, ok := child.(discordgo.TextInput); ok {
				element := map[string]any{"type": "plain_text_input", "action_id": input.CustomID, "multiline": input.Style == discordgo.TextInputParagraph, "placeholder": map[string]any{"type": "plain_text", "text": truncateSlackPlain(input.Placeholder, 150)}}
				if input.MinLength > 0 {
					element["min_length"] = input.MinLength
				}
				if input.MaxLength > 0 {
					element["max_length"] = input.MaxLength
				}
				blocks = append(blocks, map[string]any{"type": "input", "block_id": fmt.Sprintf("%s_block_%d", input.CustomID, i), "optional": !input.Required, "label": map[string]any{"type": "plain_text", "text": truncateSlackPlain(input.Label, 200)}, "element": element})
			}
		}
	}
	metadataJSON, _ := json.Marshal(metadata)
	return map[string]any{"type": "modal", "callback_id": customID, "title": map[string]any{"type": "plain_text", "text": title}, "submit": map[string]any{"type": "plain_text", "text": "Submit"}, "close": map[string]any{"type": "plain_text", "text": "Cancel"}, "private_metadata": string(metadataJSON), "blocks": blocks}, nil
}

func inlineFiles(content string, files []*discordgo.File) string {
	parts := []string{content}
	for _, file := range files {
		if file == nil {
			continue
		}
		var body string
		if file.Reader != nil {
			data, _ := io.ReadAll(file.Reader)
			body = string(data)
		}
		if len(body) > 3500 {
			body = body[:3500] + "\n... truncated ..."
		}
		parts = append(parts, fmt.Sprintf("Export: %s (%s)\n```\n%s\n```", file.Name, file.ContentType, body))
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}

func truncateSlackText(text string, max int) string {
	if len(text) <= max {
		return text
	}
	return text[:max-20] + "\n… truncated …"
}
func truncateSlackPlain(text string, max int) string {
	text = strings.TrimSpace(text)
	if text == "" {
		text = " "
	}
	if len(text) <= max {
		return text
	}
	return text[:max-1] + "…"
}
func boolMap(m map[string]any, key string) bool { v, _ := m[key].(bool); return v }

func nestedString(m map[string]any, parent, child string) string {
	return fmt.Sprint(mapValue(m[parent])[child])
}
func mapValue(v any) map[string]any {
	m, _ := v.(map[string]any)
	if m == nil {
		return map[string]any{}
	}
	return m
}
func firstMap(v any) map[string]any {
	arr, _ := v.([]any)
	if len(arr) == 0 {
		return map[string]any{}
	}
	return mapValue(arr[0])
}
func parseJSONMap(text string) map[string]any {
	var out map[string]any
	_ = json.Unmarshal([]byte(text), &out)
	if out == nil {
		return map[string]any{}
	}
	return out
}

func modalValuesFromSlack(view map[string]any) map[string]any {
	ret := map[string]any{}
	state := mapValue(view["state"])
	values := mapValue(state["values"])
	for _, rawBlock := range values {
		block := mapValue(rawBlock)
		for actionID, rawAction := range block {
			action := mapValue(rawAction)
			ret[actionID] = fmt.Sprint(action["value"])
		}
	}
	return ret
}

func slackTextFromBlocks(message map[string]any) string {
	if text := strings.TrimSpace(fmt.Sprint(message["text"])); text != "" {
		return text
	}
	blocks, _ := message["blocks"].([]any)
	parts := []string{}
	for _, raw := range blocks {
		block := mapValue(raw)
		textObj := mapValue(block["text"])
		if text := strings.TrimSpace(fmt.Sprint(textObj["text"])); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, "\n")
}
