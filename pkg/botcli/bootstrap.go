package botcli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	internalbotcli "github.com/manuel/wesen/2026-04-20--js-discord-bot/internal/botcli"
)

const (
	BotRepositoryFlag = internalbotcli.BotRepositoryFlag
	DefaultEnvVarName = "DISCORD_BOT_REPOSITORIES"
	DefaultRepoPath   = "examples/discord-bots"
)

type (
	Bootstrap     = internalbotcli.Bootstrap
	Repository    = internalbotcli.Repository
	DiscoveredBot = internalbotcli.DiscoveredBot
)

type buildOptions struct {
	workingDirectory string
	envVarName       string
	defaultRepos     []string
	flagName         string
}

// BuildOption customizes public repository bootstrap construction.
type BuildOption func(*buildOptions) error

// WithWorkingDirectory controls how relative repository paths are resolved.
func WithWorkingDirectory(dir string) BuildOption {
	return func(cfg *buildOptions) error {
		if strings.TrimSpace(dir) == "" {
			return fmt.Errorf("working directory is empty")
		}
		cfg.workingDirectory = dir
		return nil
	}
}

// WithEnvironmentVariable overrides the env var used for repository discovery.
func WithEnvironmentVariable(name string) BuildOption {
	return func(cfg *buildOptions) error {
		name = strings.TrimSpace(name)
		if name == "" {
			return fmt.Errorf("environment variable name is empty")
		}
		cfg.envVarName = name
		return nil
	}
}

// WithDefaultRepositories overrides the fallback repository paths.
func WithDefaultRepositories(paths ...string) BuildOption {
	return func(cfg *buildOptions) error {
		cfg.defaultRepos = append([]string(nil), paths...)
		return nil
	}
}

// WithRepositoryFlagName overrides the CLI flag name used during raw-argv pre-scan.
func WithRepositoryFlagName(name string) BuildOption {
	return func(cfg *buildOptions) error {
		name = strings.TrimSpace(name)
		if name == "" {
			return fmt.Errorf("repository flag name is empty")
		}
		cfg.flagName = strings.TrimLeft(name, "-")
		if cfg.flagName == "" {
			return fmt.Errorf("repository flag name is empty")
		}
		return nil
	}
}

// BuildBootstrap resolves bot repositories using CLI > env > default precedence.
func BuildBootstrap(rawArgs []string, opts ...BuildOption) (Bootstrap, error) {
	cfg := buildOptions{
		envVarName:   DefaultEnvVarName,
		defaultRepos: []string{DefaultRepoPath},
		flagName:     BotRepositoryFlag,
	}
	cwd, err := os.Getwd()
	if err != nil {
		return Bootstrap{}, fmt.Errorf("resolve cwd: %w", err)
	}
	cfg.workingDirectory = cwd
	for i, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&cfg); err != nil {
			return Bootstrap{}, fmt.Errorf("apply build option %d: %w", i, err)
		}
	}

	repos, err := repositoriesFromArgs(rawArgs, cfg.flagName, cfg.workingDirectory)
	if err != nil {
		return Bootstrap{}, err
	}
	if len(repos) > 0 {
		return Bootstrap{Repositories: repos}, nil
	}

	if cfg.envVarName != "" {
		repos, err = repositoriesFromEnv(cfg.envVarName, cfg.workingDirectory)
		if err != nil {
			return Bootstrap{}, err
		}
		if len(repos) > 0 {
			return Bootstrap{Repositories: repos}, nil
		}
	}

	repos, err = repositoriesFromDefaults(cfg.defaultRepos, cfg.workingDirectory)
	if err != nil {
		return Bootstrap{}, err
	}
	return Bootstrap{Repositories: repos}, nil
}

func repositoriesFromArgs(args []string, flagName string, cwd string) ([]Repository, error) {
	paths, err := repositoryPathsFromArgs(args, flagName)
	if err != nil {
		return nil, err
	}
	return buildRepositories(paths, "cli", "--"+flagName, cwd, true)
}

func repositoriesFromEnv(envVarName, cwd string) ([]Repository, error) {
	value := os.Getenv(envVarName)
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	paths := []string{}
	for _, path := range strings.Split(value, string(os.PathListSeparator)) {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		paths = append(paths, path)
	}
	return buildRepositories(paths, "env", envVarName, cwd, true)
}

func repositoriesFromDefaults(paths []string, cwd string) ([]Repository, error) {
	return buildRepositories(paths, "default", strings.Join(paths, ","), cwd, false)
}

func buildRepositories(paths []string, source, sourceRef, cwd string, requireExists bool) ([]Repository, error) {
	ret := []Repository{}
	seen := map[string]struct{}{}
	for _, path := range paths {
		abs, err := normalizeRepositoryPath(path, cwd)
		if err != nil {
			if !requireExists && (os.IsNotExist(err) || strings.Contains(err.Error(), "no such file or directory")) {
				continue
			}
			return nil, fmt.Errorf("resolve %s repository %q: %w", source, path, err)
		}
		if _, ok := seen[abs]; ok {
			continue
		}
		seen[abs] = struct{}{}
		ret = append(ret, Repository{
			Name:      filepath.Base(abs),
			Source:    source,
			SourceRef: sourceRef,
			RootDir:   abs,
		})
	}
	return ret, nil
}

func normalizeRepositoryPath(path string, cwd string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("repository path is empty")
	}
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		path = filepath.Join(home, strings.TrimPrefix(path, "~/"))
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(cwd, path)
	}
	path = filepath.Clean(path)
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("not a directory")
	}
	return path, nil
}

func repositoryPathsFromArgs(args []string, flagName string) ([]string, error) {
	ret := []string{}
	flagName = strings.TrimLeft(strings.TrimSpace(flagName), "-")
	if flagName == "" {
		return nil, fmt.Errorf("repository flag name is empty")
	}
	longFlag := "--" + flagName
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			break
		}
		if arg == longFlag {
			if i+1 >= len(args) {
				return nil, fmt.Errorf("flag needs an argument: %s", longFlag)
			}
			i++
			ret = append(ret, args[i])
			continue
		}
		if strings.HasPrefix(arg, longFlag+"=") {
			ret = append(ret, strings.TrimPrefix(arg, longFlag+"="))
		}
	}
	return ret, nil
}
