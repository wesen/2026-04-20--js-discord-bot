---
Title: Show Space Discord Bot
Ticket: DISCORD-BOT-022
Status: active
Topics:
  - discord
  - javascript
  - bots
  - database
  - moderation
  - announcements
  - venues
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
  - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/ping/index.js
    Note: Existing JS bot showcase with commands, events, components, modals, follow-ups, and outbound Discord operations
  - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/knowledge-base/index.js
    Note: Database-backed JS bot structure, runtime config, and command composition pattern
  - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/knowledge-base/lib/store.js
    Note: Current `require("database")` usage and SQLite persistence wrapper
  - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/moderation/lib/register-message-moderation-commands.js
    Note: Existing pin/unpin/list-pins command patterns and permission-sensitive moderation flows
  - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/README.md
    Note: Example bot inventory, runtime notes, and command invocation patterns
  - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_responses.go
    Note: Interaction response behavior, including in-place updates and follow-ups
  - Path: /home/manuel/code/wesen/2026-04-20--js-discord-bot/pkg/doc/topics/discord-js-bot-api-reference.md
    Note: Current JS API reference for commands, contexts, `ctx.discord`, `ctx.store`, `require("database")`, and `require("ui")`
ExternalSources: []
Summary: Plan and analyze a Discord bot for an artist/show space that posts show announcements, pins them, and later tracks shows in a durable database.
LastUpdated: 2026-04-22T22:11:54-04:00
WhatFor: Organize the implementation plan, current-state analysis, and phase-by-phase task list for the show-space bot.
WhenToUse: Use when building or reviewing the venue operations bot that manages upcoming shows and pinned announcements.
---

# Show Space Discord Bot

## Overview

This ticket captures the implementation plan for a Discord bot that supports an artist/show space. The bot’s first job is to keep show announcements tidy and pinned. The second job is to add durable show storage so staff can manage shows by ID and automatically archive old pins.

The implementation target in this repository is the embedded JavaScript bot runtime, not Discord.js or discord.py. That means the bot should live under `examples/discord-bots/`, use `require("discord")`, and rely on the repository’s existing host operations for messaging, pins, and persistence.

## Key Links

- `design/01-show-space-discord-bot-implementation-and-analysis-guide.md`
- `tasks.md`
- `changelog.md`
- `reference/01-diary.md`
- `reference/02-operator-runbook.md`

## Status

Current status: **active**

## Topics

- discord
- javascript
- bots
- database
- moderation
- announcements
- venues

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- `design/` — implementation and analysis guide
- `reference/` — source extracts or verbatim spec copies if needed later
- `scripts/` — any temporary migration or validation scripts
- `archive/` — deprecated notes, if the plan changes
