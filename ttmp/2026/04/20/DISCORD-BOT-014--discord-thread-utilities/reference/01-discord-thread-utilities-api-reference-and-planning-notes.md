---
Title: Discord Thread Utilities API Reference and Planning Notes
Ticket: DISCORD-BOT-014
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
      Note: The request-scoped Discord capability object exposes the thread helpers here
    - Path: internal/jsdiscord/host_ops_threads.go
      Note: Thread fetch/join/leave/start host operations live here
    - Path: internal/jsdiscord/host_ops_helpers.go
      Note: Thread start payload normalization lives here
    - Path: internal/jsdiscord/host_maps.go
      Note: Thread snapshots reuse the normalized channel snapshot helper here
    - Path: internal/jsdiscord/runtime_test.go
      Note: Runtime tests validate the thread utility APIs here
    - Path: examples/discord-bots/support/index.js
      Note: The support example bot demonstrates the implemented thread helpers here
ExternalSources: []
Summary: Quick reference for the implemented thread utility APIs, normalized payloads, and operational caveats.
LastUpdated: 2026-04-20T23:05:00-04:00
WhatFor: Provide copy/paste-ready API sketches and operator-facing notes for the implemented DISCORD-BOT-014 surface.
WhenToUse: Use when implementing, reviewing, or operating the thread utility APIs.
---

# Goal

Provide a quick reference for the implemented DISCORD-BOT-014 thread utilities.

# Quick Reference

```js
const thread = await ctx.discord.threads.fetch(threadID)
await ctx.discord.threads.join(threadID)
await ctx.discord.threads.leave(threadID)
const started = await ctx.discord.threads.start(channelID, {
  name: "Support follow-up",
  type: "public",
  autoArchiveDuration: 60,
})
```

## Current normalized thread shape

```js
{
  id: "...",
  guildID: "...",
  parentID: "...",
  ownerID: "...",
  name: "...",
  type: "...",
  archived: false,
  locked: false,
  invitable: true,
  autoArchiveDuration: 60,
  archiveTimestamp: "...",
  messageCount: 0,
  memberCount: 0,
}
```

Only fields that are present are included.

# Operational Notes

## Visibility and permission expectations

These helpers depend on the bot being able to view the target thread and, for creation/join/leave flows, the bot also needs the relevant thread participation or creation permissions for the target guild/channel.

## Current helper behavior

### `threads.fetch(threadID)`
- fetches the channel by ID
- rejects non-thread channels
- returns a normalized thread snapshot

### `threads.join(threadID)`
- joins the current bot user to the thread
- logs the join action at debug level

### `threads.leave(threadID)`
- removes the current bot user from the thread
- logs the leave action at debug level

### `threads.start(channelID, payload)`
Current implemented behavior:
- accepts a string payload as a shorthand thread name, or an object payload
- supports:
  - `name`
  - `messageId`
  - `type`
  - `autoArchiveDuration`
  - `invitable`
  - `rateLimitPerUser`
- starts a message thread when `messageId` is present
- otherwise starts a regular thread from the channel
- defaults to a public thread when no explicit type is provided for a channel-start flow

## Current ticket boundary decision

Archive/lock management does **not** belong in this ticket. This ticket now covers:
- fetch
- join
- leave
- start

Archive/lock lifecycle control should stay for a later focused follow-up rather than expanding this ticket into a broader thread-admin surface.

## Common failure modes

- wrong thread ID or channel ID
- trying to fetch a non-thread channel through `threads.fetch(...)`
- the bot lacks permission to view, join, or create the target thread
- using an invalid `type` value for thread start
- using a missing or invalid `messageId` for message-thread creation

# Usage examples

```js
command("support-fetch-thread", {
  options: {
    thread_id: { type: "string", required: true }
  }
}, async (ctx) => {
  const thread = await ctx.discord.threads.fetch(ctx.args.thread_id)
  return { content: `Fetched thread ${thread.name || thread.id}.`, ephemeral: true }
})

command("support-start-thread", {
  options: {
    name: { type: "string", required: true },
    source_message_id: { type: "string", required: false },
    type: { type: "string", required: false },
    auto_archive_duration: { type: "integer", required: false },
  }
}, async (ctx) => {
  const channelId = ctx.channel && ctx.channel.id
  const thread = await ctx.discord.threads.start(channelId, {
    name: ctx.args.name,
    messageId: ctx.args.source_message_id || "",
    type: ctx.args.type || "public",
    autoArchiveDuration: ctx.args.auto_archive_duration || 1440,
  })
  return { content: `Started thread ${thread.name || thread.id}.`, ephemeral: true }
})
```