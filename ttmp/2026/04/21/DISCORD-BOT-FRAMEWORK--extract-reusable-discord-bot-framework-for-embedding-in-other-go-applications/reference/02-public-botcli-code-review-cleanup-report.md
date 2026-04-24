---
Title: Public botcli Code Review Guide and Study Workbook
Ticket: DISCORD-BOT-FRAMEWORK
Status: active
Topics:
    - discord-bot
    - cli
    - code-quality
    - cleanup
    - refactoring
    - architecture
    - jsverbs
    - go
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/doc.go
      Note: Public package overview and runtime-customization decision ladder
    - Path: /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_root.go
      Note: Canonical public constructor and command-registration orchestration
    - Path: /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/bootstrap.go
      Note: Raw-argv bootstrap and repository precedence logic
    - Path: /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/discover.go
      Note: Bot entrypoint discovery, scan policy, and selector resolution
    - Path: /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/options.go
      Note: Runtime customization surface and option semantics
    - Path: /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/command_test.go
      Note: Public regression coverage and downstream-style review entrypoint
    - Path: /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/cmd/discord-bot/root.go
      Note: First-party dogfooding of the public package from the standalone app
Summary: Textbook-style code review guide for the final cleaned `pkg/botcli` design. Teaches not just what to inspect, but why the package is structured this way, what invariants matter, how the runtime/discovery/command layers fit together, and how to review the implementation systematically with concrete files, commands, questions, and exercises.
LastUpdated: 2026-04-23T16:00:00-04:00
---

# Public botcli Code Review Guide and Study Workbook

## How to use this guide

This document is no longer a historical cleanup report. It is a **teaching document for reviewing the final cleaned `pkg/botcli` package**. The goal is not merely to help you decide whether the code is “good” or “bad.” The real goal is to help you understand the architecture so well that a future change request—new discovery rules, new runtime hooks, new help output, new downstream embedding use cases—will feel like an understandable design exercise rather than a pile of unfamiliar files.

That distinction matters. A weak review checklist tells you where to look. A good study guide tells you **why** each place exists, what role it plays in the system, what assumptions it encodes, what could break if it were changed carelessly, and how to tell the difference between deliberate complexity and accidental complexity. This guide is written in that second style.

If you read it straight through, you should come away with three things:

1. a mental model of what `pkg/botcli` is for,
2. a concrete review sequence through the implementation,
3. a vocabulary for judging future changes with confidence.

If you are short on time, read these sections in order:

