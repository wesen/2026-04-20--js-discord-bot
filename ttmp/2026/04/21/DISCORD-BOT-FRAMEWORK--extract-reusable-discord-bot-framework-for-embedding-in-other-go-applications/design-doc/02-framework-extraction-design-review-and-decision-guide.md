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
    - Path: /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_root.go
      Note: Public repo-driven command builder and registration orchestration
    - Path: /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/options.go
      Note: Public botcli configuration and runtime-customization surface
    - Path: /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/examples/framework-combined/main.go
      Note: Combined downstream app showing both extracted layers together
    - Path: /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/ttmp/2026/04/21/DISCORD-BOT-FRAMEWORK--extract-reusable-discord-bot-framework-for-embedding-in-other-go-applications/reference/02-public-botcli-code-review-cleanup-report.md
      Note: Textbook-style code review guide and study workbook for the final botcli shape
Summary: Updated design review for the cleaned final extraction state. Helps a reviewer judge the architecture after the public botcli clean cut, not during the earlier transitional phase.
LastUpdated: 2026-04-23T16:10:00-04:00
---

# Framework Extraction Design Review and Decision Guide

## Executive summary

The framework extraction has now reached the state the earlier review work was arguing toward.

The public split is real, coherent, and substantially cleaned up:

- `pkg/framework` is the clear single-bot embedding path.
- `pkg/botcli` is the clear optional repo-driven discovery/command path.
- `internal/botcli` is gone.
- the public botcli package owns its own public types,
- exposes one canonical constructor (`NewBotsCommand(...)`),
- keeps only the canonical named-bot run shape (`bots <bot> run`),
- and now has clearer package structure and clearer runtime-customization guidance.

If the question is **“is the design direction sound?”**, the answer is **yes**.

If the question is **“is the extraction still primarily a cleanup project?”**, the answer is now **much less so than before**. The large clean-cut items have been completed. What remains is ordinary design stewardship: protecting clarity, validating future changes, and resisting regressions back into wrapper-heavy or duplicate implementations.

---

## What is now clearly true

## 1. The public split is the intended architecture, not a draft

Evidence:
- `pkg/framework/framework.go`
- `pkg/botcli/command_root.go`
- `cmd/discord-bot/root.go`
- `examples/framework-combined/main.go`

The current codebase now expresses the architecture directly:

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
  = standalone app that dogfoods the public layers
