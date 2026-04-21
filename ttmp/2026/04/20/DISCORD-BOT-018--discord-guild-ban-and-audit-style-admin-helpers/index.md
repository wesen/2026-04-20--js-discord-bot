---
Title: Discord Guild Ban and Audit-Style Admin Helpers
Ticket: DISCORD-BOT-018
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
      Note: Request-scoped Discord capability object will grow with ban/audit-style admin helpers
    - Path: internal/jsdiscord/host_ops_members.go
      Note: Existing moderation helpers may gain more consistent reason handling here
    - Path: internal/jsdiscord/runtime_test.go
      Note: Runtime tests should validate ban lookup/admin helper behavior here
    - Path: examples/discord-bots/moderation/lib/register-member-moderation-commands.js
      Note: Moderation examples can demonstrate ban inspection/admin helpers here
ExternalSources: []
Summary: Track the next Discord JS admin utilities after broader event expansion, focused on guild ban inspection and more consistent moderation reason/audit patterns.
LastUpdated: 2026-04-20T22:35:00-04:00
WhatFor: Organize planning and implementation work for guild ban lookup/admin helpers and reason consistency improvements.
WhenToUse: Use when implementing or reviewing ban/audit-style admin helpers after DISCORD-BOT-017.
---


# Discord Guild Ban and Audit-Style Admin Helpers

## Overview

This ticket captures a more operator-specialized Discord JS admin surface after the core object and event helpers are in place. The focus is to make ban inspection and audit-style moderation workflows more coherent.

## Key Links

- `design-doc/01-discord-guild-ban-and-audit-style-admin-helpers-architecture-and-implementation-guide.md`
- `reference/01-discord-guild-ban-and-audit-style-admin-helpers-api-reference-and-planning-notes.md`
- `reference/02-diary.md`
- `tasks.md`
- `changelog.md`
