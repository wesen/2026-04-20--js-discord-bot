---
Title: Discord Guild and Role Lookup Utilities Architecture and Implementation Guide
Ticket: DISCORD-BOT-011
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
      Note: Request-scoped Discord capability object will grow with guild/role lookup helpers
    - Path: internal/jsdiscord/host_ops.go
      Note: Request-scoped Discord operations are composed here and will gain guild/role lookup support
    - Path: internal/jsdiscord/host_maps.go
      Note: Guild and role normalization helpers should stay plain-map based for JS friendliness
    - Path: examples/discord-bots/moderation/lib/register-member-moderation-commands.js
      Note: Existing moderation flows are a natural consumer of role lookup context
ExternalSources: []
Summary: Detailed implementation guide for adding request-scoped guild and role lookup helpers to the Discord JS runtime.
LastUpdated: 2026-04-20T22:05:00-04:00
WhatFor: Explain the architecture, phase ordering, and implementation seams for guild and role lookup utilities.
WhenToUse: Use when implementing or reviewing DISCORD-BOT-011.
---

# Goal

Add a small, practical lookup surface that lets JavaScript moderation/admin bots inspect guild metadata and roles through request-scoped host capabilities.

# Why this ticket exists

The current moderation/admin runtime can already:

- moderate members
- inspect and moderate messages
- inspect and lightly modify channels

But moderation bots often also need lookup context before taking action, for example:

- listing roles before choosing one to add or remove
- fetching a guild snapshot to show operator-facing context
- fetching one role to confirm its name/position/flags before applying it

That is the gap this ticket fills.

# Proposed API surface

```js
const guild = await ctx.discord.guilds.fetch(guildID)
const roles = await ctx.discord.roles.list(guildID)
const role = await ctx.discord.roles.fetch(guildID, roleID)
```

# Design constraints

- Keep the APIs request-scoped under `ctx.discord`, not global.
- Return normalized plain JS maps, not raw Discordgo structs.
- Keep the first phase read-only.
- Keep the normalized data practical and compact rather than exhaustive.

# Phase ordering

## Phase 1 — guild and role lookup core

Implement:
- `guilds.fetch(guildID)`
- `roles.list(guildID)`
- `roles.fetch(guildID, roleID)`
- normalized snapshot helpers
- tests and example commands

## Phase 2 — operator docs and caveats

Document:
- expected permissions/failure modes
- normalized payload shapes
- how moderation bots can combine lookup helpers with existing member moderation APIs

# Implementation sketch

## Host capability additions

Extend `DiscordOps` with lookup functions for:
- guild fetch
- role list
- role fetch

## Host implementation

Use Discordgo session helpers to retrieve:
- guild snapshot
- guild role list

Then normalize the results into plain maps in `host_maps.go`.

## JavaScript binding

Expose two new namespaces under `ctx.discord`:

- `ctx.discord.guilds.fetch(...)`
- `ctx.discord.roles.list(...)`
- `ctx.discord.roles.fetch(...)`

## Example moderation commands

Add a few operator-facing read-only commands such as:
- `mod-fetch-guild`
- `mod-list-roles`
- `mod-fetch-role`

# Review guidance

Start review in this order:

1. `internal/jsdiscord/bot.go`
2. `internal/jsdiscord/host_ops.go`
3. `internal/jsdiscord/host_maps.go`
4. `internal/jsdiscord/runtime_test.go`
5. `examples/discord-bots/moderation/lib/register-member-moderation-commands.js`
