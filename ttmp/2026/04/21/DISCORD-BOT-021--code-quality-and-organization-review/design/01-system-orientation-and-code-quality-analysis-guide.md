---
Title: System Orientation and Code Quality Analysis Guide
Ticket: DISCORD-BOT-021
Status: active
Topics:
    - backend
    - go
    - javascript
    - architecture
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/cmd/discord-bot/root.go
      Note: CLI root that composes Glazed commands, help docs, and the bot repository runner
    - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/bot/bot.go
      Note: Live Discord session wrapper that bridges Discord gateway events into the JavaScript host
    - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/runtime.go
      Note: Runtime module registrar that exposes require("discord") to the JS runtime
    - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/bot.go
      Note: Largest runtime file; contains bot-definition DSL, dispatch boundary, context building, and capability bindings
    - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_dispatch.go
      Note: Host-side inbound event and interaction routing from Discordgo into the JS runtime
    - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_payloads.go
      Note: Payload normalization hot spot for outgoing Discord responses and operations
    - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/botcli/run_schema.go
      Note: Dynamic startup-config parsing path for bot-level runtime configuration
    - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/knowledge-base/index.js
      Note: Largest and richest example bot; key example for API ergonomics and organizational pressure
ExternalSources: []
Summary: Orient a new intern to the repo and identify the subsystems that matter most for code quality and organization review.
LastUpdated: 2026-04-21T07:35:00-04:00
WhatFor: Help a reviewer understand the system before reading the detailed cleanup and code quality report.
WhenToUse: Use when onboarding into the repo or planning refactors focused on maintainability, organization, and API clarity.
---

# Goal

This guide is the orientation document for the code quality review. It explains what the system is, how the major parts fit together, and which subsystems matter most before you read the detailed report.

This document is intentionally written for a new intern. The goal is not only to say “this code could be cleaner,” but to explain:

- what the product does,
- how the main runtime path works,
- where the complexity lives,
- and why certain files feel large, repetitive, or confusing.

## Scope and explicit non-goals

This review is about **code quality, organization, API clarity, duplication, and stale/deprecated code**.

It is **not** primarily about:

- bug hunting,
- correctness audits,
- or the colleague’s in-flight work around command/subcommand/message-command expansion.

That means this review treats the command-surface expansion as an active area and focuses instead on the structural parts that already shape the whole repository.

# What this repository is

At a high level, this repository is a **Discord bot host** that lets you implement bot behavior in **JavaScript** while keeping the outer Discord gateway/session plumbing in **Go**.

Another way to say it:

- Go owns the Discord session, process lifecycle, command sync, and runtime embedding.
- JavaScript owns the bot-definition surface (`require("discord")`) and the bot’s actual behavior.
- The repository also includes several example bots that act as both demos and integration tests.

# The most important mental model

The simplest reliable mental model is:

```text
CLI / Config
    ↓
Discord session host (Go)
    ↓
JS runtime host (Go + goja)
    ↓
require("discord") DSL for bot authors
    ↓
Example bot implementations (JS)
```

If you understand those five layers, most of the codebase becomes readable.

# Directory map

The live code is concentrated in a few places.

## Main directories

- `cmd/discord-bot/`
  - the CLI entrypoint
- `internal/config/`
  - configuration decoding and validation
- `internal/bot/`
  - live Discordgo session wrapper
- `internal/jsdiscord/`
  - the embedded JavaScript Discord runtime and host bridge
- `internal/botcli/`
  - repository discovery and `bots list|help|run`
- `examples/discord-bots/`
  - named JavaScript bot implementations
- `pkg/doc/`
  - embedded help docs for the CLI

## Size snapshot

The repo’s structural weight is not evenly distributed.

### Package-level live-code totals

