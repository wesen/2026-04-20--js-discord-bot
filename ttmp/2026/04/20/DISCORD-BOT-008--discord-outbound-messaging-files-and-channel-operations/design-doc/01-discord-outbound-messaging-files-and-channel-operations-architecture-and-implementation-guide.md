---
Title: Discord Outbound Messaging, Files, and Channel Operations Architecture and Implementation Guide
Ticket: DISCORD-BOT-008
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
    - Path: internal/jsdiscord/host.go
      Note: Payload normalization and Discord session operations should expand here
    - Path: internal/jsdiscord/bot.go
      Note: Host capability injection should happen here
    - Path: internal/bot/bot.go
      Note: Live bot wiring already owns the session and should remain the capability boundary
ExternalSources: []
Summary: Detailed design for outbound Discord operations beyond reply/edit/follow-up, including files, arbitrary sends, and channel/message utilities.
LastUpdated: 2026-04-20T16:15:00-04:00
WhatFor: Explain how to grow the JS API from interaction-bound responses into broader Discord host capabilities.
WhenToUse: Use when implementing or reviewing outbound Discord operations from JavaScript.
---

# Discord Outbound Messaging, Files, and Channel Operations Architecture and Implementation Guide

## Executive Summary

The current JS API is mostly interaction-bound: a handler can reply, defer, edit, or follow up to the event that invoked it. That is a good start, but real bots quickly need broader outbound capabilities such as sending a report to another channel, uploading files, reacting to a message, or creating a thread. This ticket defines that next capability layer. The first implementation slice uses a request-scoped `ctx.discord` object rather than a global runtime singleton, so Discord operations stay tied to a live session-backed handler context.

## Proposed API Direction

```js
await ctx.discord.channels.send(channelID, { content: "hello" })
await ctx.discord.messages.react(channelID, messageID, "✅")
await ctx.discord.messages.delete(channelID, messageID)
await ctx.discord.channels.send(channelID, {
  content: "report attached",
  files: [{ name: "report.txt", content: "..." }]
})
```

## Key Design Rules

- keep host-side Discord operations capability-based and explicit
- keep session ownership in Go, not in JavaScript
- keep payload shapes aligned with existing response payload objects where possible
- expose a small coherent surface first instead of a large accidental wrapper over the entire Discord SDK

## Implementation Plan

### Phase 1 — payload support
- files / attachments
- message references
- richer send/edit payload normalization

### Phase 2 — host operations
- `discord.channels.send`
- `discord.messages.edit`
- `discord.messages.delete`
- `discord.messages.react`

### Phase 3 — thread helpers
- thread creation
- basic fetch helpers
- example/reporting flows

## Testing Strategy

- pure normalization tests for file payloads
- runtime tests for injected capability objects
- limited live smoke tests for send/edit/react flows
