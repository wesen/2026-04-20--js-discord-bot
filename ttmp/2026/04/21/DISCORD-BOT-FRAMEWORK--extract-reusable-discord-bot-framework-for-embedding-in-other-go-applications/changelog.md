# Changelog

## 2026-04-21

- Initial workspace created


## 2026-04-21

Created framework extraction design document with three options (Interface-based, Functional Options, Hybrid). Recommended Option C (Hybrid). Uploaded to reMarkable at /ai/2026/04/22/DISCORD-BOT-FRAMEWORK.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/21/DISCORD-BOT-FRAMEWORK--extract-reusable-discord-bot-framework-for-embedding-in-other-go-applications/design-doc/01-framework-extraction-design-and-implementation-guide.md — Design document with three options


## 2026-04-22

Added jsverbs integration and RuntimeFactory override to design document. Analyzed loupedeck codebase for reference implementation pattern. Added Repository, RepositoryDiscovery, VerbRegistry, and RuntimeFactory concepts. Re-uploaded to reMarkable as v2.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/go-go-goja/pkg/jsverbs — Reference for the scan/build/invoke pipeline
- /home/manuel/code/wesen/corporate-headquarters/loupedeck/cmd/loupedeck/cmds/verbs/bootstrap.go — Reference for repository discovery pattern
- /home/manuel/code/wesen/corporate-headquarters/loupedeck/runtime/js/runtime.go — Reference for custom runtime factory pattern

## 2026-04-22

Started implementation Track A for the framework split by improving the standalone single-bot path. Added `--sync-on-start` to `discord-bot run`, wired it to sync commands before opening the gateway session, added root help coverage, and refined the framework ticket task list to treat the single-bot path as a first-class track alongside public `botcli` extraction.

### Related Files

- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/cmd/discord-bot/commands.go — Standalone `run` command now supports `--sync-on-start`
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/cmd/discord-bot/root_test.go — Root help regression test for `--sync-on-start`
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/ttmp/2026/04/21/DISCORD-BOT-FRAMEWORK--extract-reusable-discord-bot-framework-for-embedding-in-other-go-applications/tasks.md — Track A/Track B split with first single-bot subtask checked off

## 2026-04-22

Created the first public single-bot embedding package at `pkg/framework`. The package exposes `framework.New(...)` with explicit script and credentials options, an env-backed credentials option, runtime config injection, and optional sync-on-start behavior. Added focused tests using the existing example bot scripts to prove the public single-bot path can be constructed without any repository scanning or `botcli` involvement.

### Related Files

- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/framework/framework.go — Initial public single-bot framework wrapper around `internal/bot`
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/framework/framework_test.go — Public package tests for script requirement, env credentials, and runtime config wiring
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/ttmp/2026/04/21/DISCORD-BOT-FRAMEWORK--extract-reusable-discord-bot-framework-for-embedding-in-other-go-applications/tasks.md — Track A updated with initial public package progress

## 2026-04-22

Added the first public single-bot embedding example and linked it from the repository docs. The new example application imports `pkg/framework`, selects one explicit JavaScript bot script, uses env-backed credentials, injects runtime config, and runs without any repository scanning. This makes the Track A story concrete for downstream embedders instead of leaving it as an API-only package.

### Related Files

- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-single-bot/main.go — Minimal Go application using `pkg/framework`
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-single-bot/README.md — Usage notes for the embeddable single-bot example
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/discord-bots/README.md — Examples index now points readers to the public embedding path
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/README.md — Top-level repo README now mentions the public single-bot package and example

## 2026-04-22

Extended the public single-bot framework path so downstream embedders can inject custom Go-native `require()` modules without using repo-driven `botcli`. Added `framework.WithRuntimeModuleRegistrars(...)`, threaded host options through `internal/bot` and `jsdiscord.NewHost(...)`, added regression tests for scripts that require a custom module, and created a second embedding example showing one explicit bot script plus a custom `app` module.

### Related Files

- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/jsdiscord/host.go — `NewHost(...)` now accepts host options and custom runtime module registrars
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/jsdiscord/host_options.go — New host option plumbing for runtime registrars
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/bot/bot.go — Single-bot runtime now forwards host options to `jsdiscord`
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/framework/framework.go — Public `WithRuntimeModuleRegistrars(...)` option
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/framework/framework_test.go — Regression tests for missing/present custom runtime modules
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-custom-module/main.go — Explicit bot embedding example with custom `require("app")` module
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-custom-module/bot/index.js — JS bot script that consumes the injected `app` module

## 2026-04-22

