package botcli

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-go-golems/go-go-goja/pkg/jsverbs"
	"github.com/go-go-golems/discord-bot/internal/jsdiscord"
)

func DiscoverBots(ctx context.Context, bootstrap Bootstrap, hostOpts ...jsdiscord.HostOption) ([]DiscoveredBot, error) {
	ret := []DiscoveredBot{}
	seen := map[string]DiscoveredBot{}
	for _, repo := range bootstrap.Repositories {
		scripts, err := discoverScriptCandidates(repo.RootDir)
		if err != nil {
			return nil, fmt.Errorf("discover scripts in %s: %w", repo.RootDir, err)
		}
		for _, script := range scripts {
			descriptor, err := jsdiscord.InspectScript(ctx, script, hostOpts...)
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

// ScanBotRepositories discovers bot entrypoint scripts in each repository and scans
// only those scripts for jsverbs metadata. This avoids treating helper libraries under
// bot directories as standalone verbs.
func ScanBotRepositories(repos []Repository) ([]*jsverbs.Registry, error) {
	if len(repos) == 0 {
		return nil, fmt.Errorf("no repositories provided")
	}

	results := make([]*jsverbs.Registry, 0, len(repos))
	for _, repo := range repos {
		scripts, err := discoverScriptCandidates(repo.RootDir)
		if err != nil {
			return nil, fmt.Errorf("discover bot scripts in %s: %w", repo.RootDir, err)
		}
		if len(scripts) == 0 {
			results = append(results, &jsverbs.Registry{RootDir: repo.RootDir})
			continue
		}

		inputs := make([]jsverbs.SourceFile, 0, len(scripts))
		for _, script := range scripts {
			rel, err := filepath.Rel(repo.RootDir, script)
			if err != nil {
				return nil, fmt.Errorf("relpath %s: %w", script, err)
			}
			content, err := os.ReadFile(script)
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", script, err)
			}
			inputs = append(inputs, jsverbs.SourceFile{
				Path:   filepath.ToSlash(rel),
				Source: content,
			})
		}

		registry, err := jsverbs.ScanSources(inputs, jsverbs.ScanOptions{IncludePublicFunctions: false})
		if err != nil {
			return nil, fmt.Errorf("scan bot scripts in %s: %w", repo.RootDir, err)
		}
		registry.RootDir = repo.RootDir
		for _, file := range registry.Files {
			if file == nil {
				continue
			}
			file.AbsPath = filepath.Join(repo.RootDir, filepath.FromSlash(file.RelPath))
		}
		results = append(results, registry)
	}

	return results, nil
}

func ResolveBot(selector string, discovered []DiscoveredBot) (DiscoveredBot, error) {
	selector = strings.TrimSpace(selector)
	if selector == "" {
		return DiscoveredBot{}, fmt.Errorf("bot selector is empty")
	}
	matches := make([]DiscoveredBot, 0, len(discovered))
	for _, bot := range discovered {
		if bot.Name() == selector || bot.SourceLabel() == selector {
			matches = append(matches, bot)
		}
	}
	switch len(matches) {
	case 0:
		return DiscoveredBot{}, fmt.Errorf("bot %q not found", selector)
	case 1:
		return matches[0], nil
	default:
		names := make([]string, 0, len(matches))
		for _, bot := range matches {
			names = append(names, bot.Name())
		}
		sort.Strings(names)
		return DiscoveredBot{}, fmt.Errorf("bot selector %q is ambiguous: %s", selector, strings.Join(names, ", "))
	}
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
			looksLikeBot, err := looksLikeBotScript(path)
			if err != nil {
				return err
			}
			if looksLikeBot {
				ret = append(ret, path)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(ret)
	return ret, nil
}

func looksLikeBotScript(path string) (bool, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	trimmed := bytes.TrimSpace(source)
	if len(trimmed) == 0 {
		return false, nil
	}
	text := string(trimmed)
	if !strings.Contains(text, "defineBot") {
		return false, nil
	}
	if strings.Contains(text, `require("discord")`) || strings.Contains(text, `require('discord')`) {
		return true, nil
	}
	return false, nil
}
