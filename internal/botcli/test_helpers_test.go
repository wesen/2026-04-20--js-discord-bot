package botcli

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/dop251/goja"
	noderequire "github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/go-go-goja/engine"
	"github.com/stretchr/testify/require"
)

func writeBotCLIRepoBot(t *testing.T, script string) Repository {
	t.Helper()
	dir := t.TempDir()
	botDir := filepath.Join(dir, "custom-module-bot")
	require.NoError(t, os.MkdirAll(botDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(botDir, "index.js"), []byte(script), 0o644))
	return Repository{
		Name:      filepath.Base(dir),
		Source:    "test",
		SourceRef: "test",
		RootDir:   dir,
	}
}

type testAppRegistrar struct{}

func (testAppRegistrar) ID() string {
	return "test-app-registrar"
}

func (testAppRegistrar) RegisterRuntimeModules(_ *engine.RuntimeModuleContext, reg *noderequire.Registry) error {
	reg.RegisterNativeModule("app", func(vm *goja.Runtime, moduleObj *goja.Object) {
		exports := moduleObj.Get("exports").(*goja.Object)
		_ = exports.Set("name", func() string { return "app" })
		_ = exports.Set("description", func() string { return "Bot using a custom runtime module" })
		_ = exports.Set("commandDescription", func() string { return "Ping through the custom Go module" })
		_ = exports.Set("greeting", func() string { return fmt.Sprintf("hello from %s", "app") })
	})
	return nil
}
