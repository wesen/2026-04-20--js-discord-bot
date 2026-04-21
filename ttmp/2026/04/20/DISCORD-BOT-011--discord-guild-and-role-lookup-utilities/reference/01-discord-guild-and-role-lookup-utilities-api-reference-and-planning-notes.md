---
Title: Discord Guild and Role Lookup Utilities API Reference and Planning Notes
Ticket: DISCORD-BOT-011
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
      Note: The request-scoped Discord capability object exposes the guild and role lookup helpers here
    - Path: internal/jsdiscord/host_ops_guilds.go
      Note: Guild lookup host operations live here
    - Path: internal/jsdiscord/host_ops_roles.go
      Note: Role list/fetch host operations live here
    - Path: internal/jsdiscord/host_maps.go
      Note: Normalized guild and role snapshots are returned from here
ExternalSources: []
Summary: Quick reference for the implemented guild and role lookup APIs, normalized payloads, and operational caveats.
LastUpdated: 2026-04-20T22:20:00-04:00
WhatFor: Provide copy/paste-ready API sketches and operator-facing notes for the implemented DISCORD-BOT-011 surface.
WhenToUse: Use when implementing, reviewing, or operating the guild/role lookup helper APIs.
---

# Goal

Provide a quick reference for the implemented DISCORD-BOT-011 guild and role lookup utilities.

# Quick Reference

## Guild and role lookup core

```js
const guild = await ctx.discord.guilds.fetch(guildID)
const roles = await ctx.discord.roles.list(guildID)
const role = await ctx.discord.roles.fetch(guildID, roleID)
```

## Current normalized guild shape

```js
{
  id: "...",
  name: "...",
  ownerID: "...",
  description: "...",
  memberCount: 0,
  large: false,
  verificationLevel: "...",
  features: ["..."],
  icon: "...",
  afkChannelID: "...",
  widgetChannelID: "...",
}
```

Only fields that are present are included.

## Current normalized role shape

```js
{
  id: "...",
  guildID: "...",
  name: "...",
  color: 0,
  position: 0,
  permissions: "0",
  managed: false,
  mentionable: false,
  hoist: false,
  unicodeEmoji: "",
  icon: "...",
  flags: 0,
}
```

Only fields that are present are included.

# Operational Notes

## Permission and visibility expectations

These helpers are read-only, but they still depend on the bot being able to see the target guild and its roles.

## Common failure modes

- wrong guild ID
- wrong role ID
- the bot lacks visibility into the target guild
- the bot can see the guild but Discord rejects the lookup request

## Current lookup behavior

### `guilds.fetch(guildID)`

Current implemented behavior:
- fetches one guild snapshot through Discordgo
- returns a normalized plain JS map
- logs the guild lookup at debug level

### `roles.list(guildID)`

Current implemented behavior:
- fetches the guild role list through Discordgo
- normalizes each role into a plain JS map
- logs role count at debug level

### `roles.fetch(guildID, roleID)`

Current implemented behavior:
- fetches the guild role list through Discordgo
- resolves the requested role from that list
- returns a normalized plain JS map
- returns an explicit not-found error if the role is absent

# Usage examples

```js
command("mod-fetch-guild", async (ctx) => {
  const guildId = ctx.guild && ctx.guild.id
  const guild = await ctx.discord.guilds.fetch(guildId)
  return { content: `Fetched guild ${guild.name || guild.id}.`, ephemeral: true }
})

command("mod-list-roles", async (ctx) => {
  const guildId = ctx.guild && ctx.guild.id
  const roles = await ctx.discord.roles.list(guildId)
  return { content: `Found ${roles.length} role(s).`, ephemeral: true }
})

command("mod-fetch-role", {
  options: {
    role_id: { type: "string", required: true }
  }
}, async (ctx) => {
  const guildId = ctx.guild && ctx.guild.id
  const role = await ctx.discord.roles.fetch(guildId, ctx.args.role_id)
  return { content: `Fetched role ${role.name || role.id}.`, ephemeral: true }
})
