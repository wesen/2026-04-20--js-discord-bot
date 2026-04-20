---
Title: Move Sandbox Here and Implement Real JavaScript Discord Bot API
Ticket: DISCORD-BOT-002
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
        Live Discord host now loads optional JavaScript bot scripts from this repository
        Optional JavaScript bot loading integrated into the live host
    - Path: internal/jsdiscord/bot.go
      Note: |-
        Local bot-definition API, dispatch flow, and async settlement for JS-authored Discord bots
        Local bot-definition API
    - Path: internal/jsdiscord/host.go
      Note: |-
        Host-side bridge from Discord gateway events into the JS bot runtime
        Live Discord gateway bridge for JavaScript bot commands and events
    - Path: internal/jsdiscord/runtime.go
      Note: |-
        Runtime-scoped local Discord JS module and registrar now live in this repository
        Runtime-scoped local Discord JS module registrar and loader
ExternalSources: []
Summary: Move the sandbox-style JS host layer into js-discord-bot and evolve it into a real Discord-specific JavaScript bot API.
LastUpdated: 2026-04-20T14:13:00-04:00
WhatFor: Track the move of sandbox functionality into this repo and the first live Discord-integrated JS bot implementation.
WhenToUse: Use when implementing or reviewing the JavaScript-hosted Discord bot runtime in this repository.
---


# Move Sandbox Here and Implement Real JavaScript Discord Bot API

## Overview

This ticket moves the sandbox-style JS host layer out of `go-go-goja` and into `js-discord-bot`, where the real application-specific Discord API belongs. The goal is to let `go-go-goja` remain the reusable engine/jsverbs dependency while this repository owns the scriptable Discord host API, example scripts, and runtime integration.

## Key Links

- `design-doc/01-sandbox-move-and-discord-javascript-api-architecture-guide.md`
- `reference/01-discord-javascript-bot-api-reference-and-example-script.md`
- `reference/02-diary.md`
- `tasks.md`
- `changelog.md`

## Status

Current status: **active**

## Tasks

See [tasks.md](./tasks.md).

## Changelog

See [changelog.md](./changelog.md).
