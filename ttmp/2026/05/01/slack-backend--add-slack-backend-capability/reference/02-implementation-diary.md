---
Title: Implementation Diary
Ticket: slack-backend
Status: active
Topics:
    - slack
    - backend
    - javascript
    - bot-framework
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/discord-bot/root.go
      Note: Root command wiring for Slack commands (commit f55387a)
    - Path: cmd/discord-bot/slack_commands.go
      Note: CLI commands for Slack manifest and HTTP server (commit f55387a)
    - Path: internal/jsdiscord/slack_backend.go
      Note: Slack HTTP backend prototype implementation (commit f55387a)
    - Path: internal/jsdiscord/slack_backend_test.go
      Note: Slack backend unit tests (commit f55387a)
ExternalSources: []
Summary: Implementation diary for adding a Slack backend while preserving the existing JavaScript Discord-shaped bot API.
LastUpdated: 2026-05-01T16:47:42-07:00
WhatFor: Track implementation steps, validation, commits, and review guidance for the Slack backend ticket.
WhenToUse: Use when reviewing or continuing Slack backend implementation work.
---


# Diary

## Goal

Capture the implementation journey for adding a Slack backend/capability while keeping existing JavaScript bot scripts unchanged.

## Step 1: Slack HTTP Backend Prototype

This step added the first working Slack backend surface in Go. The implementation keeps the JavaScript side Discord-shaped and maps Slack slash commands, Block Kit actions, modal submissions, and app mention events into the existing `DispatchRequest` path.

It also added manifest generation, Slack request signing verification, Block Kit rendering, modal rendering, inline file export handling, and SQLite-backed adapter state for message/interaction correlation.

### Prompt Context

**User prompt (verbatim):** "ok lets implement it. Keep a diary and commit as you go."

**Assistant interpretation:** Implement the Slack backend according to the resolved design decisions, maintain a docmgr diary, and commit focused progress.

**Inferred user intent:** The user wants the design to move from documentation into code with reviewable commits and a continuation-friendly implementation record.

**Commit (code):** f55387aac78674bbf7c3727df154e3a122dd25e9 — "Add Slack HTTP backend prototype"

### What I did

- Added `internal/jsdiscord/slack_backend.go` with:
  - `SlackBackend` HTTP server for `/slack/commands`, `/slack/interactivity`, `/slack/events`, and `/healthz`.
  - Slack signature verification using `X-Slack-Signature` and `X-Slack-Request-Timestamp`.
  - Slack slash command normalization into existing command dispatch.
  - Slack `block_actions` normalization into existing component dispatch.
  - Slack `view_submission` normalization into existing modal dispatch.
  - Slack app mention event normalization into `messageCreate` dispatch.
  - Slack Web API client wrappers for `chat.postMessage`, `chat.update`, `views.open`, and `response_url` calls.
  - SQLite tables for `slack_messages`, `slack_interactions`, and `slack_modal_contexts`.
  - Block Kit rendering for content and buttons.
  - Modal rendering for text inputs.
  - Inline file/export rendering into message text.
  - Slack app manifest generation as JSON-compatible data.
- Added `internal/jsdiscord/slack_backend_test.go` covering:
  - request signature verification,
  - manifest generation,
  - button + inline-file Block Kit rendering,
  - SQLite message persistence.
- Added top-level CLI commands in `cmd/discord-bot/slack_commands.go`:
  - `slack-manifest --bot-script ... --base-url ...`
  - `slack-serve --bot-script ... --listen-addr ... --slack-state-db ...`
- Wired those commands into `cmd/discord-bot/root.go`.

### Why

- The design says JavaScript should not change; therefore Slack normalization belongs in Go around the existing JS runtime dispatch contract.
- Slack interactions need durable correlation, so message content and interaction payloads are stored in SQLite.
- Slack setup needs repeatability, so manifest generation is part of the CLI.
- Slack file upload is deferred by design, so file responses are inlined into messages.

