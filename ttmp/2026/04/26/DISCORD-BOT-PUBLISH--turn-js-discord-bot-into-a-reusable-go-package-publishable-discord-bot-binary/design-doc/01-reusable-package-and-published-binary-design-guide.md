---
title: Reusable Package and Published Binary Design Guide
description: Complete analysis, design, and implementation guide for turning the js-discord-bot prototype into a reusable Go package and publishable discord-bot binary under go-go-golems.
doc_type: design-doc
status: active
intent: long-term
topics:
  - packaging
  - go
  - reusable-package
  - publishing
  - architecture
ticket: DISCORD-BOT-PUBLISH
related_files: []
---

# Reusable Package and Published Binary Design Guide

## Table of Contents

1. Executive Summary
2. Problem Statement and Scope
3. Current-State Architecture (Evidence-Based)
   - 3.1 What js-discord-bot Is Today
   - 3.2 What go-template Provides
   - 3.3 What pinocchio Looks Like as a Finished Published Tool
   - 3.4 Gap Analysis: Current vs Target
4. Target Architecture
   - 4.1 Reusable Go Package (`pkg/`)
   - 4.2 Standalone Published Binary (`cmd/discord-bot/`)
   - 4.3 JavaScript Bot Authoring API
   - 4.4 Repository Layout and Module Naming
5. Infrastructure and CI/CD
   - 5.1 Makefile
   - 5.2 GoReleaser
   - 5.3 GitHub Actions Workflows
   - 5.4 Linting and Hooks
6. Detailed Implementation Phases
   - Phase 1: Rename and Reparent
   - Phase 2: Extract Public API Surface
   - Phase 3: Infrastructure from go-template
   - Phase 4: CI and Publishing
   - Phase 5: Polish and Documentation
7. API Reference and Pseudocode
8. Testing and Validation Strategy
9. Risks, Alternatives, and Open Questions
10. References

---

## 1. Executive Summary

This document is a complete, intern-friendly guide to transforming the `js-discord-bot` prototype (currently living at `~/code/wesen/2026-04-20--js-discord-bot`) into two deliverables:

1. **A reusable Go package** that downstream Go applications can import to embed a Discord bot runtime with a JavaScript authoring API. Think of it as "the `pinocchio` of Discord bots" — a library other tools compose into their own binaries.

2. **A standalone published binary** called `discord-bot` (or similar) that provides the CLI experience operators already use (`discord-bot bots list`, `discord-bot bots <name> run`, etc.), built and released via GoReleaser to GitHub, Homebrew, and apt/rpm repositories.

The guide is organized as a full orientation (what each piece does, how it connects), followed by a phased implementation plan with file-level instructions, pseudocode, and validation steps.

**Key design decisions** (explained in detail later):

- The repo moves under `github.com/go-go-golems/discord-bot` (or similar).
- The public Go API lives under `pkg/` (already partially extracted: `pkg/framework/` and `pkg/botcli/`).
- The `internal/jsdiscord/` runtime engine stays internal but becomes the backbone that `pkg/` wraps.
- Infrastructure (Makefile, GoReleaser, CI workflows, lint config) is adapted from the `go-template` skeleton.
- The publishing pipeline mirrors what `pinocchio` already does.

---

## 2. Problem Statement and Scope

### What we have today

The `js-discord-bot` repository is a working prototype. It has:

- A Go binary (`cmd/discord-bot/`) that hosts a Discord gateway session.
- An embedded JavaScript runtime (goja) that runs bot scripts written in JS.
- A `require("discord")` module that exposes the Discord bot authoring API to JS scripts.
- ~14,000 lines of Go code across `internal/` and `pkg/`.
- ~20 tickets of development history in `ttmp/`.
- Example bots that double as executable documentation.
- A partially-extracted public API (`pkg/framework/` and `pkg/botcli/`).

### What is missing

The project is not yet a proper Go library or a published binary. Specifically:

1. **No canonical module path.** The `go.mod` says `github.com/manuel/wesen/2026-04-20--js-discord-bot`, which is a local development path, not an importable Go module.

2. **No release infrastructure.** There is no Makefile, no GoReleaser config, no CI workflows, no linting hooks. Every other published tool in the go-go-golems ecosystem has these.

3. **No versioned releases.** There are no git tags, no GitHub releases, no Homebrew formula, no deb/rpm packages.

4. **Local dependency hack.** The `go.mod` contains a `replace` directive pointing `go-go-goja` to a local path:

   ```
   replace github.com/go-go-golems/go-go-goja => /home/manuel/code/wesen/corporate-headquarters/go-go-goja
   ```

   This works on one machine but blocks anyone else from building the project.

5. **Public API surface is young.** While `pkg/framework/` and `pkg/botcli/` exist, they have not been reviewed for stability, naming consistency, or completeness as a library interface.

### Scope

This guide covers:

- Reparenting the repository under `github.com/go-go-golems/discord-bot`.
- Adapting the `go-template` infrastructure (Makefile, GoReleaser, CI, lint).
- Validating that the public `pkg/` API is importable by a downstream application.
- Setting up the full release pipeline so that `discord-bot` ships as a binary.

This guide does **not** cover:

- Adding new Discord API features (those are tracked in separate tickets).
- Rewriting the JavaScript authoring API.
- Changing the runtime model (one bot per process).

### Success criteria

We are done when all of the following are true:

1. `go get github.com/go-go-golems/discord-bot` works from any machine.
2. A downstream Go app can import `pkg/framework/` and run a Discord bot in < 20 lines of code.
3. `goreleaser release` produces cross-platform binaries, a Homebrew formula, and deb/rpm packages.
4. CI runs `make lint && make test` on every push.
5. Tagging a version with `svu patch && git push origin --tags` triggers a full release.
