---
title: "Section 3: Current-State Architecture"
description: Evidence-based analysis of js-discord-bot, go-template, and pinocchio.
doc_type: design-doc
status: active
topics: [packaging, architecture]
ticket: DISCORD-BOT-PUBLISH
---

## 3. Current-State Architecture (Evidence-Based)

This section maps out three codebases: the prototype (js-discord-bot), the template (go-template), and the finished product (pinocchio). Understanding all three is essential because the plan is to "fill in the template around the prototype to match the finished product."

### 3.1 What js-discord-bot Is Today

#### High-level data flow

```text
Operator types:
  discord-bot bots <name> run --bot-token $TOKEN ...

What happens inside:
  cmd/discord-bot/main.go
    → parses CLI flags (Glazed + Cobra)
    → resolves bot repository (BuildBootstrap)
    → discovers named bot scripts (DiscoverBots)
    → selects the requested bot
    → creates internal/bot.Bot (NewWithScript)
      → creates discordgo.Session (Discord gateway)
      → creates jsdiscord.Host (embedded JS runtime)
        → builds goja engine.Runtime
        → registers require("discord") module
        → loads and compiles the bot script
    → syncs slash commands to Discord
    → opens gateway session
    → blocks until context is canceled

Discord events flow back:
  discordgo.Session handler → bot.handleInteractionCreate
    → jsdiscord.Host.DispatchInteraction()
      → looks up registered JS handler by command/event name
      → calls JS function with ctx object
      → normalizes JS return value into Discord response
      → sends response via discordgo.Session
```

#### Runtime architecture diagram

```text
┌─────────────────────────────────────────────────────────────┐
│                     cmd/discord-bot                         │
│  (Cobra + Glazed CLI, parses flags, wires everything)       │
└──────────────────────┬──────────────────────────────────────┘
                       │
          ┌────────────┴────────────┐
          ▼                         ▼
┌──────────────────┐  ┌──────────────────────────────────┐
│  internal/bot    │  │     pkg/botcli (public)           │
│  (Discord session│  │  Bot discovery, repo scanning,    │
│   lifecycle)     │  │  Cobra command tree mounting      │
└────────┬─────────┘  └──────────────────────────────────┘
         │
         ▼
┌─────────────────────────────────────────────────────────┐
│              internal/jsdiscord                          │
│  ┌─────────────┐  ┌──────────────┐  ┌───────────────┐  │
│  │ Host        │  │ BotCompile   │  │ Dispatch      │  │
│  │ (runtime    │→ │ (defineBot() │→ │ (event→JS     │  │
│  │  lifecycle) │  │  parser)     │  │  handler)     │  │
│  └──────┬──────┘  └──────────────┘  └───────────────┘  │
│         │                                               │
│  ┌──────┴──────────────────────────────────────────┐    │
│  │            require("discord") module             │    │
│  │  ┌────────┐ ┌────────┐ ┌────────┐ ┌──────────┐  │    │
│  │  │Runtime │ │ UI DSL │ │Context │ │Payloads  │  │    │
│  │  │State   │ │(Go-side│ │helpers │ │(message, │  │    │
│  │  │(store, │ │ builders│ │(reply, │ │ embeds,  │  │    │
│  │  │ log)   │ │ for    │ │ defer, │ │ buttons) │  │    │
│  │  │        │ │ discord│ │ edit)  │ │          │  │    │
│  │  └────────┘ └────────┘ └────────┘ └──────────┘  │    │
│  └─────────────────────────────────────────────────┘    │
│                                                          │
│  ┌──────────────────────────────────────────────────┐    │
│  │         Host Operations (outbound Discord)        │    │
│  │  messages | channels | guilds | members | roles   │    │
│  │  threads  | helpers  | command sync | responses   │    │
│  └──────────────────────────────────────────────────┘    │
└──────────────────────────────────────────────────────────┘
         │
         ▼
┌──────────────────────────────────────────────────────────┐
│              JavaScript bot script                        │
│  examples/discord-bots/<name>/index.js                    │
│                                                            │
│  const { defineBot } = require("discord")                 │
│  module.exports = defineBot(({ command, event }) => {     │
│    command("ping", ..., async () => ({ content: "pong" }))│
│  })                                                        │
└──────────────────────────────────────────────────────────┘
```

#### Directory structure (source only, no ttmp/testdata)

