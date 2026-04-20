---
Title: Discord Event Expansion and Moderation/Admin APIs Architecture and Implementation Guide
Ticket: DISCORD-BOT-009
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
    - Path: internal/bot/bot.go
      Note: New discordgo session handlers will be added here
    - Path: internal/jsdiscord/host.go
      Note: Event normalization and moderation host methods should live here
    - Path: internal/jsdiscord/bot.go
      Note: Runtime context shape should grow with richer event data and host capability objects
ExternalSources: []
Summary: Detailed design for broadening inbound Discord event support and exposing moderation/admin capabilities to JavaScript.
LastUpdated: 2026-04-20T16:15:00-04:00
WhatFor: Explain how to expand beyond slash commands and messageCreate into richer Discord automation and moderation workflows.
WhenToUse: Use when implementing or reviewing richer event and moderation support.
---

# Discord Event Expansion and Moderation/Admin APIs Architecture and Implementation Guide

## Executive Summary

The current runtime only exposes a small event set: `ready`, `guildCreate`, and `messageCreate`. That is enough to prove the host/runtime seam, but not enough for real operations bots. This ticket defines the next expansion layer: more inbound events plus explicit host-side moderation/admin methods for tasks like role changes and timeouts.

## Proposed Event Expansion

- `messageUpdate`
- `messageDelete`
- `reactionAdd`
- `reactionRemove`
- `guildMemberAdd`
- `guildMemberUpdate`
- `guildMemberRemove`
- thread/channel lifecycle events where practical

## Proposed Capability APIs

```js
await discord.members.addRole(guildID, userID, roleID)
await discord.members.removeRole(guildID, userID, roleID)
await discord.members.timeout(guildID, userID, { durationSeconds: 600, reason: "spam" })
```

## Design Rules

- events should remain declarative and runtime-scoped
- destructive admin actions should live behind explicit host methods, not be smuggled through generic maps
- failure messages should be clear because Discord permission errors are common operationally

## Implementation Plan

### Phase 1 — event normalization
- add more session handlers in `internal/bot/bot.go`
- normalize message, member, reaction, and thread payloads in `internal/jsdiscord/host.go`

### Phase 2 — moderation capabilities
- inject `discord.members.*` host methods
- start with role assignment and timeout helpers
- add precise logging around Discord API failures

### Phase 3 — examples and tests
- moderation-oriented example bot
- runtime tests for injected capability methods
- selected live smoke tests for role/timeout flows

## Risks

- permissions differ across guilds, so examples must explain prerequisite roles/permissions clearly
- destructive actions require careful logging and operator trust