```text
6661 LOC  22 files  internal/jsdiscord
5175 LOC  21 files  examples/discord-bots
1264 LOC   7 files  internal/botcli
1104 LOC   3 files  pkg/doc
 352 LOC   4 files  cmd/discord-bot
 350 LOC   1 files  internal/bot
  61 LOC   1 files  internal/config
```

### Biggest live files

```text
1289  internal/jsdiscord/bot.go
1205  internal/jsdiscord/runtime_test.go
 736  internal/jsdiscord/host_payloads.go
 664  examples/discord-bots/knowledge-base/lib/store.js
 591  examples/discord-bots/knowledge-base/index.js
 476  internal/jsdiscord/host_dispatch.go
 462  examples/discord-bots/knowledge-base/lib/search.js
 427  examples/discord-bots/knowledge-base/lib/review.js
 397  internal/jsdiscord/descriptor.go
 369  internal/jsdiscord/host_maps.go
 350  internal/bot/bot.go
 346  internal/botcli/run_schema.go
```

The package and file counts already tell us something important:

- `internal/jsdiscord` is the architectural center of gravity.
- the `knowledge-base` example is the biggest example and the most useful stress test for API design.
- the code quality review should focus there first.

# Read order for a new intern

If you are new to this repo, read the following files in this order.

## 1. CLI entrypoint

Start here:

- `cmd/discord-bot/root.go`
- `cmd/discord-bot/commands.go`

These files tell you:

- what the main operator-facing commands are,
- how Glazed and Cobra are used,
- and how the repo distinguishes between direct runtime commands and bot repository commands.

## 2. Config boundary

Then read:

- `internal/config/config.go`

This file is small but important. It shows which runtime settings are considered first-class and how CLI/env inputs become a `Settings` object.

## 3. Live Discord session host

Then read:

- `internal/bot/bot.go`

This is the Go-side shell around Discordgo. It:

- creates the session,
- loads one JavaScript bot,
- syncs application commands,
- and forwards inbound events and interactions into the JS host.

## 4. JS runtime registration

Then read:

- `internal/jsdiscord/runtime.go`

This is the first file that explains how the `discord` module gets injected into the embedded JS runtime.

## 5. Bot-definition runtime core

Then read the biggest file:

- `internal/jsdiscord/bot.go`

This file defines the JS bot authoring surface and the bridge between Go and JS.

Important entry points here are:

- `CompileBot(...)`
- `(*BotHandle).dispatch(...)`
- `buildDispatchInput(...)`
- `buildContext(...)`
- `discordOpsObject(...)`
- `(*botDraft).command(...)`
- `(*botDraft).configure(...)`
- `(*botDraft).finalize(...)`

## 6. Host-side event/interaction routing and payload normalization

Then read:

- `internal/jsdiscord/host.go`
- `internal/jsdiscord/host_dispatch.go`
- `internal/jsdiscord/host_payloads.go`
- `internal/jsdiscord/host_maps.go`

These files explain how Discordgo events become `DispatchRequest` values and how JS return values become Discord API payloads again.

## 7. Bot repository runner

Then read:

- `internal/botcli/command.go`
- `internal/botcli/bootstrap.go`
- `internal/botcli/run_schema.go`
- `internal/botcli/runtime.go`

These files explain the `discord-bot bots ...` command surface and the startup-config parsing used for selected bot implementations.

## 8. Example bots

Finally, inspect:

- `examples/discord-bots/ping/index.js`
- `examples/discord-bots/poker/index.js`
- `examples/discord-bots/knowledge-base/index.js`

Read the knowledge-base bot last. It is the richest example and the best demonstration of where the current API and organization start to feel heavy.

# Core runtime flow

The most useful control-flow diagram is this one.

