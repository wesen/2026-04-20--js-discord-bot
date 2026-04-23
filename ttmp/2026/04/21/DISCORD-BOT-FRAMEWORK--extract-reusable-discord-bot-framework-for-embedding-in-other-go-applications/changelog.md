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

