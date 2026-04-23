# Tasks

## TODO

- [x] Add tasks here

### Completed design work
- [x] Design Option B: Functional options builder for framework.Embed()
- [x] Analyze current coupling points between jsdiscord, bot, botcli, config, and cmd
- [x] Design the Option A: Interface-based extraction with BotProvider/ModuleProvider
- [x] Design Option C: Hybrid — opinionated defaults + escape hatches + custom primitives
- [x] Write the design doc with API references, pseudocode, diagrams, and file references
- [x] Update the design doc to recommend an optional public `botcli` package for embedders
- [x] Upload the updated design doc to reMarkable

### New implementation tasks

#### Track A — keep single-bot embedding first-class and simple
- [ ] Extract a clean public single-bot framework API
  - [x] Create an initial public `pkg/framework` package with `framework.New(...)`
  - [x] Add explicit `WithScript(...)` and `WithCredentials(...)` options
  - [x] Add `WithCredentialsFromEnv()` for the simple env-backed path
  - [x] Add `WithRuntimeModuleRegistrars(...)` for custom Go-native `require()` modules
  - [x] Keep the package free of repository scanning concerns
  - `framework.New(...)` / `NewHost(...)` remain the primary path
- [ ] Preserve or improve the current single-bot operator flow
  - [x] Add `--sync-on-start` to the standalone `run --bot-script ...` path
  - standalone `run --bot-script ...` remains supported
  - downstream embedders can run one explicit bot without importing `botcli`
- [x] Add single-bot embedding examples to docs
  - [x] minimal embedded app example
  - [x] one explicit bot + custom module/runtime example
- [ ] Add regression tests for the single-bot path
  - [x] simple framework startup with one explicit script
  - [x] credentials/env loading for the explicit single-bot path
  - [x] `run --help` exposes `--sync-on-start`
  - optional `--sync-on-start` behavior where applicable

#### Track B — optional public `botcli` package for repo-driven bots
- [ ] Promote `internal/botcli` into an optional public package (`pkg/botcli` or `framework/botcli`)
  - [x] Create an initial public `pkg/botcli` bootstrap package
- [ ] Define the public bootstrap / repository API
  - [x] `Repository`
  - [x] `Bootstrap`
  - [x] `BuildBootstrap(rawArgs []string, opts ...)`
  - [x] CLI/env/default repository precedence helpers
- [x] Expose a public embeddable Cobra integration entrypoint
  - [x] `NewCommand(bootstrap, opts...)`
  - or `NewRootCommand(rawArgs, opts...)`
  - [x] make downstream integration into an existing Cobra root trivial
- [x] Move the entrypoint-only explicit-verb scanning policy into the public package
  - [x] Slice A1 — add public repository scanning helpers that discover only real bot entrypoints
  - [x] Slice A2 — keep `IncludePublicFunctions: false` in the public scan path so helper functions never leak into the command tree
  - [x] Slice A3 — preserve `AbsPath` mapping in the public scan path so host-managed run commands bind to the correct script files
- [x] Keep host-managed `run` semantics in the public package
  - [x] Slice B1 — public command builder owns explicit `__verb__("run")` as a `BareCommand`
  - [x] Slice B2 — public command builder synthesizes `run` for bots without explicit run metadata
  - [x] Slice B3 — public command builder registers both `bots <bot> run` and `bots run <bot>` compatibility paths
- [ ] Make the public package configurable for downstream apps
  - [x] `WithAppName(...)` so dynamic commands use Glazed env middleware correctly
  - [x] `WithRuntimeModuleRegistrars(...)` as the first public runtime-extension hook
  - [x] Slice C1 — add `WithRuntimeFactory(...)` as a broader public hook for ordinary jsverbs runtime creation
  - [x] Slice C2 — let a custom runtime factory contribute host options used by discovery and host-managed bot runs
  - optional custom repository flag/env names for embedders
- [ ] Add regression tests for the public package
  - downstream-style root integration with raw argv pre-scan
  - env-backed `DISCORD_BOT_TOKEN` / `DISCORD_APPLICATION_ID` loading
  - [x] no leaked helper verbs from bot libraries
  - [x] compatibility aliases for both run command shapes
  - [x] custom runtime factory behavior for ordinary jsverbs plus discovery/run host-option parity
- [x] Update framework docs/examples to show downstream embedding of the public `botcli` package

#### Integration / positioning
- [x] Document the recommended split clearly
  - [x] core framework = single explicit bot is easy
  - [x] optional public `botcli` package = repository-driven multi-bot workflows are easy
- [x] Add one downstream example app that shows both layers together
  - [x] one explicit built-in bot via `framework.New(...)`
  - [x] plus repo-driven bots via the public `botcli` package

#### Cleanup / clean cut after review
- [x] Remove public aliases and wrapper constructors from `pkg/botcli`
  - [x] make `pkg/botcli` own its public model types directly
  - [x] remove the panic-based `NewCommand(...)` wrapper and switch first-party callers/tests to `NewBotsCommand(...)`
  - [x] stop dogfooding explicit default `WithAppName("discord")` where the default already applies
- [x] Remove backwards-compatibility command paths and docs
  - [x] delete the `bots run <bot>` compatibility path and keep only `bots <bot> run`
  - [x] remove compatibility-specific helper code/tests/docs
  - [x] update operator-facing help/docs/error text to the canonical command shape
- [x] Delete the duplicated `internal/botcli` implementation once the public path stands on its own
  - [x] move any still-needed test fixtures/helpers into `pkg/botcli`
  - [x] remove redundant internal tests and command/discovery helpers
  - [x] validate the repo on the public path only