### What worked

- `go test ./... -count=1` passed.
- `go run ./cmd/discord-bot slack-manifest --bot-script ./examples/discord-bots/adventure/index.js --base-url https://bot.example` produced a Slack manifest JSON.
- Existing JS runtime tests continued to pass without changing bot scripts.

### What didn't work

- Initial compile failed because I tried to set a nonexistent `Values` field on `ComponentSnapshot`:

```text
# github.com/go-go-golems/discord-bot/internal/jsdiscord [github.com/go-go-golems/discord-bot/internal/jsdiscord.test]
internal/jsdiscord/slack_backend.go:389:56: unknown field Values in struct literal of type ComponentSnapshot
FAIL	github.com/go-go-golems/discord-bot/internal/jsdiscord [build failed]
FAIL
```

- Initial manifest output omitted slash commands because `CommandDescriptor.Type` can contain the string `"<nil>"` for normal chat commands. I changed filtering to skip only explicit `user` and `message` command types.

### What I learned

- The existing `DispatchRequest` and responder callbacks are sufficient to host Slack without a JS API rename.
- `ctx.message.content` compatibility requires storing rendered content keyed by Slack `team/channel/ts`; reconstructing from Slack blocks should only be a fallback.
- For one-option Slack command support, option selection should prefer names like `prompt`, `text`, `query`, `input`, or `message` before alphabetical fallback. This keeps the current adventure bot usable even though it declares multiple Discord options.

### What was tricky to build

- Slack ACK behavior differs from Discord interaction responses. The HTTP handlers ACK immediately before async dispatch; `ctx.defer()` records intent but does not send another HTTP acknowledgement.
- Slack has multiple response paths (`response_url`, `chat.postMessage`, `chat.update`) and the responder has to pick based on whether it is handling a command, component, modal, or event.
- Slack modal submissions do not include the original message context, so `private_metadata` carries `channel_id`, `message_ts`, and `response_url`.
- The file reader is consumed when inlining files; this is acceptable for the first Slack-only rendering pass but warrants review if the same normalized response is reused elsewhere.

### What warrants a second pair of eyes

- `slackResponder.Reply/Edit/FollowUp` routing between `response_url`, `chat.postMessage`, and `chat.update` should be reviewed against live Slack behavior.
- The immediate-ACK goroutine dispatch pattern should be reviewed for process shutdown and error visibility.
- Slack command option handling intentionally supports only one option, but current implementation logs and picks a preferred first option if more exist; confirm whether this should hard-error instead.
- The Slack Block Kit renderer currently supports buttons and text inputs, not the full Discord component set.

### What should be done in the future

- Add live Slack app setup/playbook using generated manifest.
- Add integration tests with an `httptest.Server` fake Slack API for response routing.
- Decide whether multi-option commands should hard fail in Slack manifest generation.
- Add cleanup/TTL handling for SQLite interaction/message rows.

### Code review instructions

- Start with `internal/jsdiscord/slack_backend.go`:
  - `SlackBackend.handleCommand`, `handleInteractivity`, `handleEvents`
  - `dispatchSlashCommand`, `dispatchBlockAction`, `dispatchViewSubmission`
  - `slackMessagePayload`, `slackViewPayload`
  - `SlackStore.ensure`
- Then review CLI wiring in `cmd/discord-bot/slack_commands.go` and `cmd/discord-bot/root.go`.
- Validate with:

```bash
go test ./... -count=1
go run ./cmd/discord-bot slack-manifest --bot-script ./examples/discord-bots/adventure/index.js --base-url https://bot.example
```

### Technical details

Slack serve command shape:

```bash
go run ./cmd/discord-bot slack-serve \
  --bot-script ./examples/discord-bots/adventure/index.js \
  --listen-addr :8080 \
  --slack-state-db ./slack-state.sqlite
```

Environment fallbacks:

```text
SLACK_BOT_TOKEN
SLACK_SIGNING_SECRET
```
