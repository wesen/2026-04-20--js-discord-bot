---
Title: Discord Role Management APIs Architecture and Implementation Guide
Ticket: DISCORD-BOT-015
Status: active
Topics:
    - backend
    - chat
    - javascript
    - goja
DocType: design-doc
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
Summary: Detailed implementation guide for discord role management apis.
LastUpdated: 2026-04-20T22:35:00-04:00
WhatFor: Explain the architecture, phase ordering, and implementation seams for discord role management apis.
WhenToUse: Use when implementing or reviewing DISCORD-BOT-015.
---

# Goal

Add a small role-management surface for creating and updating role objects with normalized payload conventions.

# Why this ticket exists

The runtime can already inspect roles and assign them to members, but it cannot yet manage the role objects themselves.

# Proposed API surface

```js
const role = await ctx.discord.roles.create(guildID, payload)
const updated = await ctx.discord.roles.update(guildID, roleID, payload)
```

# Design constraints

- Keep the APIs request-scoped under `ctx.discord`, not global.
- Return normalized plain JS maps or normalized payload objects, not raw Discordgo structs.
- Keep the first phase focused and reviewable.

# Review guidance

Start review in this order:

1. the request-scoped Discord binding in `internal/jsdiscord/bot.go`
2. the host operation file for this surface
3. normalization helpers and runtime tests
4. the example bot updates