Merged `main` into `task/discord-bot-framework`, validated the combined branch with `go test ./...`, and then started Track B by extracting the repository bootstrap layer into a new public `pkg/botcli` package. The new package exposes `Repository`, `Bootstrap`, and `BuildBootstrap(rawArgs, opts...)` with the same CLI > env > default precedence currently used by `discord-bot`, and the root command now consumes that public bootstrap helper instead of keeping the logic private in `cmd/discord-bot/root.go`.

### Related Files

- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/bootstrap.go — Public bootstrap/repository API for repo-driven bot discovery
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/bootstrap_test.go — Regression tests for CLI/env/default precedence and custom flag/env options
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/cmd/discord-bot/root.go — Root command now uses `pkg/botcli.BuildBootstrap(...)`

## 2026-04-22

Continued Track B by exposing the repo-driven bot command tree itself from `pkg/botcli`. Added public `NewBotsCommand(...)` / `NewCommand(...)` wrappers, added downstream-style integration tests that mount the command under an arbitrary Cobra root, and switched `cmd/discord-bot/root.go` to consume the public package for both bootstrap resolution and command registration.

### Related Files

- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command.go — Public embeddable Cobra entrypoints for repo-driven bots
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_test.go — Downstream-style Cobra integration tests
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/cmd/discord-bot/root.go — Root command now consumes `pkg/botcli.NewCommand(...)`

## 2026-04-22

Added public `WithAppName(...)` configurability to `pkg/botcli` and threaded the option through the repo-driven command builder so downstream apps can control the env prefix used by Glazed for dynamic bot commands. Added internal and public regression coverage, including a parser-level test that proves a non-default prefix such as `WEZEN_BOT_TOKEN` / `WEZEN_APPLICATION_ID` is honored when the app name is set to `wezen`.

### Related Files

- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/options.go — Internal command option plumbing with `WithAppName(...)`
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/command.go — Dynamic parser config now uses the configured app name
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/command_test.go — Custom env-prefix regression coverage
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/options.go — Public `WithAppName(...)` option
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command.go — Public command wrappers now accept command options

## 2026-04-23

Completed the next Track B runtime-customization slice by adding public `WithRuntimeModuleRegistrars(...)` support to `pkg/botcli` and threading it through the repo-driven command flow. The option now affects all three relevant runtime touchpoints: bot discovery/inspection, ordinary jsverbs invocation, and host-managed bot `run` construction. Added internal and public regression tests using bot scripts that require a custom `app` module, and manually validated a tiny downstream app that mounted `pkg/botcli.NewCommand(...)` with the new option and successfully executed a `status` verb from a custom-module bot repository.

### Related Files

- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/options.go — Internal `WithRuntimeModuleRegistrars(...)` plumbing and host-option conversion
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/bootstrap.go — Bot discovery now supports host options for custom modules
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/jsverbs_invoker.go — Ordinary jsverbs runtime now receives custom registrars
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/bot_run_command.go — Host-managed run path now forwards host options into bot construction
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/command_test.go — Internal regression tests for missing/present custom runtime modules
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/options.go — Public `WithRuntimeModuleRegistrars(...)` option
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_test.go — Public package regression tests and downstream-style coverage

## 2026-04-23

Added a durable downstream example app that combines both extracted public layers in one process. The new `examples/framework-combined` application mounts a repo-driven `bots` subtree through `pkg/botcli` while also exposing a `run-builtin` command that starts one explicit built-in bot through `pkg/framework`. Updated the top-level documentation to describe the recommended split clearly and validated that the repo-driven `bots list` flow works through the combined app.

### Related Files

- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-combined/main.go — Combined downstream app using both `pkg/framework` and `pkg/botcli`
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-combined/README.md — Usage docs for the combined example
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-combined/builtin-bot/index.js — Explicit built-in bot script for the combined app
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/README.md — Recommended public split documented explicitly
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/discord-bots/README.md — Examples index now points readers at the combined downstream app

## 2026-04-23

Completed the next public-ownership slice for `pkg/botcli`. The public package now owns repository scanning policy (real bot entrypoints only, explicit verbs only, `AbsPath` preservation), host-managed `run` synthesis (explicit run verbs, synthetic run fallback, and both run command shapes), and a broader `WithRuntimeFactory(...)` hook. Added missing public regressions for helper-function leakage, both run aliases, and custom runtime-factory behavior, then manually validated both run shapes through `examples/framework-combined` and a temporary downstream app using `WithRuntimeFactory(...)` against a custom-module bot repository.

### Related Files

- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/discover.go — Public repository discovery, scan policy, and bot resolution helpers
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/commands_impl.go — Public command builder, list/help commands, and host-managed run ownership
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/options.go — Public `WithRuntimeFactory(...)` and host-option/runtime-factory configuration
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/runtime_factory.go — Default public verb runtime factory
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_test.go — Public regressions for helper leakage, both run shapes, and custom runtime factory support

