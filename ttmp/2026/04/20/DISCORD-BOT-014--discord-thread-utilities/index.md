---
Title: Discord Thread Utilities
Ticket: DISCORD-BOT-014
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
    - Path: examples/discord-bots/README.md
      Note: Example repository notes mention support thread utilities and permissions here
    - Path: examples/discord-bots/support/index.js
      Note: |-
        Support-style examples are natural consumers of thread utilities
        Support example bot demonstrates thread utility commands here
    - Path: internal/jsdiscord/bot.go
      Note: |-
        Request-scoped Discord capability object will grow with thread helpers
        Request-scoped Discord capability object exposes thread helpers
    - Path: internal/jsdiscord/host_maps.go
      Note: Normalized thread snapshots reuse and extend channel snapshot helpers here
    - Path: internal/jsdiscord/host_ops.go
      Note: Host ops builder composes thread operations here
    - Path: internal/jsdiscord/host_ops_channels.go
      Note: Thread helpers may share channel-host implementation seams here
    - Path: internal/jsdiscord/host_ops_helpers.go
      Note: Thread start payload normalization lives here
    - Path: internal/jsdiscord/host_ops_threads.go
      Note: Thread fetch/join/leave/start host operations live here
    - Path: internal/jsdiscord/runtime_test.go
      Note: |-
        Runtime tests should validate the new thread APIs here
        Runtime tests validate thread utilities and payload normalization here
ExternalSources: []
Summary: Track the next Discord JS operational utilities after message history, focused on thread fetch/join/leave/start helpers.
LastUpdated: 2026-04-20T22:35:00-04:00
WhatFor: Organize planning and implementation work for thread utility APIs.
WhenToUse: Use when implementing or reviewing the next Discord JS thread helper slice after DISCORD-BOT-013.
---



# Discord Thread Utilities

## Overview

This ticket captures the next practical Discord JS surface after message history/listing helpers. The focus is intentionally operational: support real support/community workflows that move discussions into threads.

## Key Links

- `design-doc/01-discord-thread-utilities-architecture-and-implementation-guide.md`
- `reference/01-discord-thread-utilities-api-reference-and-planning-notes.md`
- `reference/02-diary.md`
- `tasks.md`
- `changelog.md`
