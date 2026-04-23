package framework

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRequiresScript(t *testing.T) {
	_, err := New(
		WithCredentials(Credentials{
			BotToken:      "token",
			ApplicationID: "app-id",
		}),
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "framework script path is required")
}

func TestNewLoadsCredentialsFromEnv(t *testing.T) {
	t.Setenv("DISCORD_BOT_TOKEN", "token-from-env")
	t.Setenv("DISCORD_APPLICATION_ID", "app-from-env")

	bot, err := New(
		WithCredentialsFromEnv(),
		WithScript(exampleBotScript(t)),
	)
	require.NoError(t, err)
	require.NotNil(t, bot)
	require.NoError(t, bot.Close())
}

func TestNewSupportsExplicitCredentialsAndRuntimeConfig(t *testing.T) {
	bot, err := New(
		WithCredentials(Credentials{
			BotToken:      "token",
			ApplicationID: "app-id",
			GuildID:       "guild-id",
		}),
		WithScript(exampleBotScript(t)),
		WithRuntimeConfig(map[string]any{"db_path": "./data/test.sqlite", "feature_enabled": true}),
		WithSyncOnStart(true),
	)
	require.NoError(t, err)
	require.NotNil(t, bot)
	require.True(t, bot.cfg.SyncOnStart)
	require.Equal(t, "./data/test.sqlite", bot.cfg.RuntimeConfig["db_path"])
	require.Equal(t, true, bot.cfg.RuntimeConfig["feature_enabled"])
	require.NoError(t, bot.Close())
}

func exampleBotScript(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	return filepath.Join(root, "examples", "discord-bots", "unified-demo", "index.js")
}
