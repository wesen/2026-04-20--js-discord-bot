---
Title: Diary
Ticket: DISCORD-BOT-003
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
    - Path: internal/bot/bot.go
      Note: Live host integration with named bot selections is tracked here
    - Path: internal/botcli/command.go
      Note: Command surface rewrite is tracked here
    - Path: internal/botcli/command_test.go
      Note: Expanded in Step 1 to cover list/help/run behavior for named bots
    - Path: internal/jsdiscord/host.go
      Note: Multi-bot runtime composition work is tracked here
    - Path: internal/jsdiscord/multihost_test.go
      Note: Expanded in Step 1 to cover duplicate names and duplicate command rejection
ExternalSources: []
Summary: Chronological diary for the named bot repository runner work.
LastUpdated: 2026-04-20T14:55:00-04:00
WhatFor: Record how the CLI moved from function-oriented verbs to named bot implementations and how the example bot repositories were added.
WhenToUse: Use when reviewing or continuing the bot repository runner implementation.
---


# Diary

## Goal

Record the work needed to turn `discord-bot bots ...` into a named bot implementation runner.

## Step 1: Replace the verb-oriented model with named bot discovery, multi-bot composition, and example repositories

I started this ticket by keeping the design package and tasks concrete, then I implemented the whole runner in the same order the design described: discover named bot implementations, compose multiple selected bots into one runtime, rewrite the CLI surface, and add example repositories plus validation tests. The important conceptual shift is that the CLI no longer thinks in terms of JS verbs or bot-internal commands; it now thinks in terms of whole bot implementations such as `knowledge-base`, `support`, or `moderation`.

This implementation also made the runtime model honest. Multiple selected bots can now share one live Discord session while keeping isolated JS runtimes and stores. Non-command events fan out to all selected bots, while slash commands route to the bot that owns the command name. That keeps the operator UX simple while still protecting against ambiguous command ownership.

### Prompt Context

**User prompt (verbatim):** "ok, create a new ticket and an implementation guide and a detailed list of tasks. then go task by tsk and implement. please build a whole set of example bots as well to exercise the functionality we have implemented so far."

**Assistant interpretation:** Create a new ticket for the named bot repository runner, document the architecture and tasks first, then implement the discovery/runtime/CLI/example-bot work in phases and keep the diary current.

**Inferred user intent:** Replace the old function-oriented `bots` abstraction with a real bot-implementation runner and provide enough example bots to exercise the Discord JS host functionality already built in this repo.

**Commit (code):** 327c8d1 — "Run named bot implementations from repositories"

### What I did
- Created ticket `DISCORD-BOT-003` with:
  - implementation guide
  - CLI/reference doc
  - detailed task list
  - diary
- Added `internal/jsdiscord/descriptor.go` so bot scripts can be inspected through `describe()` and normalized into stable bot descriptors.
- Added `internal/jsdiscord/multihost.go` plus tests so one process can load multiple bot implementations, reject duplicate slash-command names, fan out non-command events, and route slash commands to the owning bot.
- Rewrote `internal/botcli/` around named bot discovery instead of jsverbs function discovery.
- Added `discord-bot bots list`, `discord-bot bots help <bot>`, and `discord-bot bots run <bot...>`.
- Added `--sync-on-start` plus direct Discord config flags/env handling to `bots run`.
- Updated `internal/bot/bot.go` so the live host can load multiple selected bot scripts through `jsdiscord.MultiHost`.
- Added a full example bot repository under `examples/discord-bots` with:
  - `knowledge-base`
  - `support`
  - `moderation`
  - `announcements`
- Added duplicate-name fixtures under `testdata/discord-bots-dupe-name-*`.
- Validated with:
  - `GOWORK=off go test ./internal/botcli ./internal/jsdiscord ./internal/bot ./cmd/discord-bot`
  - `GOWORK=off go test ./...`
  - `GOWORK=off go run ./cmd/discord-bot bots list --bot-repository ./examples/discord-bots`
  - `GOWORK=off go run ./cmd/discord-bot bots help knowledge-base --bot-repository ./examples/discord-bots`

