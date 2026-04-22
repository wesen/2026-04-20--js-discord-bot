package botcli

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/go-go-goja/engine"
	"github.com/go-go-golems/go-go-goja/pkg/jsverbs"

	"github.com/manuel/wesen/2026-04-20--js-discord-bot/internal/jsdiscord"
)

// botVerbInvoker executes standard jsverbs discovered in Discord bot scripts.
// It extends the default jsverbs runtime with the Discord module registrar so
// scripts that require("discord") at top level can still run as CLI verbs.
func botVerbInvoker(ctx context.Context, registry *jsverbs.Registry, verb *jsverbs.VerbSpec, parsedValues *values.Values) (interface{}, error) {
	if registry == nil {
		return nil, fmt.Errorf("registry is nil")
	}
	if verb == nil {
		return nil, fmt.Errorf("verb is nil")
	}

	absScript := ""
	if verb.File != nil {
		absScript = strings.TrimSpace(verb.File.AbsPath)
	}
	builder := engine.NewBuilder().
		WithModules(engine.DefaultRegistryModules()).
		WithRequireOptions(require.WithLoader(registry.RequireLoader())).
		WithRuntimeModuleRegistrars(jsdiscord.NewRegistrar(jsdiscord.Config{}))

	if absScript != "" {
		builder = engine.NewBuilder(
			engine.WithModuleRootsFromScript(absScript, engine.DefaultModuleRootsOptions()),
		).WithModules(engine.DefaultRegistryModules()).
			WithRequireOptions(
				require.WithLoader(registry.RequireLoader()),
				require.WithGlobalFolders(filepath.Dir(absScript), filepath.Join(filepath.Dir(absScript), "node_modules")),
			).
			WithRuntimeModuleRegistrars(jsdiscord.NewRegistrar(jsdiscord.Config{}))
	}

	factory, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("build runtime for %s: %w", verb.SourceRef(), err)
	}
	runtime, err := factory.NewRuntime(ctx)
	if err != nil {
		return nil, fmt.Errorf("create runtime for %s: %w", verb.SourceRef(), err)
	}
	defer func() {
		_ = runtime.Close(context.Background())
	}()

	return registry.InvokeInRuntime(ctx, runtime, verb, parsedValues)
}
