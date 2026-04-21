---
Title: Discord Channel, Thread, and Role Event Expansion Architecture and Implementation Guide
Ticket: DISCORD-BOT-017
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
    - Path: internal/jsdiscord/host_dispatch.go
      Note: New event dispatch entrypoints will live here
    - Path: internal/jsdiscord/host_maps.go
      Note: Normalized event payload maps will live here
    - Path: internal/jsdiscord/runtime_test.go
      Note: Runtime tests should validate the new event dispatch coverage here
    - Path: examples/discord-bots/moderation/lib/register-events.js
      Note: Event-heavy examples can demonstrate the new lifecycle handlers here
ExternalSources: []
Summary: Detailed implementation guide for discord channel, thread, and role event expansion.
LastUpdated: 2026-04-20T22:35:00-04:00
WhatFor: Explain the architecture, phase ordering, and implementation seams for discord channel, thread, and role event expansion.
WhenToUse: Use when implementing or reviewing DISCORD-BOT-017.
---

# Goal

Add a small, useful set of Discord lifecycle events for bots that need to react to structural guild changes.

# Why this ticket exists

The runtime already has good message/member/reaction coverage, but still lacks structural channel/thread/role lifecycle events.

# Proposed API surface

```js
event("channelCreate", async (ctx) => { ... })
event("threadUpdate", async (ctx) => { ... })
event("roleDelete", async (ctx) => { ... })
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
