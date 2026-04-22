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
