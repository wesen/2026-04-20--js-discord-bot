package botcli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListCommandOutputsDiscoveredBots(t *testing.T) {
	root := NewCommand()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)
	root.SetArgs([]string{"list", "--bot-repository", examplesFixtureDir(t)})

	err := root.Execute()
	require.NoError(t, err)
	output := stdout.String()
	require.Contains(t, output, "discord greet\tdiscord.js")
	require.Contains(t, output, "issues list\tissues.js")
}

func TestRunCommandExecutesStructuredBot(t *testing.T) {
	root := NewCommand()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)
	root.SetArgs([]string{"run", "discord", "greet", "--bot-repository", examplesFixtureDir(t), "Manuel", "--excited"})

	err := root.Execute()
	require.NoError(t, err)
	require.Contains(t, stdout.String(), "\"greeting\": \"Hello, Manuel!\"")
}

func TestRunCommandExecutesTextBot(t *testing.T) {
	root := NewCommand()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)
	root.SetArgs([]string{"run", "discord", "banner", "--bot-repository", examplesFixtureDir(t), "Manuel"})

	err := root.Execute()
	require.NoError(t, err)
	require.Equal(t, "*** Manuel ***\n", stdout.String())
}

func TestRunCommandSettlesAsyncBot(t *testing.T) {
	root := NewCommand()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)
	root.SetArgs([]string{"run", "math", "multiply", "--bot-repository", examplesFixtureDir(t), "6", "7"})

	err := root.Execute()
	require.NoError(t, err)
	require.Contains(t, stdout.String(), "\"product\": 42")
}

func TestRunCommandSupportsRelativeRequire(t *testing.T) {
	root := NewCommand()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)
	root.SetArgs([]string{"run", "nested", "relay", "--bot-repository", examplesFixtureDir(t), "hi", "there"})

	err := root.Execute()
	require.NoError(t, err)
	require.Contains(t, stdout.String(), "\"value\": \"hi:there\"")
}

func TestHelpCommandShowsVerbFlags(t *testing.T) {
	root := NewCommand()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)
	root.SetArgs([]string{"help", "issues", "list", "--bot-repository", examplesFixtureDir(t)})

	err := root.Execute()
	require.NoError(t, err)
	output := stdout.String()
	require.Contains(t, output, "discord-bot bots run issues list")
	require.Contains(t, output, "--state")
	require.Contains(t, output, "--labels")
}

func TestListCommandAllowsEmptyRepository(t *testing.T) {
	emptyDir := t.TempDir()
	root := NewCommand()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)
	root.SetArgs([]string{"list", "--bot-repository", emptyDir})

	err := root.Execute()
	require.NoError(t, err)
	require.Equal(t, "", stdout.String())
}

func TestListCommandRejectsDuplicateBotsAcrossRepositories(t *testing.T) {
	root := NewCommand()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)
	root.SetArgs([]string{"list", "--bot-repository", duplicateFixtureDir(t, "a"), "--bot-repository", duplicateFixtureDir(t, "b")})

	err := root.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate bot path")
	require.Contains(t, err.Error(), "discord greet")
}

func examplesFixtureDir(t *testing.T) string {
	t.Helper()
	root := repoRoot(t)
	return filepath.Join(root, "examples", "bots")
}

func duplicateFixtureDir(t *testing.T, suffix string) string {
	t.Helper()
	root := repoRoot(t)
	return filepath.Join(root, "examples", "bots-dupe-"+suffix)
}

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	_, err = os.Stat(root)
	require.NoError(t, err)
	return root
}
