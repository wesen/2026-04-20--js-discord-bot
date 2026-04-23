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

func TestNewCommandEmbedsIntoDownstreamRoot(t *testing.T) {
	root := &cobra.Command{Use: "downstream-app"}
	root.AddCommand(NewCommand(Bootstrap{Repositories: []Repository{repoFromDir(t, scannerFixtureDir(t), "scanner")}}))

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

func TestNewCommandAcceptsWithAppNameOption(t *testing.T) {
	root := &cobra.Command{Use: "downstream-app"}
	root.AddCommand(NewCommand(
		Bootstrap{Repositories: []Repository{repoFromDir(t, scannerFixtureDir(t), "scanner")}},
		WithAppName("wezen"),
	))

	output, err := executeCaptured(t, root, []string{"bots", "demo-bot", "status", "--output", "json"})
	require.NoError(t, err)
	require.Contains(t, output, `"active": true`)
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
