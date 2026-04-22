---
Title: Pass 1 Implementation Guide
Ticket: PASS1-2026-0421
Status: active
Topics:
    - code-quality
    - refactoring
    - cleanup
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/bot/bot.go
      Note: Removed dead fallback handler
    - Path: internal/jsdiscord/runtime.go
      Note: Removed unused lifecycle surfaces
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---



# Pass 1: Stale Code and Dead Branches Cleanup

## Executive Summary

Pass 1 is the lowest-risk cleanup pass. It removes stale artifacts, dead code, and unused API surfaces without changing any behavior. This pass should take approximately **1 hour** and consists of 3 tasks.

## Origin

This pass is derived from the CODEQUAL-2026-0421 code quality review, specifically findings:
- **5.4.2** — Stale example artifacts
- **5.1.4** — Dead fallback interaction code in `internal/bot/bot.go`
- **5.5.4** — Unused lifecycle surfaces in `internal/jsdiscord/runtime.go`

## Tasks

### Task 1: Delete stale example artifacts

**What to delete:**
- `examples/bots/` (old `__verb__`-based examples, orphaned)
- `examples/bots-dupe-a/` (duplicate)
- `examples/bots-dupe-b/` (duplicate)
- `examples/discord-bots/knowledge-base/lib/register-knowledge-bot.js` (unreferenced alternate registration)

**Verification:**
```bash
# Confirm these are not referenced anywhere
grep -r "bots-dupe" --include="*.go" --include="*.md" .
grep -r "register-knowledge-bot" --include="*.go" --include="*.md" --include="*.js" .
```

**Commit message:**
```
chore: remove stale example artifacts

Deletes orphaned example directories and an unreferenced JS registration
file. Current bot discovery (looksLikeBotScript in bootstrap.go) requires
`defineBot` + `require("discord")`, which the old examples do not use.

Refs: PASS1-2026-0421
```

---

### Task 2: Remove dead fallback interaction code

**What to remove:** In `internal/bot/bot.go`, within `handleInteractionCreate`, there is a fallback branch that handles hardcoded `ping` and `echo` commands when no JS bot is loaded:

```go
// Around lines 169–320 — look for:
if i.Type == discordgo.InteractionApplicationCommand {
    // fallback ping/echo handler
}
```

This code is dead because:
1. The system always loads a JS bot when running
2. The JS bot handles all interactions
3. The fallback is a development relic

**Verification:**
```bash
go test ./...
```

**Commit message:**
```
chore: remove dead fallback interaction handler

Removes the hardcoded ping/echo fallback in handleInteractionCreate.
This code was a development relic; all interactions are now handled
by the JS bot runtime.

Refs: PASS1-2026-0421
```

---

### Task 3: Document or delete unused runtime surfaces

**What to address:** In `internal/jsdiscord/runtime.go`, the following surfaces appear unused:

- `RuntimeStateContextKey` (line ~14)
- `runtimeStateByVM` (line ~16)
- `LookupRuntimeState` (line ~88)

**Decision:** After checking callers, if `LookupRuntimeState` has no callers outside `runtime.go`, choose one of:

**Option A — delete:**
```go
// Remove LookupRuntimeState, RuntimeStateContextKey, and runtimeStateByVM
// Keep only RegisterRuntimeState and UnregisterRuntimeState if they are used
```

**Option B — document:**
```go
// RuntimeStateContextKey and VM registration are retained as future extension
// seams for runtime-level inspectors. They are intentionally unused today.
const RuntimeStateContextKey = "discord.runtime"
```

**Verification:**
```bash
# Check for callers
grep -r "LookupRuntimeState" --include="*.go" .
grep -r "RuntimeStateContextKey" --include="*.go" .

go test ./...
```

**Commit message:**
```
chore: remove/document unused runtime lifecycle surfaces

LookupRuntimeState, RuntimeStateContextKey, and runtimeStateByVM had no
consumers outside runtime.go itself. [Choose: Removed them / Added an
explicit comment documenting them as extension seams.]

Refs: PASS1-2026-0421
```

## Risk Assessment

| Task | Risk | Mitigation |
|------|------|------------|
| Delete examples | Very low | Only deletes `examples/`; no production code | 
| Remove fallback handler | Low | `go test ./...` validates no test depends on it |
| Runtime surfaces | Low | Verify no callers with `grep` first |

## Success Criteria

- [ ] `examples/bots/`, `examples/bots-dupe-a/`, `examples/bots-dupe-b/` deleted
- [ ] `register-knowledge-bot.js` deleted
- [ ] `internal/bot/bot.go` fallback handler removed
- [ ] `internal/jsdiscord/runtime.go` unused surfaces addressed
- [ ] `go test ./...` passes after each commit
- [ ] `go build ./...` passes

## Related Tickets

- **CODEQUAL-2026-0421** — Parent code quality review
- **PASS2-2026-0421** — File-size cleanup (next pass)
