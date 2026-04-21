---
Title: Discord Role Management APIs API Reference and Planning Notes
Ticket: DISCORD-BOT-015
Status: active
Topics:
    - backend
    - chat
    - javascript
    - goja
DocType: reference
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
Summary: Quick reference for the planned API surface, normalized payloads, and implementation ordering for DISCORD-BOT-015.
LastUpdated: 2026-04-20T22:35:00-04:00
WhatFor: Provide copy/paste-ready API sketches and implementation-phase guidance for DISCORD-BOT-015.
WhenToUse: Use when implementing or reviewing DISCORD-BOT-015.
---

# Goal

Provide quick API sketches and a phase-by-phase reference for DISCORD-BOT-015.

# Quick Reference

```js
const role = await ctx.discord.roles.create(guildID, payload)
const updated = await ctx.discord.roles.update(guildID, roleID, payload)
```

## Expected normalized shape

```js
{
  id: "...",
  name: "...",
  position: 0
}
```
