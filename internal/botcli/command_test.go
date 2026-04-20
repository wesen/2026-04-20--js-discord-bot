package botcli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListCommandOutputsNamedBots(t *testing.T) {
	root := NewCommand()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)
	root.SetArgs([]string{"list", "--bot-repository", examplesFixtureDir(t)})

	err := root.Execute()
	require.NoError(t, err)
	output := stdout.String()
	require.Contains(t, output, "ping")
	require.Contains(t, output, "knowledge-base")
	require.Contains(t, output, "support")
	require.Contains(t, output, "moderation")
	require.Contains(t, output, "poker")
	require.Contains(t, output, "announcements")
}

func TestHelpCommandShowsBotCommandsAndEvents(t *testing.T) {
	root := NewCommand()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)
	root.SetArgs([]string{"help", "knowledge-base", "--bot-repository", examplesFixtureDir(t)})

	err := root.Execute()
	require.NoError(t, err)
	output := stdout.String()
	require.Contains(t, output, "Bot: knowledge-base")
	require.Contains(t, output, "kb-search")
	require.Contains(t, output, "messageCreate")
}

func TestHelpCommandShowsPingHelpCommand(t *testing.T) {
	root := NewCommand()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)
	root.SetArgs([]string{"help", "ping", "--bot-repository", examplesFixtureDir(t)})

	err := root.Execute()
	require.NoError(t, err)
	output := stdout.String()
	require.Contains(t, output, "Bot: ping")
	require.Contains(t, output, "ping")
	require.Contains(t, output, "feedback")
}

func TestHelpCommandShowsPokerHelpCommand(t *testing.T) {
	root := NewCommand()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)
	root.SetArgs([]string{"help", "poker", "--bot-repository", examplesFixtureDir(t)})

	err := root.Execute()
	require.NoError(t, err)
	output := stdout.String()
	require.Contains(t, output, "Bot: poker")
	require.Contains(t, output, "poker-help")
	require.Contains(t, output, "poker-action")
}

func TestRunCommandResolvesSingleNamedBot(t *testing.T) {
	root := NewCommand()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)

	var captured RunRequest
	previous := runSelectedBotsFn
	runSelectedBotsFn = func(ctx context.Context, request RunRequest) error {
		captured = request
		return nil
	}
	defer func() { runSelectedBotsFn = previous }()

	root.SetArgs([]string{
		"run", "knowledge-base",
		"--bot-repository", examplesFixtureDir(t),
		"--bot-token", "test-token",
		"--application-id", "test-app",
		"--guild-id", "123",
		"--sync-on-start",
	})

	err := root.Execute()
	require.NoError(t, err)
	require.Equal(t, "knowledge-base", captured.Bot.Name())
	require.True(t, captured.SyncOnStart)
	require.Equal(t, "test-token", captured.Config.BotToken)
	require.Equal(t, "test-app", captured.Config.ApplicationID)
	require.Equal(t, "123", captured.Config.GuildID)
}

func TestRunCommandPrintsParsedValues(t *testing.T) {
	root := NewCommand()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)
	root.SetArgs([]string{
		"run", "knowledge-base",
		"--bot-repository", examplesFixtureDir(t),
		"--bot-token", "test-token",
		"--application-id", "test-app",
		"--guild-id", "123",
		"--print-parsed-values",
	})

	err := root.Execute()
	require.NoError(t, err)
	output := stdout.String()
	require.Contains(t, output, "\"botToken\": \"test…oken\"")
	require.Contains(t, output, "\"name\": \"knowledge-base\"")
	require.Contains(t, output, "\"commands\": [")
}

func TestRunCommandParsesBotRuntimeConfig(t *testing.T) {
	repo := runConfigFixtureDir(t)
	root := NewCommand()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)

	var captured RunRequest
	previous := runSelectedBotsFn
	runSelectedBotsFn = func(ctx context.Context, request RunRequest) error {
		captured = request
		return nil
	}
	defer func() { runSelectedBotsFn = previous }()

	root.SetArgs([]string{
		"run", "config-bot",
		"--bot-repository", repo,
		"--bot-token", "test-token",
		"--application-id", "test-app",
		"--index-path", "./docs",
		"--read-only",
	})

	err := root.Execute()
	require.NoError(t, err)
	require.Equal(t, "./docs", captured.RuntimeConfig["indexPath"])
	require.Equal(t, true, captured.RuntimeConfig["readOnly"])
}

func TestHelpCommandShowsRunConfig(t *testing.T) {
	repo := runConfigFixtureDir(t)
	root := NewCommand()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)
	root.SetArgs([]string{"help", "config-bot", "--bot-repository", repo})

	err := root.Execute()
	require.NoError(t, err)
	output := stdout.String()
	require.Contains(t, output, "Run config:")
	require.Contains(t, output, "indexPath (--index-path)")
	require.Contains(t, output, "readOnly (--read-only)")
}

func TestRunCommandRejectsMultipleSelectedBots(t *testing.T) {
	root := NewCommand()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)
	root.SetArgs([]string{
		"run", "knowledge-base", "support",
		"--bot-repository", examplesFixtureDir(t),
		"--bot-token", "test-token",
		"--application-id", "test-app",
	})

	err := root.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "accepts exactly one bot selector")
}

func TestListCommandRejectsDuplicateBotNames(t *testing.T) {
	root := NewCommand()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)
	root.SetArgs([]string{"list", "--bot-repository", duplicateNameFixtureDir(t, "a"), "--bot-repository", duplicateNameFixtureDir(t, "b")})

	err := root.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate bot name")
}

func examplesFixtureDir(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "examples", "discord-bots")
}

func duplicateNameFixtureDir(t *testing.T, suffix string) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "testdata", "discord-bots-dupe-name-"+suffix)
}

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	return root
}

func runConfigFixtureDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	source := `const { defineBot } = require("discord");
module.exports = defineBot(({ command, configure }) => {
  configure({
    name: "config-bot",
    description: "Bot with runtime config",
    run: {
      fields: {
        indexPath: { type: "string", help: "Path to the docs index" },
        readOnly: { type: "bool", help: "Disable writes" }
      }
    }
  });
  command("ping", async (ctx) => ({ content: String(ctx.config && ctx.config.indexPath || "") }))
});`
	path := filepath.Join(dir, "index.js")
	require.NoError(t, os.WriteFile(path, []byte(source), 0o644))
	return dir
}