## 2026-04-23

Added the documentation/review pass for the extracted framework surface. Wrote a detailed code-review cleanup report focused on duplication, deprecation candidates, unclear code, bloated files, backwards-compatibility burden, and legacy wrappers; wrote a separate design review and decision guide to help judge the current extraction state; updated the public docs to call out the main `pkg/botcli` customization hooks; validated ticket vocabulary with `docmgr doctor`; and prepared the review deliverables for reMarkable upload.

### Related Files

- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/ttmp/2026/04/21/DISCORD-BOT-FRAMEWORK--extract-reusable-discord-bot-framework-for-embedding-in-other-go-applications/reference/02-public-botcli-code-review-cleanup-report.md — Detailed cleanup/code-review report with file-backed evidence and cleanup sketches
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/ttmp/2026/04/21/DISCORD-BOT-FRAMEWORK--extract-reusable-discord-bot-framework-for-embedding-in-other-go-applications/design-doc/02-framework-extraction-design-review-and-decision-guide.md — Design-state review to help judge whether the extraction is coherent and what should happen next
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/README.md — Stabilization/docs pass for the public `pkg/botcli` option surface
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-combined/README.md — Combined example now documents the advanced public botcli customization hooks


## 2026-04-23

Started the post-review clean-cut work by removing the first set of public aliases and wrappers from `pkg/botcli`. The public package now owns its repository/bot model types directly, the panic-based `NewCommand(...)` convenience wrapper is gone, and first-party callers now use `NewBotsCommand(...)` directly without redundantly passing the default app name.

### Related Files

- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/model.go — Public `Bootstrap`, `Repository`, `DiscoveredBot`, and `BotRepositoryFlag` now live in the public package directly
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/bootstrap.go — Public bootstrap logic no longer aliases internal botcli types
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/commands_impl.go — Removed the panic-based `NewCommand(...)` wrapper so `NewBotsCommand(...)` is the only constructor
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_test.go — Public command tests now use the canonical constructor directly
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/cmd/discord-bot/root.go — Standalone app now mounts `pkg/botcli.NewBotsCommand(...)`
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-combined/main.go — Combined example now mounts `pkg/botcli.NewBotsCommand(...)`

## 2026-04-23

Completed the next clean-cut slice by removing the legacy `bots run <bot>` compatibility path from the public command tree. The canonical repo-driven run path is now only `bots <bot> run`, the compatibility helper code is gone from `pkg/botcli`, and the operator-facing README/help docs/error text now point to the canonical command shape.

### Related Files

- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/commands_impl.go — Removed the compatibility alias registration loop for legacy `bots run <bot>` paths
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/run_description.go — Removed the compatibility-specific run-description helper
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_test.go — Public regression now asserts only the canonical run shape is exposed
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/README.md — Named-bot examples now use `bots <bot> run`
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/doc/topics/discord-js-bot-api-reference.md — API reference now documents the canonical run shape only
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/doc/tutorials/building-and-running-discord-js-bots.md — Tutorial commands now use `bots <bot> run`
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/bot/bot.go — Direct-run missing-script error now points to `bots <bot> run`

## 2026-04-23

Completed the clean cut by deleting the duplicated `internal/botcli` package entirely. The remaining shared scanner fixture now lives under `pkg/botcli/testdata`, package/root tests point at the public fixture path, the README/example validation snippets no longer mention the deleted internal package, and the repository validates on the public-path-only tree with `go test ./...`.

### Related Files

- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/testdata/scanner-repo/demo-bot.js — Shared scanner fixture now owned by the public package
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_test.go — Public package tests now resolve the scanner fixture from `pkg/botcli/testdata`
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/cmd/discord-bot/root_test.go — Standalone root test now uses the public scanner fixture path
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/README.md — Project layout and validation commands now reflect the public-path-only tree
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-combined/README.md — Combined example now points at the public scanner fixture path
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/ttmp/2026/04/21/DISCORD-BOT-FRAMEWORK--extract-reusable-discord-bot-framework-for-embedding-in-other-go-applications/index.md — Removed the stale related-file entry that referenced deleted `internal/botcli/runtime.go`

## 2026-04-23

Started the dedicated `pkg/botcli` cleanup pass by splitting the oversized command builder into focused files. The public constructor now lives in `command_root.go`, the list/help/run command implementations live in their own files, the shared bool flag helper moved into runtime helpers, and the placeholder `pkg/botcli/command.go` file is gone.

### Related Files

- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_root.go — Public `NewBotsCommand(...)` and command-registration orchestration
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_list.go — Focused list command implementation
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_help.go — Focused help command implementation
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_run.go — Focused host-managed run command implementation
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/runtime_helpers.go — Shared bool/runtime-config helpers after the split