```text
cmd/
  discord-bot/              CLI entrypoint (commands.go, root.go)

internal/
  bot/                      Discordgo session wrapper (313 lines)
  config/                   Host config: Settings struct, Validate() (51 lines)
  jsdiscord/                Embedded JS runtime engine (~11,700 lines total)
    host.go                   Top-level Host: NewHost(), Dispatch*(), Close()
    runtime.go                Registrar + RuntimeState for require("discord")
    descriptor.go             BotDescriptor, CommandDescriptor, etc. (367 lines)
    bot_compile.go            CompileBot(): parses JS defineBot() return (728 lines)
    bot_dispatch.go           Routes Discord events to JS handlers (225 lines)
    bot_context.go            Request-scoped ctx object for JS
    bot_store.go              MemoryStore for per-bot key-value state
    bot_logging.go            ctx.log.* implementation
    bot_ops.go                ctx.discord.* outbound operation routing
    host_dispatch.go          Top-level event dispatch (535 lines)
    host_ops.go               Host-level outbound operation dispatch
    host_ops_messages.go      ctx.discord.messages.* implementations
    host_ops_channels.go      ctx.discord.channels.* implementations
    host_ops_guilds.go        ctx.discord.guilds.* implementations
    host_ops_members.go       ctx.discord.members.* implementations
    host_ops_roles.go         ctx.discord.roles.* implementations
    host_ops_threads.go       ctx.discord.threads.* implementations
    host_ops_helpers.go       Shared helpers for host ops (303 lines)
    host_commands.go          Application command sync (243 lines)
    host_maps.go              Snapshot/map conversion (423 lines)
    host_options.go           HostOption functional options
    host_responses.go         Response normalization (251 lines)
    store.go                  MemoryStore implementation
    payload_*.go              Payload type definitions
    snapshot_types.go         Snapshot structs (302 lines)
    snapshot_builders.go      Snapshot construction (228 lines)
    ui_*.go                   Go-side UI DSL for components, embeds, forms

pkg/
  framework/                Simple single-bot embedding API (~200 lines)
  botcli/                   Repo-driven named-bot CLI layer (14 files)
    bootstrap.go              BuildBootstrap(): resolves repos from CLI/env/defaults
    discover.go               DiscoverBots(): scans repos, loads descriptors
    command_root.go           NewBotsCommand(): builds Cobra command tree
    command_run.go            botRunCommand: runs one named bot
    command_list.go           listBotsCommand: lists all discovered bots
    command_help.go           helpBotsCommand: shows one bot's metadata
    model.go                  Bootstrap, Repository, DiscoveredBot types
    options.go                CommandOption, RuntimeFactory, HostOptionsProvider
    runtime_factory.go        Runtime factory hooks
    runtime_helpers.go        Runtime helper functions
    doc.go                    Package documentation
  doc/                       Embedded help pages (topics + tutorials)

examples/
  discord-bots/             Named JS bot implementations
  framework-single-bot/     Minimal embedding example
  framework-custom-module/  Custom require("app") module example
  framework-combined/       Combined built-in + repo-driven example
```

#### Key Go module dependencies

From `go.mod`:

```text
module github.com/manuel/wesen/2026-04-20--js-discord-bot
go 1.26.1

require (
    github.com/bwmarrin/discordgo    v0.29.0      # Discord gateway/session
    github.com/dop251/goja_nodejs     v0.0.0-...   # JS require() support
    github.com/go-go-golems/glazed    v1.2.3       # CLI framework (Cobra + structured output)
    github.com/rs/zerolog             v1.35.0      # Structured logging
    github.com/spf13/cobra            v1.10.2      # CLI command framework
    github.com/stretchr/testify       v1.11.1      # Test assertions
)

replace github.com/go-go-golems/go-go-goja => /home/manuel/code/wesen/corporate-headquarters/go-go-goja
```

The `replace` directive is a **blocker** for publishing. It means the go-go-goja dependency is resolved from a local directory, not from the Go module proxy.

#### The two embedding paths

The public API (`pkg/`) intentionally offers two layers:

**Path A: Simple single-bot embedding** (`pkg/framework/`)

For when you want to embed exactly one bot in your Go application:

```go
bot, err := framework.New(
    framework.WithCredentialsFromEnv(),
    framework.WithScript("./my-bot/index.js"),
    framework.WithSyncOnStart(true),
)
bot.Run(ctx)
```

**Path B: Repo-driven multi-bot CLI** (`pkg/botcli/`)

For when you want the full `bots list / bots help / bots <name> run` experience inside your Cobra command tree:

