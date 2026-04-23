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
- [ ] Move the entrypoint-only explicit-verb scanning policy into the public package
  - scan only real bot entrypoints
  - set `IncludePublicFunctions: false`
  - preserve correct `AbsPath` mapping for host-managed run commands
- [ ] Keep host-managed `run` semantics in the public package
  - explicit `__verb__("run")` becomes a `BareCommand`
  - bots without explicit run metadata still get a synthetic `run` command
  - register both `bots <bot> run` and `bots run <bot>` compatibility paths
- [ ] Make the public package configurable for downstream apps
  - `WithAppName(...)` so dynamic commands use Glazed env middleware correctly
  - `WithRuntimeFactory(...)` / framework hooks for custom runtimes and modules
  - optional custom repository flag/env names for embedders
- [ ] Add regression tests for the public package
  - downstream-style root integration with raw argv pre-scan
  - env-backed `DISCORD_BOT_TOKEN` / `DISCORD_APPLICATION_ID` loading
  - no leaked helper verbs from bot libraries
  - compatibility aliases for both run command shapes
- [ ] Update framework docs/examples to show downstream embedding of the public `botcli` package

#### Integration / positioning
- [ ] Document the recommended split clearly
  - core framework = single explicit bot is easy
  - optional public `botcli` package = repository-driven multi-bot workflows are easy
- [ ] Add one downstream example app that shows both layers together
  - one explicit built-in bot via `framework.New(...)`
  - plus repo-driven bots via the public `botcli` package
