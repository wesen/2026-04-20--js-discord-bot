---
Title: Tasks
Ticket: PASS2-2026-0421
---

# Tasks

## Task 1: Split `internal/jsdiscord/bot.go`
- [x] Create `bot_compile.go` — CompileBot, botDraft, finalize, snapshot helpers
- [x] Create `bot_dispatch.go` — BotHandle dispatch + promise settlement
- [x] Create `bot_context.go` — DispatchRequest and context builders
- [x] Create `bot_store.go` — storeObject
- [x] Create `bot_ops.go` — DiscordOps builder
- [x] Create `bot_logging.go` — loggerObject, applyFields
- [x] Fix missing package declarations (goimports quirk)
- [x] Fix wrong log import in bot_logging.go
- [x] Run `go test ./...` and `go build ./...`
- [x] Commit (f64f16a)

## Task 2: Split `internal/jsdiscord/host_payloads.go`
- [x] Create `payload_model.go` — normalizedResponse + top-level normalize helpers + shared type helpers
- [x] Create `payload_embeds.go` — embed normalization
- [x] Create `payload_components.go` — button/select/menu/text-input normalization
- [x] Create `payload_mentions.go` — allowedMentions normalization
- [x] Create `payload_files.go` — file attachment normalization
- [x] Create `payload_message.go` — message reference normalization
- [x] Fix missing package declarations
- [x] Run `go test ./...` and `go build ./...`
- [x] Commit (c5335fa)

## Task 3: Split `internal/jsdiscord/runtime_test.go`
- [x] Create `runtime_helpers_test.go` — loadTestBot, writeBotScript
- [x] Create `runtime_dispatch_test.go` — command dispatch, async, components, modals
- [x] Create `runtime_events_test.go` — message/reaction/member events
- [x] Create `runtime_descriptor_test.go` — application command descriptors
- [x] Create `runtime_payloads_test.go` — payload normalization
- [x] Create `runtime_ops_messages_test.go` — message CRUD ops
- [x] Create `runtime_ops_threads_test.go` — thread ops
- [x] Create `runtime_ops_members_test.go` — member lookup and admin ops
- [x] Create `runtime_ops_guilds_test.go` — guild/role/channel ops
- [x] Fix helper leak into wrong file
- [x] Fix missing package declarations
- [x] Run `go test ./...`
- [x] Commit (8e75a3d)
- [x] Delete temporary Python scripts
