---
Title: Discord Event Expansion and Moderation/Admin APIs
Ticket: DISCORD-BOT-009
Status: active
Topics:
    - backend
    - chat
    - javascript
    - goja
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/bot/bot.go
      Note: |-
        Live Discord session handlers need to fan more event types into the JS host
        New discordgo session handlers will be added here
    - Path: internal/jsdiscord/bot.go
      Note: |-
        Runtime context objects should support richer guild/member/channel information
        Runtime context shape and host capability injection should grow here
    - Path: internal/jsdiscord/host.go
      Note: |-
        Event-specific request normalization should grow here
        Event normalization and moderation host methods should grow here
ExternalSources: []
Summary: Plan the next major expansion of inbound Discord events plus moderation and administration host APIs.
LastUpdated: 2026-04-20T16:15:00-04:00
WhatFor: Track future event expansion and moderation/admin capability work.
WhenToUse: Use when planning richer event and moderation support.
---


# Discord Event Expansion and Moderation/Admin APIs

## Overview

This ticket groups the next operationally important Discord surfaces after richer interactions: more inbound events and host-side moderation/admin operations.
