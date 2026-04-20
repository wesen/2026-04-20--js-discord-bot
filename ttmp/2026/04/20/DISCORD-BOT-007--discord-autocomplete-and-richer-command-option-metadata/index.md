---
Title: Discord Autocomplete and Richer Command Option Metadata
Ticket: DISCORD-BOT-007
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
      Note: Example bot should demonstrate autocomplete flows
    - Path: internal/jsdiscord/bot.go
      Note: |-
        JS builder/runtime layer needs autocomplete registrations
        JS runtime contract needs autocomplete registrations
    - Path: internal/jsdiscord/descriptor.go
      Note: Discovery/help output should surface autocomplete and richer option metadata
    - Path: internal/jsdiscord/host.go
      Note: |-
        Application command option normalization and autocomplete response handling belong here
        Command option normalization and autocomplete response handling live here
    - Path: internal/jsdiscord/runtime_test.go
      Note: Runtime tests should prove focused option dispatch and choice normalization
ExternalSources: []
Summary: Add autocomplete handlers and richer slash-command option metadata to the local JavaScript Discord bot API.
LastUpdated: 2026-04-20T16:15:00-04:00
WhatFor: Track the design and implementation of Discord autocomplete and richer option schemas.
WhenToUse: Use when implementing or reviewing autocomplete support and richer command option metadata.
---


# Discord Autocomplete and Richer Command Option Metadata

## Overview

This ticket adds runtime autocomplete handling and expands the command option schema beyond the current basic type/required/description shape.
