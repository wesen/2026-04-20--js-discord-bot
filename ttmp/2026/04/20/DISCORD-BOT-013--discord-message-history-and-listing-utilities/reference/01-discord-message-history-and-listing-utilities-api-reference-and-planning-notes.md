---
Title: Discord Message History and Listing Utilities API Reference and Planning Notes
Ticket: DISCORD-BOT-013
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
Summary: Quick reference for the planned API surface, normalized payloads, and implementation ordering for DISCORD-BOT-013.
LastUpdated: 2026-04-20T22:35:00-04:00
WhatFor: Provide copy/paste-ready API sketches and implementation-phase guidance for DISCORD-BOT-013.
WhenToUse: Use when implementing or reviewing DISCORD-BOT-013.
---

# Goal

Provide quick API sketches and a phase-by-phase reference for DISCORD-BOT-013.

# Quick Reference

```js
const messages = await ctx.discord.messages.list(channelID)
const before = await ctx.discord.messages.list(channelID, { before: "msg-1", limit: 25 })
const after = await ctx.discord.messages.list(channelID, { after: "msg-1", limit: 25 })
const around = await ctx.discord.messages.list(channelID, { around: "msg-1", limit: 25 })
```

## Expected normalized shape

```js
{
  id: "...",
  content: "...",
  guildID: "...",
  channelID: "...",
  author: { id: "...", username: "..." }
}
```
