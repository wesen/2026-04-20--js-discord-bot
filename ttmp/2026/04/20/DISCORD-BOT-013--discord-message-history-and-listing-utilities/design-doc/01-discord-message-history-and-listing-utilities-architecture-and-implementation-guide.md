---
Title: Discord Message History and Listing Utilities Architecture and Implementation Guide
Ticket: DISCORD-BOT-013
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
      Note: Request-scoped Discord capability object will grow with message history and listing helpers
    - Path: internal/jsdiscord/host_ops_messages.go
      Note: Message history/listing host operations will live alongside existing message moderation operations here
    - Path: internal/jsdiscord/host_ops_helpers.go
      Note: Message list option normalization will live here
    - Path: internal/jsdiscord/runtime_test.go
      Note: Runtime tests should validate the new message history/listing APIs here
    - Path: examples/discord-bots/moderation/lib/register-message-moderation-commands.js
      Note: The moderation example should demonstrate message history/listing helpers here
ExternalSources: []
Summary: Detailed implementation guide for discord message history and listing utilities.
LastUpdated: 2026-04-20T22:35:00-04:00
WhatFor: Explain the architecture, phase ordering, and implementation seams for discord message history and listing utilities.
WhenToUse: Use when implementing or reviewing DISCORD-BOT-013.
---

# Goal

Add a small, practical message-history surface that lets JavaScript moderation/admin bots inspect recent or anchored channel history through request-scoped host capabilities.

# Why this ticket exists

The runtime can already fetch one message and perform message moderation actions, but it still cannot browse surrounding channel history in a useful bounded way.

# Proposed API surface

```js
const messages = await ctx.discord.messages.list(channelID)
const before = await ctx.discord.messages.list(channelID, { before: "msg-1", limit: 25 })
const after = await ctx.discord.messages.list(channelID, { after: "msg-1", limit: 25 })
const around = await ctx.discord.messages.list(channelID, { around: "msg-1", limit: 25 })
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
