package botcli

import (
	"fmt"

	"github.com/go-go-golems/go-go-goja/pkg/jsverbs"
)

// ScanBotRepositories walks each bot repository directory and discovers
// jsverbs (__verb__ / __section__ / __package__) metadata using the
// go-go-goja Tree-sitter scanner.
//
// Returns a slice of per-repo scan results. Callers can iterate over
// each registry to build commands.
func ScanBotRepositories(repos []Repository) ([]*jsverbs.Registry, error) {
	if len(repos) == 0 {
		return nil, fmt.Errorf("no repositories provided")
	}

	results := make([]*jsverbs.Registry, 0, len(repos))
	for _, repo := range repos {
		registry, err := jsverbs.ScanDir(repo.RootDir)
		if err != nil {
			return nil, fmt.Errorf("scan %s: %w", repo.RootDir, err)
		}
		results = append(results, registry)
	}

	return results, nil
}
