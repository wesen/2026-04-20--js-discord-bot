package botcli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildBootstrapPrefersCLIOverEnvAndDefault(t *testing.T) {
	root := t.TempDir()
	cliRepo := mustMkdir(t, filepath.Join(root, "cli-repo"))
	envRepo := mustMkdir(t, filepath.Join(root, "env-repo"))
	_ = mustMkdir(t, filepath.Join(root, "examples", "discord-bots"))

	t.Setenv(DefaultEnvVarName, envRepo)

	bootstrap, err := BuildBootstrap(
		[]string{"--bot-repository", cliRepo, "bots", "list"},
		WithWorkingDirectory(root),
	)
	require.NoError(t, err)
	require.Len(t, bootstrap.Repositories, 1)
	require.Equal(t, cliRepo, bootstrap.Repositories[0].RootDir)
	require.Equal(t, "cli", bootstrap.Repositories[0].Source)
}

func TestBuildBootstrapUsesEnvWhenCLIAbsent(t *testing.T) {
	root := t.TempDir()
	envRepo := mustMkdir(t, filepath.Join(root, "env-repo"))
	_ = mustMkdir(t, filepath.Join(root, "examples", "discord-bots"))

	t.Setenv(DefaultEnvVarName, envRepo)

	bootstrap, err := BuildBootstrap(nil, WithWorkingDirectory(root))
	require.NoError(t, err)
	require.Len(t, bootstrap.Repositories, 1)
	require.Equal(t, envRepo, bootstrap.Repositories[0].RootDir)
	require.Equal(t, "env", bootstrap.Repositories[0].Source)
}

func TestBuildBootstrapUsesDefaultWhenCLIAndEnvAbsent(t *testing.T) {
	root := t.TempDir()
	defaultRepo := mustMkdir(t, filepath.Join(root, "examples", "discord-bots"))
	t.Setenv(DefaultEnvVarName, "")

	bootstrap, err := BuildBootstrap(nil, WithWorkingDirectory(root))
	require.NoError(t, err)
	require.Len(t, bootstrap.Repositories, 1)
	require.Equal(t, defaultRepo, bootstrap.Repositories[0].RootDir)
	require.Equal(t, "default", bootstrap.Repositories[0].Source)
}

func TestBuildBootstrapDedupesRepeatedCLIRepositories(t *testing.T) {
	root := t.TempDir()
	repo := mustMkdir(t, filepath.Join(root, "repo"))

	bootstrap, err := BuildBootstrap(
		[]string{"--bot-repository", repo, "--bot-repository=" + repo, "bots", "list"},
		WithWorkingDirectory(root),
	)
	require.NoError(t, err)
	require.Len(t, bootstrap.Repositories, 1)
	require.Equal(t, repo, bootstrap.Repositories[0].RootDir)
}

func TestBuildBootstrapSupportsCustomFlagAndEnvVar(t *testing.T) {
	root := t.TempDir()
	envRepo := mustMkdir(t, filepath.Join(root, "env-repo"))
	cliRepo := mustMkdir(t, filepath.Join(root, "cli-repo"))
	t.Setenv("APP_BOT_REPOS", envRepo)

	bootstrap, err := BuildBootstrap(
		[]string{"--repo-root=" + cliRepo},
		WithWorkingDirectory(root),
		WithRepositoryFlagName("repo-root"),
		WithEnvironmentVariable("APP_BOT_REPOS"),
		WithDefaultRepositories(),
	)
	require.NoError(t, err)
	require.Len(t, bootstrap.Repositories, 1)
	require.Equal(t, cliRepo, bootstrap.Repositories[0].RootDir)
	require.Equal(t, "cli", bootstrap.Repositories[0].Source)
}

func mustMkdir(t *testing.T, path string) string {
	t.Helper()
	require.NoError(t, os.MkdirAll(path, 0o755))
	return path
}
