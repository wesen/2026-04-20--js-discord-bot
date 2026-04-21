package jsdiscord

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/go-go-goja/engine"
)

func loadTestBot(t *testing.T, scriptPath string) *BotHandle {
	t.Helper()
	factory, err := engine.NewBuilder(
		engine.WithModuleRootsFromScript(scriptPath, engine.DefaultModuleRootsOptions()),
	).WithModules(engine.DefaultRegistryModules()).
		WithRuntimeModuleRegistrars(NewRegistrar(Config{})).
		Build()
	if err != nil {
		t.Fatalf("build factory: %v", err)
	}
	rt, err := factory.NewRuntime(context.Background())
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	t.Cleanup(func() { _ = rt.Close(context.Background()) })
	value, err := rt.Require.Require(scriptPath)
	if err != nil {
		t.Fatalf("require bot script: %v", err)
	}
	handle, err := CompileBot(rt.VM, value)
	if err != nil {
		t.Fatalf("compile bot: %v", err)
	}
	return handle
}

func writeBotScript(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "bot.js")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write bot script: %v", err)
	}
	return path
}

func repoRootJSDiscord(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	return root
}