```text
Operator runs CLI
    │
    ├── cmd/discord-bot/root.go
    │       builds root command, help system, and subcommands
    │
    ├── cmd/discord-bot/commands.go
    │       decodes config for run / validate-config / sync-commands
    │
    └── internal/botcli/*
            resolves named bots from repositories for `bots list|help|run`

Selected bot script path
    │
    └── internal/bot/bot.go::NewWithScript
            creates Discord session
            creates jsdiscord host
            attaches Discord handlers

Discord event / interaction arrives
    │
    └── internal/bot/bot.go::handle...
            forwards to internal/jsdiscord/host_dispatch.go

Host dispatch
    │
    └── internal/jsdiscord/host_dispatch.go
            converts Discordgo structs to plain maps
            builds DispatchRequest
            attaches reply/edit/followUp/defer/showModal funcs

JS runtime boundary
    │
    └── internal/jsdiscord/bot.go::BotHandle.dispatch
            converts DispatchRequest to JS object
            calls JS handler
            settles Promise results if needed

JS bot author code
    │
    └── examples/discord-bots/<bot>/index.js
            command(...)
            event(...)
            component(...)
            modal(...)
            autocomplete(...)

JS returns payload / action
    │
    └── internal/jsdiscord/host_payloads.go
            normalizes embeds/components/files/mentions/etc.
            sends response through Discordgo
```

# API surface map

The repo has a few key internal APIs that are worth memorizing.

## `internal/config`

### `type Settings`
File: `internal/config/config.go`

Purpose:
- strongly named Go-side config object for the live bot host.

Key methods:
- `FromValues(...)`
- `Validate()`
- `HasGuild()`
- `RedactedToken()`

## `internal/bot`

### `NewWithScript(cfg, script, runtimeConfig)`
File: `internal/bot/bot.go:28`

Purpose:
- create one Discord session and one selected JavaScript bot host.

### `SyncCommands()`
File: `internal/bot/bot.go:102`

Purpose:
- bulk overwrite Discord application commands based on the bot descriptor.

## `internal/jsdiscord/runtime`

### `NewRegistrar(...)`
File: `internal/jsdiscord/runtime.go`

Purpose:
- inject the runtime-native `discord` module into the goja runtime.

### `RuntimeState.Loader(...)`
Purpose:
- expose `defineBot(...)` to JS code.

## `internal/jsdiscord/bot`

### `CompileBot(...)`
File: `internal/jsdiscord/bot.go:135`

Purpose:
- compile the JS-exported bot object into callable Go-side handles.

### `type DispatchRequest`
File: `internal/jsdiscord/bot.go:97`

Purpose:
- common transport object for commands, events, components, modals, autocomplete, and request-scoped Discord operations.

### `buildContext(...)`
File: `internal/jsdiscord/bot.go:942`

Purpose:
- shape the JS `ctx` object exposed to handlers.

### `discordOpsObject(...)`
File: `internal/jsdiscord/bot.go:1027`

Purpose:
- expose request-scoped host operations under `ctx.discord.*`.

## `internal/jsdiscord/host_dispatch`

### `DispatchInteraction(...)`
File: `internal/jsdiscord/host_dispatch.go:251`

Purpose:
- central entrypoint for Discord interactions, including command routing and reply lifecycle.

### `DispatchMessageCreate(...)`, `DispatchReactionAdd(...)`, etc.
Purpose:
- convert Discord gateway events into `DispatchRequest` values.

## `internal/jsdiscord/host_payloads`

### `normalizePayload(...)`
File: `internal/jsdiscord/host_payloads.go:184`

Purpose:
- turn JS-friendly object literals into Discordgo payload structs.

This file is especially important because it is where ergonomic JS API design collides with Discord’s stricter payload shapes.

## `internal/botcli`

### `DiscoverBots(...)`
File: `internal/botcli/bootstrap.go:113`

Purpose:
- find named bot scripts in a repository.

### `preparseRunArgs(...)`
File: `internal/botcli/run_schema.go:45`

Purpose:
- manually peel off static runner flags before handing the remaining args to dynamic Glazed parsing.

### `buildRunSchema(...)`
File: `internal/botcli/run_schema.go:198`

Purpose:
- translate bot metadata into a Glazed schema for runtime config.

