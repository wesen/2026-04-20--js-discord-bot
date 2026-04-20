---
Title: Discord Event Expansion and Moderation/Admin APIs API Reference and Planning Notes
Ticket: DISCORD-BOT-009
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
    - Path: internal/bot/bot.go
      Note: Session event handlers should expand here
    - Path: internal/jsdiscord/host.go
      Note: Event payload normalization and moderation methods should live here
ExternalSources: []
Summary: Quick reference for future event names and moderation/admin capability APIs.
LastUpdated: 2026-04-20T16:15:00-04:00
WhatFor: Provide copy/paste-oriented sketches for event and moderation APIs.
WhenToUse: Use when planning richer event and moderation support.
---

# Quick API Sketches

## Event handlers

```js
event("messageUpdate", async (ctx) => {
  ctx.log.info("message updated", {
    before: ctx.before && ctx.before.content,
    after: ctx.message && ctx.message.content,
  })
})

event("reactionAdd", async (ctx) => {
  ctx.log.info("reaction added", {
    emoji: ctx.reaction && ctx.reaction.emoji && ctx.reaction.emoji.name,
    userId: ctx.user && ctx.user.id,
  })
})

event("guildMemberAdd", async (ctx) => {
  ctx.log.info("member joined", {
    userId: ctx.user && ctx.user.id,
    roles: ctx.member && ctx.member.roles,
  })
})
```

## Member moderation APIs

```js
await ctx.discord.members.addRole(guildID, userID, roleID)
await ctx.discord.members.removeRole(guildID, userID, roleID)
await ctx.discord.members.timeout(guildID, userID, {
  durationSeconds: 600,
})
await ctx.discord.members.timeout(guildID, userID, {
  clear: true,
})
```

## Current payload notes

### `ctx.before`

When Discordgo has cached prior state, update-style events may expose a prior snapshot through `ctx.before`.

Current examples:
- `messageUpdate`
- `messageDelete`
- `guildMemberUpdate`

The field may be empty when Discordgo does not have cached previous state.

### `ctx.reaction`

Current normalized reaction shape:

```js
{
  userId: "123",
  messageId: "456",
  channelId: "789",
  guildId: "012",
  emoji: {
    id: "",
    name: "🔥",
    animated: false,
  }
}
```

### `ctx.member`

Current normalized member shape includes:

```js
{
  id: "123",
  guildId: "456",
  nick: "mod-user",
  roles: ["role-a", "role-b"],
  pending: false,
  deaf: false,
  mute: false,
  joinedAt: "2026-04-20T19:30:00Z",
  user: {
    id: "123",
    username: "manuel",
    discriminator: "0001",
    bot: false,
  }
}
```

## Current operational notes

### Gateway intents

The current DISCORD-BOT-009 event slices rely on these intents in the live bot:

- `GuildMessages`
- `GuildMessageReactions`
- `GuildMembers`
- `MessageContent`

Without the relevant intents, some event families simply will not arrive.

### Permission expectations for host moderation APIs

The first moderation/admin host methods require Discord permissions roughly equivalent to:

- add/remove role operations → role-management permission and valid role hierarchy
- member timeout operations → timeout/moderation permission and valid role hierarchy

Common failure mode categories include:

- the bot lacks the required guild permission
- the bot’s highest role is lower than the target role/member
- the command is used outside a guild context
- the provided Discord IDs are wrong or do not belong to the selected guild

### Current limitation

The first `timeout(...)` slice supports:

- `durationSeconds`
- `until` (RFC3339 timestamp)
- `clear: true`

It does **not** yet carry an audit-log reason parameter through to Discord.
