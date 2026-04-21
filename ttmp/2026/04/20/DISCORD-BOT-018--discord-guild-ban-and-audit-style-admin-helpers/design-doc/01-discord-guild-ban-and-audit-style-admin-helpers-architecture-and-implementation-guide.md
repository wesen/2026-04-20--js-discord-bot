---
Title: Discord Guild Ban and Audit-Style Admin Helpers Architecture and Implementation Guide
Ticket: DISCORD-BOT-018
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
      Note: Request-scoped Discord capability object will grow with ban/audit-style admin helpers
    - Path: internal/jsdiscord/host_ops_members.go
      Note: Existing moderation helpers may gain more consistent reason handling here
    - Path: internal/jsdiscord/runtime_test.go
      Note: Runtime tests should validate ban lookup/admin helper behavior here
    - Path: examples/discord-bots/moderation/lib/register-member-moderation-commands.js
      Note: Moderation examples can demonstrate ban inspection/admin helpers here
ExternalSources: []
Summary: Detailed implementation guide for discord guild ban and audit-style admin helpers.
LastUpdated: 2026-04-20T22:35:00-04:00
WhatFor: Explain the architecture, phase ordering, and implementation seams for discord guild ban and audit-style admin helpers.
WhenToUse: Use when implementing or reviewing DISCORD-BOT-018.
---

# Goal

Add a small operator-focused admin surface for ban inspection and more consistent moderation reason handling.

# Why this ticket exists

The runtime already performs bans and unbans, but it does not yet help operators inspect guild ban state or reason consistency.

# Proposed API surface

```js
const bans = await ctx.discord.bans.list(guildID)
const ban = await ctx.discord.bans.fetch(guildID, userID)
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
