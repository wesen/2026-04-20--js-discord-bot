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
      Note: Request-scoped Discord capability object will grow with thread helpers
    - Path: internal/jsdiscord/host_ops_channels.go
      Note: Thread helpers may share channel-host implementation seams here
    - Path: internal/jsdiscord/runtime_test.go
      Note: Runtime tests should validate the new thread APIs here
    - Path: examples/discord-bots/support/index.js
      Note: Support-style examples are natural consumers of thread utilities
ExternalSources: []
Summary: Quick reference for the planned API surface, normalized payloads, and implementation ordering for DISCORD-BOT-014.
LastUpdated: 2026-04-20T22:35:00-04:00
WhatFor: Provide copy/paste-ready API sketches and implementation-phase guidance for DISCORD-BOT-014.
WhenToUse: Use when implementing or reviewing DISCORD-BOT-014.
---

# Goal

Provide quick API sketches and a phase-by-phase reference for DISCORD-BOT-014.

# Quick Reference

```js
const thread = await ctx.discord.threads.fetch(threadID)
await ctx.discord.threads.join(threadID)
await ctx.discord.threads.leave(threadID)
```

## Expected normalized shape

```js
{
  id: "...",
  parentID: "...",
  name: "...",
  archived: false
}
```
