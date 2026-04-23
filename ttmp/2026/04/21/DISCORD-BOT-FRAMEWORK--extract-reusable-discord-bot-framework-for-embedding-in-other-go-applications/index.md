---
Title: Extract Reusable Discord Bot Framework for Embedding in Other Go Applications
Ticket: DISCORD-BOT-FRAMEWORK
Status: active
Topics:
    - discord
    - goja
    - javascript
    - go
    - framework
    - embedding
    - glazed
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/bot/bot.go
      Note: Live bot host wiring — session creation
    - Path: pkg/botcli/command_root.go
      Note: Public repo-driven command tree and host-managed bot run orchestration
    - Path: internal/config/config.go
      Note: Settings struct — Glazed-backed bot credentials
    - Path: internal/jsdiscord/bot_compile.go
      Note: BotHandle
    - Path: internal/jsdiscord/host.go
      Note: Core Host struct — owns goja runtime and BotHandle
    - Path: internal/jsdiscord/runtime.go
      Note: Registrar and RuntimeState — module registration via goja require()
ExternalSources: []
Summary: ""
LastUpdated: 2026-04-21T22:38:26.882232734-04:00
WhatFor: ""
WhenToUse: ""
---


# Extract Reusable Discord Bot Framework for Embedding in Other Go Applications

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- discord
- goja
- javascript
- go
- framework
- embedding
- glazed

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
