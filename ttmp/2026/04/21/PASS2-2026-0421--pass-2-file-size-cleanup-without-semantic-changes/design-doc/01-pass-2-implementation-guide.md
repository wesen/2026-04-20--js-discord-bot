---
Title: Pass 2 Implementation Guide
Ticket: PASS2-2026-0421
Status: active
Topics:
    - code-quality
    - refactoring
    - cleanup
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/jsdiscord/bot_compile.go
      Note: Split from bot.go
    - Path: internal/jsdiscord/payload_model.go
      Note: Split from host_payloads.go
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---



# Pass 2: File-size Cleanup Without Semantic Changes

## Executive Summary

Pass 2 splits the three largest files in the codebase into smaller, responsibility-focused files. These are **pure move refactors** — no logic changes, no behavior changes, no API changes. The goal is to make the code navigable for new interns and reviewers.

Estimated time: **~5 hours**.

## Origin

Derived from CODEQUAL-2026-0421 findings:
- **5.1.1** — `internal/jsdiscord/bot.go` at 1,293 lines
- **5.1.2** — `internal/jsdiscord/host_payloads.go` at 736 lines
- **5.2.4** — `internal/jsdiscord/runtime_test.go` at 1,205 lines

## Task 1: Split `internal/jsdiscord/bot.go`

**Current:** 1,293 lines. Contains:
- `BotHandle` struct and methods
- `botDraft` struct and `finalize`
- `CompileBot`
- `DispatchRequest` and context builders
- `storeObject` and KV store bridge
- `DiscordOps` builder
- `loggerObject` and `applyFields`
- `waitForPromise` / `settleValue`

**Target layout:**
```text
internal/jsdiscord/
  bot_compile.go    // CompileBot, botDraft, finalize, bot registration
  bot_dispatch.go   // BotHandle.dispatch, settleValue, waitForPromise
  bot_context.go    // DispatchRequest, buildDispatchInput, buildContext
  bot_store.go      // storeObject
  bot_ops.go        // DiscordOps, discordOpsObject
  bot_logging.go    // loggerObject, applyFields
```

**Rules:**
- Move types and their methods together.
- Keep unexported helpers near their primary consumer.
- Do not change any function signatures.
- Do not change any logic.

**Verification:**
```bash
go test ./...
go build ./...
```

**Commit message:**
```
refactor: split bot.go into responsibility-focused files

Splits the 1,293-line bot.go monolith into 6 focused files:
- bot_compile.go: CompileBot, botDraft, finalize
- bot_dispatch.go: BotHandle dispatch + promise settlement
- bot_context.go: DispatchRequest and context builders
- bot_store.go: storeObject
- bot_ops.go: DiscordOps builder
- bot_logging.go: loggerObject and applyFields

No behavior changes. Pure move refactor.

Refs: PASS2-2026-0421
```

---

## Task 2: Split `internal/jsdiscord/host_payloads.go`

**Current:** 736 lines. Contains normalization for 12 Discord payload types.

**Target layout:**
```text
internal/jsdiscord/
  payload_model.go      // normalizedResponse, shared types
  payload_message.go    // message send/edit/delete normalization
  payload_embeds.go     // embed normalization
  payload_components.go // button/select/menu normalization
  payload_files.go      // attachment/file normalization
  payload_mentions.go   // allowedMentions normalization
```

**Rules:**
- Each file should be < 200 lines.
- Shared types (e.g., `normalizedResponse`) go in `payload_model.go`.
- Keep normalization functions unexported.

**Verification:**
```bash
go test ./...
go build ./...
```

**Commit message:**
```
refactor: split host_payloads.go by payload concern

Splits the 736-line payload normalization monolith into focused files
by Discord payload type. No behavior changes.

Refs: PASS2-2026-0421
```

---

## Task 3: Split `internal/jsdiscord/runtime_test.go`

**Current:** 1,205 lines. Tests everything: descriptors, dispatch, events, payloads, ops, threads, moderation.

**Target layout:**
```text
internal/jsdiscord/
  runtime_descriptor_test.go      // bot descriptor tests
  runtime_dispatch_test.go        // command dispatch + async settlement
  runtime_events_test.go          // event handler tests
  runtime_payloads_test.go        // payload normalization tests
  runtime_ops_messages_test.go    // message ops tests
  runtime_ops_members_test.go     // member ops tests
  runtime_ops_threads_test.go     // thread ops tests
  runtime_knowledge_base_test.go  // knowledge base integration test
```

**Rules:**
- Extract shared helpers (`loadTestBot`, `writeBotScript`) into `internal/jsdiscord/testutil/` or a `jsdiscord_test` package.
- Each test file should test one behavior family.
- Table-driven tests stay intact; just move them.

**Verification:**
```bash
go test ./...
```

**Commit message:**
```
refactor: split runtime_test.go by behavior family

Extracts shared test helpers to testutil/ and splits the 1,205-line
runtime_test.go into focused behavior-family test files.

Refs: PASS2-2026-0421
```

## Risk Assessment

| Task | Risk | Mitigation |
|------|------|------------|
| Split bot.go | Low | Pure move; `go test` catches any missed references |
| Split host_payloads.go | Low | Same |
| Split runtime_test.go | Low | Same; extract helpers first |

## Success Criteria

- [ ] `bot.go` split into 6 files, all < 400 lines
- [ ] `host_payloads.go` split into 6 files, all < 200 lines
- [ ] `runtime_test.go` split into 8 files + testutil package
- [ ] `go test ./...` passes
- [ ] `go build ./...` passes
- [ ] Each commit is a pure move with no logic changes

## Related Tickets

- **PASS1-2026-0421** — Previous pass (stale code removal)
- **PASS3-2026-0421** — Next pass (API clarity)
