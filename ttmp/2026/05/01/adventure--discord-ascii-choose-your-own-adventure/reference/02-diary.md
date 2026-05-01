---
Title: Diary
Ticket: adventure
Status: active
Topics:
    - discord
    - game
    - adventure
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: examples/discord-bots/adventure/README.md
      Note: Run instructions for adventure bot (commit 3feb8ef)
    - Path: examples/discord-bots/adventure/index.js
      Note: Adventure bot command/component/modal entrypoint (commit 3feb8ef)
    - Path: examples/discord-bots/adventure/lib/engine.js
      Note: Adventure state advancement and LLM orchestration (commit 3feb8ef)
    - Path: examples/discord-bots/adventure/lib/render.js
      Note: ASCII scene card renderer and UI payloads (commit 3feb8ef)
    - Path: examples/discord-bots/adventure/lib/schema.js
      Note: JSON parsing and validation for LLM contracts (commit 3feb8ef)
    - Path: examples/discord-bots/adventure/lib/store.js
      Note: SQLite store for seeds
    - Path: internal/jsdiscord/host.go
      Note: Registers OpenRouter module in default JS bot host (commit f9a47bf)
    - Path: internal/jsdiscord/openrouter_module.go
      Note: Go-owned OpenRouter adapter exposed to JS as adventure_llm.completeJson (commit f9a47bf)
    - Path: internal/jsdiscord/openrouter_module_test.go
      Note: Focused tests for OpenRouter adapter request/response behavior (commit f9a47bf)
    - Path: pkg/botcli/runtime_factory.go
      Note: Registers OpenRouter module in botcli-created runtimes (commit f9a47bf)
ExternalSources: []
Summary: Implementation diary for the adventure Discord bot milestone.
LastUpdated: 2026-05-01T13:33:30-07:00
WhatFor: Track implementation steps, commits, validation commands, failures, and review notes for the adventure ticket.
WhenToUse: Use when reviewing or continuing the adventure implementation.
---



# Diary

## Goal

Capture the implementation of Milestone 1 for the adventure Discord bot: a Go-owned OpenRouter adapter plus a JS framework bot that uses SQLite-defined scenes, Discord buttons/modals, and structured JSON LLM contracts.

## Step 1: Add Go-Owned OpenRouter Adapter

This step implemented the narrow host-side LLM adapter that JavaScript can call without controlling provider details. The new `adventure_llm` module exposes `completeJson(input)`, while Go/process configuration owns the OpenRouter API key, base URL, model, max tokens, temperature, and request headers.

The implementation is intentionally small: JS supplies prompt text and lightweight metadata, and Go returns a safe `{ ok, text, provider, usage }` result or a safe error object. This establishes the boundary requested before building the JS adventure app.

### Prompt Context

**User prompt (verbatim):** "OK looks good lets try it. Build openrouter adapter first, then the JS app. Keep a diary and commit as you go"

**Assistant interpretation:** Implement the OpenRouter host adapter as the first focused commit, then continue toward the JS app, while recording a docmgr diary and committing incremental work.

**Inferred user intent:** Keep provider secrets/model controls in Go, unblock JS adventure development with a narrow LLM function, and maintain an auditable implementation trail.

**Commit (code):** f9a47bfb59258055389b7ab1a5d71d8b7183dd2e — "Add OpenRouter LLM module"

### What I did

- Added `internal/jsdiscord/openrouter_module.go`.
- Registered `require("adventure_llm")` in default JS bot hosts via `internal/jsdiscord/host.go`.
- Registered the same module in `pkg/botcli/runtime_factory.go` for botcli-created runtimes.
- Added `internal/jsdiscord/openrouter_module_test.go` with an `httptest` OpenRouter-compatible server.
- Verified JS cannot override model/max tokens through the call payload; tests assert the request uses Go/env-owned values.

### Why

- The JS runtime has no `fetch`, `http`, `process`, npm packages, or file access.
- OpenRouter calls require HTTP and secret-bearing headers, which should stay in Go.
- The adventure JS app needs a simple LLM boundary before implementing prompts and scene generation.

### What worked

- `go test ./internal/jsdiscord -run TestOpenRouter -count=1` passed.
- The adapter returns safe, Discord-displayable error objects when the API key is missing.
- The adapter sends OpenRouter chat completion requests to `/chat/completions` with `Authorization`, `Content-Type`, and `X-Title` headers.

### What didn't work

- Broader test commands exposed pre-existing/ambient failures around example bots requiring modules:

```text
go test ./internal/jsdiscord -count=1
--- FAIL: TestKnowledgeBaseBotUsesSQLiteStoreForCaptureSearchAndReview (0.00s)
    knowledge_base_runtime_test.go:14: require bot script: GoError: Invalid module at github.com/dop251/goja_nodejs/require.(*RequireModule).require-fm (native)
...
FAIL github.com/go-go-golems/discord-bot/internal/jsdiscord
```

