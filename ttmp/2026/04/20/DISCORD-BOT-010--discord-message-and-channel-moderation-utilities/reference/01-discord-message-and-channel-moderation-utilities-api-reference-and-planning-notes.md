---
Title: Discord Message and Channel Moderation Utilities API Reference and Planning Notes
Ticket: DISCORD-BOT-010
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
      Note: The request-scoped Discord capability object will expose the new message and channel moderation methods here
    - Path: internal/jsdiscord/host.go
      Note: Host implementations and normalization helpers for fetched messages/channels will live here
ExternalSources: []
Summary: Quick reference for the planned message and channel moderation utility APIs, payloads, and phase ordering.
LastUpdated: 2026-04-20T20:25:00-04:00
WhatFor: Provide copy/paste-ready API sketches and implementation-phase guidance for DISCORD-BOT-010.
WhenToUse: Use when implementing or reviewing the message/channel moderation utility ticket.
---

# Goal

Provide quick API sketches and a phase-by-phase reference for the next moderation/admin ticket after DISCORD-BOT-009.

# Context

The runtime already supports:

- outbound channel/message operations (`send`, `edit`, `delete`, `react`)
- member moderation operations (`addRole`, `removeRole`, `timeout`, `kick`, `ban`, `unban`)
- richer inbound moderation events

This ticket adds the next practical moderation/operator utilities around messages and channels.

# Quick Reference

## Phase 1 — message inspection and pinning

```js
const message = await ctx.discord.messages.fetch(channelID, messageID)
await ctx.discord.messages.pin(channelID, messageID)
await ctx.discord.messages.unpin(channelID, messageID)
const pinned = await ctx.discord.messages.listPinned(channelID)
```

## Phase 2 — message bulk deletion

```js
await ctx.discord.messages.bulkDelete(channelID, ["m1", "m2", "m3"])
```

## Phase 3 — channel utilities

```js
const channel = await ctx.discord.channels.fetch(channelID)
await ctx.discord.channels.setTopic(channelID, "Escalation queue")
await ctx.discord.channels.setSlowmode(channelID, 30)
```

## Expected normalized message shape

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

## Expected normalized channel shape

```js
{
  id: "...",
  guildID: "...",
  parentID: "...",
  name: "...",
  type: "...",
  topic: "...",
  nsfw: false,
  rateLimitPerUser: 0,
}
```

## Priority ordering

1. `messages.fetch`, `pin`, `unpin`, `listPinned`
2. `messages.bulkDelete`
3. `channels.fetch`, `setTopic`, `setSlowmode`
4. docs / playbook / operator guidance

# Usage Examples

## Example moderation-bot direction

```js
command("mod-pin", async (ctx) => {
  await ctx.discord.messages.pin(ctx.channel.id, ctx.args.message_id)
  return { content: "Pinned message.", ephemeral: true }
})

command("mod-bulk-delete", async (ctx) => {
  await ctx.discord.messages.bulkDelete(ctx.channel.id, ctx.args.message_ids)
  return { content: "Deleted messages.", ephemeral: true }
})

command("mod-set-slowmode", async (ctx) => {
  await ctx.discord.channels.setSlowmode(ctx.channel.id, ctx.args.seconds)
  return { content: "Updated slowmode.", ephemeral: true }
})
```