```go
bootstrap, _ := botcli.BuildBootstrap(os.Args[1:])
cmd, _ := botcli.NewBotsCommand(bootstrap)
rootCmd.AddCommand(cmd)
```

#### The JavaScript bot authoring API

From JavaScript, bots use:

```javascript
const { defineBot } = require("discord")

module.exports = defineBot(({ command, event, configure }) => {
  configure({ name: "demo", description: "A demo bot" })

  event("ready", async (ctx) => {
    ctx.log.info("ready", { user: ctx.me && ctx.me.username })
  })

  command("ping", { description: "Reply pong" }, async () => {
    return { content: "pong" }
  })
})
```

The `ctx` (context) object in handlers exposes:

- `ctx.reply(...)`, `ctx.defer(...)`, `ctx.edit(...)`, `ctx.followUp(...)` — response helpers
- `ctx.showModal(...)` — modal presentation
- `ctx.log.*(...)` — structured logging
- `ctx.store.*(...)` — per-bot key-value store
- `ctx.discord.*(...)` — outbound Discord operations (channels, messages, members, guilds, roles, threads)
- `ctx.config` — runtime config from CLI flags
- `ctx.me` — current bot user snapshot

### 3.2 What go-template Provides

The `go-template` repository is a skeleton that every go-go-golems tool starts from. It provides the **infrastructure boilerplate** that js-discord-bot is missing.

#### File listing

```text
go-template/
  cmd/XXX/main.go            # Empty main() placeholder
  pkg/doc.go                 # Package doc placeholder
  .goreleaser.yaml           # GoReleaser config (linux + darwin, deb, rpm, brew, fury.io)
  .golangci.yml              # Lint config (errcheck, govet, staticcheck, etc.)
  .golangci-lint-version     # Pin file
  Makefile                   # All standard targets
  lefthook.yml               # Git hooks (pre-commit lint+test, pre-push release+lint+test)
  AGENT.md                   # Instructions for AI agents
  README.md                  # Template README
  LICENSE                    # MIT license
  .github/workflows/
    release.yaml              # Split GoReleaser (linux + darwin, merge, GPG sign, brew, fury)
    push.yml                  # CI: generate + test on every push/PR
    lint.yml                  # Lint workflow
    codeql-analysis.yml       # CodeQL security analysis
    secret-scanning.yml       # Secret scanning
    dependency-scanning.yml   # Dependency scanning
```

#### Key patterns from go-template

1. **GoReleaser split build.** The release workflow builds linux and darwin binaries in separate jobs, then merges them. This is because CGO cross-compilation requires different compilers per platform.

2. **Makefile targets.** Every project has: `make lint`, `make test`, `make build`, `make goreleaser`, `make tag-patch`, `make release`.

3. **Lefthook pre-commit/pre-push hooks.** These catch issues before they reach CI.

4. **Module path.** The template uses `github.com/go-go-golems/XXX` — you replace `XXX` with your binary name.

#### go-template Makefile (key targets)

```makefile
VERSION=v0.1.14
GORELEASER_ARGS ?= --skip=sign --snapshot --clean
GOLANGCI_LINT_BIN ?= $(CURDIR)/.bin/golangci-lint

lint: golangci-lint-install
	$(GOLANGCI_LINT_BIN) config verify
	$(GOLANGCI_LINT_BIN) run -v $(GOLANGCI_LINT_ARGS)

test:
	GOWORK=off go test ./...

build:
	GOWORK=off go generate ./...
	GOWORK=off go build ./...

goreleaser:
	GOWORK=off goreleaser release $(GORELEASER_ARGS) $(GORELEASER_TARGET)

tag-major / tag-minor / tag-patch:
	git tag $(shell svu <major|minor|patch>)

release:
	git push origin --tags
	GOPROXY=proxy.golang.org go list -m github.com/go-go-golems/XXX@$(shell svu current)
```

#### go-template GoReleaser config (structure)

```yaml
project_name: XXX
builds:
  - id: XXX-linux          # CGO_ENABLED=1, cross-compile arm64
    main: ./cmd/XXX
  - id: XXX-darwin         # CGO_ENABLED=1
    main: ./cmd/XXX
brews:                     # Homebrew tap
  - repository:
      owner: go-go-golems
      name: homebrew-go-go-go
nfpms:                     # deb + rpm
  - formats: [deb, rpm]
publishers:                # fury.io apt/rpm repo
  - name: fury.io
```

#### go-template release workflow (key pattern)