```text
go test ./pkg/framework ./pkg/botcli ./cmd/discord-bot -count=1
ok   github.com/go-go-golems/discord-bot/pkg/framework 0.302s
--- FAIL: TestNewBotsCommandExposesRunHelp (0.01s)
    command_test.go:26: Received unexpected error:
    discover bots: inspect bot script /Users/kball/git/go-go-golems/discord-bot/examples/discord-bots/knowledge-base/index.js: load js bot script: GoError: Invalid module at github.com/dop251/goja_nodejs/require.(*RequireModule).require-fm (native)
FAIL github.com/go-go-golems/discord-bot/pkg/botcli
FAIL github.com/go-go-golems/discord-bot/cmd/discord-bot
```

The focused OpenRouter tests pass; the broader failures appear tied to existing module loading for example bots rather than the OpenRouter adapter itself.

### What I learned

- The JS framework can consume Go-native modules through `engine.RuntimeModuleRegistrar` and `reg.RegisterNativeModule(...)`.
- Returning a plain object from Go is enough for JS `await` usage, even though the current implementation is synchronous/blocking inside the runtime call.
- Keeping the module result shape simple will make the JS app easier to test and degrade gracefully.

### What was tricky to build

- The API boundary needed to avoid exposing model/token/provider knobs to JS while still allowing tests to control Go-owned settings. I used process environment variables for Go-owned configuration so the settings remain outside `ctx.config` and outside the JS call payload.
- The adapter returns safe errors instead of throwing, because Discord handlers can display or log those errors without tripping the interaction response lifecycle.

### What warrants a second pair of eyes

- Whether the adapter should be synchronous as implemented or should return a Goja promise / run HTTP off the runtime owner goroutine.
- Whether environment variable names are the right Go-owned configuration surface.
- Whether `OPENROUTER_MODEL` should remain overrideable by process env or be hard-pinned to Haiku for the prototype.
- Whether `adventure_llm` should become a generic `llm` module later.

### What should be done in the future

- Add JS app integration using `require("adventure_llm").completeJson(...)`.
- Consider adding JSON-mode / provider-specific response format options once OpenRouter model compatibility is verified.
- Add metrics/logging around latency and token usage.

### Code review instructions

- Start with `internal/jsdiscord/openrouter_module.go`, especially `OpenRouterRegistrar`, `openRouterLoader`, and `(*openRouterClient).complete`.
- Check default registration in `internal/jsdiscord/host.go` and `pkg/botcli/runtime_factory.go`.
- Validate with:

```bash
go test ./internal/jsdiscord -run TestOpenRouter -count=1
```

### Technical details

JS-facing call shape:

```js
const llm = require("adventure_llm")
const result = await llm.completeJson({
  purpose: "scene_patch",
  system: "Return JSON only.",
  user: "Generate the next scene.",
  metadata: { sessionId, turn }
})
```

Go-owned configuration:

- `OPENROUTER_API_KEY`
- `OPENROUTER_BASE_URL` defaulting to `https://openrouter.ai/api/v1`
- `OPENROUTER_MODEL` defaulting to `anthropic/claude-3.5-haiku`
- `OPENROUTER_MAX_TOKENS` defaulting to `1200`
- `OPENROUTER_TEMPERATURE` defaulting to `0.7`
- `OPENROUTER_HTTP_REFERER`
- `OPENROUTER_APP_TITLE` defaulting to `discord-bot-adventure`

## Step 2: Add JS Adventure Bot Prototype

This step built the first playable JavaScript adventure bot on top of the existing Discord framework. The bot is repo-discovered, declares slash commands/components/modals, stores seeds/sessions/scenes/choices/audit rows in SQLite, and calls the Go-owned `adventure_llm` module for JSON scene generation and free-form action interpretation.

The prototype keeps dynamic component IDs intentionally simple because current component dispatch is exact-match based. To still protect against stale clicks, rendered scene content includes the turn number and component handlers compare the clicked message's turn against the canonical session turn before advancing.

### Prompt Context

**User prompt (verbatim):** "I updated the underlying code. Try again"

**Assistant interpretation:** Re-run validation after framework dependency/code updates, then continue implementing the JS adventure app.

**Inferred user intent:** Confirm the framework now supports the needed modules and complete the first JS milestone after the Go OpenRouter adapter.

**Commit (code):** 3feb8efcc0adbfdfac3751b93b82412907e89530 — "Add adventure Discord bot prototype"

### What I did

