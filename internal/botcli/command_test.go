package botcli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	glazed_cli "github.com/go-go-golems/glazed/pkg/cli"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	appconfig "github.com/manuel/wesen/2026-04-20--js-discord-bot/internal/config"
	"github.com/manuel/wesen/2026-04-20--js-discord-bot/internal/jsdiscord"
)

func TestListCommandOutputsNamedBots(t *testing.T) {
	root := NewCommand(Bootstrap{Repositories: []Repository{repoFromDir(examplesFixtureDir(t), "examples")}})
	output, err := executeCaptured(t, root, []string{"list", "--output", "json"})
	require.NoError(t, err)
	require.Contains(t, output, `"name": "ping"`)
	require.Contains(t, output, `"name": "knowledge-base"`)
	require.Contains(t, output, `"name": "support"`)
	require.Contains(t, output, `"name": "moderation"`)
	require.Contains(t, output, `"name": "poker"`)
	require.Contains(t, output, `"name": "announcements"`)
}

func TestHelpCommandShowsBotCommandsAndEvents(t *testing.T) {
	root := NewCommand(Bootstrap{Repositories: []Repository{repoFromDir(examplesFixtureDir(t), "examples")}})
	output, err := executeCaptured(t, root, []string{"help", "knowledge-base", "--output", "json"})
	require.NoError(t, err)
	require.Contains(t, output, `"kind": "command"`)
	require.Contains(t, output, `"name": "kb-search"`)
	require.Contains(t, output, `"kind": "event"`)
	require.Contains(t, output, `"name": "messageCreate"`)
}

func TestScannedStatusVerbRegistersUnderParent(t *testing.T) {
	root := NewCommand(Bootstrap{Repositories: []Repository{repoFromDir(scannerFixtureDir(t), "scanner")}})
	output, err := executeCaptured(t, root, []string{"demo-bot", "status", "--output", "json"})
	require.NoError(t, err)
	require.Contains(t, output, `"active": true`)
	require.Contains(t, output, `"commands": 1`)
}

func TestScannedRunVerbHelpShowsFields(t *testing.T) {
	root := NewCommand(Bootstrap{Repositories: []Repository{repoFromDir(scannerFixtureDir(t), "scanner")}})
	output, err := executeCaptured(t, root, []string{"demo-bot", "run", "--help"})
	require.NoError(t, err)
	require.Contains(t, output, "--bot-token")
	require.Contains(t, output, "--api-key")
	require.Contains(t, output, "Run the demo bot")
}

