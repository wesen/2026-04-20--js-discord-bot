---
Title: Discord Message and Channel Moderation Utilities
Ticket: DISCORD-BOT-010
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
    - Path: examples/discord-bots/moderation/index.js
      Note: |-
        The moderation example bot should demonstrate the new message and channel utilities
        Moderation example bot will demonstrate the new utilities
    - Path: internal/jsdiscord/bot.go
      Note: |-
        The request-scoped Discord capability object will gain the new message and channel moderation APIs here
        Request-scoped Discord capability object will grow with message and channel moderation utilities
    - Path: internal/jsdiscord/host.go
      Note: |-
        Host operations and normalization helpers for fetched messages and channels will grow here
        Host moderation operations and normalization helpers will grow here
    - Path: internal/jsdiscord/runtime_test.go
      Note: |-
        Runtime tests should validate the new moderation utility APIs here
        Runtime tests will validate the new message and channel moderation APIs
ExternalSources: []
Summary: Track the next moderation/admin Discord JS APIs after member moderation, focused on message inspection/moderation and small channel utility helpers.
LastUpdated: 2026-04-20T20:25:00-04:00
WhatFor: Organize planning and implementation work for message and channel moderation utility APIs.
WhenToUse: Use when implementing or reviewing the next Discord moderation utility slice after DISCORD-BOT-009.
---


# Discord Message and Channel Moderation Utilities

## Overview

This ticket captures the next admin-oriented Discord JS APIs after member moderation. The focus is intentionally narrow and practical: common message moderation utilities and small channel helper operations.

## Key Links

- `design-doc/01-discord-message-and-channel-moderation-utilities-architecture-and-implementation-guide.md`
- `reference/01-discord-message-and-channel-moderation-utilities-api-reference-and-planning-notes.md`
- `reference/02-diary.md`
- `playbook/01-debugging-message-and-channel-moderation-flows.md`
- `tasks.md`
- `changelog.md`
