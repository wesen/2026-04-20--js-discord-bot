package botcli

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/go-go-goja/engine"
	"github.com/go-go-golems/go-go-goja/pkg/jsverbs"

	"github.com/go-go-golems/discord-bot/internal/jsdiscord"
)

// defaultRuntimeFactory builds the standard ordinary-jsverb runtime used when
// callers do not provide WithRuntimeFactory(...).
func defaultRuntimeFactory(cfg commandOptions) RuntimeFactory {
	return RuntimeFactoryFunc(func(ctx context.Context, registry *jsverbs.Registry, verb *jsverbs.VerbSpec) (*engine.Runtime, error) {
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

		runtimeRegistrars := []engine.RuntimeModuleRegistrar{jsdiscord.NewRegistrar(jsdiscord.Config{})}
		runtimeRegistrars = append(runtimeRegistrars, cfg.runtimeModuleRegistrars...)

		builder := engine.NewBuilder().
			WithModules(engine.DefaultRegistryModules()).
			WithRequireOptions(require.WithLoader(registry.RequireLoader())).
			WithRuntimeModuleRegistrars(runtimeRegistrars...)

		if absScript != "" {
			builder = engine.NewBuilder(
				engine.WithModuleRootsFromScript(absScript, engine.DefaultModuleRootsOptions()),
			).WithModules(engine.DefaultRegistryModules()).
				WithRequireOptions(
					require.WithLoader(registry.RequireLoader()),
					require.WithGlobalFolders(filepath.Dir(absScript), filepath.Join(filepath.Dir(absScript), "node_modules")),
				).
				WithRuntimeModuleRegistrars(runtimeRegistrars...)
		}

		factory, err := builder.Build()
		if err != nil {
			return nil, fmt.Errorf("build runtime for %s: %w", verb.SourceRef(), err)
		}
		runtime, err := factory.NewRuntime(ctx)
		if err != nil {
			return nil, fmt.Errorf("create runtime for %s: %w", verb.SourceRef(), err)
		}
		return runtime, nil
	})
}
