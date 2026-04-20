---
Title: Diary
Ticket: DISCORD-BOT-002
Status: active
Topics:
    - backend
    - chat
    - javascript
    - goja
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: examples/js-bots/ping.js
      Note: |-
        First example JS Discord bot script added in Step 1
        First example script for the local Discord JS API
    - Path: internal/bot/bot.go
      Note: Host integration with optional JavaScript bot script added in Step 1
    - Path: internal/config/config.go
      Note: Added the bot-script setting during the first local Discord JS host implementation step
    - Path: internal/jsdiscord/bot.go
      Note: Local bot-definition API, context, store, and async dispatch flow added in Step 1
    - Path: internal/jsdiscord/host.go
      Note: Live Discord host bridge added in Step 1
    - Path: internal/jsdiscord/runtime.go
      Note: Runtime-scoped local Discord JS module registrar added in Step 1
ExternalSources: []
Summary: Chronological diary for moving the sandbox here and implementing the first real JavaScript Discord bot API slice.
LastUpdated: 2026-04-20T14:13:00-04:00
WhatFor: Record why the local Discord JS host layer was added here and how the first implementation slice was validated.
WhenToUse: Use when reviewing or continuing the local JavaScript Discord bot runtime work.
---


# Diary

## Goal

Record the move of the sandbox-style JS host layer into `js-discord-bot` and the first real Discord-integrated JavaScript bot API implementation.

## Step 1: Create the move ticket and implement the first local Discord JS host slice

After clarifying that the sandbox is really a host-facing script API rather than core runtime machinery, I created a new ticket here and implemented the first real local Discord JS bot slice in this repository. The main goal was to put the application-facing JavaScript API where it belongs: next to the live Discord host, not in the shared engine repo.

The code path I chose is intentionally vertical instead of purely infrastructural. A JavaScript bot script can now be loaded by the live host, can define slash commands via `require("discord")`, can sync those commands to Discord, and can handle live slash-command interactions plus the `ready` event.

### Prompt Context

**User prompt (verbatim):** "ok, the sandbox needs to move to this repo as well, so we can basically leave go-go-goja as is . Create a ticket to move it all over, and start implementing a proper real discord implementation of the bot API"

**Assistant interpretation:** Create a new ticket in `js-discord-bot`, move the sandbox-style JS host layer here, and start implementing a real Discord-specific JavaScript bot API integrated with the existing live host.

**Inferred user intent:** Keep `go-go-goja` as a reusable dependency repo while making this repository the true home of the application-facing JavaScript Discord bot API.

**Commit (code):** d75ace3 — "Add local JavaScript Discord bot host API"

### What I did
- Created ticket `DISCORD-BOT-002`.
- Added a design doc, API reference, and diary for the move-and-implement work.
- Added `internal/jsdiscord/` with:
  - `runtime.go`
  - `store.go`
  - `bot.go`
  - `host.go`
  - `runtime_test.go`
- Exposed a local CommonJS entrypoint via `require("discord")`.
- Ported the runtime-local store, bot-definition DSL, dispatch path, and async Promise settlement from the earlier sandbox concept into this repo.
- Integrated the live host in `internal/bot/bot.go` so `DISCORD_BOT_SCRIPT` / `bot-script` can load a JS bot script.
- Added `examples/js-bots/ping.js` as the first example script.
- Added `DISCORD_BOT_SCRIPT` to `.envrc.example` and the CLI config surface.
- Validated with focused tests and a full repository `go test` run.

### Why
- The sandbox/API layer belongs with the application host that actually uses it.
- The live Discord integration is easier to reason about when the JS API and host bridge are in the same repository.
- This keeps `go-go-goja` closer to its intended role as engine/runtime dependency code.

### What worked
- The local `internal/jsdiscord` package compiled cleanly.
- The tests covering async settlement and bot compilation passed.
- The live host now has a clear optional path for loading a JS bot script.
- Full validation passed with `GOWORK=off go test ./...`.

### What didn't work
- While writing the local host layer, I initially placed a temporary example script under `internal/jsdiscord/example.js`, which was the wrong home for a repo-facing example. I removed it and kept the canonical example under `examples/js-bots/ping.js`.
- There were also a couple of small implementation nits that surfaced during compile/test cleanup, such as import cleanup and making the command snapshot conversion robust for the exported describe shape.

### What I learned
- The earlier sandbox design ports over cleanly when treated as a local package rather than a global reusable module.
- The right abstraction boundary is now much clearer: generic runtime in `go-go-goja`, app-facing Discord API in this repo.
- A narrow first slice (ping/echo/ready + simple response payloads) is enough to prove the architecture end-to-end.

### What was tricky to build
- The trickiest part was keeping the new JS API genuinely Discord-specific without prematurely over-modeling the full Discord interaction surface. The symptom of overreach would have been a large, hard-to-validate response schema and too much host complexity before the first live round-trip worked.
- I approached that by keeping the first response contract intentionally small (`string`, `{ content }`, `{ content, ephemeral }`, `ctx.reply`, `ctx.defer`) while still wiring the actual live Discord path.

### What warrants a second pair of eyes
- The exact long-term JS API naming: whether `require("discord")` should stay this small or grow richer helper namespaces.
- The response payload normalization path in `internal/jsdiscord/host.go`.
- Whether more Discord events should be exposed before the response payload contract grows further.

### What should be done in the future
- Add richer reply payload support.
- Add a script-inspection or dry-run command surface.
- Reconcile the old sandbox code in `go-go-goja` now that this repo is the better source of truth for the bot API.

### Code review instructions
- Start with `internal/jsdiscord/runtime.go`, `internal/jsdiscord/bot.go`, and `internal/jsdiscord/host.go`.
- Then review `internal/bot/bot.go`, `internal/config/config.go`, and `cmd/discord-bot/commands.go`.
- Validate with:
  - `GOWORK=off go test ./internal/jsdiscord ./internal/bot ./cmd/discord-bot`
  - `GOWORK=off go test ./...`

### Technical details
- New env/config surface: `DISCORD_BOT_SCRIPT` / `bot-script`.
- JS entrypoint: `require("discord")`.
- Example script: `examples/js-bots/ping.js`.
- The first supported JS-to-Discord response shapes are intentionally minimal.

## Related

- `design-doc/01-sandbox-move-and-discord-javascript-api-architecture-guide.md`
- `reference/01-discord-javascript-bot-api-reference-and-example-script.md`
