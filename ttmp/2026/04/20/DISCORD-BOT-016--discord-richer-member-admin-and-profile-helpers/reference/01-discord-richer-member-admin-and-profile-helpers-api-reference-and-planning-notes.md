---
Title: Discord Richer Member Admin and Profile Helpers API Reference and Planning Notes
Ticket: DISCORD-BOT-016
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
      Note: Request-scoped Discord capability object will grow with richer member admin/profile helpers
    - Path: internal/jsdiscord/host_ops_members.go
      Note: Richer member admin/profile helpers will live here
    - Path: internal/jsdiscord/runtime_test.go
      Note: Runtime tests should validate the richer member helpers here
    - Path: examples/discord-bots/moderation/lib/register-member-moderation-commands.js
      Note: Moderation/admin examples can demonstrate nickname and related helpers here
ExternalSources: []
Summary: Quick reference for the planned API surface, normalized payloads, and implementation ordering for DISCORD-BOT-016.
LastUpdated: 2026-04-20T22:35:00-04:00
WhatFor: Provide copy/paste-ready API sketches and implementation-phase guidance for DISCORD-BOT-016.
WhenToUse: Use when implementing or reviewing DISCORD-BOT-016.
---

# Goal

Provide quick API sketches and a phase-by-phase reference for DISCORD-BOT-016.

# Quick Reference

```js
await ctx.discord.members.setNick(guildID, userID, "new-nick")
await ctx.discord.members.clearNick(guildID, userID)
```

## Expected normalized shape

```js
{
  id: "...",
  nick: "..."
}
```
