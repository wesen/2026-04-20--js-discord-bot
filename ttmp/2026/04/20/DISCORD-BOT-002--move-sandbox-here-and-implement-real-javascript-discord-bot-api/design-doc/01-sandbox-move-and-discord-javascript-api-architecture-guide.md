---
Title: Sandbox Move and Discord JavaScript API Architecture Guide
Ticket: DISCORD-BOT-002
Status: active
Topics:
    - backend
    - chat
    - javascript
    - goja
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/jsdiscord/runtime.go
      Note: Local runtime-scoped registrar and CommonJS module loader
    - Path: internal/jsdiscord/bot.go
      Note: Builder API, bot compilation, dispatch, and JS context construction
    - Path: internal/jsdiscord/host.go
      Note: Discord-specific application-command syncing and interaction dispatch
    - Path: internal/bot/bot.go
      Note: Live Discord host integration point for the JavaScript bot runtime
ExternalSources: []
Summary: Architecture guide for moving the sandbox-style JS host layer into js-discord-bot and evolving it into a real Discord bot API.
LastUpdated: 2026-04-20T14:13:00-04:00
WhatFor: Explain why the sandbox belongs here, how the new local Discord JS API works, and what was implemented first.
WhenToUse: Use when extending or reviewing the JavaScript-hosted Discord runtime.
---

# Sandbox Move and Discord JavaScript API Architecture Guide

## Executive Summary

The generic runtime and jsverbs layers belong in `go-go-goja`, but the application-facing JavaScript bot API belongs in `js-discord-bot`. This ticket moves the sandbox-style host functionality into this repository and reshapes it into a real Discord-specific CommonJS API exposed as `require("discord")`.

The first implementation slice is intentionally vertical: a JavaScript bot script can now define slash commands and a `ready` event, the host can sync those commands to Discord, and live interactions can dispatch into JavaScript handlers while the Go host still owns the Discord session, credentials, and lifecycle.

## Problem Statement

The earlier sandbox work in `go-go-goja` proved that a runtime-scoped JS host API was feasible, but it left an architectural ambiguity: was `go-go-goja` meant to be a shared runtime library, or the place where app-specific bot APIs lived? For this project, the answer is now clear.

This repository should own:

- the real Discord-flavored JS entrypoint
- the script-visible bot API
- the example scripts
- the live Discord gateway integration

`go-go-goja` should remain the reusable dependency that provides the engine, runtime ownership, `require()` support, and jsverbs.

## Proposed Solution

Implement a local `internal/jsdiscord` package with four responsibilities:

1. **runtime-scoped module registration**
   - register `require("discord")` per runtime
2. **bot-definition DSL**
   - `defineBot`, `command`, `event`, `configure`
3. **runtime-local state + context**
   - `ctx.args`, `ctx.reply`, `ctx.defer`, `ctx.log`, `ctx.store`
4. **Discord host integration**
   - sync application commands from JS metadata
   - dispatch live slash commands and `ready` events into JS

## Design Decisions

### 1. Keep the JS API here, not in `go-go-goja`

This repo is the application. It should own the application-facing API.

### 2. Use a dedicated `require("discord")` entrypoint

This is clearer than a generic `sandbox` name now that the host is concretely Discord-specific.

### 3. Keep the host capability-based

Go still owns:

- Discord session lifecycle
- credentials
- command sync API calls
- socket ownership
- process shutdown

JavaScript owns:

- command behavior
- event behavior
- small in-memory state
- reply/defer decisions

### 4. Keep the first response shape intentionally small

The first implementation supports:

- string return values
- `{ content: string }`
- `{ content: string, ephemeral: true }`
- explicit `ctx.reply(...)`
- explicit `ctx.defer()`

That is enough to prove the full path end-to-end without prematurely modeling every Discord response feature.

## Architecture

```mermaid
flowchart TD
    Operator[Operator] --> CLI[discord-bot CLI]
    CLI --> Host[internal/bot]
    Host --> DiscordGo[discordgo session]
    Host --> JSHost[internal/jsdiscord.Host]
    JSHost --> Engine[go-go-goja engine runtime]
    Engine --> Module[require("discord")]
    Module --> Script[JS bot script]
    DiscordGo --> Host
    Host --> JSHost
    JSHost --> Script
```

## Current First-Slice Behavior

A JS file can now do this:

```js
const { defineBot } = require("discord")

module.exports = defineBot(({ command, event, configure }) => {
  configure({ name: "ping-bot" })

  command("ping", {
    description: "Reply with pong from JavaScript"
  }, async () => {
    return { content: "pong" }
  })

  command("echo", {
    description: "Echo text back from JavaScript",
    options: {
      text: {
        type: "string",
        description: "Text to echo back",
        required: true,
      }
    }
  }, async (ctx) => {
    return { content: ctx.args.text }
  })

  event("ready", async (ctx) => {
    ctx.log.info("js discord bot connected", { user: ctx.me && ctx.me.username })
  })
})
```

## Implementation Plan

### Phase 1 — local runtime package
- copy the sandbox-style runtime-local store and dispatch ideas into `internal/jsdiscord`
- keep async Promise settlement
- keep runtime-local state per VM

### Phase 2 — real Discord host integration
- add `DISCORD_BOT_SCRIPT` / `bot-script`
- load the JS bot during host construction
- derive slash command metadata from JS command specs
- dispatch interactions and `ready` into JS handlers

### Phase 3 — richer Discord surface
- richer response payloads
- more command option types
- more events
- maybe script inspection / dry-run CLI commands

## Alternatives Considered

### Keep the generic `sandbox` API in `go-go-goja`

Rejected for this app-facing surface. It keeps the wrong ownership boundary and makes the shared repo feel more application-specific than it should.

### Keep all bot behavior hard-coded in Go

Rejected because it blocks the main project goal: authoring bot behavior in JavaScript while Go retains host ownership.

## Review Notes

Start review with:

- `internal/jsdiscord/runtime.go`
- `internal/jsdiscord/bot.go`
- `internal/jsdiscord/host.go`
- `internal/bot/bot.go`

Validate with:

```bash
GOWORK=off go test ./internal/jsdiscord ./internal/bot ./cmd/discord-bot
GOWORK=off go test ./...
```
