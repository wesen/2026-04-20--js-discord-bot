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

```js
event("reactionAdd", async (ctx) => {
  ctx.log.info("reaction added", { emoji: ctx.reaction && ctx.reaction.emoji })
})

event("guildMemberAdd", async (ctx) => {
  await discord.members.addRole(ctx.guild.id, ctx.member.user.id, WELCOME_ROLE_ID)
})

await discord.members.timeout(guildID, userID, {
  durationSeconds: 600,
  reason: "spam"
})
```