1. [Chapter 1: What `pkg/botcli` is actually responsible for](#chapter-1-what-pkgbotcli-is-actually-responsible-for)
2. [Chapter 4: The command tree is assembled in two phases](#chapter-4-the-command-tree-is-assembled-in-two-phases)
3. [Chapter 5: Discovery and scanning are intentionally conservative](#chapter-5-discovery-and-scanning-are-intentionally-conservative)
4. [Chapter 7: Runtime customization follows a smallest-hook-first ladder](#chapter-7-runtime-customization-follows-a-smallest-hook-first-ladder)
5. [Chapter 10: The review workbook](#chapter-10-the-review-workbook)

---

## Teaching goal

By the end of this guide, you should be able to answer the following questions without guessing:

- Why does `pkg/botcli` exist separately from `pkg/framework`?
- Why must repository selection happen **before** Cobra parses subcommands?
- Why does discovery use both `InspectScript(...)` and `jsverbs.ScanSources(...)` instead of only one mechanism?
- Why does the package synthesize some `run` commands in Go even when jsverbs metadata is present elsewhere?
- Why is `WithRuntimeModuleRegistrars(...)` the preferred hook most of the time, and when is `WithRuntimeFactory(...)` truly warranted?
- Why are the tests written in a downstream-mounting style rather than only testing internal helpers?

If those questions stop feeling mysterious, the review has done its job.

---

# Chapter 1: What `pkg/botcli` is actually responsible for

The most important thing to understand first is that `pkg/botcli` is **not** the Discord runtime itself. The runtime lives lower in the stack (`internal/jsdiscord`, `internal/bot`). `pkg/botcli` is the optional layer that takes a filesystem-oriented, repository-oriented view of the world and turns it into a public Cobra/Glazed command surface.

That sounds abstract, so let us make it concrete. Suppose a downstream application wants to do two things:

1. discover named bot scripts from one or more directories,
2. mount a `bots` command subtree that lets operators list bots, inspect bots, invoke ordinary jsverbs, and run a chosen bot with host-managed lifecycle.

That is precisely what `pkg/botcli` exists to do. It is the glue that connects:

- repository selection,
- bot discovery,
- jsverbs scanning,
- command registration,
- host-managed `run` behavior,
- and optional runtime customization.

A useful way to think about the package is this:

```text
pkg/framework
  = "I already know which script I want to run."

pkg/botcli
  = "I need the host to discover scripts and expose them as a command tree."
```

That distinction is the foundation of the whole review. If you forget it, the package will seem more complicated than it really is, because you will keep asking why repository/bootstrap/command concerns are not folded into the simple single-bot path. They are not folded in because they are not the same problem.

### The short version

- `pkg/framework` is for an explicit bot.
- `pkg/botcli` is for a **repo-driven command surface**.
- It is allowed to know about Cobra, Glazed, scanning, argv pre-processing, and discovery policy.
- It is not trying to be a generic runtime abstraction layer.

### Code evidence

The package says this directly in its public package documentation:

- `pkg/botcli/doc.go:1-21`

The first-party app and example both dogfood the public package in exactly this role:

- `cmd/discord-bot/root.go:15-75`
- `examples/framework-combined/main.go:25-49`

### Key points to internalize

- `pkg/botcli` is an **integration layer**, not the lowest-level runtime layer.
- Its complexity comes from mapping repository structure to a CLI tree, not from Discord protocol handling.
- You should review it as a package that encodes **policy**: bootstrap policy, discovery policy, command-tree policy, and runtime-extension policy.

---

# Chapter 2: The package is small enough to study, but only if you group it correctly

A common mistake in reviews is to walk files in lexical order and hope understanding emerges from accumulation. That works poorly here. The package is not huge, but it contains several distinct sub-problems, and those sub-problems are easier to understand if you group files by responsibility rather than by filename.

Here is the right grouping.

## 2.1 Package map by responsibility

| Responsibility | Files | Why this grouping matters |
| --- | --- | --- |
| Package contract and high-level guidance | `doc.go`, `model.go` | These tell you what the package is and what its public nouns mean. |
| Repository bootstrap | `bootstrap.go` | This is where raw argv, env vars, defaults, and path normalization live. |
| Discovery and scan policy | `discover.go` | This is where the package decides what counts as a bot entrypoint and what should be scanned. |
| Command tree assembly | `command_root.go`, `command_list.go`, `command_help.go`, `command_run.go` | These form the public CLI surface. |
| Run-schema synthesis and runtime value shaping | `run_description.go`, `runtime_helpers.go` | These bridge metadata and runtime config into a host-managed `run`. |
| Runtime customization hooks | `options.go`, `runtime_factory.go`, `invoker.go` | These are the extension seams for downstream embedders. |
| Review evidence | `command_test.go`, `bootstrap_test.go`, `test_helpers_test.go` | These show how the package is meant to be mounted and validated. |

That grouping already teaches something important: the package is not a random pile of helpers. It is a layered integration package whose main responsibilities are explicit and separable.

## 2.2 Current size hot spots

At the time of this review refresh, the remaining larger files are:

- `pkg/botcli/bootstrap.go` — repository precedence and normalization logic
- `pkg/botcli/discover.go` — discovery, inspection, scanning, selector resolution
- `pkg/botcli/command_test.go` — public regression coverage
- `pkg/botcli/run_description.go` — host-managed run-schema shaping

That is a much healthier shape than the earlier pre-cleanup state, because the package no longer has:

- an oversized `commands_impl.go`,
- a placeholder `command.go`,
- public aliases to internal types,
- or a shadow `internal/botcli` implementation.

In other words: the package is now reviewable on its own terms.

---

# Chapter 3: Bootstrap happens before Cobra, and that is not an accident

The bootstrap path is where many downstream embedders underestimate the problem. They assume repository selection can happen during ordinary command execution, after Cobra has already decided what command the user meant. That assumption is wrong here, and `pkg/botcli` exists partly because it encodes the correct answer.

## 3.1 The fundamental problem

Imagine the user types something like this:

```bash
myapp --bot-repository ./examples/discord-bots bots knowledge-base run --help
```

At the moment Cobra tries to parse `knowledge-base run`, the host must already know:

- which repositories to scan,
- which bots exist,
- which jsverbs exist,
- and therefore which commands should even be registered.

That means repository selection cannot be deferred until “inside” a command. It changes the **shape of the command tree itself**. This is the central bootstrap insight, and everything in `bootstrap.go` exists to preserve that invariant.

## 3.2 How `BuildBootstrap(...)` solves it

Read this function carefully:

- `pkg/botcli/bootstrap.go:71-115`

The order of precedence is the whole point:

1. raw argv CLI flags,
2. environment variable,
3. default repositories.

The function is not trying to be fancy. It is trying to be predictable.

```go
repos, err := repositoriesFromArgs(rawArgs, cfg.flagName, cfg.workingDirectory)
...
repos, err = repositoriesFromEnv(cfg.envVarName, cfg.workingDirectory)
...
repos, err = repositoriesFromDefaults(cfg.defaultRepos, cfg.workingDirectory)
```

This is a strong review point: **bootstrap policy is encoded centrally and deterministically**. The downstream app should not have to re-implement precedence rules, home-directory expansion, duplicate filtering, or default behavior.

## 3.3 What a reviewer should inspect here

### A. Raw argv parsing is intentionally narrow

- `pkg/botcli/bootstrap.go:196-220`

The parser looks only for the configured long flag, stops at `--`, and supports both:

- `--bot-repository path`
- `--bot-repository=path`

That narrowness is good. The function is not trying to become a second CLI framework. It extracts only the information needed to decide what commands must exist before real parsing begins.

### B. Path normalization is part of correctness, not cleanup polish

- `pkg/botcli/bootstrap.go:170-194`

A weaker implementation would pass raw strings downstream and let later stages fail. This implementation instead normalizes:

- whitespace,
- `~` home expansion,
- relative paths against the chosen working directory,
- filesystem existence,
- and directory-ness.

That gives downstream callers a much cleaner contract: once you have a `Bootstrap`, its repositories are already real directories or the call failed.

### C. Duplicate elimination is intentional

- `pkg/botcli/bootstrap.go:145-168`

This is not a micro-optimization. It is a correctness guard. Without deduplication, the same repository path could be injected twice through mixed sources and produce duplicate discovery/scanning work or duplicate bot-name errors that are artifacts of bootstrap, not real repo ambiguity.

## 3.4 Pseudocode for the mental model

```text
BuildBootstrap(rawArgs, opts...):
    load defaults
    apply caller-supplied build options
    resolve CLI repositories from raw argv
    if any CLI repos exist:
        return Bootstrap(CLI repos)
    resolve env repositories
    if any env repos exist:
        return Bootstrap(env repos)
    resolve default repositories
    return Bootstrap(default repos)
```

The important thing is not the algorithm’s cleverness. It is the **stage ordering**.

## 3.5 Review questions

- If a caller changes the repository flag name, does raw-argv scanning still honor it?
- If a default repository is absent on disk, does that fail or quietly skip as designed?
- If both CLI and env specify repositories, does CLI always win?
- If two repository paths normalize to the same absolute path, is duplication removed before discovery?

### Key points to internalize

- Bootstrap is part of command-tree construction, not merely command execution.
- `BuildBootstrap(...)` is valuable because it centralizes precedence and normalization rules.
- A downstream embedder should rarely, if ever, bypass it.

---

# Chapter 4: The command tree is assembled in two phases

The command tree is easier to understand when you stop thinking of it as one big registration function. It is really two different operations:

1. add the **static** commands that always exist,
2. add the **discovered** commands that depend on repository contents.

That split is visible in the cleaned package layout and is worth studying because it reflects the design, not just the file organization.

## 4.1 Start at the public constructor

- `pkg/botcli/command_root.go:23-44`

```go
func NewBotsCommand(bootstrap Bootstrap, opts ...CommandOption) (*cobra.Command, error) {
    cfg, err := applyCommandOptions(opts...)
    ...
    if err := addStaticBotsCommands(...); err != nil { ... }
    if err := addDiscoveredBotsCommands(...); err != nil { ... }
    return root, nil
}
```

The reason this reads well is that the function now shows the package’s high-level architecture directly:

- configuration first,
- parser config next,
- static subtree,
- discovered subtree,
- return the assembled root.

This is precisely the kind of structural clarity you want from a public package.

## 4.2 Static commands are the stable skeleton

- `pkg/botcli/command_root.go:46-79`
- `pkg/botcli/command_list.go:13-39`
- `pkg/botcli/command_help.go:16-73`

The static commands are:

- `bots list`
- `bots help <bot>`

They exist whether any repository currently contains jsverbs verbs or not. That is important because they are **inventory and inspection tools**, not discovered script verbs themselves.

The static commands also teach a design principle: the package makes strong use of Glazed command descriptions and processors even for simple inventory flows. This keeps the CLI surface structurally consistent with the discovered command surface.

## 4.3 Discovered commands are where the package earns its keep

- `pkg/botcli/command_root.go:81-149`

This is the most important review section in the package. It ties together:

- discovery,
- jsverbs registries,
- host-managed `run` behavior,
- synthetic `run` fallback,
- and Glazed/Cobra registration.

The function `discoveredBotsCommands(...)` is doing several conceptually distinct jobs, but they all belong in the same conceptual phase: “turn discovered repositories into command objects.”

### The algorithm in prose

1. Inspect repositories to find discovered bots.
2. Build a map from script path to discovered bot.
3. Scan repositories for explicit jsverbs metadata.
4. For each discovered jsverb:
   - if it is `run`, create a host-managed `botRunCommand` with schema defaults,
   - otherwise, create an ordinary jsverb command with the custom invoker.
5. For each discovered bot that did not expose an explicit `run` verb:
   - synthesize a host-managed `run` command in Go.
6. Register the resulting command list into the Cobra root via Glazed helpers.

That algorithm reveals the package’s central promise: **named bots get a coherent command tree whether or not they authored an explicit jsverbs `run` verb**.

### Why not just invoke the JS `run()` function?

This is one of the most important design decisions in the whole system. A bot’s `run` path is not merely another short-lived verb invocation. It is host lifecycle orchestration:

- decode credentials,
- apply runtime config,
- construct the bot host,
- optionally sync commands,
- open the gateway session,
- block until context cancellation.

That is why `run` is handled as a special case in:

- `pkg/botcli/command_root.go:117-128`
- `pkg/botcli/command_run.go:22-50`

This is good architecture. The JS metadata declares schema and intent; the Go host owns long-lived orchestration.

## 4.4 Why the split into `command_root`, `command_list`, `command_help`, `command_run` matters

Earlier in the project, these ideas were all crowded into one oversized file. That made every review thread harder, because static registration, discovered registration, help/list data emission, and runtime orchestration were all interleaved.

Now the structure communicates intent:

- `command_root.go` — orchestration and registration,
- `command_list.go` — inventory execution,
- `command_help.go` — metadata emission,
- `command_run.go` — long-lived host-managed lifecycle.

This is not merely aesthetic. It reduces review coupling.

### Key points to internalize

- The command tree is assembled in **two phases**, not one undifferentiated pass.
- Static commands describe the stable control surface.
- Discovered commands encode repository-dependent behavior.
- `run` is special because it is host lifecycle orchestration, not ordinary jsverb invocation.

---

# Chapter 5: Discovery and scanning are intentionally conservative

A weaker discovery system would try to scan every `.js` file in sight and hope for the best. This package does something much more disciplined. It discovers candidate bot entrypoints conservatively, inspects those entrypoints as bots, and only then scans those same entrypoints for explicit jsverbs metadata.

That discipline exists because looser discovery caused real problems earlier in the project: helper libraries leaked into the command tree, fake verbs appeared, and the host started mistaking implementation detail for operator-facing API.

## 5.1 The first discovery pass: “Is this a bot entrypoint?”

- `pkg/botcli/discover.go:124-183`

The heuristics are intentionally simple:

- walk the repository,
- skip hidden directories and `node_modules`,
- consider top-level `.js` files and `index.js` entrypoints,
- then require that the file contains both `defineBot` and `require("discord")`.

This is not trying to solve general JavaScript program classification. It is trying to identify probable bot entrypoints with a bias toward **not** over-discovering.

That bias is correct for a command-building package. False negatives are usually easier to diagnose than false positives that surface nonsense commands to operators.

## 5.2 The second pass: inspect as a bot

- `pkg/botcli/discover.go:17-44`

Each candidate entrypoint is inspected through `jsdiscord.InspectScript(...)`, not through jsverbs scanning alone. This matters because bot identity is not only about verbs. The package wants:

- bot name,
- description,
- command metadata,
- event/component/modal descriptors,
- run schema,
- and canonical script path.

That is why `DiscoveredBot` wraps a `jsdiscord.BotDescriptor`, not merely a jsverbs verb list.

## 5.3 The third pass: scan only the entrypoints for explicit verbs

- `pkg/botcli/discover.go:46-96`

This section is easy to skim past, but it encodes a very strong invariant:

```go
registry, err := jsverbs.ScanSources(inputs, jsverbs.ScanOptions{IncludePublicFunctions: false})
```

That option, `IncludePublicFunctions: false`, is a policy statement. It says:

> only explicit verb declarations should become commands.

That is exactly the right rule for this package. It prevents helper functions from being accidentally promoted into operator-facing commands.

## 5.4 Why `DiscoverBots(...)` and `ScanBotRepositories(...)` are separate

At first glance, you might ask: why not merge these? Both walk repositories. Both care about entrypoints. Why not do inspection and jsverbs scanning in one combined operation?

The answer is that they serve different conceptual outputs:

- `DiscoverBots(...)` produces **bot identities and descriptors**,
- `ScanBotRepositories(...)` produces **jsverbs registries**.

The command builder needs both. It uses descriptors to reason about bot identity and fallback `run` behavior, and registries to reason about explicit jsverbs verbs and their schemas. Keeping the functions separate preserves that conceptual clarity.

## 5.5 Selector resolution is part of UX quality

- `pkg/botcli/discover.go:98-122`

`ResolveBot(...)` accepts either the canonical bot name or the source label, and it distinguishes among three states:

- not found,
- exactly one match,
- ambiguous match.

That is good operator-facing behavior. It gives the CLI a human-useful error vocabulary instead of collapsing all selector problems into “not found.”

### Review questions

- Does discovery err on the side of under-discovery rather than helper leakage?
- Does scanning remain explicit-verb-only?
- Are `AbsPath` values reconstructed correctly for host-managed run binding?
- If two bots share a name across repositories, is the failure precise and early?
- Are selector ambiguity errors understandable enough for operators?

### Key points to internalize

- Discovery is intentionally conservative.
- Bot inspection and jsverbs scanning are related but not identical responsibilities.
- The package is opinionated that only explicit verbs should surface publicly.

---

# Chapter 6: Host-managed run synthesis is where CLI metadata meets lifecycle orchestration

The `run` path is the place where this package most clearly stops being a generic command wrapper and becomes a bot-specific orchestration layer.

## 6.1 Two sources of `run`

A bot may get a `run` command in one of two ways:

1. it explicitly declares a `__verb__("run")` shape in JS,
2. or the host synthesizes a `run` description in Go from the bot descriptor.

You can see both paths in:

- explicit run handling: `pkg/botcli/command_root.go:117-128`
- synthetic fallback: `pkg/botcli/command_root.go:139-145`

This is a strong design. It lets authors express runtime schema in JS when they want to, but it does not make host-managed operation disappear for older or simpler bots.

## 6.2 The run-description layer does more than add fields

- `pkg/botcli/run_description.go:14-126`

This file is worth reading line by line because it encodes several subtle behaviors that are easy to miss:

### A. Core host fields are always added

- `bot-token`
- `application-id`
- `guild-id`
- `sync-on-start`

These are host concerns, not per-bot concerns. The host owns them even when JS contributes additional fields.

### B. Bot run-schema fields are merged in

- `pkg/botcli/run_description.go:69-86`

This is how JS-declared run fields become CLI flags on the host-managed path.

### C. Env-backed credential fields are deliberately relaxed

- `pkg/botcli/run_description.go:98-113`

This is a subtle but important design choice. Parser-level required flags are relaxed for `bot-token` and `application-id` so Glazed env middleware can still supply them before final validation. Without this, the parser could reject a command before env loading had a chance to do its job.

A good reviewer should pause here. This is exactly the kind of line that can look odd in isolation and completely reasonable in context.

## 6.3 Runtime config shaping is deliberately mechanical

- `pkg/botcli/runtime_helpers.go:10-71`

The host converts flag names into runtime config keys by a simple kebab-case to snake_case rule and then walks all parsed values into a config map.

This is a good example of a boring mechanism being preferable to cleverness. The code does not try to invent a complicated mapping language. It applies a predictable normalization rule and gets out of the way.

## 6.4 The actual long-lived run path

- `pkg/botcli/command_run.go:22-50`

Review this in order:

1. decode settings from parsed values,
2. validate settings,
3. build runtime config,
4. read `sync-on-start`,
5. construct the bot host with script + config + host options,
6. optionally sync commands,
7. open the session,
8. block until context cancellation.

That sequence is not accidental. It is the real contract of “run a bot” in this host.

### Pseudocode

```text
Run(ctx, parsedValues):
    cfg := decodeDiscordSettings(parsedValues)
    validate cfg
    runtimeConfig := buildRuntimeConfig(parsedValues)
    syncOnStart := boolField(parsedValues, "sync-on-start")
    bot := bot.NewWithScript(cfg, scriptPath, runtimeConfig, hostOpts...)
    if syncOnStart:
        bot.SyncCommands()
    bot.Open()
    wait until ctx.Done()
```

The important thing to notice is that there is no branch where JS “takes over” the long-lived run process. The host always owns that lifecycle.

### Key points to internalize

- `run` is host-owned lifecycle orchestration.
- JS metadata can shape the schema, but not replace the host’s responsibilities.
- The env-backed relaxation in `run_description.go` is a deliberate parsing/middleware compatibility move.

---

# Chapter 7: Runtime customization follows a smallest-hook-first ladder

This is the chapter to sit with if you want to understand the public extensibility story correctly.

A common failure mode in API reviews is to see the most powerful hook, decide that it is “the real one,” and mentally demote the smaller hooks to trivialities. That is backwards here. The package explicitly wants callers to choose the **smallest hook that solves the real problem**.

## 7.1 Read the package doc first

- `pkg/botcli/doc.go:1-21`

This package comment is short, but it is doing real architectural work. It teaches the decision ladder directly:

- `WithAppName(...)` when env-prefix behavior changes,
- `WithRuntimeModuleRegistrars(...)` when scripts just need extra native modules,
- `WithRuntimeFactory(...)` only when runtime creation itself must change,
- `HostOptionsProvider` when the same choice should also affect discovery and host-managed runs.

That ordering is not arbitrary. It expresses the design philosophy of the package.

## 7.2 Why `WithRuntimeModuleRegistrars(...)` is the preferred hook most of the time

- `pkg/botcli/options.go:75-91`

Many customization requests are actually very modest. A downstream app often just wants a bot script or jsverb to say:

```js
const app = require("app")
```

and get some Go-native helper module. In that case, the default runtime creation is still correct. The only missing piece is an extra registrar.

That is why `WithRuntimeModuleRegistrars(...)` exists and why the docs now explicitly say to prefer it over `WithRuntimeFactory(...)` whenever possible.

This is the healthy API design here: common extension cases should not force callers into the most powerful abstraction.

## 7.3 What `RuntimeFactory` is really for

- `pkg/botcli/options.go:13-20`
- `pkg/botcli/options.go:93-110`
- `pkg/botcli/runtime_factory.go:16-60`
- `pkg/botcli/invoker.go:11-29`

A `RuntimeFactory` is not “the module registrar, but more complicated.” It is the point where the caller can replace the entire ordinary-jsverb runtime construction process.

That matters when you need to change things like:

- module root selection,
- require loader behavior,
- builder configuration,
- runtime lifecycle behavior,
- or other runtime-creation mechanics that registrars alone do not touch.

The default runtime factory shows exactly what would otherwise happen:

- choose module roots from script path when available,
- use registry require loader,
- add default registry modules,
- add the Discord registrar,
- add custom runtime registrars,
- build factory,
- create runtime.

A reviewer should compare a proposed custom `RuntimeFactory` against that baseline and ask: **what truly needs to differ?**

## 7.4 Why `HostOptionsProvider` exists separately

- `pkg/botcli/options.go:29-36`
- `pkg/botcli/options.go:112-121`

This is one of the most subtle but well-justified design choices in the package.

Ordinary jsverb invocation uses a runtime created by the `RuntimeFactory`. But discovery and host-managed bot runs do not go through that same ordinary-jsverb invocation path. They go through `jsdiscord` host options.

That means there are really two related but distinct extension surfaces:

1. ordinary jsverb runtime creation,
2. jsdiscord host configuration for discovery and host-managed runs.

`HostOptionsProvider` is the bridge that lets one caller-defined concept influence both when needed.

The package deliberately does **not** force every runtime factory to understand host options. That is good design. It keeps the more powerful parity behavior opt-in.

## 7.5 The right review questions for customization hooks

### Good questions

- Could this use case be solved with `WithRuntimeModuleRegistrars(...)` alone?
- Does the proposed custom runtime factory really need to change runtime creation itself?
- If the same customization must affect discovery and host-managed runs, has the factory also implemented `HostOptionsProvider`?
- Does the proposed design preserve the package’s “smallest hook first” philosophy?

### Bad questions

- “Why not just use `RuntimeFactory` for everything?”
- “Why not collapse all customization into one super-interface immediately?”

Those questions ignore the design goal of keeping the common path simple.

### Key points to internalize

- The package wants callers to choose the smallest sufficient hook.
- Module registrars are the common extension seam.
- `RuntimeFactory` is the advanced seam for changing runtime construction itself.
- `HostOptionsProvider` exists to keep advanced customization consistent across all runtime touchpoints.

---

# Chapter 8: The tests are part of the architecture argument

A code review that stops at the implementation files misses half the picture. In this package, the public tests are not merely regression protection. They are evidence for the intended public usage model.

## 8.1 Why the tests mount into a downstream root

- `pkg/botcli/command_test.go:14-22`

The very first test does not call hidden helpers. It mounts `NewBotsCommand(...)` into an arbitrary Cobra root and executes it as a downstream app would.

That is an architectural statement:

> this package is supposed to be embedded.

This style of test is much more meaningful than a purely internal helper test would be, because it proves the public contract actually works when consumed through Cobra.

## 8.2 What the regression set is trying to prove

Read the current tests as a narrative, not a list:

- `pkg/botcli/command_test.go:14-22` — downstream embedding works
- `pkg/botcli/command_test.go:24-35` — canonical run help exposes bot schema
- `pkg/botcli/command_test.go:37-47` — app-name override changes env prefix behavior
- `pkg/botcli/command_test.go:49-70` — module registrars are enough for custom native modules
- `pkg/botcli/command_test.go:72-94` — runtime factory supports deeper customization
- `pkg/botcli/command_test.go:96-110` — only canonical `bots <bot> run` remains
- `pkg/botcli/command_test.go:112-120` — helper functions do not leak into command help

This is a very nice public regression story because it tracks the package’s main promises directly.

## 8.3 The first-party app and combined example are living review artifacts

Do not ignore:

- `cmd/discord-bot/root.go:15-75`
- `examples/framework-combined/main.go:25-49`

The standalone app proves that the project itself now dogfoods the public path. The combined example proves that a downstream application can mount the package while also using `pkg/framework` for a separate explicit built-in bot.

That combination is not incidental. It is one of the strongest pieces of evidence that the public split is real rather than aspirational.

### Key points to internalize

- Public tests here are part of the design argument.
- A good review of `pkg/botcli` should always include the tests and first-party embeddings.
- The package is healthiest when its public path is continuously exercised by the app itself.

---

# Chapter 9: What already got cleaned up, and why that matters to your review stance

A good reviewer should not review the package as though it were still in its transitional extraction state. That would be reviewing the ghosts of old problems instead of the current code.

The following issues have already been resolved and should now be treated as **historical context**, not active defects:

- `internal/botcli` duplication is gone.
- Public aliases to internal botcli types are gone.
- The panic-based `NewCommand(...)` wrapper is gone.
- The legacy `bots run <bot>` compatibility path is gone.
- The old oversized command implementation has been split into focused files.
- The `WithRuntimeFactory(...)` guidance gap has been reduced substantially through package docs and option comments.

Why does this matter? Because a reviewer’s posture should change after cleanup. Earlier, the right stance was “find the duplication and remove it.” Now the better stance is:

- verify the cleaned architecture stays coherent,
- check whether new complexity is deliberate and justified,
- and guard against regression back into wrapper-driven or shadow-package design.

This is a different phase of review. It is less about large-scale extraction and more about **keeping the public package honest**.

---

# Chapter 10: The review workbook

This chapter turns the earlier explanation into an actual review procedure.

## 10.1 Recommended reading order

Read the files in this order:

1. `pkg/botcli/doc.go`
2. `pkg/botcli/model.go`
3. `pkg/botcli/bootstrap.go`
4. `pkg/botcli/discover.go`
5. `pkg/botcli/command_root.go`
6. `pkg/botcli/command_run.go`
7. `pkg/botcli/run_description.go`
8. `pkg/botcli/options.go`
9. `pkg/botcli/runtime_factory.go`
10. `pkg/botcli/invoker.go`
11. `pkg/botcli/command_test.go`
12. `cmd/discord-bot/root.go`
13. `examples/framework-combined/main.go`

That order mirrors the package’s actual conceptual flow:

- what the package is,
- what nouns it exposes,
- how repositories are chosen,
- how bots are discovered,
- how commands are assembled,
- how run is orchestrated,
- how runtime customization works,
- how the public contract is validated.

## 10.2 Review checklist by concern

### Concern A: Public API clarity

Questions:

- Are the public nouns (`Bootstrap`, `Repository`, `DiscoveredBot`) sufficient and unsurprising?
- Does `NewBotsCommand(...)` feel like the canonical constructor for mounting the subtree?
- Are option names understandable without code archaeology?
- Does the package documentation make the intended usage model visible quickly?

Evidence:

- `pkg/botcli/doc.go`
- `pkg/botcli/model.go`
- `pkg/botcli/options.go`

### Concern B: Bootstrap correctness

Questions:

- Is the raw-argv pre-scan still minimal and correct?
- Are CLI/env/default precedence rules explicit and deterministic?
- Are path normalization and duplicate elimination done before discovery?

Evidence:

- `pkg/botcli/bootstrap.go`
- `pkg/botcli/bootstrap_test.go`

### Concern C: Discovery conservatism

Questions:

- Could helper libraries leak into command discovery again?
- Is the “entrypoint only + explicit verbs only” policy still intact?
- Do duplicate bot-name failures happen early and clearly?

Evidence:

- `pkg/botcli/discover.go`
- `pkg/botcli/command_test.go:112-120`

### Concern D: Command-tree coherence

Questions:

- Are static commands and discovered commands kept conceptually separate?
- Is `run` still handled as host lifecycle orchestration rather than ordinary jsverb invocation?
- Does synthetic `run` fallback still exist for bots without explicit run metadata?

Evidence:

- `pkg/botcli/command_root.go`
- `pkg/botcli/command_run.go`
- `pkg/botcli/run_description.go`

### Concern E: Runtime customization discipline

Questions:

- Does the package still encourage the smallest sufficient hook?
- Are callers likely to misuse `WithRuntimeFactory(...)` when registrars would do?
- Is parity across discovery, run, and ordinary jsverbs maintained when needed?

Evidence:

- `pkg/botcli/doc.go`
- `pkg/botcli/options.go`
- `pkg/botcli/runtime_factory.go`
- `pkg/botcli/invoker.go`

### Concern F: Public dogfooding

Questions:

- Does the first-party app still use the public package directly?
- Does the combined example still prove the intended split between `pkg/framework` and `pkg/botcli`?

Evidence:

- `cmd/discord-bot/root.go`
- `examples/framework-combined/main.go`

## 10.3 Hands-on review commands

Run these in order:

```bash
cd /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot

go test ./...

go doc ./pkg/botcli | head -n 40

go run ./cmd/discord-bot bots list --output json

go run ./examples/framework-combined bots list --output json

go run ./examples/framework-combined bots ui-showcase run --help

go run ./examples/framework-combined bots run ui-showcase --help
```

What each command teaches:

| Command | What it proves |
| --- | --- |
| `go test ./...` | The public path is the real path and the repository still validates. |
| `go doc ./pkg/botcli` | The package contract is readable to embedders. |
| `discord-bot bots list` | The standalone app really dogfoods the public package. |
| `framework-combined bots list` | The downstream embedding story is real. |
| `ui-showcase run --help` | Canonical run help still exposes the host-managed run surface. |
| `bots run ui-showcase --help` | The legacy compatibility path is no longer supported as run help. |

## 10.4 Exercises for deeper understanding

These are not busywork. They are the kinds of small thought experiments that reveal whether you really understand the design.

### Exercise 1: Explain the need for raw-argv pre-scan

Without looking at the code, explain why repository selection must happen before Cobra subcommand parsing. Then confirm your explanation against `bootstrap.go` and `cmd/discord-bot/root.go`.

### Exercise 2: Justify the split between bot inspection and jsverbs scanning

Why can the package not simply scan for jsverbs and call that “discovery”? Write the answer in terms of bot identity, descriptors, and host-managed `run` behavior.

### Exercise 3: Decide between hooks

For each scenario, decide which hook is correct and why:

1. “I need `require("app")` to exist.”
2. “I need to change module roots and require behavior.”
3. “I need discovery and host-managed runs to see the same module extension as ordinary jsverbs.”
4. “I only want a different env prefix.”

A correct answer should land on:

1. `WithRuntimeModuleRegistrars(...)`
2. `WithRuntimeFactory(...)`
3. `WithRuntimeFactory(...)` + `HostOptionsProvider`
4. `WithAppName(...)`

### Exercise 4: Review a hypothetical regression

Suppose a future patch changes discovery to scan all public functions again. What tests would you expect to fail, and why would that regression matter to operators?

### Exercise 5: Review a hypothetical convenience API proposal

Suppose someone proposes reintroducing a panic-based `NewCommand(...)` convenience wrapper. What would you ask before accepting it? Frame your answer in terms of public API honesty, failure handling, and whether the wrapper truly improves the package.

---

# Chapter 11: What good future changes should look like

This chapter is about review taste. Once a package reaches a cleaned, coherent state, future changes should be judged not only by whether they work but by whether they preserve the package’s design discipline.

## 11.1 Signs of a good change

A strong future patch to `pkg/botcli` will usually have these qualities:

- It preserves the smallest-hook-first philosophy for customization.
- It keeps discovery conservative rather than broad and magical.
- It uses the public package path in tests and examples rather than inventing parallel private helpers.
- It makes command-tree behavior clearer, not more surprising.
- It keeps repository/bootstrap rules centralized instead of scattering them into apps.

## 11.2 Warning signs

A weaker patch will often smell like one of these:

- reintroducing wrapper APIs that hide real errors,
- adding a new “temporary” internal mirror of public behavior,
- making discovery broader in ways that invite helper leakage,
- bypassing `BuildBootstrap(...)` from first-party callers,
- or forcing common extension cases into the most powerful runtime hook.

These warnings are not theoretical. They correspond closely to the historical cleanup work the project just finished.

---

# Closing: what this guide wants you to notice

The deepest lesson of `pkg/botcli` is not about Discord bots specifically. It is about how a public integration package can stay healthy.

A healthy public integration package does four things at once:

1. it hides awkward but necessary orchestration in one well-defined place,
2. it exposes a public contract that feels smaller than the internal complexity it manages,
3. it teaches callers to use the smallest sufficient hook,
4. and it continuously proves its own design through first-party dogfooding and downstream-style tests.

`pkg/botcli` is in that much healthier state now. That does not mean it will never need more changes. It means the package has crossed the more important threshold: it is now understandable enough to review as a coherent system rather than as an extraction-in-progress.

When you review it, review it in that spirit.
