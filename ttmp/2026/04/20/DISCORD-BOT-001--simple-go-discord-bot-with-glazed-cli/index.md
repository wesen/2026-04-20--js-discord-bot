---
Title: Simple Go Discord Bot with Glazed CLI
Ticket: DISCORD-BOT-001
Status: active
Topics:
    - backend
    - chat
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: examples/bots/README.md
      Note: Local example bot repository and smoke commands for the moved bot CLI
    - Path: internal/botcli/command.go
      Note: Local host-side bot CLI command surface now lives in this repository
    - Path: ttmp/2026/04/20/DISCORD-BOT-001--simple-go-discord-bot-with-glazed-cli/analysis/01-glazed-documentation-feedback-and-env-loading-gotcha.md
      Note: Maintainer-facing writeup of the Glazed documentation gap and env-loading gotcha
    - Path: ttmp/2026/04/20/DISCORD-BOT-001--simple-go-discord-bot-with-glazed-cli/design-doc/01-implementation-and-architecture-guide.md
      Note: Primary architecture and implementation guide for the Discord bot
    - Path: ttmp/2026/04/20/DISCORD-BOT-001--simple-go-discord-bot-with-glazed-cli/playbook/01-local-validation-and-smoke-test-checklist.md
      Note: Local smoke-test checklist for the bot CLI and gateway flow
    - Path: ttmp/2026/04/20/DISCORD-BOT-001--simple-go-discord-bot-with-glazed-cli/reference/01-diary.md
      Note: Chronological work log for the ticket
    - Path: ttmp/2026/04/20/DISCORD-BOT-001--simple-go-discord-bot-with-glazed-cli/reference/02-discord-credentials-and-setup.md
      Note: Credential checklist and setup instructions
    - Path: ttmp/2026/04/20/DISCORD-BOT-001--simple-go-discord-bot-with-glazed-cli/sources/ARTICLE - Playbook - Adding jsverbs to Arbitrary Go Glazed Tools.md
      Note: Ticket-local copy of the reusable jsverbs integration playbook
    - Path: ttmp/2026/04/20/DISCORD-BOT-001--simple-go-discord-bot-with-glazed-cli/sources/PROJ - JS Discord Bot - Adding jsverbs Support.md
      Note: Ticket-local copy of the jsverbs-support project report
    - Path: ttmp/2026/04/20/DISCORD-BOT-001--simple-go-discord-bot-with-glazed-cli/sources/PROJ - JS Discord Bot - Building a Discord Bot with a JavaScript API.md
      Note: Ticket-local copy of the JS API project report
ExternalSources: []
Summary: Ticket overview for the simple Go Discord bot plan, docs, and working notes.
LastUpdated: 2026-04-20T10:04:14.202445006-04:00
WhatFor: Track the main documents and status for the Discord bot ticket.
WhenToUse: Use as the landing page for the ticket workspace.
---






# Simple Go Discord Bot with Glazed CLI

## Overview

This ticket captures the first-pass plan for a simple Discord bot written in Go. The bot is designed to start from a Glazed command tree, keep the runtime minimal, and use clear documentation for setup and credentials.

## Key Links

- [Implementation and Architecture Guide](./design-doc/01-implementation-and-architecture-guide.md)
- [Discord Credentials and Setup](./reference/02-discord-credentials-and-setup.md)
- [Diary](./reference/01-diary.md)
- [Local Validation and Smoke Test Checklist](./playbook/01-local-validation-and-smoke-test-checklist.md)
- [Glazed Documentation Feedback and Env-Loading Gotcha](./analysis/01-glazed-documentation-feedback-and-env-loading-gotcha.md)
- [Copied source note: jsverbs integration playbook](./sources/ARTICLE%20-%20Playbook%20-%20Adding%20jsverbs%20to%20Arbitrary%20Go%20Glazed%20Tools.md)
- [Copied source note: Discord bot with JS API](./sources/PROJ%20-%20JS%20Discord%20Bot%20-%20Building%20a%20Discord%20Bot%20with%20a%20JavaScript%20API.md)
- [Copied source note: Discord bot jsverbs support](./sources/PROJ%20-%20JS%20Discord%20Bot%20-%20Adding%20jsverbs%20Support.md)

## Status

Current status: **active**

## Topics

- backend
- chat

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design-doc/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
