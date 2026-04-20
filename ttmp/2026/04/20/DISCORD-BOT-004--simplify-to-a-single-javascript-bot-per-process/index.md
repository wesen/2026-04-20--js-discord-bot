---
Title: Simplify to a Single JavaScript Bot Per Process
Ticket: DISCORD-BOT-004
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
    - Path: internal/bot/bot.go
      Note: |-
        Live host currently supports loading multiple scripts and should return to one selected script/runtime
        Live host currently supports multiple scripts and should return to one selected script/runtime
    - Path: internal/botcli/command.go
      Note: |-
        Current named-bot runner CLI that should be simplified back to one selected bot per process
        Current bot runner CLI surface that should simplify back to single-bot run semantics
    - Path: internal/botcli/runtime.go
      Note: |-
        Current bot runner path where parsed-values output and run orchestration live
        Current run path where parsed-values output and future dynamic run-schema parsing would live
    - Path: internal/jsdiscord/host.go
      Note: Live single-bot host and interaction lifecycle logging after the multi-host deletion
ExternalSources: []
Summary: Simplify the Discord bot runtime back to a single JavaScript bot implementation per process and treat in-JS composition as the preferred way to combine behavior.
LastUpdated: 2026-04-20T15:28:00-04:00
WhatFor: Track the rollback from multi-bot runtime composition and define the simpler single-bot runner architecture.
WhenToUse: Use when implementing or reviewing the simplification from multi-bot runtime composition to one selected JS bot per process.
---


# Simplify to a Single JavaScript Bot Per Process

## Overview

This ticket captures a design pivot: stop supporting multiple separately discovered JavaScript bot implementations in one process. Instead, `discord-bot bots run <bot>` should select exactly one bot implementation, start exactly one JS runtime, and let authors compose multiple sub-behaviors inside that bot if they want aggregation.

## Key Links

- `design-doc/01-single-javascript-bot-per-process-architecture-and-implementation-guide.md`
- `analysis/01-ping-bot-search-failure-postmortem.md`
- `reference/01-single-bot-runner-reference-and-migration-notes.md`
- `reference/02-diary.md`
- `tasks.md`
- `changelog.md`
