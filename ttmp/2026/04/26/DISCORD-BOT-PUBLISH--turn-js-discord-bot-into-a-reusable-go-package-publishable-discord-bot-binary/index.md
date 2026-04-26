---
Title: Turn js-discord-bot into a reusable Go package + publishable discord-bot binary
Ticket: DISCORD-BOT-PUBLISH
Status: active
Topics:
    - packaging
    - go
    - reusable-package
    - publishing
    - architecture
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go.mod
      Note: Module definition
    - Path: internal/bot/bot.go
      Note: Discord session wrapper
    - Path: internal/jsdiscord/host.go
      Note: JS runtime host lifecycle
    - Path: pkg/botcli/options.go
      Note: RuntimeFactory
    - Path: pkg/framework/framework.go
      Note: Public single-bot embedding API
ExternalSources: []
Summary: ""
LastUpdated: 2026-04-26T09:01:15.266401924-04:00
WhatFor: ""
WhenToUse: ""
---



# Turn js-discord-bot into a reusable Go package + publishable discord-bot binary

## Overview

This ticket tracks the work to transform the js-discord-bot prototype into:

1. **A reusable Go package** importable as `github.com/go-go-golems/discord-bot/pkg/{framework,botcli}`.
2. **A standalone published binary** (`discord-bot`) released via GoReleaser to Homebrew, deb, rpm, and GitHub Releases.

The design guide is in `design-doc/` (6 files covering current-state analysis, target architecture, infrastructure, implementation phases, API reference, testing, and risks).

Key blocker: the local `go-go-goja` replace directive in go.mod must be removed (requires go-go-goja to be published first).

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- packaging
- go
- reusable-package
- publishing
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
