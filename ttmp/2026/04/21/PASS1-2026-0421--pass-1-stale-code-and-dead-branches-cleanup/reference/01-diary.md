---
Title: Diary
Ticket: PASS1-2026-0421
Status: active
Topics:
    - code-quality
    - refactoring
    - cleanup
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/bot/bot.go
      Note: Commit 8f49893 removed fallback handler
    - Path: internal/jsdiscord/runtime.go
      Note: Commit 6cb680b removed unused surfaces
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---



# Diary

## Goal

Step-by-step implementation of Pass 1: stale code and dead branches cleanup. Three tasks, each committed independently.

---

## Step 1: Delete stale example artifacts

Removed four stale artifacts that had no consumers in the current codebase.

### Prompt Context

**User prompt (verbatim):** "Create a new docmgr ticket for each pass, and add an implementaiton guide and a list of tasks to cover the refactor. Then work on the first pass, task by task, committing at appropriate intervals and keeping a diary."

**Assistant interpretation:** Create 4 tickets (one per pass), write implementation guides, then implement Pass 1 task by task with commits and diary entries.

**Inferred user intent:** Turn the code quality recommendations into actionable, tracked work with proper documentation.

**Commit (code):** d51b4a3 — "chore: remove stale example artifacts"

### What I did
- Verified no code references to `examples/bots/`, `examples/bots-dupe-a/`, `examples/bots-dupe-b/` outside of the directories themselves
- Verified `register-knowledge-bot.js` had zero references
- Deleted all four artifacts
- Ran `go test ./...` — all passed
- Ran `go build ./...` — no errors
- Committed

### Why
These directories were orphaned by the transition from `__verb__`-based bots to `defineBot`-based bots. The discovery code (`looksLikeBotScript` in `bootstrap.go`) explicitly requires `defineBot` + `require("discord")`, which the old examples don't use.

### What worked
- `grep -r` confirmed zero external references before deletion
- Tests passed immediately after deletion

### What didn't work
- N/A

### What I learned
- The `bots-dupe-a` and `bots-dupe-b` directories appear to have been manual test fixtures, but the actual duplicate-name tests now use `testdata/discord-bots-dupe-name-*` instead.

### What was tricky to build
- N/A

### What warrants a second pair of eyes
- N/A — pure deletions, no logic changes

### What should be done in the future
- N/A

### Code review instructions
- Check `git show d51b4a3 --stat` to confirm only the expected files were deleted
- Verify `go test ./...` still passes

---

## Step 2: Remove dead fallback interaction handler

Removed 37 lines of unreachable fallback code in `handleInteractionCreate`.

**Commit (code):** 8f49893 — "chore: remove dead fallback interaction handler"

### What I did
- Located the fallback `ping`/`echo`/`unknown` handler in `internal/bot/bot.go`
- Confirmed it was unreachable: `NewWithScript` always loads a JS bot and sets `jsHost`
- The function returns early when `jsHost != nil`, making everything after `return` dead code
- Removed the entire fallback block
- Ran `go test ./...` — all passed
- Ran `go build ./...` — no errors
- Committed

### Why
The fallback was a development relic from before the JS bot runtime was mandatory. It created false confidence that the bot could function without a JS host.

### What worked
- The edit was a simple block deletion
- `go build` confirmed `fmt` and `strings` imports were still needed elsewhere in the file

### What didn't work
- N/A

### What I learned
- `handleInteractionCreate` went from ~45 lines to ~8 lines — much clearer intent

### What was tricky to build
- N/A

### What warrants a second pair of eyes
- N/A — pure deletion of unreachable code

### What should be done in the future
- N/A

### Code review instructions
- Check `git show 8f49893` to see the deleted block
- Verify the remaining `handleInteractionCreate` only dispatches to `jsHost`

---

## Step 3: Remove unused runtime lifecycle surfaces

Removed unused VM registration/unregistration/lookup code from `runtime.go`.

**Commit (code):** 6cb680b — "chore: remove unused runtime lifecycle surfaces"

### What I did
- Checked callers of `LookupRuntimeState` — zero callers outside `runtime.go`
- Checked callers of `RegisterRuntimeState`/`UnregisterRuntimeState` — only called from `RegisterRuntimeModules` inside `runtime.go`
- Checked if `RuntimeStateContextKey` is read anywhere — never read, only set
- Removed:
  - `var runtimeStateByVM sync.Map`
  - `func RegisterRuntimeState(...)`
  - `func UnregisterRuntimeState(...)`
  - `func LookupRuntimeState(...)`
  - The registration/unregistration calls in `RegisterRuntimeModules`
- Kept `RuntimeStateContextKey` with an explicit comment documenting it as an extension seam
- Removed unused `context` and `sync` imports
- Ran `go test ./...` — all passed
- Ran `go build ./...` — no errors
- Committed

### Why
These surfaces implied an API contract (`LookupRuntimeState` is exported) that had no consumers. A new intern might assume they were important and preserve them unnecessarily.

### What worked
- `grep -r` quickly confirmed zero external usage
- The file shrank from 147 lines to 103 lines (44% reduction)

### What didn't work
- Initial build failed because `context` and `sync` imports became unused — caught immediately by the compiler

### What I learned
- `ctx.SetValue(RuntimeStateContextKey, state)` stores the state in the engine context, but nothing currently reads it. The comment documents this as an intentional extension point.

### What was tricky to build
- Deciding whether to keep `RuntimeStateContextKey` or delete it. Chose to keep with a comment because it's part of the engine context API and could be useful for debugging/runtime inspection.

### What warrants a second pair of eyes
- Confirm that `ctx.SetValue` without a corresponding `GetValue` is acceptable. It is — the engine context is a generic key-value store.

### What should be done in the future
- If a runtime inspector or debugger is built, it can use `RuntimeStateContextKey` to retrieve the state.

### Code review instructions
- Check `git show 6cb680b` to see the removed functions
- Verify `go test ./...` passes
- Confirm `RuntimeStateContextKey` comment is accurate

---

## Final handoff summary

- **Ticket path:** `ttmp/2026/04/21/PASS1-2026-0421--pass-1-stale-code-and-dead-branches-cleanup/`
- **Commits:**
  - `d51b4a3` — Delete stale example artifacts
  - `8f49893` — Remove dead fallback interaction handler
  - `6cb680b` — Remove unused runtime lifecycle surfaces
- **Lines deleted:** 489 (examples) + 37 (bot.go) + 41 (runtime.go) = **567 lines**
- **Tests:** `go test ./...` passes after each commit
- **Build:** `go build ./...` passes
- **Next:** PASS2-2026-0421 (file-size cleanup)
