---
Title: Framework Extraction Design Review and Decision Guide
Ticket: DISCORD-BOT-FRAMEWORK
Status: active
Topics:
    - discord-bot
    - architecture
    - api-design
    - framework
    - cli
    - jsverbs
    - goja
    - glazed
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/framework/framework.go
      Note: Public single-bot surface
    - Path: /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/commands_impl.go
      Note: Public repo-driven command builder and host-managed run semantics
    - Path: /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/options.go
      Note: Public botcli configuration surface
    - Path: /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-combined/main.go
      Note: Combined downstream app showing both extracted layers together
Summary: Design review for the current framework extraction state, focused on helping a reviewer judge whether the public split is coherent, what remains unfinished, what is stable enough to treat as the intended design, and where cleanup or reduction should happen next.
LastUpdated: 2026-04-23T12:35:00-04:00
---

# Framework Extraction Design Review and Decision Guide

## Executive summary

The extraction has crossed the line from “promising refactor” to “usable public design.” The repo now presents two distinct public layers that match the intended architecture:

- `pkg/framework` — the simple explicit single-bot path
- `pkg/botcli` — the optional repo-driven multi-bot/jsverbs command layer

This split is no longer just theoretical. It is exercised by:

- public package tests,
- the standalone app (`cmd/discord-bot`),
- dedicated embedding examples, and
- a combined downstream example app that uses both layers in one process.

If the question is **“is the design direction sound enough to keep investing in?”**, my answer is **yes**.

If the question is **“is the codebase ready to freeze as-is?”**, my answer is **not yet**. The main remaining issue is no longer architectural uncertainty — it is cleanup and reduction of duplicate/internal legacy structures.

## What is now clearly true

### 1. The public split is coherent

Evidence:
- `pkg/framework/framework.go` provides the explicit single-bot path.
- `pkg/botcli/commands_impl.go` provides the repo-driven command tree.
- `examples/framework-combined/main.go` uses both in one app.

This means the architecture now has a clear story for both kinds of embedders:

#### Simple embedder
A downstream app that wants exactly one bot script can use:
- `framework.New(...)`
- explicit `WithScript(...)`
- env-backed or explicit credentials
- runtime config and optional sync-on-start

#### Advanced / operator-heavy embedder
A downstream app that wants repository-driven discovery and jsverbs can use:
- `botcli.BuildBootstrap(...)`
- `botcli.NewCommand(...)`
- dynamic commands registered before Cobra parses subcommands
- repo-driven discovery and host-managed run semantics

That division is clean and understandable.

---

### 2. The public botcli package is no longer just a façade

This is important. Earlier in the extraction, `pkg/botcli` was mostly wrappers around `internal/botcli`. That is no longer the case.

Evidence:
- `pkg/botcli/discover.go` owns scan policy and discovery.
- `pkg/botcli/run_description.go` owns run synthesis helpers.
- `pkg/botcli/commands_impl.go` owns the public command builder and host-managed run behavior.
- `pkg/botcli/runtime_factory.go` and `pkg/botcli/invoker.go` own public runtime customization for ordinary jsverbs.

This means the public package now contains meaningful behavior, not just naming indirection.

---

### 3. The extraction now supports real customization, not only the happy path

The public surface now supports:

#### On `pkg/framework`
- `WithRuntimeModuleRegistrars(...)`

#### On `pkg/botcli`
- `WithAppName(...)`
- `WithRuntimeModuleRegistrars(...)`
- `WithRuntimeFactory(...)`

The important design implication is that the public API is not trapped in the default runtime anymore. That was one of the biggest original architectural concerns.

---

## What a reviewer should judge positively

### A. The architecture boundary is finally visible in code

The current repo has a much better conceptual map than it did at the start of the ticket:

```text
pkg/framework
  = explicit one-bot embedding path

pkg/botcli
  = optional repo-driven command/discovery layer

internal/jsdiscord
  = runtime/host core

internal/bot
  = Discord session host wiring

cmd/discord-bot
  = standalone app using the public layers
```

That is a good architecture. It is legible and teachable.

### B. The combined downstream example is the strongest design proof so far

`examples/framework-combined/main.go` is especially valuable because it proves:
- the public split is not mutually exclusive,
- one process can host both a built-in bot and repo-driven bots,
- the public API is expressive enough for a downstream root command.

That is exactly the kind of design evidence a reviewer should trust more than purely abstract docs.

### C. The manual validations reduce design risk materially

The recent manual validations matter because they were not just happy-path tests:
- env-backed credentials worked through both the standalone and downstream botcli paths,
- custom runtime module registrars worked in repo-driven paths,
- custom runtime factory behavior worked in a downstream temporary app,
- both run command shapes worked through the combined example.

This is strong evidence that the new public layers are behaviorally real.

---

## What a reviewer should still be cautious about

### 1. The codebase still carries too much duplication

This is the biggest caution flag.

The design can be “right” while the implementation still carries too much transitional duplication. That is where the repo is now.

Most important examples:
- `pkg/botcli/run_description.go` and `internal/botcli/run_description.go`
- `pkg/botcli/discover.go` and `internal/botcli/bootstrap.go` discovery sections
- `pkg/botcli/commands_impl.go` and `internal/botcli/command.go`

This does **not** invalidate the design. It just means the next major value is cleanup rather than more feature surface.

