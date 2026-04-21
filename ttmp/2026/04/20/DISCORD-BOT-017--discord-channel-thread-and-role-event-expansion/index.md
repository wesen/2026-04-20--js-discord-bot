---
Title: Discord Channel, Thread, and Role Event Expansion
Ticket: DISCORD-BOT-017
Status: active
Topics:
    - backend
    - chat
    - javascript
    - goja
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/jsdiscord/host_dispatch.go
      Note: New event dispatch entrypoints will live here
    - Path: internal/jsdiscord/host_maps.go
      Note: Normalized event payload maps will live here
    - Path: internal/jsdiscord/runtime_test.go
      Note: Runtime tests should validate the new event dispatch coverage here
    - Path: examples/discord-bots/moderation/lib/register-events.js
      Note: Event-heavy examples can demonstrate the new lifecycle handlers here
ExternalSources: []
Summary: Track the next Discord JS event expansion after admin object helpers, focused on channel/thread/role lifecycle events.
LastUpdated: 2026-04-20T22:35:00-04:00
WhatFor: Organize planning and implementation work for the next lifecycle event expansion.
WhenToUse: Use when implementing or reviewing channel/thread/role event expansion after DISCORD-BOT-016.
---


# Discord Channel, Thread, and Role Event Expansion

## Overview

This ticket captures the next event-oriented Discord JS surface after the current object/action helper expansion. The focus is to add a few high-value lifecycle events rather than broad event coverage all at once.

## Key Links

- `design-doc/01-discord-channel-thread-and-role-event-expansion-architecture-and-implementation-guide.md`
- `reference/01-discord-channel-thread-and-role-event-expansion-api-reference-and-planning-notes.md`
- `reference/02-diary.md`
- `tasks.md`
- `changelog.md`