```

That is a stable and teachable architecture.

## 2. `pkg/botcli` is now the canonical repo-driven layer

This matters because the earlier review period spent a lot of time asking whether the public package was still mostly a façade. That question is now answered.

Evidence:
- `pkg/botcli/bootstrap.go` owns public bootstrap behavior.
- `pkg/botcli/discover.go` owns public discovery and scan policy.
- `pkg/botcli/command_root.go` plus the focused command files own public command construction.
- `pkg/botcli/options.go`, `runtime_factory.go`, and `invoker.go` own the public runtime-extension story.
- `internal/botcli` no longer exists as a shadow implementation.

This is the strongest possible sign that the extraction is no longer halfway public and halfway private.

## 3. The public API now communicates its intended usage more clearly

The cleaned package now teaches its intended usage in three reinforcing ways:

- through its package doc (`pkg/botcli/doc.go`),
- through the option/interface comments in `pkg/botcli/options.go`,
- and through first-party docs/examples (`README.md`, `examples/framework-combined/README.md`).

That is especially important for the runtime-extension story, where the package now makes the “smallest hook first” rule explicit:

- `WithAppName(...)`
- `WithRuntimeModuleRegistrars(...)`
- `WithRuntimeFactory(...)`
- optional `HostOptionsProvider`

---

## What a reviewer should judge positively now

### A. The awkward integration rules are centralized instead of scattered

Repository precedence, raw-argv pre-scan, discovery heuristics, explicit-verb-only scanning, host-managed run synthesis, and runtime extension hooks all live in one public package.

That is good architecture because it gives embedders one place to adopt the repo-driven behavior rather than forcing each downstream app to rediscover the same awkward rules.

### B. The first-party app is still the best dogfooding proof

`cmd/discord-bot/root.go` now mounts `pkg/botcli.NewBotsCommand(...)` directly.

This is one of the strongest health signals in the repository. It means the public API is not only exported; it is trusted enough that the project itself depends on it for its real command tree.

### C. The combined example clarifies the division of labor

`examples/framework-combined/main.go` remains the clearest proof of the intended split:

- `run-builtin` proves the explicit single-bot path,
- `bots ...` proves the optional repo-driven path,
- and using both in the same process proves they are complementary rather than competing abstractions.

### D. The cleanup cuts improved not just code structure, but reviewability

Several earlier review concerns were directly addressed:

- no public aliases to internal botcli types,
- no panic wrapper constructor,
- no legacy `bots run <bot>` compatibility path,
- no duplicated internal package,
- no oversized command “god file,”
- and better guidance around advanced runtime hooks.

That means the public package is now much easier to review as a coherent system.

---

## What a reviewer should still pay attention to

The fact that the major cleanup is complete does **not** mean the package should now be treated as untouchable. It means the review posture changes.

The right questions are no longer mainly about extraction debt. They are now about **preserving design quality**.

### 1. Does future work preserve the smallest-hook-first philosophy?

The package has a good extensibility ladder now. A reviewer should be cautious about future changes that push common use cases toward `WithRuntimeFactory(...)` unnecessarily.

### 2. Does discovery remain conservative?

The package is healthiest when it is biased toward avoiding helper leakage and fake commands. Future broadening of scan policy should be reviewed carefully.

### 3. Does the public command tree stay legible?

The command split is clearer now, but future changes could still re-accumulate orchestration, execution, and registration logic into one place. A reviewer should guard against that drift.

### 4. Does the project continue to dogfood the public path?

The standalone app and combined example should keep exercising the public package. If first-party code starts bypassing the public layer again, that is an architectural smell.

---

## Current design status by area

## 1. Single-bot public framework path

### Status
**Strong and stable.**

### Why
- clear constructor/options surface,
- strong examples,
- custom runtime-module support,
- intentionally free of repository-scanning concerns.

### Judgment
Treat this as the stable simple path.

---

## 2. Public botcli bootstrap/configuration path

### Status
**Strong and much cleaner than before.**

### Why
- public bootstrap owns its own types,
- precedence rules are centralized,
- custom flag/env/default behavior is supported,
- advanced runtime customization is present but now better explained.

### Judgment
This is now a credible public configuration surface rather than a transitional wrapper layer.

---

## 3. Public botcli command/discovery/run behavior

### Status
**Strong and now structurally cleaner.**

### Why
- command registration is public and focused,
- discovery is public and conservative,
- host-managed run semantics are public,
- tests cover downstream mounting, canonical run shape, helper-leak prevention, and runtime customization.

### Judgment
This is now clean enough to be treated as the intended design, not merely an extraction milestone.

---

## 4. Remaining internal/runtime layers beneath the public API

### Status
**Appropriately internal.**

### Why
The layers below the public API now look like proper implementation layers rather than unfinished public candidates.

### Judgment
No major architecture inversion is still pending here.

---

## Decision guide: how to judge where we are now

### Question 1: Is the public split credible?
**Answer: yes.**
It is exercised by the standalone app, examples, tests, and public docs.

### Question 2: Has the large cleanup debt been paid down enough to trust the package shape?
**Answer: yes, substantially.**
The biggest earlier blockers were real, and they have now been removed.

### Question 3: Should the next work focus on more extraction, or on normal design stewardship?
**Answer: normal design stewardship.**
The package is now in the phase where good review discipline matters more than large-scale structural surgery.

### Question 4: Is this ready to be treated as a stable intended public API?
**Answer: yes, with ordinary healthy caution.**
That is a much stronger answer than the earlier “partially” state.

---

## Recommended next steps

### Option A — maintain and validate (recommended)

Focus on:
1. keeping the public package dogfooded by the app and examples,
2. reviewing future discovery/runtime-extension changes against the current design principles,
3. continuing to improve docs/examples only when they sharpen understanding rather than increase surface area unnecessarily.

### Option B — add one more advanced educational example

This would only be worthwhile if we want to teach `RuntimeFactoryFunc` in a more hands-on way. It is not currently required for the architecture to feel complete.

### Option C — continue adding new public botcli capabilities

Possible, but should be justified narrowly. The package no longer needs breadth for its own sake.

---

## My current judgment

If I had to summarize the state in one sentence now, it would be this:

> The framework extraction is now no longer interesting mainly as an extraction; it is interesting as a cleaned, coherent public design that should be preserved and taught well.

That is a good place to be.

---

## Reviewer checklist

If you want the shortest high-value review path, inspect these in order:

1. `pkg/framework/framework.go`
2. `pkg/botcli/doc.go`
3. `pkg/botcli/command_root.go`
4. `pkg/botcli/discover.go`
5. `pkg/botcli/options.go`
6. `pkg/botcli/command_test.go`
7. `cmd/discord-bot/root.go`
8. `examples/framework-combined/main.go`
9. the detailed textbook-style review guide in:
   - `reference/02-public-botcli-code-review-cleanup-report.md`

And validate with:

```bash
go test ./...
go doc ./pkg/botcli | head -n 40
go run ./examples/framework-combined bots list --output json
go run ./examples/framework-combined bots ui-showcase run --help
```

That is now enough to understand both the architecture and the current design quality.
