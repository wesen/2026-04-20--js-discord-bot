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
      Note: The request-scoped Discord capability object exposes the message history/listing helpers here
    - Path: internal/jsdiscord/host_ops_helpers.go
      Note: Message list option normalization lives here
    - Path: internal/jsdiscord/host_ops_messages.go
      Note: Message listing host operations live here
    - Path: internal/jsdiscord/runtime_test.go
      Note: Runtime tests validate message listing helpers here
ExternalSources: []
Summary: Quick reference for the implemented message history/listing APIs, normalized payloads, and operational caveats.
LastUpdated: 2026-04-20T22:55:00-04:00
WhatFor: Provide copy/paste-ready API sketches and operator-facing notes for the implemented DISCORD-BOT-013 surface.
WhenToUse: Use when implementing, reviewing, or operating the message history/listing helper APIs.
---

# Goal

Provide a quick reference for the implemented DISCORD-BOT-013 message history and listing utilities.

# Quick Reference

```js
const messages = await ctx.discord.messages.list(channelID)
const before = await ctx.discord.messages.list(channelID, { before: "msg-1", limit: 25 })
const after = await ctx.discord.messages.list(channelID, { after: "msg-1", limit: 25 })
const around = await ctx.discord.messages.list(channelID, { around: "msg-1", limit: 25 })
```

## Current normalized message shape

```js
{
  id: "...",
  content: "...",
  guildID: "...",
  channelID: "...",
  author: {
    id: "...",
    username: "...",
    discriminator: "...",
    bot: false,
  }
}
```

# Operational Notes

## Visibility expectations

These helpers are read-only, but they still depend on the bot being able to read the target channel and its history.

## Current listing behavior

### `messages.list(channelID, payload?)`

Current implemented behavior:
- accepts `nil` or `{ before, after, around, limit }`
- allows at most one anchor out of `before`, `after`, or `around`
- defaults to `limit: 25`
- clamps limit to at most `100`
- returns normalized plain JS message maps
- logs list options and returned count at debug level

## Common failure modes

- wrong channel ID
- the bot lacks channel visibility or read/history access
- payload includes more than one anchor (`before`, `after`, `around`)
- anchor message ID is invalid or outside a useful channel history range

# Usage examples

```js
command("mod-list-messages", {
  options: {
    before_message_id: { type: "string", required: false },
    after_message_id: { type: "string", required: false },
    around_message_id: { type: "string", required: false },
    limit: { type: "integer", required: false },
  }
}, async (ctx) => {
  const channelId = ctx.channel && ctx.channel.id
  const messages = await ctx.discord.messages.list(channelId, {
    before: ctx.args.before_message_id || "",
    after: ctx.args.after_message_id || "",
    around: ctx.args.around_message_id || "",
    limit: ctx.args.limit || 10,
  })
  return { content: `Fetched ${messages.length} message(s).`, ephemeral: true }
})
```