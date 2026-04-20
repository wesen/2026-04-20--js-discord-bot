---
Title: Discord Outbound Messaging, Files, and Channel Operations API Reference and Planning Notes
Ticket: DISCORD-BOT-008
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
    - Path: internal/jsdiscord/host.go
      Note: Outbound payload normalization and session methods live here
ExternalSources: []
Summary: Quick reference for the future outbound messaging and file APIs.
LastUpdated: 2026-04-20T16:15:00-04:00
WhatFor: Provide copy/paste-oriented API sketches for future outbound operations.
WhenToUse: Use when planning outbound Discord operations.
---

# Quick API Sketches

```js
await ctx.discord.channels.send(channelID, { content: "hello" })
await ctx.discord.channels.send(channelID, {
  content: "report",
  files: [{ name: "report.txt", content: "hello" }]
})
await discord.messages.edit(channelID, messageID, { content: "updated" })
await ctx.discord.messages.delete(channelID, messageID)
await ctx.discord.messages.react(channelID, messageID, "✅")
```
