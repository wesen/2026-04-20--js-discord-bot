package botcli

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/go-go-golems/go-go-goja/pkg/jsverbs"
)

func TestScanBotRepositoriesDiscoversVerbs(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	rootDir := filepath.Join(filepath.Dir(file), "testdata", "scanner-repo")

	repos := []Repository{{Name: "test", Source: "test", SourceRef: "test", RootDir: rootDir}}
	results, err := ScanBotRepositories(repos)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("results = %d", len(results))
	}

	registry := results[0]
	verbs := registry.Verbs()
	if len(verbs) != 2 {
		t.Fatalf("verbs = %d, want 2", len(verbs))
	}

	verbNames := map[string]bool{}
	for _, v := range verbs {
		verbNames[v.Name] = true
	}
	if !verbNames["status"] {
		t.Fatalf("missing 'status' verb")
	}
	if !verbNames["run"] {
		t.Fatalf("missing 'run' verb")
	}
}

func TestScanBotRepositoriesRunVerbHasFields(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	rootDir := filepath.Join(filepath.Dir(file), "testdata", "scanner-repo")

	repos := []Repository{{Name: "test", Source: "test", SourceRef: "test", RootDir: rootDir}}
	results, err := ScanBotRepositories(repos)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	registry := results[0]
	var runVerb *jsverbs.VerbSpec
	for _, v := range registry.Verbs() {
		if v.Name == "run" {
			runVerb = v
			break
		}
	}
	if runVerb == nil {
		t.Fatalf("'run' verb not found")
	}
	if len(runVerb.Fields) != 2 {
		t.Fatalf("run fields = %d, want 2", len(runVerb.Fields))
	}
	if _, ok := runVerb.Fields["bot-token"]; !ok {
		t.Fatalf("missing 'bot-token' field")
	}
	if _, ok := runVerb.Fields["api-key"]; !ok {
		t.Fatalf("missing 'api-key' field")
	}
}
