---
title: "Sections 7-9: API Reference, Testing, Risks"
description: Pseudocode, testing strategy, risks, alternatives, and open questions.
doc_type: design-doc
status: active
topics: [packaging, testing, risks]
ticket: DISCORD-BOT-PUBLISH
---

## 7. API Reference and Pseudocode

### 7.1 Downstream embedding example (Path A: framework)

This is what a downstream Go application looks like when embedding discord-bot:

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/go-go-golems/discord-bot/pkg/framework"
)

func main() {
    bot, err := framework.New(
        framework.WithCredentialsFromEnv(),
        framework.WithScript("./bots/my-bot/index.js"),
        framework.WithRuntimeConfig(map[string]any{
            "db_path": "./data/my-bot.sqlite",
        }),
        framework.WithSyncOnStart(true),
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Block until context is canceled (SIGINT, etc.)
    if err := bot.Run(ctx); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
}
```

### 7.2 Downstream embedding example (Path B: botcli)

This is what a downstream Go application looks like when using the repo-driven CLI:

```go
package main

import (
    "os"

    "github.com/go-go-golems/discord-bot/pkg/botcli"
    "github.com/spf13/cobra"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "my-app",
        Short: "My application with embedded Discord bot support",
    }

    // Resolve bot repositories from CLI flags, env vars, or defaults
    bootstrap, err := botcli.BuildBootstrap(
        os.Args[1:],
        botcli.WithDefaultRepositories("./bots"),
    )
    if err != nil {
        // Handle error
    }

    // Mount the bots command tree
    botsCmd, err := botcli.NewBotsCommand(bootstrap)
    if err != nil {
        // Handle error
    }
    rootCmd.AddCommand(botsCmd)

    rootCmd.Execute()
}
```

### 7.3 Downstream embedding example (Path C: combined)

When an application wants both a built-in bot and repo-driven discovery:

```go
package main

import (
    "os"

    "github.com/go-go-golems/discord-bot/pkg/botcli"
    "github.com/go-go-golems/discord-bot/pkg/framework"
    "github.com/spf13/cobra"
)

var version = "dev"

func main() {
    rootCmd := &cobra.Command{
        Use:     "my-combined-app",
        Version: version,
    }

    // Built-in bot (always available)
    rootCmd.AddCommand(&cobra.Command{
        Use:   "run-built-in",
        Short: "Run the built-in bot",
        RunE: func(cmd *cobra.Command, args []string) error {
            bot, err := framework.New(
                framework.WithCredentialsFromEnv(),
                framework.WithScript("./internal/bot/index.js"),
            )
            if err != nil {
                return err
            }
            return bot.Run(cmd.Context())
        },
    })

    // Repo-driven bots (optional discovery)
    bootstrap, _ := botcli.BuildBootstrap(
        os.Args[1:],
        botcli.WithDefaultRepositories("./bots"),
    )
    botsCmd, _ := botcli.NewBotsCommand(bootstrap)
    rootCmd.AddCommand(botsCmd)

    rootCmd.Execute()
}
```

### 7.4 Custom native module example

When a downstream app needs to expose custom Go functionality to JS bot scripts:

```go
package main

import (
    "github.com/dop251/goja"
    "github.com/go-go-golems/discord-bot/pkg/botcli"
    "github.com/go-go-golems/go-go-goja/engine"
)

// MyAppModule exposes application-specific functions to JS bot scripts.
type MyAppModule struct{}

func (m *MyAppModule) ID() string {
    return "my-app-module"
}

func (m *MyAppModule) RegisterRuntimeModules(
    ctx *engine.RuntimeModuleContext,
    registry *require.Registry,
) error {
    registry.RegisterNativeModule("app", func(vm *goja.Runtime, module *goja.Object) {
        exports := module.NewObject()
        exports.Set("getVersion", func() string { return "1.0.0" })
        exports.Set("queryDatabase", func(sql string) (any, error) {
            // ... real database query
            return results, nil
        })
        module.Set("exports", exports)
    })
    return nil
}

func main() {
    bootstrap, _ := botcli.BuildBootstrap(os.Args[1:])
    botsCmd, _ := botcli.NewBotsCommand(
        bootstrap,
        botcli.WithRuntimeModuleRegistrars(&MyAppModule{}),
    )
    // ...
}
```

### 7.5 Dependency resolution flow (pseudocode)

```text
BuildBootstrap(rawArgs):
  1. Check rawArgs for --bot-repository flag
     → If found: use those paths, mark source="cli"
  2. Else check DISCORD_BOT_REPOSITORIES env var
     → If found: use those paths, mark source="env"
  3. Else use default paths (e.g., "examples/discord-bots")
     → Mark source="default"
  4. For each path:
     a. Resolve to absolute path (handle ~, relative)
     b. Verify it exists and is a directory
     c. Create Repository{Name: baseDir, RootDir: absPath, Source: source}
  5. Return Bootstrap{Repositories: repos}

NewBotsCommand(bootstrap, opts):
  1. Apply CommandOptions (appName, runtimeModuleRegistrars, runtimeFactory)
  2. Create "bots" root Cobra command
  3. Add static subcommands: "list", "help"
  4. For each repository in bootstrap:
     a. Scan directory for bot scripts (*.js files and */index.js)
     b. Load each script to extract BotDescriptor (name, commands, events)
     c. Create a "run" subcommand per discovered bot
     d. Create verb commands for any jsverb scripts
  5. Return the command tree

botRunCommand.Run(ctx, parsedValues):
  1. Extract Settings from parsedValues (bot-token, application-id, etc.)
  2. Validate settings (require token + application-id)
  3. Build runtime config from parsedValues
  4. Create internal/bot.Bot via NewWithScript(settings, scriptPath, runtimeConfig)
     → This creates discordgo.Session + jsdiscord.Host
  5. Optionally sync slash commands
  6. Open gateway session
  7. Block until context canceled
