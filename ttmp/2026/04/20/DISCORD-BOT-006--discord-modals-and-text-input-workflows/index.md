---
Title: Discord Modals and Text Input Workflows
Ticket: DISCORD-BOT-006
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
      Note: |-
        Example bot can demonstrate button -> modal -> submit flow
        Example bot should demonstrate modal workflows
    - Path: internal/jsdiscord/bot.go
      Note: |-
        JS builder/runtime layer needs modal registrations and `showModal` support
        JS runtime contract needs modal registration and showModal support
    - Path: internal/jsdiscord/host.go
      Note: |-
        Modal response emission and modal submit handling should live here
        Modal response emission and modal submit handling belong here
    - Path: internal/jsdiscord/runtime_test.go
      Note: Runtime tests should prove modal payloads and submit dispatch
ExternalSources: []
Summary: Add Discord modal presentation and modal-submit handling to the local JavaScript bot API.
LastUpdated: 2026-04-20T16:15:00-04:00
WhatFor: Track the design and implementation of modal workflows.
WhenToUse: Use when implementing or reviewing modal support.
---


# Discord Modals and Text Input Workflows

## Overview

This ticket adds the modal half of richer Discord interactions. It covers both sides of the flow: opening a modal from a slash command or component interaction, and handling the resulting modal submit event inside JavaScript.
