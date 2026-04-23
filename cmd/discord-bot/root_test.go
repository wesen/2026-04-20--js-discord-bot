package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRootHelpLoadsEmbeddedDocs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains string
	}{
		{
			name:     "api reference",
			args:     []string{"help", "discord-js-bot-api-reference"},
			contains: "Discord JavaScript Bot API Reference",
		},
		{
			name:     "tutorial",
			args:     []string{"help", "build-and-run-discord-js-bots"},
			contains: "Build and Run Discord JavaScript Bots",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := newRootCommand(tt.args...)
			require.NoError(t, err)

			var stdout bytes.Buffer
			root.SetOut(&stdout)
			root.SetErr(&stdout)
			root.SetArgs(tt.args)

			err = root.Execute()
			require.NoError(t, err)
			require.Contains(t, stdout.String(), tt.contains)
		})
	}
}

func TestRootLevelBotRepositoryFlagRegistersDiscoveredStatusVerb(t *testing.T) {
	args := []string{"--bot-repository", scannerFixtureDir(t), "bots", "demo-bot", "status", "--output", "json"}
	root, err := newRootCommand(args...)
	require.NoError(t, err)

	output, err := executeCaptured(t, root, args)
	require.NoError(t, err)
	require.Contains(t, output, `"active": true`)
	require.Contains(t, output, `"commands": 1`)
}

func TestRootLevelBotRepositoryFlagRegistersKnowledgeBaseRunVerb(t *testing.T) {
	args := []string{"--bot-repository", examplesFixtureDir(t), "bots", "knowledge-base", "run", "--help"}
	root, err := newRootCommand(args...)
	require.NoError(t, err)

	output, err := executeCaptured(t, root, args)
	require.NoError(t, err)
	require.Contains(t, output, "--db-path")
	require.Contains(t, output, "--capture-enabled")
	require.Contains(t, output, "--review-limit")
}

func TestStandaloneRunHelpShowsSyncOnStart(t *testing.T) {
	args := []string{"run", "--help"}
	root, err := newRootCommand(args...)
	require.NoError(t, err)

	output, err := executeCaptured(t, root, args)
	require.NoError(t, err)
	require.Contains(t, output, "--sync-on-start")
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

func examplesFixtureDir(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "examples", "discord-bots")
}

func scannerFixtureDir(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "pkg", "botcli", "testdata", "scanner-repo")
}

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	return root
}
