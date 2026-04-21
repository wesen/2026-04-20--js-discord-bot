---
Title: Discord Guild and Role Lookup Utilities
Ticket: DISCORD-BOT-011
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
      Note: Example repository notes mention guild and role lookup helpers
    - Path: examples/discord-bots/moderation/index.js
      Note: |-
        Moderation example bot should demonstrate the new lookup utilities
        Moderation example bot registers the guild and role lookup commands
    - Path: examples/discord-bots/moderation/lib/register-guild-role-lookup-commands.js
      Note: Example guild and role lookup commands live here
    - Path: internal/jsdiscord/bot.go
      Note: |-
        Request-scoped Discord capability object will grow with guild and role lookup helpers
        Request-scoped Discord capability object exposes guild and role lookup helpers
    - Path: internal/jsdiscord/host_maps.go
      Note: |-
        Normalized guild and role snapshots will live here
        Normalized guild and role snapshots live here
    - Path: internal/jsdiscord/host_ops.go
      Note: Host ops builder composes guild and role lookup helpers
    - Path: internal/jsdiscord/host_ops_guilds.go
      Note: Guild lookup host operations live here
    - Path: internal/jsdiscord/host_ops_members.go
      Note: Member/admin flows often need guild and role lookup context alongside moderation operations
    - Path: internal/jsdiscord/host_ops_roles.go
      Note: Role list and role fetch host operations live here
    - Path: internal/jsdiscord/runtime_test.go
      Note: |-
        Runtime tests should validate the new guild and role lookup APIs here
        Runtime tests validate guild and role lookup helpers
ExternalSources: []
Summary: Track the next Discord JS admin utilities after message/channel moderation, focused on guild snapshots and role lookup/list helpers.
LastUpdated: 2026-04-20T22:05:00-04:00
WhatFor: Organize planning and implementation work for guild and role lookup APIs used by moderation/admin bots.
WhenToUse: Use when implementing or reviewing the next Discord JS moderation helper slice after DISCORD-BOT-010.
---


# Discord Guild and Role Lookup Utilities

## Overview

This ticket captures the next practical Discord JS admin surface after message and channel moderation utilities. The focus is intentionally narrow and operational: make it easier for JavaScript moderation bots to inspect guild metadata and roles before taking moderation actions.

## Key Links

- `design-doc/01-discord-guild-and-role-lookup-utilities-architecture-and-implementation-guide.md`
- `reference/01-discord-guild-and-role-lookup-utilities-api-reference-and-planning-notes.md`
- `reference/02-diary.md`
- `tasks.md`
- `changelog.md`
