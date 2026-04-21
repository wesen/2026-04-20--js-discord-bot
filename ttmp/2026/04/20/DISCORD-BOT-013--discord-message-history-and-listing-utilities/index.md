---
Title: Discord Message History and Listing Utilities
Ticket: DISCORD-BOT-013
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
    - Path: examples/discord-bots/README.md
      Note: Example repository notes mention the message history helper
    - Path: examples/discord-bots/moderation/lib/register-message-moderation-commands.js
      Note: |-
        The moderation example should demonstrate message history/listing helpers here
        Example message history command lives here
    - Path: internal/jsdiscord/bot.go
      Note: |-
        Request-scoped Discord capability object will grow with message history and listing helpers
        Request-scoped Discord capability object exposes the message history/listing helper
    - Path: internal/jsdiscord/host_ops_helpers.go
      Note: |-
        Message list option normalization will live here
        Message list option normalization lives here
    - Path: internal/jsdiscord/host_ops_messages.go
      Note: |-
        Message history/listing host operations will live alongside existing message moderation operations here
        Message listing host operations live here
    - Path: internal/jsdiscord/runtime_test.go
      Note: |-
        Runtime tests should validate the new message history/listing APIs here
        Runtime tests validate message history/listing helpers
ExternalSources: []
Summary: Track the next Discord JS moderation/admin utilities after member lookup, focused on message history and bounded message listing helpers.
LastUpdated: 2026-04-20T22:35:00-04:00
WhatFor: Organize planning and implementation work for message history/listing APIs used by moderation/admin bots.
WhenToUse: Use when implementing or reviewing the next Discord JS moderation helper slice after DISCORD-BOT-012.
---



# Discord Message History and Listing Utilities

## Overview

This ticket captures the next practical Discord JS admin surface after guild/role/member lookup helpers. The focus is intentionally narrow and operational: make it easier for JavaScript moderation bots to inspect recent or anchored message history before taking action.

## Key Links

- `design-doc/01-discord-message-history-and-listing-utilities-architecture-and-implementation-guide.md`
- `reference/01-discord-message-history-and-listing-utilities-api-reference-and-planning-notes.md`
- `reference/02-diary.md`
- `tasks.md`
- `changelog.md`