- Added `examples/discord-bots/adventure/index.js` as a repo-discovered `defineBot` entrypoint.
- Added `lib/seeds.js` with the `haunted-gate` seed as JS data.
- Added `lib/store.js` with SQLite schema and persistence for seeds, sessions, scenes, choices, and audit logs.
- Added `lib/schema.js` with JSON extraction/parsing and validators for scene patches and interpreted free-form actions.
- Added `lib/prompts.js` for scene and free-form interpretation prompts.
- Added `lib/llm.js` as the JS adapter around `require("adventure_llm").completeJson(...)`.
- Added `lib/engine.js` for session advancement, effect clamping/application, LLM orchestration, and audit writes.
- Added `lib/render.js` for ASCII scene cards, buttons, state output, and errors.
- Added `examples/discord-bots/adventure/README.md` with run instructions and OpenRouter configuration notes.

### Why

- The ticket design moved from YAML files to SQLite-defined scenes because the JS runtime has database support but no filesystem/YAML implementation.
- JS can own Discord orchestration and SQLite state while Go owns OpenRouter/provider access.
- A small prototype validates the full command/button/modal loop before hardening.

### What worked

- Bot discovery/help now succeeds:

```bash
go run ./cmd/discord-bot bots help adventure --bot-repository ./examples/discord-bots --output json
```

- Focused OpenRouter tests still pass:

```bash
go test ./internal/jsdiscord -run TestOpenRouter -count=1
```

- Broader tests pass after the underlying framework update:

```bash
go test ./pkg/framework ./pkg/botcli ./cmd/discord-bot -count=1
go test ./... -count=1
```

### What didn't work

- Before the user's framework update, bot discovery failed with:

```text
error building command: discover bots: inspect bot script /Users/kball/git/go-go-golems/discord-bot/examples/discord-bots/adventure/index.js: load js bot script: GoError: Invalid module at github.com/dop251/goja_nodejs/require.(*RequireModule).require-fm (native)
exit status 1
```

- After the update, the same discovery command succeeded.

### What I learned

- The current component registration model is exact-match, so the prototype uses fixed IDs like `adv:choice:0` rather than dynamic `session:turn` IDs.
- Stale click protection can still be approximated by parsing the turn number from the clicked message content and comparing it with canonical SQLite session state.
- `require("database")` is enough for a durable JS-owned store with JSON columns.

### What was tricky to build

- The LLM output needs to be treated as untrusted text even though the prompt asks for JSON. `schema.js` extracts fenced or raw JSON, parses it, normalizes choices, clamps lengths, and returns validation errors instead of throwing.
- Discord interaction response flow differs between slash commands, components, and modals. The bot uses `ctx.defer`/`ctx.edit` around LLM work and direct returns for lightweight state/error responses.
- Since JS does not receive the model/provider settings, `lib/llm.js` only handles request/result shape and never attempts to configure OpenRouter.

### What warrants a second pair of eyes

- Stale component handling: parsing `Turn N` from message content is pragmatic but brittle. A framework-level wildcard/dynamic component routing feature would be cleaner.
- The OpenRouter Go call is synchronous from JS's perspective; long LLM latency could block the runtime owner.
- SQLite schema uses JSON text columns heavily; this is flexible but less queryable than fully normalized state.
- The first prototype has no automated JS-level tests yet.

### What should be done in the future

- Add runtime tests for the adventure bot dispatch flow using mocked `adventure_llm` responses.
- Consider wildcard component/modal routing or custom ID prefixes in the framework.
- Add explicit JSON response-format support in the OpenRouter adapter if supported by the selected model/provider.
- Add group/voting mode and thread-based sessions.

### Code review instructions

- Start with `examples/discord-bots/adventure/index.js` to understand the command/component/modal flow.
- Review `examples/discord-bots/adventure/lib/store.js` for persistence/schema correctness.
- Review `examples/discord-bots/adventure/lib/engine.js` and `lib/schema.js` for validation and state transition safety.
- Validate with:

```bash
go run ./cmd/discord-bot bots help adventure --bot-repository ./examples/discord-bots --output json
go test ./... -count=1
```

### Technical details

Core runtime command:

```bash
export DISCORD_BOT_TOKEN=...
export DISCORD_APPLICATION_ID=...
export DISCORD_GUILD_ID=...
export OPENROUTER_API_KEY=...

GOWORK=off go run ./cmd/discord-bot bots adventure run \
  --bot-repository ./examples/discord-bots \
  --sync-on-start \
  --session-db-path ./examples/discord-bots/adventure/data/adventure.sqlite
```

Primary files:

- `examples/discord-bots/adventure/index.js`
- `examples/discord-bots/adventure/lib/store.js`
- `examples/discord-bots/adventure/lib/engine.js`
- `examples/discord-bots/adventure/lib/schema.js`
- `examples/discord-bots/adventure/lib/render.js`
