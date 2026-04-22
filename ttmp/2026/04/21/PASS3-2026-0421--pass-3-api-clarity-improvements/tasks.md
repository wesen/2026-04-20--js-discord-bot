---
Title: Tasks
Ticket: PASS3-2026-0421
---

# Tasks

## Task 1: Introduce typed internal envelopes
- [x] Define typed snapshot structs (UserSnapshot, MemberSnapshot, MessageSnapshot, InteractionSnapshot, ReactionSnapshot, EmojiSnapshot, EmbedSnapshot, AttachmentSnapshot, MessageReferenceSnapshot, ComponentSnapshot, FocusedOptionSnapshot)
- [x] Add ToMap() conversion methods
- [x] Update DispatchRequest to use typed fields internally
- [x] Update host_dispatch.go to build typed snapshots
- [x] Update buildDispatchInput to convert typed structs to JS maps
- [x] Update all tests to use typed structs
- [x] Run `go test ./...` and `go build ./...`
- [x] Commit (db5015f)

## Task 2: Refactor host_dispatch.go with base request builders
- [x] Add `Host.baseDispatchRequest(session)` method
- [x] Add `withChannelResponder(req, responder)` helper
- [x] Add `withInteractionResponder(req, responder)` helper
- [x] Replace all 18 DispatchRequest constructions with builder calls
- [x] Run `go test ./...` and `go build ./...`
- [x] Commit (6e4bf6f)

## Task 3: Make botcli parsing phases explicit
- [x] Rename `preparseRunArgs` → `parseStaticRunnerArgs`
- [x] Rename `preParsedRunArgs` → `staticRunnerArgs`
- [x] Create `run_static_args.go` with static parsing
- [x] Create `run_dynamic_schema.go` with schema building
- [x] Create `run_help.go` with help rendering
- [x] Delete `run_schema.go`
- [x] Update `command.go` reference
- [x] Run `go test ./...` and `go build ./...`
- [x] Commit (2e5f9aa)

## Task 4: Refactor DiscordOps into registry pattern
- [x] Define generic nil-guard wrappers: op1, op2, op1E, op2E, op3E, op1A, op1ASlice, op2A
- [x] Replace ~60 closure definitions with single-line wrapper calls
- [x] Handle unique signature for ChannelSetSlowmode separately
- [x] Run `go test ./...` and `go build ./...`
- [x] Commit (7a24a6a)

## Task 5: Extract test helpers into shared test file
- [x] Create `internal/jsdiscord/helpers_test.go` with consolidated helpers
- [x] Move `loadTestBot`, `writeBotScript`, `repoRootJSDiscord`
- [x] Remove duplicate `repoRootJSDiscord` from `knowledge_base_runtime_test.go`
- [x] Run `go test ./...`
- [x] Commit (c71a0b6)
- **Note:** testutil subpackage abandoned due to Go test import cycle restriction

## Task 6: Collapse descriptor parsers with generic helper
- [x] Define generic `parseDescriptors[T any]` helper
- [x] Replace 6 near-identical `parse*Descriptors` functions
- [x] Run `go test ./...`
- [x] Commit (3d80797)

## Task 7: Generic sort helper for option drafts
- [x] Extract common sort logic from required/optional draft sorting in `host_commands.go`
- [x] Run `go test ./...`
- [x] Commit (coming next)

## Task 8: Extract DispatchInteraction branches into private methods
- [x] Extract `dispatchApplicationCommandInteraction`
- [x] Extract `dispatchMessageComponentInteraction`
- [x] Extract `dispatchModalSubmitInteraction`
- [x] Extract `dispatchAutocompleteInteraction`
- [x] Run `go test ./...`
- [x] Commit (fc8b62e)

## Task 9: Extract dispatch closures from finalize into helpers
- [x] Extract `describeClosure`
- [x] Extract `dispatchCommandClosure`
- [x] Extract `dispatchSubcommandClosure`
- [x] Extract `dispatchEventClosure`
- [x] Extract `dispatchComponentClosure`
- [x] Extract `dispatchModalClosure`
- [x] Extract `dispatchAutocompleteClosure`
- [x] Run `go test ./...`
- [x] Commit (ba93e56)