func TestNewBotsCommandFailsWithoutRequiredCustomRuntimeModule(t *testing.T) {
	repo := writeBotCLIRepoBot(t, `
const { defineBot } = require("discord");
const app = require("app");
module.exports = defineBot(({ configure }) => {
  configure({ name: "custom-module-bot", description: app.description() });
});
function status() { return { active: true, module: app.name() }; }
__verb__("status", { output: "glaze", short: "Show custom module status" });
`)

	_, err := NewBotsCommand(Bootstrap{Repositories: []Repository{repo}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "Invalid module")
}

func TestNewBotsCommandSupportsCustomRuntimeModuleRegistrars(t *testing.T) {
	repo := writeBotCLIRepoBot(t, `
const { defineBot } = require("discord");
const app = require("app");
module.exports = defineBot(({ configure }) => {
  configure({ name: "custom-module-bot", description: app.description() });
});
function status() { return { active: true, module: app.name() }; }
__verb__("status", { output: "glaze", short: "Show custom module status" });
`)

	root := NewCommand(
		Bootstrap{Repositories: []Repository{repo}},
		WithRuntimeModuleRegistrars(testAppRegistrar{}),
	)
	output, err := executeCaptured(t, root, []string{"custom-module-bot", "status", "--output", "json"})
	require.NoError(t, err)
	require.Contains(t, output, `"active": true`)
	require.Contains(t, output, `"module": "app"`)
}

func TestUnifiedDemoHelpShowsCommandAndEventRows(t *testing.T) {
	root := NewCommand(Bootstrap{Repositories: []Repository{repoFromDir(examplesFixtureDir(t), "examples")}})
	output, err := executeCaptured(t, root, []string{"help", "unified-demo", "--output", "json"})
	require.NoError(t, err)
	require.Contains(t, output, `"name": "unified-ping"`)
	require.Contains(t, output, `"kind": "event"`)
	require.Contains(t, output, `"name": "ready"`)
}

func TestUnifiedDemoRunHelpShowsConfigFlags(t *testing.T) {
	root := NewCommand(Bootstrap{Repositories: []Repository{repoFromDir(examplesFixtureDir(t), "examples")}})
	output, err := executeCaptured(t, root, []string{"unified-demo", "run", "--help"})
	require.NoError(t, err)
	require.Contains(t, output, "--bot-token")
	require.Contains(t, output, "--application-id")
	require.Contains(t, output, "--db-path")
	require.Contains(t, output, "--api-key")
}

func TestKnowledgeBaseRunHelpShowsMigratedJsverbsFields(t *testing.T) {
	root := NewCommand(Bootstrap{Repositories: []Repository{repoFromDir(examplesFixtureDir(t), "examples")}})
	output, err := executeCaptured(t, root, []string{"knowledge-base", "run", "--help"})
	require.NoError(t, err)
	require.Contains(t, output, "--db-path")
	require.Contains(t, output, "--capture-enabled")
	require.Contains(t, output, "--capture-threshold")
	require.Contains(t, output, "--review-limit")
	require.Contains(t, output, "--trusted-reviewer-role-ids")
	require.Contains(t, output, "--sync-on-start")
}

func TestLegacyRunSyntaxWorksForUiShowcase(t *testing.T) {
	root := NewCommand(Bootstrap{Repositories: []Repository{repoFromDir(examplesFixtureDir(t), "examples")}})
	output, err := executeCaptured(t, root, []string{"run", "ui-showcase", "--help"})
	require.NoError(t, err)
	require.Contains(t, output, "bots run ui-showcase")
	require.Contains(t, output, "--bot-token")
	require.Contains(t, output, "--application-id")
	require.Contains(t, output, "--sync-on-start")
}

func TestUiShowcaseParentDoesNotExposeLeakedHelperFunctions(t *testing.T) {
	root := NewCommand(Bootstrap{Repositories: []Repository{repoFromDir(examplesFixtureDir(t), "examples")}})
	output, err := executeCaptured(t, root, []string{"ui-showcase"})
	require.NoError(t, err)
	require.Contains(t, output, "run         Run the ui-showcase Discord bot")
	require.NotContains(t, output, "first-value")
}

func TestLegacyRunSyntaxLoadsDiscordEnvVarsViaGlazedMiddleware(t *testing.T) {
	t.Setenv("DISCORD_BOT_TOKEN", "token-from-env")
	t.Setenv("DISCORD_APPLICATION_ID", "app-from-env")
	desc := buildSyntheticBotRunDescription(DiscoveredBot{Descriptor: &jsdiscord.BotDescriptor{Name: "ui-showcase"}}, "ui-showcase")
	parserConfig := botCLIParserConfig(defaultCommandOptions().appName)
	parser, err := glazed_cli.NewCobraParserFromSections(desc.Schema.Clone(), &glazed_cli.CobraParserConfig{
		AppName:           parserConfig.AppName,
		ShortHelpSections: parserConfig.ShortHelpSections,
	})
	require.NoError(t, err)
	cmd := glazed_cli.NewCobraCommandFromCommandDescription(desc)
	require.NoError(t, parser.AddToCobraCommand(cmd))
	require.NoError(t, cmd.ParseFlags(nil))

	parsed, err := parser.Parse(cmd, nil)
	require.NoError(t, err)
	cfg, err := appconfig.FromValues(parsed)
	require.NoError(t, err)
	require.Equal(t, "token-from-env", cfg.BotToken)
	require.Equal(t, "app-from-env", cfg.ApplicationID)
}

func TestLegacyRunSyntaxLoadsCustomAppEnvVarsViaGlazedMiddleware(t *testing.T) {
	t.Setenv("WEZEN_BOT_TOKEN", "token-from-custom-env")
	t.Setenv("WEZEN_APPLICATION_ID", "app-from-custom-env")
	desc := buildSyntheticBotRunDescription(DiscoveredBot{Descriptor: &jsdiscord.BotDescriptor{Name: "ui-showcase"}}, "ui-showcase")
	parserConfig := botCLIParserConfig("wezen")
	parser, err := glazed_cli.NewCobraParserFromSections(desc.Schema.Clone(), &glazed_cli.CobraParserConfig{
		AppName:           parserConfig.AppName,
		ShortHelpSections: parserConfig.ShortHelpSections,
	})
	require.NoError(t, err)
	cmd := glazed_cli.NewCobraCommandFromCommandDescription(desc)
	require.NoError(t, parser.AddToCobraCommand(cmd))
	require.NoError(t, cmd.ParseFlags(nil))

	parsed, err := parser.Parse(cmd, nil)
	require.NoError(t, err)
	cfg, err := appconfig.FromValues(parsed)
	require.NoError(t, err)
	require.Equal(t, "token-from-custom-env", cfg.BotToken)
	require.Equal(t, "app-from-custom-env", cfg.ApplicationID)
}

func executeCaptured(t *testing.T, root *cobra.Command, args []string) (string, error) {
	t.Helper()
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	os.Stderr = w
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	root.SetOut(w)
	root.SetErr(w)
	root.SetArgs(args)
	execErr := root.Execute()
	_ = w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()
	return buf.String(), execErr
}

func repoFromDir(dir, source string) Repository {
	return Repository{
		Name:      filepath.Base(dir),
		Source:    source,
		SourceRef: source,
		RootDir:   dir,
	}
}

func examplesFixtureDir(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "examples", "discord-bots")
}

func scannerFixtureDir(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "internal", "botcli", "testdata", "scanner-repo")
}

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	return root
}
