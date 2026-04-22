package botcli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/go-go-goja/pkg/jsverbs"
)

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

		registry, err := jsverbs.ScanSources(inputs)
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