```

## 8. Testing and Validation Strategy

### 8.1 Existing test coverage

The project already has substantial test coverage in `internal/jsdiscord/`:

```text
Test files (in internal/jsdiscord/):
  descriptor_test.go             Bot descriptor parsing
  runtime_descriptor_test.go     Runtime descriptor validation
  runtime_dispatch_test.go       Event dispatch to JS handlers
  runtime_events_test.go         Event handler registration and invocation
  runtime_payloads_test.go       Response payload normalization
  runtime_jsverbs_polyfill_test.go  JS polyfill compatibility
  runtime_ops_guilds_test.go     Guild operation tests
  runtime_ops_members_test.go    Member operation tests
  runtime_ops_messages_test.go   Message operation tests
  runtime_ops_threads_test.go    Thread operation tests
  helpers_test.go                Shared test helpers
  knowledge_base_runtime_test.go Integration test for knowledge-base bot
  show_space_runtime_test.go     Integration test for show-space bot
  ui_showcase_runtime_test.go    Integration test for UI showcase bot
  ui_builders_test.go            UI DSL builder tests
  ui_module_test.go              UI module tests
  ui_phase34_test.go             UI component phase 3/4 tests

Test files (in pkg/):
  pkg/botcli/bootstrap_test.go   Bootstrap resolution tests
  pkg/botcli/command_test.go     Command tree tests
  pkg/botcli/runtime_helpers_test.go  Runtime helper tests
  pkg/framework/framework_test.go  Framework embedding tests (if exists)
```

### 8.2 New tests to add for packaging

1. **Module import test** — Verify the public API is importable:
   ```go
   // pkg/framework/framework_test.go
   func TestPublicAPISurface(t *testing.T) {
       // Verify all Option functions return non-nil
       opts := []framework.Option{
           framework.WithScript("./testdata/test-bot.js"),
           framework.WithCredentials(framework.Credentials{}),
           framework.WithCredentialsFromEnv(),
           framework.WithRuntimeConfig(map[string]any{}),
           framework.WithSyncOnStart(false),
       }
       for _, opt := range opts {
           assert.NotNil(t, opt)
       }
   }
   ```

2. **Build test** — Verify the binary compiles:
   ```bash
   go build -o /dev/null ./cmd/discord-bot
   ```

3. **Version injection test** — Verify ldflags work:
   ```bash
   go build -ldflags "-X main.version=test-123" ./cmd/discord-bot
   ./discord-bot --version | grep test-123
   ```

### 8.3 Validation checklist

Before each release, verify:

- [ ] `make lint` passes
- [ ] `make test` passes
- [ ] `make build` produces a working binary
- [ ] `make goreleaser` (snapshot) succeeds
- [ ] `discord-bot --version` shows the correct version
- [ ] `discord-bot bots list --bot-repository ./examples/discord-bots` discovers bots
- [ ] `discord-bot bots help ping --bot-repository ./examples/discord-bots` shows metadata
- [ ] A downstream `go get github.com/go-go-golems/discord-bot` works
- [ ] The examples in `examples/` all compile

## 9. Risks, Alternatives, and Open Questions

### 9.1 Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| go-go-goja not yet published | **High** | Keep `replace` directive during transition, remove after publish |
| Import path rename breaks something subtle | Medium | Mechanical replacement + comprehensive `go build ./...` verification |
| CGO requirement limits platform support | Medium | Document clearly; GoReleaser already handles cross-compilation |
| Public API needs breaking change after first release | Medium | Start at v0.x to signal instability; follow semver pre-release convention |
| GPG signing keys not available in new repo | Low | Copy secrets from existing org-level configuration |

### 9.2 Alternatives Considered

1. **Keep the repo where it is, just add infrastructure.**
   - Rejected: The `github.com/manuel/wesen/2026-04-20--js-discord-bot` path is not importable and contains a date prefix that makes it a poor module name.

2. **Use the repo name `go-discord-bot` instead of `discord-bot`.**
   - Rejected: The convention in go-go-golems is to use the binary name without a `go-` prefix (cf. `pinocchio`, not `go-pinocchio`).

3. **Publish only the binary, not the Go package.**
   - Rejected: The Go package is a core part of the value proposition. Downstream apps like show-space, knowledge-base, and future integrations need to embed it.

4. **Move `internal/jsdiscord/` to `pkg/jsdiscord/`.**
   - Rejected: The runtime engine should stay internal. The public API is `pkg/framework/` and `pkg/botcli/`, which wrap the internal engine. This allows internal refactoring without breaking the public API.

### 9.3 Open Questions

1. **Should the Homebrew formula be named `discord-bot` or `go-discord-bot`?**
   - Recommendation: `discord-bot` (matches the binary name, follows pinocchio precedent).
   - Decision needed from @manuel.

2. **Should examples be in the same repo or a separate repo?**
   - Recommendation: Same repo. They serve as executable documentation and integration tests.
   - The examples are small and don't change independently of the runtime.

3. **What version should the first release be?**
   - Recommendation: `v0.1.0` (pre-release semver, signals that the API may still change).
   - Pinocchio is at `v0.10.x` and has not yet declared API stability.

4. **Should the ttmp/ directory be included in the published repo?**
   - Recommendation: Yes, but add it to `.gitignore` for the published repo if it contains only local development state.
   - Alternatively, keep ttmp/ in the repo as documentation of the development history.

5. **Do we need a Go workspace (go.work) file?**
   - Recommendation: No. The published module should be standalone.
   - A local go.work can be used during development (and gitignored).

6. **Should we set up Dependabot?**
   - Recommendation: Yes, add `.github/dependabot.yml` (copy from pinocchio).
