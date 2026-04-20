package botcli

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/manuel/wesen/2026-04-20--js-discord-bot/internal/jsdiscord"
)

func DiscoverBootstrapFromCommand(cmd *cobra.Command) (Bootstrap, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return Bootstrap{}, fmt.Errorf("resolve cwd: %w", err)
	}
	repositories, err := repositoriesFromCommand(cmd, cwd)
	if err != nil {
		return Bootstrap{}, err
	}
	return bootstrapFromRepositories(repositories)
}

func bootstrapFromRepositories(repositories []Repository) (Bootstrap, error) {
	if len(repositories) == 0 {
		return Bootstrap{}, fmt.Errorf("at least one --%s must be provided", BotRepositoryFlag)
	}
	seen := map[string]struct{}{}
	ret := make([]Repository, 0, len(repositories))
	for _, repo := range repositories {
		identity := filepath.Clean(repo.RootDir)
		if _, ok := seen[identity]; ok {
			continue
		}
		seen[identity] = struct{}{}
		ret = append(ret, repo)
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].RootDir < ret[j].RootDir })
	return Bootstrap{Repositories: ret}, nil
}

func repositoriesFromCommand(cmd *cobra.Command, cwd string) ([]Repository, error) {
	if cmd == nil {
		return nil, nil
	}
	paths := []string{}
	if cmd.Flags().Lookup(BotRepositoryFlag) != nil {
		flagPaths, err := cmd.Flags().GetStringArray(BotRepositoryFlag)
		if err != nil {
			return nil, err
		}
		paths = append(paths, flagPaths...)
	}
	if cmd.InheritedFlags().Lookup(BotRepositoryFlag) != nil {
		flagPaths, err := cmd.InheritedFlags().GetStringArray(BotRepositoryFlag)
		if err != nil {
			return nil, err
		}
		paths = append(paths, flagPaths...)
	}
	return repositoriesFromCLIPaths(paths, cwd)
}

func repositoriesFromCLIPaths(paths []string, cwd string) ([]Repository, error) {
	ret := []Repository{}
	for _, raw := range paths {
		normalized, err := normalizeFilesystemRepositoryPath(raw, cwd)
		if err != nil {
			return nil, fmt.Errorf("CLI repository %q: %w", raw, err)
		}
		ret = append(ret, Repository{
			Name:      filepath.Base(normalized),
			Source:    "cli",
			SourceRef: "--" + BotRepositoryFlag,
			RootDir:   normalized,
		})
	}
	return ret, nil
}

func normalizeFilesystemRepositoryPath(path string, baseDir string) (string, error) {
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
		path = filepath.Join(baseDir, path)
	}
	path = filepath.Clean(path)
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", path)
	}
	return path, nil
}

func DiscoverBots(ctx context.Context, bootstrap Bootstrap) ([]DiscoveredBot, error) {
	ret := []DiscoveredBot{}
	seen := map[string]DiscoveredBot{}
	for _, repo := range bootstrap.Repositories {
		scripts, err := discoverScriptCandidates(repo.RootDir)
		if err != nil {
			return nil, fmt.Errorf("discover scripts in %s: %w", repo.RootDir, err)
		}
		for _, script := range scripts {
			descriptor, err := jsdiscord.InspectScript(ctx, script)
			if err != nil {
				return nil, fmt.Errorf("inspect bot script %s: %w", script, err)
			}
			candidate := DiscoveredBot{Repository: repo, Descriptor: descriptor}
			if err := candidate.Validate(); err != nil {
				return nil, fmt.Errorf("inspect bot script %s: %w", script, err)
			}
			name := candidate.Name()
			if previous, ok := seen[name]; ok {
				return nil, fmt.Errorf("duplicate bot name %q from %s and %s", name, previous.ScriptPath(), candidate.ScriptPath())
			}
			seen[name] = candidate
			ret = append(ret, candidate)
		}
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].Name() < ret[j].Name() })
	return ret, nil
}

func discoverScriptCandidates(root string) ([]string, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, fmt.Errorf("repository root is empty")
	}
	ret := []string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		name := d.Name()
		if d.IsDir() {
			if name == "node_modules" || strings.HasPrefix(name, ".") {
				if path == root {
					return nil
				}
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.EqualFold(filepath.Ext(name), ".js") {
			return nil
		}
		parent := filepath.Dir(path)
		if parent == root || strings.EqualFold(name, "index.js") {
			ret = append(ret, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(ret)
	return ret, nil
}
