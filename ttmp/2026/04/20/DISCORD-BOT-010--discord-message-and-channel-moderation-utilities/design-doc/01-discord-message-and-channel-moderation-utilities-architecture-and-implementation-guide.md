---
Title: Discord Message and Channel Moderation Utilities Architecture and Implementation Guide
Ticket: DISCORD-BOT-010
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
      Note: The request-scoped Discord capability object will grow with message and channel moderation utilities here
    - Path: internal/jsdiscord/host.go
      Note: Host moderation operations and normalization helpers for fetched messages/channels should live here
    - Path: internal/jsdiscord/runtime_test.go
      Note: Runtime tests should cover the new fetch, pin, unpin, bulk-delete, and channel utility APIs
    - Path: examples/discord-bots/moderation/index.js
      Note: The moderation example bot should expose practical commands for the new utilities
ExternalSources: []
Summary: Detailed design for the next admin-oriented Discord JS APIs after member moderation: message inspection/moderation and small channel utility helpers.
LastUpdated: 2026-04-20T20:25:00-04:00
WhatFor: Explain the priority ordering and implementation plan for message and channel moderation utilities in the JavaScript Discord runtime.
WhenToUse: Use when implementing or reviewing the next admin-oriented Discord host APIs after member moderation.
---

# Discord Message and Channel Moderation Utilities Architecture and Implementation Guide

## Executive Summary

After DISCORD-BOT-009, the JavaScript runtime already supports a meaningful moderation surface for members:

- role assignment/removal
- timeout
- kick
- ban
- unban

The next most valuable admin slice is not more guild-wide configuration. It is **message and channel moderation utilities** that complement the existing moderation bot workflows:

- fetch a message
- pin / unpin a message
- list pinned messages
- bulk delete messages
- fetch a channel
- set channel topic
- set channel slowmode

This ticket defines those APIs, prioritizes them in safe, reviewable phases, and keeps them request-scoped under `ctx.discord` just like the existing host capability model.

## Problem Statement

The current moderation bot can react to events and perform member actions, but it still lacks several day-to-day moderation tools that are often more common than kick/ban workflows:

- inspecting a specific message before acting
- pinning or unpinning a key message
- bulk deleting a spam burst
- checking channel metadata
- adjusting topic or slowmode for active moderation situations

Without these utilities, the moderation API surface is skewed toward member actions while still missing common message/channel operational tasks.

## Proposed Solution

Add a new ticket-sized admin surface under the existing request-scoped capability object:

```js
await ctx.discord.messages.fetch(channelID, messageID)
await ctx.discord.messages.pin(channelID, messageID)
await ctx.discord.messages.unpin(channelID, messageID)
await ctx.discord.messages.listPinned(channelID)
await ctx.discord.messages.bulkDelete(channelID, messageIDs)

await ctx.discord.channels.fetch(channelID)
await ctx.discord.channels.setTopic(channelID, topic)
await ctx.discord.channels.setSlowmode(channelID, seconds)
```

The implementation should remain intentionally narrow:

- only APIs that map cleanly to common moderation/operator tasks
- explicit Go-owned Discord session methods
- explicit normalization of fetched message/channel data into JavaScript-friendly maps
- no attempt to mirror the entire Discordgo surface automatically

## Priority Ordering

The phases below are ordered by practical usefulness and implementation risk.

### Phase 1 — message inspection and pinning

Highest priority, lowest risk.

Add:
- `ctx.discord.messages.fetch(channelID, messageID)`
- `ctx.discord.messages.pin(channelID, messageID)`
- `ctx.discord.messages.unpin(channelID, messageID)`
- `ctx.discord.messages.listPinned(channelID)`

Why first:
- highly useful for moderation/reporting workflows
- no destructive bulk behavior yet
- easy to validate in tests and examples
- introduces the “return fetched Discord objects into JS” pattern cleanly

### Phase 2 — message bulk deletion

Add:
- `ctx.discord.messages.bulkDelete(channelID, messageIDs)`

Why second:
- operationally important for spam cleanup
- slightly riskier because it is destructive and works on many messages
- benefits from the same message normalization / message-targeting mental model as Phase 1

### Phase 3 — channel utilities

Add:
- `ctx.discord.channels.fetch(channelID)`
- `ctx.discord.channels.setTopic(channelID, topic)`
- `ctx.discord.channels.setSlowmode(channelID, seconds)`

Why third:
- still useful and moderation-adjacent
- less urgent than direct message moderation flows
- introduces a second fetched-object normalization family (`channel`)

### Phase 4 — examples, docs, and operator playbook

