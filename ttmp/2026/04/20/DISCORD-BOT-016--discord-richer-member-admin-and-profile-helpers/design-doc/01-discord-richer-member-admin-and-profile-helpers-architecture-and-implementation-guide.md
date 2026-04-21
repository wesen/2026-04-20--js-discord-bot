---
Title: Discord Richer Member Admin and Profile Helpers Architecture and Implementation Guide
Ticket: DISCORD-BOT-016
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
      Note: Request-scoped Discord capability object will grow with richer member admin/profile helpers
    - Path: internal/jsdiscord/host_ops_members.go
      Note: Richer member admin/profile helpers will live here
    - Path: internal/jsdiscord/runtime_test.go
      Note: Runtime tests should validate the richer member helpers here
    - Path: examples/discord-bots/moderation/lib/register-member-moderation-commands.js
      Note: Moderation/admin examples can demonstrate nickname and related helpers here
ExternalSources: []
Summary: Detailed implementation guide for discord richer member admin and profile helpers.
LastUpdated: 2026-04-20T22:35:00-04:00
WhatFor: Explain the architecture, phase ordering, and implementation seams for discord richer member admin and profile helpers.
WhenToUse: Use when implementing or reviewing DISCORD-BOT-016.
---

# Goal

Add a small, practical member-admin surface that goes beyond bans/timeouts into profile and lightweight admin helpers.

# Why this ticket exists

The current member moderation surface is strong, but it still lacks some everyday operator helpers such as nickname management.

# Proposed API surface

```js
await ctx.discord.members.setNick(guildID, userID, "new-nick")
await ctx.discord.members.clearNick(guildID, userID)
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
