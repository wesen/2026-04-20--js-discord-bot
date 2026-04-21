---
Title: Discord Member Lookup Utilities Architecture and Implementation Guide
Ticket: DISCORD-BOT-012
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
      Note: Request-scoped Discord capability object will grow with member lookup helpers
    - Path: internal/jsdiscord/host_ops_members.go
      Note: Member lookup and member moderation helpers should stay grouped in one host-ops file
    - Path: internal/jsdiscord/host_maps.go
      Note: Member normalization helpers will be reused for read-only member lookup snapshots
    - Path: examples/discord-bots/moderation/lib/register-member-moderation-commands.js
      Note: Existing moderation flows are a natural consumer of member lookup context
ExternalSources: []
Summary: Detailed implementation guide for adding request-scoped member lookup helpers to the Discord JS runtime.
LastUpdated: 2026-04-20T22:35:00-04:00
WhatFor: Explain the architecture, phase ordering, and implementation seams for member lookup utilities.
WhenToUse: Use when implementing or reviewing DISCORD-BOT-012.
---

# Goal

Add a small, practical lookup surface that lets JavaScript moderation/admin bots inspect one member or a small member list through request-scoped host capabilities.

# Why this ticket exists

The runtime already supports:

- member moderation operations
- guild lookup
- role lookup
- message/channel moderation helpers

But moderation bots often also need read-only member details before taking action, for example:

- fetch one member to confirm roles, nickname, pending state, or joined time
- list a small page of members for operator triage or preview

That is the gap this ticket fills.

# Proposed API surface

```js
const member = await ctx.discord.members.fetch(guildID, userID)
const members = await ctx.discord.members.list(guildID)
const preview = await ctx.discord.members.list(guildID, { after: "user-123", limit: 25 })
```

# Design constraints

- Keep the APIs request-scoped under `ctx.discord`, not global.
- Return normalized plain JS maps, not raw Discordgo structs.
- Keep the first phase read-only.
- Keep list input narrow and predictable.

# Phase ordering

## Phase 1 — member lookup core

Implement:
- `members.fetch(guildID, userID)`
- `members.list(guildID, payload?)`
- tests and example commands

## Phase 2 — operator docs and caveats

Document:
- expected permissions/failure modes
- normalized payload shapes
- how member lookup combines with existing moderation helpers

# Review guidance

Start review in this order:

1. `internal/jsdiscord/bot.go`
2. `internal/jsdiscord/host_ops_members.go`
3. `internal/jsdiscord/host_maps.go`
4. `internal/jsdiscord/runtime_test.go`
5. `examples/discord-bots/moderation/lib/register-member-moderation-commands.js`
