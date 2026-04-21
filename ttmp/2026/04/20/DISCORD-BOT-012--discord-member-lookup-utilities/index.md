---
Title: Discord Member Lookup Utilities
Ticket: DISCORD-BOT-012
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
      Note: Example repository notes mention member lookup helpers
    - Path: examples/discord-bots/moderation/lib/register-member-moderation-commands.js
      Note: |-
        Moderation example commands should demonstrate member lookup paired with moderation actions
        Example member lookup and moderation commands live here
    - Path: internal/jsdiscord/bot.go
      Note: |-
        Request-scoped Discord capability object will grow with member lookup helpers
        Request-scoped Discord capability object exposes member lookup helpers
    - Path: internal/jsdiscord/host_maps.go
      Note: Normalized member snapshots already live here and will support read-only member lookup
    - Path: internal/jsdiscord/host_ops_helpers.go
      Note: Member list option normalization lives here
    - Path: internal/jsdiscord/host_ops_members.go
      Note: |-
        Member lookup and member moderation helpers will live side by side here
        Member fetch/list and member moderation helpers live here
    - Path: internal/jsdiscord/runtime_test.go
      Note: |-
        Runtime tests should validate the new member lookup APIs here
        Runtime tests validate member lookup helpers
ExternalSources: []
Summary: Track the next Discord JS admin utilities after guild/role lookup, focused on read-only member fetch/list helpers.
LastUpdated: 2026-04-20T22:35:00-04:00
WhatFor: Organize planning and implementation work for member lookup APIs used by moderation/admin bots.
WhenToUse: Use when implementing or reviewing the next Discord JS moderation helper slice after DISCORD-BOT-011.
---


# Discord Member Lookup Utilities

## Overview

This ticket captures the next practical Discord JS admin surface after guild and role lookup helpers. The focus is intentionally narrow and operational: make it easier for JavaScript moderation bots to inspect one member or a small member list before applying existing moderation actions.

## Key Links

- `design-doc/01-discord-member-lookup-utilities-architecture-and-implementation-guide.md`
- `reference/01-discord-member-lookup-utilities-api-reference-and-planning-notes.md`
- `reference/02-diary.md`
- `tasks.md`
- `changelog.md`
