---
Title: Discord Component Interactions and Message Components
Ticket: DISCORD-BOT-005
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
    - Path: examples/discord-bots/ping/index.js
      Note: |-
        Example bot should demonstrate buttons and select menus with handlers
        Example bot should demonstrate component flows
    - Path: internal/jsdiscord/bot.go
      Note: |-
        JS builder API and runtime dispatch contract need new component surfaces
        JS builder/runtime contract needs component registration and dispatch
    - Path: internal/jsdiscord/descriptor.go
      Note: |-
        Discovery/help output should surface component metadata
        Discovery/help output should parse component descriptors
    - Path: internal/jsdiscord/host.go
      Note: Current interaction dispatch and outgoing component normalization live here
ExternalSources: []
Summary: Add first-class Discord component interactions to the JavaScript bot API, including outbound message components and inbound custom-id based handlers.
LastUpdated: 2026-04-20T16:15:00-04:00
WhatFor: Track the design and implementation of button and select-menu driven interaction workflows.
WhenToUse: Use when implementing or reviewing Discord message component support in the local JS API.
---


# Discord Component Interactions and Message Components

## Overview

This ticket adds the missing half of Discord components. The current JS API can emit simple button payloads, but it cannot route component clicks back into JavaScript. This ticket defines the handler model, payload model, descriptor model, and implementation plan for buttons and select menus.

## Key Docs

- `design-doc/01-discord-component-interactions-and-message-components-architecture-and-implementation-guide.md`
- `reference/01-discord-component-interactions-and-message-components-api-reference-and-planning-notes.md`
- `tasks.md`
- `changelog.md`