```yaml
jobs:
  goreleaser-linux:        # Build on ubuntu, cross-compile arm64
  goreleaser-darwin:       # Build on macos
  goreleaser-merge:        # Merge artifacts, GPG sign, publish brew+fury
    needs: [goreleaser-linux, goreleaser-darwin]
    environment: release   # Requires manual approval in GitHub
```

The `environment: release` means releases require a human approval step in the GitHub UI before artifacts are published.

### 3.3 What pinocchio Looks Like as a Finished Published Tool

Pinocchio is the best reference for what a finished go-go-golems tool looks like. It is a much larger project (~50+ Go packages, web frontend, protobuf schemas, SQLite persistence) but follows the same patterns.

#### Key facts about pinocchio

```text
Module:    github.com/go-go-golems/pinocchio
go 1.26.1
Binary:    ./cmd/pinocchio/main.go
Lines:     ~30,000+ Go (excluding generated protobuf)
Release:   v0.10.13-next (actively versioned)
Homebrew:  go-go-golems/homebrew-go-go-go
Publishes: deb, rpm (via fury.io), Homebrew, GitHub Releases
```

#### pinocchio main.go structure (pattern to follow)

```go
package main

import (
    "embed"
    "os"

    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/go-go-golems/glazed/pkg/help"
    // ... internal packages
)

var version = "dev"  // Injected by GoReleaser via ldflags

//go:embed prompts/*
var promptsFS embed.FS

var rootCmd = &cobra.Command{
    Use:     "pinocchio",
    Short:   "pinocchio is a tool to run LLM applications",
    Version: version,
}

func main() {
    helpSystem, err := initRootCmd()
    cobra.CheckErr(err)
    // ... command loading, repository scanning
    cobra.CheckErr(rootCmd.Execute())
}
```

Key patterns:

1. **`var version = "dev"`** — GoReleaser injects the real version via `-ldflags -X main.version=...`.
2. **`//go:embed prompts/*`** — Static assets are embedded into the binary.
3. **Cobra + Glazed** — The CLI framework is the same one js-discord-bot already uses.
4. **`initRootCmd()`** — A single function that wires everything: help system, command loading, repository scanning.
5. **Help system** — `pkg/doc/` contains embedded markdown that becomes `pinocchio help <topic>`.

#### pinocchio infrastructure (what to replicate)

```text
pinocchio/
  .goreleaser.yaml      ✅ Full config: linux+darwin, brew, deb, rpm, fury
  Makefile               ✅ Full targets: lint, test, build, proto-gen, goreleaser, release
  lefthook.yml           ✅ Pre-commit + pre-push hooks
  .golangci.yml          ✅ Custom rules (excludes for generated protobuf)
  .github/workflows/
    release.yml          ✅ Split build + merge + sign + publish
    push.yml             ✅ CI on every push
    lint.yml             ✅ Lint workflow
    codeql-analysis.yml  ✅ Security scanning
    secret-scanning.yml  ✅ Secret scanning
    webchat-check.yml    # (domain-specific: web frontend checks)
```

### 3.4 Gap Analysis: Current vs Target

| Aspect | js-discord-bot (current) | Target (like pinocchio) | Work needed |
|--------|--------------------------|------------------------|-------------|
| Module path | `github.com/manuel/wesen/2026-04-20--js-discord-bot` | `github.com/go-go-golems/discord-bot` | Rename in go.mod + all imports |
| Local replace | `go-go-goja => /home/.../go-go-goja` | No replace directive | Remove after go-go-goja is published |
| Go version | 1.26.1 | 1.26.1 | Already aligned |
| Makefile | None | Full targets | Copy from go-template, adapt |
| GoReleaser | None | Full config | Copy from go-template, adapt |
| CI workflows | None | push + release + lint + security | Copy from go-template, adapt |
| Lint config | None | .golangci.yml + lefthook.yml | Copy from go-template, adapt |
| Version injection | None | `var version = "dev"` + ldflags | Add to cmd/discord-bot/main.go |
| Embed FS | None | For help docs | Already has pkg/doc/ embeds |
| cmd/ structure | Single binary | Single binary (already correct) | Minor cleanup |
| pkg/ API | framework + botcli (exists but young) | Stable public API | Review and stabilize |
| GitHub repo | Local only | github.com/go-go-golems/discord-bot | Push to org |
| Homebrew | None | Formula in homebrew-go-go-go | GoReleaser handles this |
| deb/rpm | None | Via fury.io | GoReleaser handles this |