# Why the quality review naturally centers on a few files

A new intern might ask: why not just review everything evenly?

Because this system is not evenly shaped.

The main stress points are concentrated in a few places:

## 1. `internal/jsdiscord/bot.go`
This is where several responsibilities converge:

- JS bot-definition DSL registration
- runtime compilation
- dispatch bridging
- promise settlement
- context construction
- logger binding
- in-memory store binding
- request-scoped Discord capability binding

That is too much conceptual weight for one file.

## 2. `internal/jsdiscord/host_payloads.go`
This is the normalization layer between friendly JS payloads and strict Discordgo payloads.

That boundary is valuable, but it is also where:

- shape conversion,
- option parsing,
- array/object switching,
- and many Discord-specific component details

all accumulate.

## 3. `internal/botcli/run_schema.go`
This file is not the biggest package hot spot, but it is one of the clearest **clarity** hot spots. It manually re-implements a parsing phase around Cobra/Glazed, which makes it harder for a newcomer to know which parser is authoritative.

## 4. `examples/discord-bots/knowledge-base/`
This example is important because it is effectively executable documentation for the JS API.

If the biggest example bot becomes repetitive or confusing, that is not only an example-bot problem. It is a signal that the API surface may need cleanup.

# What makes a good cleanup here

The repo does **not** need a giant rewrite.

The highest-quality cleanup would instead follow three rules:

## Rule 1: split by responsibility, not by abstraction fashion

Bad cleanup:
- making new packages just because files are large.

Good cleanup:
- extracting units that have distinct reasons to change.

For example:
- payload normalization helpers
- JS context/capability binding
- descriptor parsing
- event dispatch request construction

## Rule 2: keep the runtime boundary explicit

This codebase is successful when it stays clear about these boundaries:

- Discordgo structs on the Go side
- plain maps / simple objects at the JS boundary
- normalized payloads when crossing back to Discordgo

A cleanup that blurs those boundaries will make the system harder to debug.

## Rule 3: improve the canonical example when the API improves

The knowledge-base bot is not “just an example.” It is also the best signal for whether the JS authoring API is pleasant.

If a cleanup makes the runtime code better but leaves the example bot repetitive and heavy, the cleanup is incomplete.

# Suggested reading questions for the intern

When you read the code, keep these questions in mind.

## About `internal/jsdiscord`
- Which functions are shaping external API, and which are just doing conversion work?
- Which maps are effectively informal structs?
- Where are we returning no-op functions instead of explicit capability errors?
- Which responsibilities are grouped because they are truly related, and which ones are grouped only because they were added over time?

## About `internal/bot`
- Are event handlers intentionally repetitive, or are they repeating transport boilerplate?
- Is there dead or fallback code that no longer matches the single-bot architecture?

## About `internal/botcli`
- Which parser is authoritative: Cobra, Glazed, or the custom pre-parser?
- Do users see the same contract in help text that the code actually enforces?

## About example bots
- Does the largest example read like a bot authoring guide, or like framework internals leaked into user code?
- Are duplicate aliases and repetitive component handlers teaching good patterns or preserving accidental complexity?

# Recommended review order for the detailed report

After this guide, read the detailed report in this order:

1. runtime core and JS bridge issues
2. host dispatch and payload normalization issues
3. CLI parsing clarity issues
4. stale/deprecated artifact issues
5. knowledge-base example organization issues
6. phased cleanup plan

# Short conclusion

The repo already has a strong architecture idea:

- one Go host process,
- one selected JS bot,
- request-scoped Discord operations,
- and a rich local JS bot-definition surface.

The code quality challenge is not that the architecture is wrong.
The challenge is that several high-value files have become **accumulation points**.

That means the right next step is not invention. It is careful re-organization:

- smaller files around the live seams,
- clearer internal boundaries,
- removal of stale artifacts,
- and better example-bot ergonomics.
