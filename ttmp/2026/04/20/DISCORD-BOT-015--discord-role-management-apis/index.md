---
Title: Discord Role Management APIs
Ticket: DISCORD-BOT-015
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
      Note: Request-scoped Discord capability object will grow with role management helpers
    - Path: internal/jsdiscord/host_ops_roles.go
      Note: Role management host operations will live alongside role lookup helpers here
    - Path: internal/jsdiscord/runtime_test.go
      Note: Runtime tests should validate role management helpers here
    - Path: examples/discord-bots/moderation/lib/register-guild-role-lookup-commands.js
      Note: Moderation/admin examples can demonstrate role creation/update helpers here
ExternalSources: []
Summary: Track the next Discord JS admin utilities after thread support, focused on creating and updating role objects rather than only assigning them to members.
LastUpdated: 2026-04-20T22:35:00-04:00
WhatFor: Organize planning and implementation work for role management APIs.
WhenToUse: Use when implementing or reviewing role create/update/delete helpers after DISCORD-BOT-014.
---


# Discord Role Management APIs

## Overview

This ticket captures the next practical Discord JS admin surface after thread utilities. The focus is intentionally narrow: extend the runtime from role inspection and member-role assignment into managing the role objects themselves.

## Key Links

- `design-doc/01-discord-role-management-apis-architecture-and-implementation-guide.md`
- `reference/01-discord-role-management-apis-api-reference-and-planning-notes.md`
- `reference/02-diary.md`
- `tasks.md`
- `changelog.md`
