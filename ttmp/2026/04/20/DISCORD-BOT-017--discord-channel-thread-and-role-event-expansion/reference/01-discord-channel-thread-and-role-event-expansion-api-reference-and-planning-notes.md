---
Title: Discord Channel, Thread, and Role Event Expansion API Reference and Planning Notes
Ticket: DISCORD-BOT-017
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
    - Path: internal/jsdiscord/host_dispatch.go
      Note: New event dispatch entrypoints will live here
    - Path: internal/jsdiscord/host_maps.go
      Note: Normalized event payload maps will live here
    - Path: internal/jsdiscord/runtime_test.go
      Note: Runtime tests should validate the new event dispatch coverage here
    - Path: examples/discord-bots/moderation/lib/register-events.js
      Note: Event-heavy examples can demonstrate the new lifecycle handlers here
ExternalSources: []
Summary: Quick reference for the planned API surface, normalized payloads, and implementation ordering for DISCORD-BOT-017.
LastUpdated: 2026-04-20T22:35:00-04:00
WhatFor: Provide copy/paste-ready API sketches and implementation-phase guidance for DISCORD-BOT-017.
WhenToUse: Use when implementing or reviewing DISCORD-BOT-017.
---

# Goal

Provide quick API sketches and a phase-by-phase reference for DISCORD-BOT-017.

# Quick Reference

```js
event("channelCreate", async (ctx) => { ... })
event("threadUpdate", async (ctx) => { ... })
event("roleDelete", async (ctx) => { ... })
```

## Expected normalized shape

```js
{
  id: "...",
  name: "..."
}
```