### 2. Public aliases to internal types are a weak point

`pkg/botcli/bootstrap.go` still exports aliases to internal types:
- `Bootstrap = internalbotcli.Bootstrap`
- `Repository = internalbotcli.Repository`
- `DiscoveredBot = internalbotcli.DiscoveredBot`

That is a transitional move, not a fully mature public boundary.

As long as that remains, the public package still structurally depends on internal type ownership.

### 3. The runtime customization story is powerful but slightly conceptually split

`WithRuntimeFactory(...)` plus optional `HostOptionsProvider` is useful, but it is not instantly obvious to a first-time reader.

A reviewer should ask:
- is this the stable intended public API,
- or a transition step that should later collapse into a single richer runtime integration contract?

My current judgment: it is acceptable as-is, but should be documented more explicitly before being treated as final.

---

## Current design status by area

## 1. Single-bot public framework path

### Status
**Good and mostly stable.**

### Why
- clear constructor surface,
- strong examples,
- custom module support exists,
- no repository scanning required,
- matches the design goal well.

### Remaining uncertainty
- whether a lower-level public `NewHost(...)` is still needed,
- whether `framework` should grow additional lower-level escape hatches or stay intentionally opinionated.

### Judgment
This part of the design is strong enough that I would treat it as the stable “easy path.”

---

## 2. Public botcli bootstrap/configuration path

### Status
**Good, with some cleanup needed.**

### Why
- `BuildBootstrap(...)` exists,
- CLI/env/default precedence is public,
- custom flag/env names are possible,
- `WithAppName(...)`, registrars, and runtime factory exist.

### Remaining uncertainty
- public aliases to internal types,
- whether the public configuration surface should stay option-based or later become a config struct for more advanced consumers.

### Judgment
Architecturally sound. Needs cleanup more than redesign.

---

## 3. Public botcli command/discovery/run behavior

### Status
**Good direction, but not yet “finished.”**

### Why
- command tree behavior is public,
- scan policy is public,
- host-managed run semantics are public,
- helper leakage and run-shape regressions are covered.

### Remaining uncertainty
- how aggressively to retire the remaining `internal/botcli` implementation,
- whether the public package should be split into smaller files for clarity,
- whether `NewCommand(...)` should remain panic-oriented.

### Judgment
Good enough to use; not yet clean enough to call complete.

---

## 4. Internal/runtime layers beneath the public API

### Status
**Still transitional.**

### Why
The public architecture is coherent, but the implementation still depends on internal layers that have not been fully reduced or repackaged.

That is normal for this stage, but it means the extraction is not “done done.”

### Judgment
No urgent redesign required. Focus on reduction and cleanup instead of new architectural invention.

---

## Decision guide: how to judge where we are

If you want to decide what to do next, I think the right questions are:

### Question 1: Is the public split credible?
**Answer: yes.**
The examples and validations show that it is.

### Question 2: Should we keep investing in this extraction direction?
**Answer: yes.**
The design is much stronger now than when the ticket began.

### Question 3: Do we need more features before cleanup?
**Answer: probably not.**
The bigger need now is reduction of duplication and documentation/stabilization.

### Question 4: Is this ready to be treated as stable public API?
**Answer: partially.**
The shape is credible; the cleanup story is not finished.

---

## Recommended next steps

## Option A — stabilization mode (my recommendation)

Focus next on:
1. public API documentation pass (especially `WithRuntimeFactory(...)`),
2. code cleanup / de-duplication,
3. reduction of `internal/botcli`.

Why this is best:
- architecture risk is now lower than maintenance risk,
- cleanup will make future judgments easier,
- fewer moving parts makes later public API decisions safer.

## Option B — more extraction cleanup immediately

Focus next on:
1. replacing public aliases in `pkg/botcli/bootstrap.go` with owned public structs,
2. making `internal/botcli` a delegating shim or removing big chunks of it.

Why this is attractive:
- biggest reduction in technical debt,
- clearest signal that the extraction is truly complete.

## Option C — keep adding advanced capabilities

Possible, but I would not recommend this yet.

Why not:
- the public design is now broad enough,
- each new feature increases the cost of later cleanup,
- the risk is drifting from “finish the extraction” into “keep expanding the transitional state.”

---

## My current judgment

If I had to summarize the state in one sentence:

> The framework extraction is now design-sound and publicly usable, but it has entered the phase where cleanup and stabilization are more valuable than additional feature growth.

So if the question is **“have we achieved enough design clarity to judge this direction?”** — yes.

If the question is **“should we now optimize for reduction and clarity instead of more surface area?”** — also yes.

---

## Reviewer checklist

A reviewer trying to make a judgment should inspect these in order:

1. `pkg/framework/framework.go`
2. `pkg/botcli/commands_impl.go`
3. `pkg/botcli/options.go`
4. `examples/framework-combined/main.go`
5. `pkg/botcli/command_test.go`
6. the code review report in:
   - `reference/02-public-botcli-code-review-cleanup-report.md`

And validate with:

```bash
go test ./...
go run ./examples/framework-combined bots list --output json
go run ./examples/framework-combined bots ui-showcase run --help
go run ./examples/framework-combined bots run ui-showcase --help
```

That is enough to understand both the design and the current maintenance tradeoffs.
