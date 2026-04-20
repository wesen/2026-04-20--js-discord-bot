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
      Note: The request-scoped Discord capability object exposes the implemented message and channel moderation methods here
    - Path: internal/jsdiscord/host.go
      Note: Host implementations and normalization helpers for fetched messages/channels live here
ExternalSources: []
Summary: Quick reference for the implemented message and channel moderation utility APIs, payloads, and operational caveats.
LastUpdated: 2026-04-20T21:00:00-04:00
WhatFor: Provide copy/paste-ready API sketches and operator-facing notes for the implemented DISCORD-BOT-010 surface.
WhenToUse: Use when implementing, reviewing, or operating the message/channel moderation utility APIs.
---

# Goal

Provide a quick reference for the implemented DISCORD-BOT-010 message and channel moderation utilities.

# Context

The runtime already supports:

- outbound channel/message operations (`send`, `edit`, `delete`, `react`)
- member moderation operations (`addRole`, `removeRole`, `timeout`, `kick`, `ban`, `unban`)
- richer inbound moderation events

DISCORD-BOT-010 adds the next practical moderation/operator utilities around messages and channels.

# Quick Reference

## Message moderation utilities

```js
const message = await ctx.discord.messages.fetch(channelID, messageID)
await ctx.discord.messages.pin(channelID, messageID)
await ctx.discord.messages.unpin(channelID, messageID)
const pinned = await ctx.discord.messages.listPinned(channelID)
await ctx.discord.messages.bulkDelete(channelID, ["m1", "m2", "m3"])
```

## Channel utilities

```js
const channel = await ctx.discord.channels.fetch(channelID)
await ctx.discord.channels.setTopic(channelID, "Escalation queue")
await ctx.discord.channels.setSlowmode(channelID, 30)
```

## Accepted `bulkDelete(...)` payload forms

```js
await ctx.discord.messages.bulkDelete(channelID, ["m1", "m2"])
await ctx.discord.messages.bulkDelete(channelID, ["m1", "m2", "m2"])
await ctx.discord.messages.bulkDelete(channelID, { messageIds: ["m1", "m2"] })
```

Current behavior:
- trims whitespace
- removes empty IDs
- deduplicates repeated IDs
- errors if no message IDs remain after normalization

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

This shape is currently used for:
- `messages.fetch(...)`
- `messages.listPinned(...)`
- existing inbound message event payloads

## Current normalized channel shape

```js
{
  id: "...",
  guildID: "...",
  parentID: "...",
  name: "...",
  type: "...",
  topic: "...",
  nsfw: false,
  position: 0,
  rateLimitPerUser: 0,
  lastMessageID: "...",
  lastPinTimestamp: "2026-04-20T21:00:00Z",
}
```

Only fields that are present are included.

# Operational Notes

## Permission expectations

The currently implemented message/channel utilities generally require Discord permissions roughly equivalent to:

- message fetch / pinned list → read access to the channel and message history
- pin / unpin → message-management permission in the channel
- bulk delete → message-management permission in the channel
- set topic / set slowmode → channel-management permission in the channel

## Common failure modes

- wrong channel ID or message ID
- bot lacks permission in the current channel
- the command is used in the wrong channel/guild context
- bulk delete contains invalid or non-existent message IDs
- slowmode/topic edits are blocked by channel permissions or hierarchy/configuration constraints

## Current limitation notes

### `bulkDelete(...)`

Current implemented behavior:
- accepts the narrow input forms documented above
- logs message count and channel ID at debug level

Current caveat:
- the host does not currently add extra age-based preflight checks before delegating to Discord
- if Discord rejects the request, the operator will see the Discord API error through the usual host path

### `setTopic(...)`

Current implemented behavior:
- sets the topic string directly through Discordgo channel edit

### `setSlowmode(...)`

Current implemented behavior:
- sets `RateLimitPerUser` directly through Discordgo channel edit

# Usage Examples

## Example moderation-bot direction

```js
command("mod-pin", {
  options: {
    message_id: { type: "string", required: true }
  }
}, async (ctx) => {
  await ctx.discord.messages.pin(ctx.channel.id, ctx.args.message_id)
  return { content: "Pinned message.", ephemeral: true }
})

command("mod-bulk-delete", {
  options: {
    message_ids: { type: "string", required: true }
  }
}, async (ctx) => {
  const messageIds = String(ctx.args.message_ids)
    .split(",")
    .map((id) => id.trim())
    .filter(Boolean)
  await ctx.discord.messages.bulkDelete(ctx.channel.id, messageIds)
  return { content: "Deleted messages.", ephemeral: true }
})

command("mod-set-slowmode", {
  options: {
    seconds: { type: "integer", required: true }
  }
}, async (ctx) => {
  await ctx.discord.channels.setSlowmode(ctx.channel.id, ctx.args.seconds)
  return { content: "Updated slowmode.", ephemeral: true }
})
```