Add/refresh:
- moderation example commands for the new APIs
- permissions/failure-mode notes
- debugging/runbook guidance

Why fourth:
- keeps the code slices reviewable first
- then ensures the new surfaces are actually operable and documented

## Design Rules

### 1. Keep the APIs request-scoped under `ctx.discord`

The project already chose a request-scoped capability model. These new message/channel utilities should follow that pattern rather than introducing a global Discord singleton.

### 2. Return normalized maps, not raw Discordgo structs

Fetched messages and channels should be exposed to JavaScript as plain normalized objects, not as opaque Go structs.

### 3. Prefer small exact APIs over a generic admin wrapper

This ticket should not expose a giant “admin anything” object. It should add a minimal coherent surface for the next moderation/operator tasks.

### 4. Log destructive operations clearly

Pinning is low risk, but bulk deletion is destructive enough that the host should emit explicit structured debug logs with channel IDs, message IDs/counts, and action labels.

## Proposed API Sketches

### Phase 1

```js
const message = await ctx.discord.messages.fetch(channelID, messageID)
await ctx.discord.messages.pin(channelID, messageID)
await ctx.discord.messages.unpin(channelID, messageID)
const pinned = await ctx.discord.messages.listPinned(channelID)
```

### Phase 2

```js
await ctx.discord.messages.bulkDelete(channelID, ["m1", "m2", "m3"])
```

### Phase 3

```js
const channel = await ctx.discord.channels.fetch(channelID)
await ctx.discord.channels.setTopic(channelID, "Escalation queue")
await ctx.discord.channels.setSlowmode(channelID, 30)
```

## Normalization Guidelines

### Message fetch / pinned messages

Return a normalized map that at minimum includes:

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

This can be expanded later, but the first slice should keep the shape consistent with existing event payloads.

### Channel fetch

Return a normalized map that at minimum includes:

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

## Implementation Plan

### Phase 1 — message inspection and pinning

Files:
- `internal/jsdiscord/bot.go`
- `internal/jsdiscord/host.go`
- `internal/jsdiscord/runtime_test.go`
- `examples/discord-bots/moderation/index.js`
- `examples/discord-bots/README.md`

Work:
- add `MessageFetch`, `MessagePin`, `MessageUnpin`, `MessageListPinned` to `DiscordOps`
- expose those methods through `ctx.discord.messages.*`
- add normalization helpers for fetched message data
- add runtime tests
- add moderation example commands demonstrating the APIs

### Phase 2 — message bulk deletion

Files:
- `internal/jsdiscord/bot.go`
- `internal/jsdiscord/host.go`
- `internal/jsdiscord/runtime_test.go`
- `examples/discord-bots/moderation/index.js`

Work:
- add `MessageBulkDelete`
- support `[]string` and `[]any` of string IDs as input
- add validation/logging
- add runtime tests and example command

### Phase 3 — channel utilities

Files:
- `internal/jsdiscord/bot.go`
- `internal/jsdiscord/host.go`
- `internal/jsdiscord/runtime_test.go`
- `examples/discord-bots/moderation/index.js`

Work:
- add `ChannelFetch`, `ChannelSetTopic`, `ChannelSetSlowmode`
- add channel normalization helper
- add tests and example commands

### Phase 4 — docs and operator guidance

Files:
- ticket docs
- example README
- playbook

Work:
- document permissions and common failure modes
- document destructive-operation caveats for bulk delete
- add a debugging playbook for message/channel moderation flows

## Alternatives Considered

### Continue broadening member moderation first

Possible, but lower leverage now that the core member action surface already exists.

### Jump straight to full role/channel admin CRUD

Too broad for the next slice. It would introduce more power than necessary before the common moderation utilities are in place.

### Wrap all Discordgo admin functions generically

Rejected. That would make the JS API large, inconsistent, and harder to document and review.

## Risks and Tradeoffs

### Bulk delete is destructive

The API should be explicit and well-logged. The playbook should also remind operators to test against known message IDs first.

### Fetch helpers increase normalization surface area

This is acceptable, but the first slice should keep the returned message/channel maps intentionally small.

### Channel edit helpers can still be operationally disruptive

Changing a topic or slowmode is not as destructive as deleting messages, but it still affects users directly. Logging remains important.

## Review Instructions

When reviewing implementation work for this ticket, start with:

- `internal/jsdiscord/bot.go`
- `internal/jsdiscord/host.go`
- `internal/jsdiscord/runtime_test.go`
- `examples/discord-bots/moderation/index.js`

Then verify that the ticket docs and diary match the implemented phase ordering.
