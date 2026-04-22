---
Title: Tasks
Ticket: PASS1-2026-0421
---

# Tasks

## Task 1: Delete stale example artifacts
- [x] Verify no references to `examples/bots/`, `examples/bots-dupe-a/`, `examples/bots-dupe-b/` in code or docs
- [x] Verify `register-knowledge-bot.js` is unreferenced
- [x] Delete `examples/bots/`
- [x] Delete `examples/bots-dupe-a/`
- [x] Delete `examples/bots-dupe-b/`
- [x] Delete `examples/discord-bots/knowledge-base/lib/register-knowledge-bot.js`
- [x] Run `go test ./...` and `go build ./...`
- [x] Commit (d51b4a3)

## Task 2: Remove dead fallback interaction code
- [x] Locate fallback handler in `internal/bot/bot.go` (`handleInteractionCreate` ping/echo branch)
- [x] Confirm no tests depend on it
- [x] Remove the fallback code
- [x] Run `go test ./...` and `go build ./...`
- [x] Commit (8f49893)

## Task 3: Document or delete unused runtime surfaces
- [x] Check callers of `LookupRuntimeState` across codebase — none found
- [x] Check callers of `RegisterRuntimeState`/`UnregisterRuntimeState` — only internal
- [x] Check if `runtimeStateByVM` is read outside `runtime.go` — no
- [x] Decide: delete unused surfaces, keep context key with comment
- [x] Apply the change
- [x] Run `go test ./...` and `go build ./...`
- [x] Commit (6cb680b)
