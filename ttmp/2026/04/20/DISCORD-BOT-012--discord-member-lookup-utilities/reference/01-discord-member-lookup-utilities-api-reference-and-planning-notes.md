---
Title: Discord Member Lookup Utilities API Reference and Planning Notes
Ticket: DISCORD-BOT-012
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
      Note: The request-scoped Discord capability object exposes the member lookup helpers here
    - Path: internal/jsdiscord/host_ops_members.go
      Note: Member fetch/list host operations live here
    - Path: internal/jsdiscord/host_maps.go
      Note: Normalized member snapshots are returned from here
ExternalSources: []
Summary: Quick reference for the implemented member lookup APIs, normalized payloads, and operational caveats.
LastUpdated: 2026-04-20T22:45:00-04:00
WhatFor: Provide copy/paste-ready API sketches and operator-facing notes for the implemented DISCORD-BOT-012 surface.
WhenToUse: Use when implementing, reviewing, or operating the member lookup helper APIs.
---

# Goal

Provide a quick reference for the implemented DISCORD-BOT-012 member lookup utilities.

# Quick Reference

## Member lookup core

```js
const member = await ctx.discord.members.fetch(guildID, userID)
const members = await ctx.discord.members.list(guildID)
const page = await ctx.discord.members.list(guildID, { after: "user-123", limit: 25 })
```

## Current normalized member shape

```js
{
  id: "...",
  guildId: "...",
  nick: "...",
  roles: ["..."],
  pending: false,
  deaf: false,
  mute: false,
  joinedAt: "...",
  user: {
    id: "...",
    username: "...",
    discriminator: "...",
    bot: false,
  }
}
```

Only fields that are present are included.

# Operational Notes

## Visibility expectations

These helpers are read-only, but they still depend on the bot being able to see the target guild and its members.

## Current lookup behavior

### `members.fetch(guildID, userID)`

Current implemented behavior:
- fetches one member snapshot through Discordgo
- returns a normalized plain JS map
- logs the member lookup at debug level

### `members.list(guildID, payload?)`

Current implemented behavior:
- accepts `nil` or `{ after, limit }`
- defaults to `limit: 25`
- clamps limit into a practical bounded range
- fetches a page of guild members through Discordgo
- returns normalized plain JS maps
- logs list options and returned count at debug level

## Common failure modes

- wrong guild ID
- wrong user ID
- the bot lacks visibility into the target guild members
- list pagination cursor does not match a useful page boundary

# Usage examples

```js
command("mod-fetch-member", {
  options: {
    user_id: { type: "string", required: true }
  }
}, async (ctx) => {
  const guildId = ctx.guild && ctx.guild.id
  const member = await ctx.discord.members.fetch(guildId, ctx.args.user_id)
  return { content: `Fetched member ${member.id}.`, ephemeral: true }
})

command("mod-list-members", {
  options: {
    limit: { type: "integer", required: false },
    after_user_id: { type: "string", required: false },
  }
}, async (ctx) => {
  const guildId = ctx.guild && ctx.guild.id
  const members = await ctx.discord.members.list(guildId, {
    after: ctx.args.after_user_id || "",
    limit: ctx.args.limit || 10,
  })
  return { content: `Fetched ${members.length} member(s).`, ephemeral: true }
})
