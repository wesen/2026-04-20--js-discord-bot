---
Title: Discord Guild Ban and Audit-Style Admin Helpers API Reference and Planning Notes
Ticket: DISCORD-BOT-018
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
      Note: Request-scoped Discord capability object will grow with ban/audit-style admin helpers
    - Path: internal/jsdiscord/host_ops_members.go
      Note: Existing moderation helpers may gain more consistent reason handling here
    - Path: internal/jsdiscord/runtime_test.go
      Note: Runtime tests should validate ban lookup/admin helper behavior here
    - Path: examples/discord-bots/moderation/lib/register-member-moderation-commands.js
      Note: Moderation examples can demonstrate ban inspection/admin helpers here
ExternalSources: []
Summary: Quick reference for the planned API surface, normalized payloads, and implementation ordering for DISCORD-BOT-018.
LastUpdated: 2026-04-20T22:35:00-04:00
WhatFor: Provide copy/paste-ready API sketches and implementation-phase guidance for DISCORD-BOT-018.
WhenToUse: Use when implementing or reviewing DISCORD-BOT-018.
---

# Goal

Provide quick API sketches and a phase-by-phase reference for DISCORD-BOT-018.

# Quick Reference

```js
const bans = await ctx.discord.bans.list(guildID)
const ban = await ctx.discord.bans.fetch(guildID, userID)
```

## Expected normalized shape

```js
{
  userID: "...",
  reason: "..."
}
```
