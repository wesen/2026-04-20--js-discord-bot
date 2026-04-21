---
Title: Discord Archive Helper Bot — Thread & Message Markdown Downloader
Ticket: DISCORD-ARCHIVE-BOT
Status: active
Topics:
    - javascript
    - backend
    - architecture
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/21/DISCORD-ARCHIVE-BOT--discord-archive-helper-bot-thread-message-markdown-downloader/design-doc/01-analysis-design-implementation-guide.md:Comprehensive guide for new interns
    - /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/21/DISCORD-ARCHIVE-BOT--discord-archive-helper-bot-thread-message-markdown-downloader/reference/01-diary.md:Implementation diary
ExternalSources: []
Summary: "Build a Discord bot that downloads channels and threads as clean Markdown archives. Target: new-intern-friendly documentation and offline reMarkable reading."
LastUpdated: 2026-04-21T08:08:53.931206202-04:00
WhatFor: "Onboarding resource and implementation blueprint for a Discord archive helper bot."
WhenToUse: "When starting implementation, onboarding a new intern, or reviewing the archive bot design."
---

# Discord Archive Helper Bot — Thread & Message Markdown Downloader

## Overview

This ticket tracks the design and implementation of a **Discord Archive Helper Bot** — a Node.js CLI tool that connects to Discord, fetches messages from channels and threads, and writes them as structured Markdown files for offline archival. The primary deliverable is a comprehensive, intern-friendly guide covering Discord API concepts, architecture, rendering logic, and a step-by-step implementation plan. The guide has been uploaded to reMarkable for offline annotation.

## Key Documents

| Document | Path | Description |
|----------|------|-------------|
| **Analysis, Design & Implementation Guide** | [design-doc/01-analysis-design-implementation-guide.md](design-doc/01-analysis-design-implementation-guide.md) | 31KB comprehensive guide: Discord primer, architecture diagrams, pseudocode, rendering details, file output structure, 7-phase implementation plan, API cheat sheet, glossary. |
| **Diary** | [reference/01-diary.md](reference/01-diary.md) | Step-by-step record of ticket creation, document authoring, and reMarkable upload. |

## reMarkable Upload

- **Remote path:** `/ai/2026/04/21/DISCORD-ARCHIVE-BOT`
- **Document:** `DISCORD-ARCHIVE-BOT Guide.pdf`
- **Contents:** Bundled design guide + diary with clickable table of contents (depth 2).

## Status

Current status: **active**

## Topics

- javascript
- backend
- architecture

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
