---
Title: Custom KB Discord Bot Implementation Guide
Ticket: DISCORD-BOT-CUSTOM-KB
Status: active
Topics:
    - backend
    - chat
    - javascript
    - goja
    - sqlite
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - Path: Makefile
      Note: bump-glazed target keeps go-go-golems dependencies current
    - Path: examples/discord-bots/custom-kb/index.js
      Note: Main custom KB bot command
    - Path: examples/discord-bots/custom-kb/lib/store.js
      Note: SQLite link store patterned after the knowledge-base bot
    - Path: go.mod
      Note: records upgraded go-go-golems module versions used by the bot runtime
    - Path: internal/jsdiscord/custom_kb_runtime_test.go
      Note: Runtime test for adding and searching links
    - Path: pkg/doc/tutorials/building-and-running-discord-js-bots.md
      Note: UI DSL-first starter tutorial update related to the custom KB bot authoring pattern
ExternalSources: []
Summary: Guide for the custom-kb Discord bot that stores links in SQLite and renders add/search flows through a small UI DSL.
LastUpdated: 2026-05-01T15:25:00-04:00
WhatFor: ""
WhenToUse: ""
---




# Custom KB Discord Bot Implementation Guide

## Goal

Build `examples/discord-bots/custom-kb/`, a focused Discord bot for saving useful links into a SQLite database and searching them later from Discord.

## Docs and code found

- `examples/discord-bots/knowledge-base/` — canonical SQLite-backed knowledge bot.
- `examples/discord-bots/knowledge-base/lib/store.js` — store factory pattern, runtime `dbPath`, schema creation, SQL queries.
- `pkg/doc/topics/discord-js-bot-api-reference.md` — `require("database")` API: `configure`, `query`, `exec`, `close`.
- `pkg/doc/tutorials/using-the-go-side-ui-dsl-for-discord-bots.md` — UI builder mental model for message/embed/button/select/modal composition.
- `examples/discord-bots/ui-showcase/index.js` — practical UI DSL patterns for forms, selects, embeds, and stateful screens.

## Implementation

The new bot lives at `examples/discord-bots/custom-kb/`.

Commands:

- `/kb-add` opens an add-link modal.
- `/kb-link url title summary tags` stores a link directly.
- `/kb-search query` searches URL, title, summary, and tags.
- `/kb-list` lists recent links.

Storage:

- SQLite table: `kb_links`.
- Configurable DB path: `dbPath` / `db_path`, defaulting to `examples/discord-bots/custom-kb/data/custom-kb.sqlite`.
- Durable fields: `id`, `url`, `title`, `summary`, `tags_json`, Discord provenance, timestamps.

UI DSL:

- The bot uses the host `require("ui")` module with `message`, `embed`, `button`, `select`, and `form` builders.
- This keeps the command code declarative while returning normalized Discord payload objects.

## Module registration note

The current dependency (`go-go-goja v0.4.12`) does not expose every default-registry module automatically. Its engine always registers data-only modules, but host-access modules like `database` must be opted in with `engine.DefaultRegistryModulesNamed("database")`.

The newer `../corporate-headquarters/go-go-goja` branch changed this again: a plain `NewBuilder().Build()` enables all default-registry modules unless middleware narrows them, and `modules/database` registers both `database` and `db`. This repo is still on the older explicit-registration model, so the fix here is to opt in to `database` at each Discord runtime factory site.

## Validation

Added `internal/jsdiscord/custom_kb_runtime_test.go` to exercise `/kb-link` and `/kb-search` against a temp SQLite path.

Validated with:

- `go test ./internal/jsdiscord -run TestCustomKB -count=1`
- `go test ./internal/jsdiscord -run TestKnowledgeBaseBotUsesSQLiteStoreForCaptureSearchAndReview -count=1`
- `go test ./pkg/botcli ./internal/jsdiscord -count=1`
- `go run ./cmd/discord-bot bots help custom-kb --bot-repository ./examples/discord-bots`
- live tmux run in session `custom-kb-bot`, including Discord command sync and ready event with the configured SQLite DB
