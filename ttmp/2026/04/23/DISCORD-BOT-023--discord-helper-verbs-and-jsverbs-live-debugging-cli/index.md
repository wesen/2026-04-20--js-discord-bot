---
Title: Discord helper verbs and jsverbs live-debugging CLI
Ticket: DISCORD-BOT-023
Status: active
Topics:
    - discord
    - jsverbs
    - cli
    - tooling
    - diagnostics
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Research ticket for adding a standalone jsverbs-powered helper CLI to the discord-bot repository for live Discord inspection, payload debugging, and bot simulation workflows."
LastUpdated: 2026-04-23T10:15:00-04:00
WhatFor: "Track the analysis and design work for a future helper-verbs subsystem distinct from the existing defineBot-based named bot runtime."
WhenToUse: "Open this ticket when planning or reviewing Discord helper verbs, jsverbs integration, or CLI-side simulation/probing workflows."
---

# Discord helper verbs and jsverbs live-debugging CLI

## Overview

This ticket captures the research and design work for a new helper-verb subsystem in the `discord-bot` repository. The goal is to let developers write standalone JavaScript helper scripts, discover them statically through `jsverbs`, expose them through the CLI, and use them for live Discord inspection and bot simulation workflows.

The main design recommendation is to keep two distinct models:

1. `defineBot(...)` for interactive Discord bots.
2. `jsverbs` for standalone tooling verbs.

That split keeps the bot runtime clean while making room for reusable, scriptable debugging and inspection tools.

## Key Links

- Primary design guide:
  - `design-doc/01-discord-helper-verbs-and-jsverbs-live-debugging-cli-design-and-implementation-guide.md`
- Investigation diary:
  - `reference/01-investigation-diary.md`
- Current task list:
  - `tasks.md`
- Changelog:
  - `changelog.md`

## Status

Current status: **active**

Research/design deliverables are in progress. Implementation is intentionally deferred to future ticket work.

## Topics

- discord
- jsverbs
- cli
- tooling
- diagnostics

## Structure

- `design-doc/` — architecture and implementation guidance
- `reference/` — diary and supporting references
- `playbooks/` — future operator procedures if needed
- `scripts/` — ticket-local helper scripts if needed later
- `sources/` — source extracts and external notes if needed later

## Expected outcome

This ticket should leave the repository with:

1. a clear design for helper-verb discovery and CLI wiring,
2. a runtime strategy for live Discord probing and bot simulation,
3. a phased implementation plan for future engineering work,
4. a bundle suitable for review on reMarkable.
