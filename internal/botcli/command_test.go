package botcli

import (
	"bytes"
	"context"
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
	require.Contains(t, output, "knowledge-base")
	require.Contains(t, output, "support")
	require.Contains(t, output, "moderation")
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

func TestRunCommandResolvesMultipleNamedBots(t *testing.T) {
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
		"run", "knowledge-base", "support",
		"--bot-repository", examplesFixtureDir(t),
		"--bot-token", "test-token",
		"--application-id", "test-app",
		"--guild-id", "123",
		"--sync-on-start",
	})

	err := root.Execute()
	require.NoError(t, err)
	require.Len(t, captured.Bots, 2)
	require.Equal(t, "knowledge-base", captured.Bots[0].Name())
	require.Equal(t, "support", captured.Bots[1].Name())
	require.True(t, captured.SyncOnStart)
	require.Equal(t, "test-token", captured.Config.BotToken)
	require.Equal(t, "test-app", captured.Config.ApplicationID)
	require.Equal(t, "123", captured.Config.GuildID)
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
