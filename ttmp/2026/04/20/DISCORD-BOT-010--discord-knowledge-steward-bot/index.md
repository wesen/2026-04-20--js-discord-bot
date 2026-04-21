---
Title: Discord Knowledge Steward Bot
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
RelatedFiles: []
ExternalSources: []
Summary: >
  Ticket workspace for designing a transparent, community-first Discord bot that listens to chat, records candidate knowledge, and routes it through human review before it becomes canonical.
LastUpdated: 2026-04-20T22:40:00-04:00
WhatFor: Track the design and implementation plan for a shared knowledge steward bot.
WhenToUse: Use when planning or implementing a Discord bot that captures, curates, and surfaces group knowledge from chat.
---

# Discord Knowledge Steward Bot

## Overview

This ticket defines a bot that behaves like a modern IRC knowledge helper: it listens to chat, records promising knowledge candidates, and turns them into durable, cited entries that the whole channel can refine over time. The key design goal is to make knowledge capture feel lightweight and conversational while still keeping humans in the loop for verification and cleanup.

The workspace holds the architecture guide, implementation guide, diary, tasks, and changelog for the bot effort. The bot is intentionally scoped as a community tool rather than a private assistant: the primary unit of value is a shared channel or guild, not an individual DM session.

## Key Links

- [Design guide](./design/01-discord-knowledge-steward-bot-architecture-and-implementation-guide.md)
- [Implementation guide](./reference/01-discord-knowledge-steward-bot-implementation-guide-and-api-sketches.md)
- [Diary](./reference/02-diary.md)
- [Tasks](./tasks.md)
- [Changelog](./changelog.md)

## Status

Current status: **active**

## Topics

- backend
- chat
- javascript
- goja

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - API sketches, implementation guides, and diary notes
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
