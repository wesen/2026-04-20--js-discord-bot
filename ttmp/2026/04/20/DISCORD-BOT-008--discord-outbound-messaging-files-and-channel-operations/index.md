---
Title: Discord Outbound Messaging, Files, and Channel Operations
Ticket: DISCORD-BOT-008
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
        Live bot runtime should expose richer Discord operations through the JS host
        Live session ownership should remain in Go while exposing richer capabilities
    - Path: internal/jsdiscord/bot.go
      Note: |-
        New host capability objects should be injected here
        Host capability objects should be injected here
    - Path: internal/jsdiscord/host.go
      Note: |-
        Current reply/edit/follow-up payload normalization should grow into broader outbound operations
        Outbound payload normalization and Discord session operations should expand here
ExternalSources: []
Summary: Plan the next host-side capability layer for files, arbitrary sends, message edits, reactions, and channel operations from JavaScript.
LastUpdated: 2026-04-20T16:15:00-04:00
WhatFor: Track future outbound Discord operations beyond direct interaction replies.
WhenToUse: Use when planning richer outbound Discord operations.
---


# Discord Outbound Messaging, Files, and Channel Operations

## Overview

This ticket is the home for the next major host-capability expansion after interactions are richer: sending files, posting to arbitrary channels, editing or deleting messages, reacting, and eventually thread/channel helpers.
