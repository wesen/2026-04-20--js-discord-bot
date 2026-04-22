---
Title: Implementation Diary
Ticket: DISCORD-BOT-FRAMEWORK
Status: active
Topics:
    - discord
    - goja
    - javascript
    - go
    - framework
    - embedding
    - glazed
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: design-doc/01-framework-extraction-design-and-implementation-guide.md
      Note: The main design document this diary accompanies
ExternalSources: []
Summary: Chronological diary for the framework extraction design ticket.
LastUpdated: 2026-04-22T10:00:00-04:00
WhatFor: Record the analysis steps, design decisions, and deliverables for making the discord-bot a reusable framework.
WhenToUse: Use when resuming or reviewing the DISCORD-BOT-FRAMEWORK work.
---

# Diary

## Goal

Record the analysis, design, and documentation work for extracting the JS Discord bot runtime into a reusable Go framework that other applications can embed — for example, a web server with a database that also exposes a Discord bot surface.

## Step 1: Create ticket and analyze current coupling

### What I did

- Created docmgr ticket `DISCORD-BOT-FRAMEWORK`.
- Read every Go file in the project to map coupling points between packages.
- Read all relevant docmgr diaries from DISCORD-BOT-002 through DISCORD-BOT-020, plus the UI-DSL design tickets and the PASS1–4 cleanup passes.
- Read the go-go-goja engine, modules, and runtimeowner packages to understand the module registration path.
- Started the design document.

### Key findings

The current codebase has five packages with tight but discoverable coupling:

1. **`internal/jsdiscord/`** — The core value. This package owns the goja runtime, the `defineBot` DSL, `BotHandle`, `DispatchRequest`, `DiscordOps`, payload normalization, and event routing. It is the main extraction target. It currently depends only on `discordgo`, `goja`, and `go-go-goja/engine`.

2. **`internal/bot/`** — The live host wiring. This package creates the `discordgo.Session`, registers all Discord event handlers, and routes them to `jsdiscord.Host.Dispatch*` methods. It depends on `config.Settings` for credentials and on `jsdiscord` for the host. This is the package that must become pluggable.

3. **`internal/botcli/`** — The CLI orchestration. Discovery, bot resolution, dynamic Glazed schema building, and the `bots run` execution path. This is application-level code, not framework code. Embedders would replace this with their own wiring.

4. **`internal/config/`** — A Glazed-backed `Settings` struct for Discord credentials. Simple and already fairly independent. The framework would expose its own config interface.

5. **`cmd/discord-bot/`** — The CLI entrypoint. Wires Glazed commands, help system, and the `botcli` commands together. This stays as the standalone application.

### What I learned

- The `jsdiscord.Host` struct is *almost* already a framework API. It takes a script path, creates a runtime, loads the bot, and exposes `Describe()`, `DispatchInteraction()`, `DispatchReady()`, etc. The main coupling is that `NewHost()` calls `engine.NewBuilder()` internally and that `buildDiscordOps()` reaches directly into the `discordgo.Session`.

- The `DiscordOps` struct is the key extensibility seam. It is already a bag of function pointers (`GuildFetch`, `ChannelSend`, etc.). Today these are populated by `build*Ops()` helpers that close over a `*discordgo.Session`. A framework would let embedders provide their own ops — or additional ops — without touching the host.

- The `DispatchRequest` struct carries everything a JS handler sees: args, config, discord ops, reply/followUp/edit/defer/showModal callbacks. This is the second key seam. Today the responder callbacks are built from `discordgo.Interaction` and `discordgo.Session`, but they could just as well be built from an HTTP handler or a test mock.

### Next step

Write the design document with three design options, full API references, pseudocode, and diagrams.

## Step 2: Write the design document

### What I did

- Wrote a comprehensive design document with:
  - Executive summary
  - Detailed problem statement with the five coupling points
  - Full current architecture walkthrough with package map, dependency graph, and file-by-file coupling analysis
  - Three design options with full pseudocode, diagrams, and trade-off analysis
  - Detailed implementation plan with file mapping table and code diffs
  - Complete API reference
  - Five concrete embedding examples
  - Open questions and alternatives considered

### What I found

The codebase is already well-structured for extraction. The main work is:

1. Parameterizing `NewHost()` to accept engine builder configuration
2. Parameterizing `baseDispatchRequest()` to accept ops customizers
3. Parameterizing `buildContext()` to accept context extenders
4. Growing `DiscordOps.Extensions` for custom namespace injection
5. Wrapping everything in a `Framework` convenience layer

### Next step

Upload the design document to reMarkable.

## Step 3: Add jsverbs integration and RuntimeFactory override

### What I did

- Read the entire loupedeck codebase to understand how it creates custom goja runtimes and integrates with jsverbs.
- Read the `go-go-goja/pkg/jsverbs` package in full — `model.go`, `scan.go`, `runtime.go`, `binding.go`, `command.go` — to understand the scan/build/invoke pipeline.
- Read loupedeck's `verbs/bootstrap.go` to understand repository discovery from CLI, env, config, and embedded FS.
- Read loupedeck's `verbs/command.go` to understand how verbs become Cobra commands with custom invokers.
- Read loupedeck's `runtime/js/runtime.go` and `registrar.go` to understand the custom runtime factory pattern.

### Key findings

The loupedeck pattern reveals a 4th coupling point I missed in the original analysis:

- **Runtime creation is hardcoded.** The discord-bot's `jsdiscord.NewHost()` calls `engine.NewBuilder()` internally with fixed configuration. Loupedeck needs to control this — it registers its own modules (ui, gfx, easing, anim, present, state, metrics) and its own environment object. The framework must let embedders replace the entire runtime creation process.

The jsverbs system provides three reusable pieces:
1. `jsverbs.ScanDir()` / `ScanFS()` — find annotated JS files
2. `Registry.CommandsWithInvoker()` — build Glazed commands with custom execution
3. `Registry.RequireLoader()` — provide a custom require loader for verb invocation

### What I added to the design

Three new concepts:
1. **RuntimeFactory** interface — lets embedders control how goja runtimes are created for both bot scripts and verb invocations
2. **Repository** + **RepositoryDiscovery** — unified repository model for both bot discovery and jsverb scanning
3. **VerbRegistry** + **RegisterVerbs()** — framework helper for scanning repos and registering jsverb CLI commands

Updated the comparison table and the file mapping.

### Next step

Re-upload the updated design document to reMarkable.
