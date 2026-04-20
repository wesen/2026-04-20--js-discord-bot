---
Title: Run Named Bot Implementations from Repositories
Ticket: DISCORD-BOT-003
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
      Note: Example repository for named bot implementations
    - Path: internal/bot/bot.go
      Note: Live Discord host entrypoint that will need to load one or more named bot scripts
    - Path: internal/botcli/bootstrap.go
      Note: Repository scanning and named bot discovery live here
    - Path: internal/botcli/command.go
      Note: Bot-oriented CLI surface to be rewritten around named bot implementations
    - Path: internal/jsdiscord/descriptor.go
      Note: Bot descriptor and script inspection model added in this ticket
    - Path: internal/jsdiscord/host.go
      Note: Existing single-bot Discord JS host that will be extended toward multi-bot composition
    - Path: internal/jsdiscord/multihost.go
      Note: Multi-bot runtime composition and command routing live here
ExternalSources: []
Summary: Replace the verb-oriented bots CLI with a named bot implementation runner that can discover, inspect, and run one or more bot scripts from repositories.
LastUpdated: 2026-04-20T14:55:00-04:00
WhatFor: Track the bot-repository runner work and the example bot implementations used to validate it.
WhenToUse: Use when implementing or reviewing named bot discovery, help, and run behavior in the Discord bot CLI.
---


# Run Named Bot Implementations from Repositories

## Overview

This ticket changes the meaning of `discord-bot bots ...` from “run individual JS functions” to “discover and run named bot implementations.” A bot implementation is now the top-level runtime unit. Running a bot starts the real long-lived Discord bot host for that implementation; later, multiple implementations may run in one process.

## Key Links

- `design-doc/01-bot-repository-runner-architecture-and-implementation-guide.md`
- `reference/01-bot-repository-cli-reference-and-example-repositories.md`
- `reference/02-diary.md`
- `tasks.md`
- `changelog.md`
