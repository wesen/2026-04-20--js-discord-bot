package botcli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestNewBotsCommandEmbedsIntoDownstreamRoot(t *testing.T) {
	root := &cobra.Command{Use: "downstream-app"}
	root.AddCommand(mustNewBotsCommand(t, Bootstrap{Repositories: []Repository{repoFromDir(t, scannerFixtureDir(t), "scanner")}}))

	output, err := executeCaptured(t, root, []string{"bots", "demo-bot", "status", "--output", "json"})
	require.NoError(t, err)
	require.Contains(t, output, `"active": true`)
	require.Contains(t, output, `"commands": 1`)
}

func TestNewBotsCommandExposesRunHelp(t *testing.T) {
	cmd, err := NewBotsCommand(Bootstrap{Repositories: []Repository{repoFromDir(t, examplesFixtureDir(t), "examples")}})
	require.NoError(t, err)

	root := &cobra.Command{Use: "downstream-app"}
	root.AddCommand(cmd)

	output, err := executeCaptured(t, root, []string{"bots", "knowledge-base", "run", "--help"})
	require.NoError(t, err)
	require.Contains(t, output, "--db-path")
	require.Contains(t, output, "--capture-enabled")
}

func TestNewBotsCommandAcceptsWithAppNameOption(t *testing.T) {
	root := &cobra.Command{Use: "downstream-app"}
	root.AddCommand(mustNewBotsCommand(t,
		Bootstrap{Repositories: []Repository{repoFromDir(t, scannerFixtureDir(t), "scanner")}},
		WithAppName("wezen"),
	))

	output, err := executeCaptured(t, root, []string{"bots", "demo-bot", "status", "--output", "json"})
	require.NoError(t, err)
	require.Contains(t, output, `"active": true`)
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

	root := &cobra.Command{Use: "downstream-app"}
	root.AddCommand(mustNewBotsCommand(t,
		Bootstrap{Repositories: []Repository{repo}},
		WithRuntimeModuleRegistrars(testAppRegistrar{}),
	))

	output, err := executeCaptured(t, root, []string{"bots", "custom-module-bot", "status", "--output", "json"})
	require.NoError(t, err)
	require.Contains(t, output, `"active": true`)
	require.Contains(t, output, `"module": "app"`)
}

func TestNewBotsCommandSupportsCustomRuntimeFactory(t *testing.T) {
	repo := writeBotCLIRepoBot(t, `
const { defineBot } = require("discord");
const app = require("app");
module.exports = defineBot(({ configure }) => {
  configure({ name: "custom-module-bot", description: app.description() });
});
function status() { return { active: true, module: app.name(), description: app.description() }; }
__verb__("status", { output: "glaze", short: "Show custom module status" });
`)

	root := &cobra.Command{Use: "downstream-app"}
	root.AddCommand(mustNewBotsCommand(t,
		Bootstrap{Repositories: []Repository{repo}},
		WithRuntimeFactory(customRuntimeFactory{}),
	))

	output, err := executeCaptured(t, root, []string{"bots", "custom-module-bot", "status", "--output", "json"})
	require.NoError(t, err)
	require.Contains(t, output, `"active": true`)
	require.Contains(t, output, `"module": "app"`)
	require.Contains(t, output, `"description": "Bot using a custom runtime module"`)
}

func TestNewBotsCommandExposesOnlyCanonicalRunShape(t *testing.T) {
	root := &cobra.Command{Use: "downstream-app"}
	root.AddCommand(mustNewBotsCommand(t, Bootstrap{Repositories: []Repository{repoFromDir(t, examplesFixtureDir(t), "examples")}}))

	output, err := executeCaptured(t, root, []string{"bots", "ui-showcase", "run", "--help"})
	require.NoError(t, err)
	require.Contains(t, output, "--bot-token")
	require.Contains(t, output, "--application-id")

	output, err = executeCaptured(t, root, []string{"bots", "run", "ui-showcase", "--help"})
	require.NoError(t, err)
	require.Contains(t, output, "Available Commands:")
	require.NotContains(t, output, "--bot-token")
	require.NotContains(t, output, "bots run ui-showcase")
}

func TestNewBotsCommandDoesNotLeakHelperFunctions(t *testing.T) {
	root := &cobra.Command{Use: "downstream-app"}
	root.AddCommand(mustNewBotsCommand(t, Bootstrap{Repositories: []Repository{repoFromDir(t, examplesFixtureDir(t), "examples")}}))

	output, err := executeCaptured(t, root, []string{"bots", "ui-showcase"})
	require.NoError(t, err)
	require.Contains(t, output, "run         Run the ui-showcase Discord bot")
	require.NotContains(t, output, "first-value")
}

func executeCaptured(t *testing.T, root interface {
	SetArgs([]string)
	Execute() error
	SetOut(io.Writer)
	SetErr(io.Writer)
}, args []string) (string, error) {
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

func mustNewBotsCommand(t *testing.T, bootstrap Bootstrap, opts ...CommandOption) *cobra.Command {
	t.Helper()
	cmd, err := NewBotsCommand(bootstrap, opts...)
	require.NoError(t, err)
	return cmd
}

func repoFromDir(t *testing.T, dir, source string) Repository {
	t.Helper()
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
