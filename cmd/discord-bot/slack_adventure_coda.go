package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"

	"github.com/go-go-golems/discord-bot/internal/jsdiscord"
)

type adventureCodaSession struct {
	ID             string
	ChannelID      string
	Turn           int
	CurrentSceneID string
	Title          string
	AsciiArt       string
	Narration      string
	StatsJSON      string
	InventoryJSON  string
	RawPatchJSON   string
}

type adventureCodaMessage struct {
	TeamID    string
	ChannelID string
	TS        string
	Content   string
}

func newSlackAdventureCodaBackfillCommand() *cobra.Command {
	var adventureDB string
	var slackStateDB string
	var botToken string
	var apply bool
	cmd := &cobra.Command{
		Use:   "slack-adventure-coda-backfill",
		Short: "Replace old Slack adventure export messages with coda/lookback messages",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(botToken) == "" {
				botToken = os.Getenv("SLACK_BOT_TOKEN")
			}
			return runSlackAdventureCodaBackfill(cmd.Context(), adventureDB, slackStateDB, botToken, apply)
		},
	}
	cmd.Flags().StringVar(&adventureDB, "adventure-db", "./examples/discord-bots/adventure/data/adventure.sqlite", "Adventure SQLite database path")
	cmd.Flags().StringVar(&slackStateDB, "slack-state-db", "./var/slack-adventure.sqlite", "Slack adapter SQLite database path")
	cmd.Flags().StringVar(&botToken, "slack-bot-token", "", "Slack bot token; defaults to SLACK_BOT_TOKEN")
	cmd.Flags().BoolVar(&apply, "apply", false, "Actually update Slack messages; default is dry-run")
	return cmd
}

func runSlackAdventureCodaBackfill(ctx context.Context, adventureDBPath, slackDBPath, botToken string, apply bool) error {
	adventureDB, err := sql.Open("sqlite3", adventureDBPath)
	if err != nil {
		return err
	}
	defer adventureDB.Close()
	slackDB, err := sql.Open("sqlite3", slackDBPath)
	if err != nil {
		return err
	}
	defer slackDB.Close()

	sessions, err := completedAdventureSessions(adventureDB)
	if err != nil {
		return err
	}
	client := jsdiscord.NewSlackClient(botToken)
	updated := 0
	for _, session := range sessions {
		messages, err := matchingSlackExportMessages(slackDB, session)
		if err != nil {
			return err
		}
		if len(messages) == 0 {
			continue
		}
		lookback, err := adventureLookback(adventureDB, session.ID)
		if err != nil {
			return err
		}
		text := renderAdventureCoda(session, lookback)
		for _, message := range messages {
			fmt.Printf("%s session=%s channel=%s ts=%s title=%q\n", dryRunLabel(apply), session.ID, message.ChannelID, message.TS, session.Title)
			if !apply {
				continue
			}
			payload := slackCodaPayload(message.ChannelID, message.TS, text, session.Turn)
			if _, err := client.UpdateMessage(ctx, payload); err != nil {
				return fmt.Errorf("update slack message channel=%s ts=%s: %w", message.ChannelID, message.TS, err)
			}
			if err := updateStoredSlackMessage(slackDB, message, text); err != nil {
				return err
			}
			updated++
		}
	}
	fmt.Printf("matched_sessions=%d updated_messages=%d apply=%v\n", len(sessions), updated, apply)
	return nil
}

