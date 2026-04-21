---
Title: Discord Thread Utilities Architecture and Implementation Guide
Ticket: DISCORD-BOT-014
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
      Note: Request-scoped Discord capability object will grow with thread helpers
    - Path: internal/jsdiscord/host_ops_channels.go
      Note: Thread helpers may share channel-host implementation seams here
    - Path: internal/jsdiscord/runtime_test.go
      Note: Runtime tests should validate the new thread APIs here
    - Path: examples/discord-bots/support/index.js
      Note: Support-style examples are natural consumers of thread utilities
ExternalSources: []
Summary: Detailed implementation guide for discord thread utilities.
LastUpdated: 2026-04-20T22:35:00-04:00
WhatFor: Explain the architecture, phase ordering, and implementation seams for discord thread utilities.
WhenToUse: Use when implementing or reviewing DISCORD-BOT-014.
---

# Goal

Add a small, practical thread utility surface for Discord workflows that use threads as the next step after a slash command or moderation/support action.

# Why this ticket exists

Threads are a major Discord workflow primitive, especially for support and moderation follow-up, but the runtime currently has no thread-aware host helpers.

# Proposed API surface

```js
const thread = await ctx.discord.threads.fetch(threadID)
await ctx.discord.threads.join(threadID)
await ctx.discord.threads.leave(threadID)
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
