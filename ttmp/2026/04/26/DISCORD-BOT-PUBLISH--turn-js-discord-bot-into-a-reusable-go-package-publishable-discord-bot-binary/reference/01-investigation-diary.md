---
title: Investigation Diary
description: Chronological investigation diary for DISCORD-BOT-PUBLISH ticket.
doc_type: reference
status: active
topics: [packaging, diary]
ticket: DISCORD-BOT-PUBLISH
---

# Diary

## Goal

Capture the investigation and design work for turning js-discord-bot into a reusable Go package and published binary.

## Step 1: Three-codebase analysis and design guide authoring

### Prompt Context

**User prompt (verbatim):** "Analyze this package and ~/code/wesen/corporate-headquarters/go-template and a finished published tool like ~/code/wesen/corporate-headquarters/pinocchio and create a docmgr ticket on how to turn discord-bot into a reusable package + published discord-bot binary."

**Assistant interpretation:** Analyze three codebases (the prototype, the template skeleton, and a finished published tool), then create a comprehensive docmgr ticket with a design document that serves as an intern-friendly implementation guide.

**Inferred user intent:** Produce a complete, actionable plan that an unfamiliar engineer can follow to transform the prototype into a properly published Go package and binary, understanding all the pieces along the way.

### What I did

- Read and mapped all three codebases: js-discord-bot (prototype), go-template (infrastructure skeleton), pinocchio (finished product).
- Created docmgr ticket DISCORD-BOT-PUBLISH.
- Wrote a multi-section design document broken into separate files for readability:
  - Section 1-2: Executive summary, problem statement, scope
  - Section 3: Current-state architecture with runtime diagrams, directory listings, dependency analysis, gap analysis table
  - Section 4: Target architecture with API surface definitions, module naming, repository layout
  - Section 5: Infrastructure and CI/CD specifications (Makefile, GoReleaser, GitHub Actions, linting)
  - Section 6: Phased implementation plan with exact steps
  - Section 7-9: API reference pseudocode, testing strategy, risks/alternatives/open questions
  - Section 10: File reference index
- Created three helper scripts in scripts/:
  - `01-check-module-importability.sh` — verifies the module can be imported
  - `02-rename-module-path.sh` — mechanical module path rename
  - `03-copy-infrastructure.sh` — copies infrastructure files from go-template

### Why

The design doc needs to be comprehensive enough for an intern who has never seen any of these codebases. Every claim is anchored to concrete files and line counts.

### What worked

- Breaking the document into multiple files avoided write size limits.
- The runtime architecture diagram (ASCII) makes the internal structure clear at a glance.
- The gap analysis table directly shows what needs to happen.

### What didn't work

- Attempting to write the entire document as one edit failed due to tool constraints. Splitting into per-section files worked well.

### What I learned

- The js-discord-bot prototype is architecturally very close to being publishable. The main gaps are infrastructure (Makefile, GoReleaser, CI), module path, and the local go-go-goja replace directive.
- The go-template and pinocchio CI/release pipelines are well-tested and can be adapted almost mechanically.

### What was tricky to build

- The gap analysis required cross-referencing all three codebases simultaneously to identify exactly what differs.

### What warrants a second pair of eyes

- The go-go-goja replace directive status: is the discord registrar support already published in a tagged go-go-goja release?
- The open questions in Section 9.3 need decisions from @manuel.

### What should be done in the future

- Actually execute Phase 1 (module rename) once go-go-goja is published.
- Add Dependabot config (mentioned but not yet created).
- Consider adding a `doc.go` for the root package.

### Code review instructions

- Start with `design-doc/01-reusable-package-and-published-binary-design-guide.md` (executive summary).
- Read Section 3 (current state) to understand the evidence base.
- Jump to Section 6 (implementation phases) for the step-by-step plan.
- Check scripts/ for the helper scripts.

### Technical details

- Total design document size: ~50KB across 6 files.
- Key blocker: `replace github.com/go-go-golems/go-go-goja => /home/manuel/...` in go.mod.