func firstNonEmptyLocal(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func dryRunLabel(apply bool) string {
	if apply {
		return "UPDATE"
	}
	return "DRY-RUN"
}

func completedAdventureSessions(db *sql.DB) ([]adventureCodaSession, error) {
	rows, err := db.Query(`SELECT s.id, s.channel_id, s.turn, s.current_scene_id, sc.title, sc.ascii_art, sc.narration, sc.stats_json, sc.inventory_json, sc.raw_patch_json
FROM adventure_sessions s
JOIN adventure_scenes sc ON sc.id = s.current_scene_id
WHERE s.status = 'completed'
ORDER BY s.updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ret []adventureCodaSession
	for rows.Next() {
		var session adventureCodaSession
		if err := rows.Scan(&session.ID, &session.ChannelID, &session.Turn, &session.CurrentSceneID, &session.Title, &session.AsciiArt, &session.Narration, &session.StatsJSON, &session.InventoryJSON, &session.RawPatchJSON); err != nil {
			return nil, err
		}
		ret = append(ret, session)
	}
	return ret, rows.Err()
}

func matchingSlackExportMessages(db *sql.DB, session adventureCodaSession) ([]adventureCodaMessage, error) {
	patterns := []string{"%adventure-" + session.ID + ".json%", "%\"id\": \"" + session.ID + "\"%", "%" + session.ID + "%"}
	seen := map[string]bool{}
	var ret []adventureCodaMessage
	for _, pattern := range patterns {
		rows, err := db.Query(`SELECT team_id, channel_id, message_ts, content FROM slack_messages WHERE channel_id = ? AND content LIKE ? ORDER BY updated_at DESC`, session.ChannelID, pattern)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var message adventureCodaMessage
			if err := rows.Scan(&message.TeamID, &message.ChannelID, &message.TS, &message.Content); err != nil {
				_ = rows.Close()
				return nil, err
			}
			key := message.ChannelID + ":" + message.TS
			if !seen[key] {
				seen[key] = true
				ret = append(ret, message)
			}
		}
		if err := rows.Close(); err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func adventureLookback(db *sql.DB, sessionID string) ([]string, error) {
	rows, err := db.Query(`SELECT turn, title FROM adventure_scenes WHERE session_id = ? ORDER BY turn ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ret []string
	for rows.Next() {
		var turn int
		var title string
		if err := rows.Scan(&turn, &title); err != nil {
			return nil, err
		}
		ret = append(ret, fmt.Sprintf("Turn %d: %s", turn, firstNonEmptyLocal(title, "Untitled")))
	}
	return ret, rows.Err()
}

func renderAdventureCoda(session adventureCodaSession, lookback []string) string {
	summary := endingSummary(session.RawPatchJSON)
	if summary == "" {
		summary = "The adventure has ended."
	}
	body := strings.TrimSpace(strings.Join([]string{session.AsciiArt, session.Narration}, "\n\n"))
	stats := statLineFromJSON(session.StatsJSON, session.InventoryJSON)
	return truncateText(strings.Join([]string{
		"```",
		"╔═ " + firstNonEmptyLocal(session.Title, "Adventure"),
		fmt.Sprintf("Turn %d", session.Turn),
		"",
		truncateText(body, 1000),
		"",
		stats,
		"```",
		"",
		"*Coda*",
		truncateText(summary, 500),
		"",
		"*Look back*",
		truncateText(strings.Join(lookback, "\n"), 700),
		"",
		"Use the navigation buttons to scroll through previous scenes.",
	}, "\n"), 2900)
}

func slackCodaPayload(channelID, ts, text string, turn int) map[string]any {
	blocks := []map[string]any{{"type": "section", "text": map[string]any{"type": "mrkdwn", "text": truncateText(text, 2900)}}}
	if turn > 0 {
		blocks = append(blocks, map[string]any{"type": "actions", "block_id": "row_0", "elements": []map[string]any{{"type": "button", "action_id": "adv:history:prev", "value": "adv:history:prev", "text": map[string]any{"type": "plain_text", "text": "← Previous", "emoji": true}}}})
	}
	return map[string]any{"channel": channelID, "ts": ts, "text": text, "blocks": blocks}
}

func updateStoredSlackMessage(db *sql.DB, message adventureCodaMessage, content string) error {
	_, err := db.Exec(`UPDATE slack_messages SET content = ?, updated_at = datetime('now') WHERE channel_id = ? AND message_ts = ?`, content, message.ChannelID, message.TS)
	return err
}

func endingSummary(raw string) string {
	var root map[string]any
	if err := json.Unmarshal([]byte(raw), &root); err != nil {
		return ""
	}
	patch, _ := root["scene_patch"].(map[string]any)
	ending, _ := patch["ending"].(map[string]any)
	return strings.TrimSpace(fmt.Sprint(ending["summary"]))
}

func statLineFromJSON(statsJSON, inventoryJSON string) string {
	var stats map[string]any
	_ = json.Unmarshal([]byte(statsJSON), &stats)
	keys := make([]string, 0, len(stats))
	for key := range stats {
		keys = append(keys, key)
	}
	sortStrings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, strings.ToUpper(key)+": "+fmt.Sprint(stats[key]))
	}
	var inventory []any
	_ = json.Unmarshal([]byte(inventoryJSON), &inventory)
	items := make([]string, 0, len(inventory))
	for _, item := range inventory {
		items = append(items, fmt.Sprint(item))
	}
	if len(items) == 0 {
		items = append(items, "empty")
	}
	return strings.Join(parts, "  ") + "\nInventory: " + strings.Join(items, ", ")
}

func sortStrings(values []string) {
	for i := 0; i < len(values); i++ {
		for j := i + 1; j < len(values); j++ {
			if values[j] < values[i] {
				values[i], values[j] = values[j], values[i]
			}
		}
	}
}

func truncateText(text string, max int) string {
	if len(text) <= max {
		return text
	}
	if max < 20 {
		return text[:max]
	}
	return text[:max-16] + "\n… truncated …"
}
