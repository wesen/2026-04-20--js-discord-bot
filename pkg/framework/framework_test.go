package framework

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/dop251/goja"
	noderequire "github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/go-go-goja/engine"
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

func TestNewFailsWhenScriptNeedsMissingCustomModule(t *testing.T) {
	_, err := New(
		WithCredentials(Credentials{
			BotToken:      "token",
			ApplicationID: "app-id",
		}),
		WithScript(writeFrameworkBotScript(t, `
const { defineBot } = require("discord");
const app = require("app");
module.exports = defineBot(({ configure, command }) => {
  configure({
    name: "custom-module-bot",
    description: app.description(),
  });
  command("custom-ping", {
    description: app.commandDescription(),
  }, async () => ({ content: app.greeting() }));
});
`)),
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Invalid module")
}

func TestNewSupportsCustomRuntimeModuleRegistrars(t *testing.T) {
	bot, err := New(
		WithCredentials(Credentials{
			BotToken:      "token",
			ApplicationID: "app-id",
		}),
		WithScript(writeFrameworkBotScript(t, `
const { defineBot } = require("discord");
const app = require("app");
module.exports = defineBot(({ configure, command }) => {
  configure({
    name: "custom-module-bot",
    description: app.description(),
  });
  command("custom-ping", {
    description: app.commandDescription(),
  }, async () => ({ content: app.greeting() }));
});
`)),
		WithRuntimeModuleRegistrars(testAppRegistrar{}),
	)
	require.NoError(t, err)
	require.NotNil(t, bot)
	require.NoError(t, bot.Close())
}

func exampleBotScript(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	return filepath.Join(root, "examples", "discord-bots", "unified-demo", "index.js")
}

func writeFrameworkBotScript(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "index.js")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

type testAppRegistrar struct{}

func (testAppRegistrar) ID() string {
	return "test-app-registrar"
}

func (testAppRegistrar) RegisterRuntimeModules(_ *engine.RuntimeModuleContext, reg *noderequire.Registry) error {
	reg.RegisterNativeModule("app", func(vm *goja.Runtime, moduleObj *goja.Object) {
		exports := moduleObj.Get("exports").(*goja.Object)
		_ = exports.Set("description", func() string { return "Bot using a custom runtime module" })
		_ = exports.Set("commandDescription", func() string { return "Ping through the custom Go module" })
		_ = exports.Set("greeting", func() string { return fmt.Sprintf("hello from %s", "app") })
	})
	return nil
}