### Why
- The old CLI abstraction answered the wrong question. Operators want to run whole bot implementations, not individual JS functions inside those bots.
- Repository discovery should reuse the real local Discord JS bot API contract instead of introducing a second metadata system.
- A full example repository is necessary to keep the new runner grounded in realistic bot packages rather than synthetic fixtures only.

### What worked
- Descriptor-based discovery through the existing `describe()` contract worked well and avoided duplicated metadata plumbing.
- Multi-bot composition stayed manageable once command ownership and event fan-out were made explicit.
- The new CLI smoke commands worked directly against the example repository.
- Full repository tests passed after the rewrite.

### What didn't work
- The first compile/test pass failed because of an unused import left in the rewritten `internal/botcli/command.go`:

  `internal/botcli/command.go:7:2: "strings" imported and not used`

- A second discovery bug showed up in `TestFallbackBotNameUsesDirectoryForIndex`, where a missing metadata field was rendered as `"<nil>"` instead of falling back to the directory-derived bot name. The failure was:

  `multihost_test.go:67: name = "<nil>"`

- I fixed that by adding a small `mapString(...)` helper in `internal/jsdiscord/descriptor.go` so missing metadata values normalize to empty strings instead of `"<nil>"`.

### What I learned
- The existing local Discord JS host API was already the right foundation; the missing layer was discovery/composition, not another scripting contract.
- The repository runner becomes much easier to reason about when bot ownership and command ownership are treated as separate concepts.
- Root-level `.js` plus `index.js` packages are enough to give bot authors flexibility without making discovery rules too magical.

### What was tricky to build
- The trickiest part was changing the mental model without dragging the old jsverbs assumptions along. It is very easy to keep thinking in terms of “call a command inside a bot” instead of “run a bot implementation.” The symptoms were all in the previous CLI structure: verb-like names, function-style resolution, and a runtime model centered on invocation instead of long-lived ownership.
- The fix was architectural, not cosmetic. I replaced the descriptor type, the discovery logic, and the runner behavior together so the code now reflects the intended operator model end-to-end.

### What warrants a second pair of eyes
- The repository discovery heuristics in `internal/botcli/bootstrap.go`, especially the rule that top-level `.js` files and `index.js` package entrypoints are discoverable.
- The duplicate command rejection path in `internal/jsdiscord/multihost.go`.
- The CLI UX around `--sync-on-start` and whether it should remain optional or eventually become the default for named bot runs.

### What should be done in the future
- Add richer option validation/normalization for more Discord slash-command option shapes.
- Add a dedicated inspect/dry-run playbook or command surface for bot repositories.
- Clean up the now-obsolete bot/jsverbs-oriented leftovers if any remain elsewhere in this repo or `go-go-goja`.

### Code review instructions
- Start with `internal/jsdiscord/descriptor.go` and `internal/jsdiscord/multihost.go`.
- Then review `internal/botcli/bootstrap.go`, `internal/botcli/command.go`, and `internal/botcli/command_test.go`.
- Then review `internal/bot/bot.go` and the example repository under `examples/discord-bots`.
- Validate with:
  - `GOWORK=off go test ./internal/botcli ./internal/jsdiscord ./internal/bot ./cmd/discord-bot`
  - `GOWORK=off go test ./...`
  - `GOWORK=off go run ./cmd/discord-bot bots list --bot-repository ./examples/discord-bots`
  - `GOWORK=off go run ./cmd/discord-bot bots help knowledge-base --bot-repository ./examples/discord-bots`

### Technical details
- Discoverable bot scripts are:
  - root-level `.js` files
  - `index.js` package entrypoints under subdirectories
- Command routing is single-owner; event dispatch is fan-out.
- Multi-bot command-name collisions are rejected at composition time.
