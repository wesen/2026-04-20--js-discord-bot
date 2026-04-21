---
Title: Discord Richer Member Admin and Profile Helpers
Ticket: DISCORD-BOT-016
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
    - Path: internal/jsdiscord/bot.go
      Note: Request-scoped Discord capability object will grow with richer member admin/profile helpers
    - Path: internal/jsdiscord/host_ops_members.go
      Note: Richer member admin/profile helpers will live here
    - Path: internal/jsdiscord/runtime_test.go
      Note: Runtime tests should validate the richer member helpers here
    - Path: examples/discord-bots/moderation/lib/register-member-moderation-commands.js
      Note: Moderation/admin examples can demonstrate nickname and related helpers here
ExternalSources: []
Summary: Track the next Discord JS admin utilities after role management, focused on richer member profile/admin helpers such as nickname and lightweight voice-state actions.
LastUpdated: 2026-04-20T22:35:00-04:00
WhatFor: Organize planning and implementation work for richer member admin/profile helpers.
WhenToUse: Use when implementing or reviewing richer member admin APIs after DISCORD-BOT-015.
---


# Discord Richer Member Admin and Profile Helpers

## Overview

This ticket captures the next practical Discord JS admin surface after role management. The focus is to round out member operations beyond basic moderation by exposing a few high-value profile/admin helpers.

## Key Links

- `design-doc/01-discord-richer-member-admin-and-profile-helpers-architecture-and-implementation-guide.md`
- `reference/01-discord-richer-member-admin-and-profile-helpers-api-reference-and-planning-notes.md`
- `reference/02-diary.md`
- `tasks.md`
- `changelog.md`
