---
Title: Public botcli Code Review and Cleanup Report
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
    - Path: /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/commands_impl.go
      Note: Main public command builder and current public ownership hot spot
    - Path: /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/pkg/botcli/discover.go
      Note: Public scan policy and discovery ownership
    - Path: /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/command.go
      Note: Legacy internal command builder still mirrors the public path closely
    - Path: /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/internal/botcli/bootstrap.go
      Note: Legacy internal discovery implementation retained after public extraction
Summary: Evidence-based code review of the framework extraction work, focused on deprecation candidates, unclear code, bloated files/packages, backwards-compatibility burden, and legacy wrappers that should be clean-cut rather than kept indefinitely.
LastUpdated: 2026-04-23T12:30:00-04:00
---

# Public botcli Code Review and Cleanup Report

## Executive summary

The framework extraction has reached an important threshold: `pkg/botcli` now owns the major public behaviors that downstream embedders care about — repository bootstrap, scan policy, command registration, host-managed run semantics, and runtime customization. That is the good news.

The code quality problem is not that the design failed. The problem is that the repo now carries **two overlapping implementations** of the same botcli subsystem: one in `pkg/botcli` and one in `internal/botcli`. The public package is no longer a thin façade, but the internal package still contains a large amount of the old machinery. That duplication is now the dominant maintenance risk.

My strongest recommendation is a **clean cut** rather than incremental indefinite coexistence:

1. treat `pkg/botcli` as the canonical command/discovery implementation,
2. shrink `internal/botcli` down to the truly app-private pieces or remove it,
3. avoid adding more backwards-compatibility wrappers unless they protect an externally consumed API.

## Scope reviewed

Reviewed surfaces:

- `pkg/botcli/*`
- `internal/botcli/*`
- `cmd/discord-bot/root.go`
- `examples/framework-combined/*`

Inspection commands used:

```bash
find pkg/botcli internal/botcli examples/framework-combined -maxdepth 3 -type f | sort
wc -l pkg/botcli/*.go internal/botcli/*.go | sort -nr | head -n 40
rg -n "deprecated|legacy|compat|compatibility|wrapper|TODO|panic\(|type \w+ = internal|internalbotcli|WithRuntimeFactory|WithRuntimeModuleRegistrars|WithAppName|NewCommand\(|NewBotsCommand\(|BuildBootstrap\(" pkg/botcli internal/botcli cmd/discord-bot -g '*.go'
```

File size hot spots:

- `pkg/botcli/commands_impl.go` — 276 lines
- `internal/botcli/command_test.go` — 235 lines
- `pkg/botcli/bootstrap.go` — 230 lines
- `internal/botcli/bootstrap.go` — 201 lines
- `pkg/botcli/discover.go` — 183 lines
- `internal/botcli/command.go` — 173 lines
- `pkg/botcli/command_test.go` — 175 lines
- `pkg/botcli/run_description.go` — 138 lines
- `internal/botcli/run_description.go` — 138 lines

Those counts matter because the bloat is not just in one file — it is duplicated across the public/internal split.

---

## 1. Largest cleanup opportunity: duplicated command/discovery ownership

### 1.1 Public and internal command builders now overlap too much

**Problem:** The public package owns real behavior now, but the internal package still carries a parallel command builder and discovery path. This is no longer “public wrappers over internals”; it is two command systems with overlapping responsibilities.

**Where to look:**
- `pkg/botcli/commands_impl.go:157-276`
- `internal/botcli/command.go:21-173`
- `pkg/botcli/discover.go:17-183`
- `internal/botcli/bootstrap.go:113-201`

**Example:**

`pkg/botcli/commands_impl.go:157-166`
```go
func NewBotsCommand(bootstrap Bootstrap, opts ...CommandOption) (*cobra.Command, error) {
    cfg, err := applyCommandOptions(opts...)
    if err != nil {
        return nil, err
    }
    parserConfig := botCLIParserConfig(cfg.appName)
    hostOpts := cfg.hostOptions()
```

`internal/botcli/command.go:24-31`
```go
func NewBotsCommand(bootstrap Bootstrap, opts ...CommandOption) (*cobra.Command, error) {
    cfg, err := applyCommandOptions(opts...)
    if err != nil {
        return nil, err
    }
    parserConfig := botCLIParserConfig(cfg.appName)
```

`pkg/botcli/discover.go:17-27`
```go
func DiscoverBots(ctx context.Context, bootstrap Bootstrap, hostOpts ...jsdiscord.HostOption) ([]DiscoveredBot, error) {
    ret := []DiscoveredBot{}
    seen := map[string]DiscoveredBot{}
    for _, repo := range bootstrap.Repositories {
        scripts, err := discoverScriptCandidates(repo.RootDir)
```

`internal/botcli/bootstrap.go:113-123`
```go
func DiscoverBots(ctx context.Context, bootstrap Bootstrap, hostOpts ...jsdiscord.HostOption) ([]DiscoveredBot, error) {
    ret := []DiscoveredBot{}
    seen := map[string]DiscoveredBot{}
    for _, repo := range bootstrap.Repositories {
        scripts, err := discoverScriptCandidates(repo.RootDir)
```

**Why it matters:**
- bug fixes must be mirrored in two places,
- tests split attention between the public and legacy path,
- reviewers cannot immediately tell which implementation is canonical,
- future refactors will drag both packages unless there is a deliberate cut-over.

**Cleanup sketch:**

```text
pkg/botcli/
  bootstrap.go        # keep
  discover.go         # keep
  commands_impl.go    # keep
  options.go          # keep
  runtime_factory.go  # keep

internal/botcli/
  model.go            # delete or alias to pkg/botcli types temporarily
  command.go          # delete after callers move
  bootstrap.go        # delete after callers move
  run_description.go  # delete after callers move
  ...tests trimmed...
```

If full deletion is too abrupt, make `internal/botcli` a **single delegating shim** that calls `pkg/botcli`, not a second implementation.

---

### 1.2 Run description logic is duplicated almost verbatim

**Problem:** `pkg/botcli/run_description.go` and `internal/botcli/run_description.go` are effectively duplicates.

**Where to look:**
- `pkg/botcli/run_description.go:14-138`
- `internal/botcli/run_description.go:14-138`

**Example:**

Both files contain the same `buildSyntheticBotRunDescription`, `buildCompatibilityRunAliasDescription`, `ensureRunCommandDefaults`, `addCoreRunFields`, `relaxEnvBackedRequiredFlags`, and `glazedFieldType` functions.

**Why it matters:**
- this is high-confidence duplicated maintenance cost,
- any semantic change to run flags or required/default behavior must be applied twice,
- it obscures whether host-managed run semantics are “public-owned” or “still internal”.

**Cleanup sketch:**

```go
// internal/botcli/run_description.go
package botcli

import publicbotcli ".../pkg/botcli"

var (
    buildSyntheticBotRunDescription = publicbotcli.BuildSyntheticBotRunDescriptionForInternalUse
    // or better: stop referencing internal path from app entirely
)
```

Better still: remove internal references to these helpers and let the app use `pkg/botcli` directly (which it already does at the root level).

---

## 2. Backwards compatibility burden: keep only what protects real users

### 2.1 Both run command shapes are valuable, but they should not justify broad duplication

**Problem:** The compatibility requirement (`bots <bot> run` and `bots run <bot>`) is real, but the codebase risks using “backwards compatibility” to justify too many duplicate paths.

**Where to look:**
- `pkg/botcli/commands_impl.go:244-259`
- `pkg/botcli/run_description.go:28-38`

**Example:**
```go
aliasDesc := buildCompatibilityRunAliasDescription(baseDesc, discoveredBot.Name())
discoveredCommands = append(discoveredCommands, &botRunCommand{
    CommandDescription: aliasDesc,
    scriptPath:         discoveredBot.ScriptPath(),
    hostOpts:           hostOpts,
})
```

**Why it matters:**
- the run-shape compatibility itself is okay,
- but additional wrapper packages or duplicated implementations are not justified merely because these aliases exist,
- compatibility should live in one canonical place and nowhere else.

**Cleanup sketch:**
- Keep the alias behavior in `pkg/botcli` only.
- Remove any duplicated internal implementation of the same compatibility logic.
- Document the two supported shapes explicitly and stop treating this as an excuse to preserve parallel machinery.

---

### 2.2 Public aliases to internal types are now a design smell

**Problem:** `pkg/botcli/bootstrap.go` exports aliases to internal types instead of owning its own public structs.

**Where to look:**
- `pkg/botcli/bootstrap.go:12-22`

**Example:**
```go
const (
    BotRepositoryFlag = internalbotcli.BotRepositoryFlag
)

type (
    Bootstrap     = internalbotcli.Bootstrap
    Repository    = internalbotcli.Repository
    DiscoveredBot = internalbotcli.DiscoveredBot
)
```

**Why it matters:**
- this reverses the desired layering: the public package still structurally depends on internal package definitions,
- it makes deprecation and cleanup of `internal/botcli` harder,
- it prevents `pkg/botcli` from being a truly independent public contract.

**Cleanup sketch:**

```go
type Bootstrap struct {
    Repositories []Repository
}

type Repository struct {
    Name      string
    Source    string
    SourceRef string
    RootDir   string
}

type DiscoveredBot struct {
    Repository Repository
    Descriptor *jsdiscord.BotDescriptor
}
```

Then add explicit conversion at the few boundaries still talking to internal code during transition.

This is the clean-cut de-legacy step I would prioritize early.

---

## 3. Unclear / misleading code surfaces

### 3.1 `pkg/botcli/command.go` is a placeholder file with no real API content

**Problem:** `pkg/botcli/command.go` now contains only a comment saying the real implementation lives elsewhere.

**Where to look:**
- `pkg/botcli/command.go:1-3`

**Example:**
```go
package botcli

// Public command construction lives in commands_impl.go.
```

**Why it matters:**
- this is surprising navigation for readers,
- it suggests unfinished refactoring rather than intentional structure,
- the filename no longer matches the implementation location.

**Cleanup sketch:**
- Either move `commands_impl.go` back to `command.go`, or
- rename files consistently around real conceptual boundaries, e.g.:

```text
pkg/botcli/
  cobra_commands.go
  discovery.go
  bootstrap.go
  runtime_factory.go
  options.go
```

The current placeholder is not harmful at runtime, but it is documentation debt.

---

### 3.2 `WithRuntimeFactory` + `HostOptionsProvider` is powerful but conceptually split

**Problem:** The public runtime customization story spans two separate concepts:
- runtime creation for ordinary jsverbs,
- host-option contribution for discovery and host-managed runs.

That may be correct, but it is not especially obvious.

**Where to look:**
- `pkg/botcli/options.go:13-25`
- `pkg/botcli/options.go:77-99`
- `pkg/botcli/invoker.go:11-28`

**Example:**
```go
type RuntimeFactory interface {
    NewRuntimeForVerb(ctx context.Context, registry *jsverbs.Registry, verb *jsverbs.VerbSpec) (*engine.Runtime, error)
}

type HostOptionsProvider interface {
    HostOptions() []jsdiscord.HostOption
}
```

**Why it matters:**
- downstream implementers have to discover by reading code that implementing both interfaces is the “full parity” path,
- the API can look incomplete if you only notice `RuntimeFactory`,
- design review becomes harder because the mental model is split across two extension hooks.

**Cleanup sketch:**
Either:

1. keep the split, but document it explicitly in a public API guide, or
2. wrap both responsibilities in a single richer factory/options interface.

For example:

```go
type RuntimeIntegration interface {
    NewRuntimeForVerb(ctx context.Context, registry *jsverbs.Registry, verb *jsverbs.VerbSpec) (*engine.Runtime, error)
    HostOptions() []jsdiscord.HostOption
}
```

I would **not** rush that refactor unless the docs remain confusing after the stabilization pass.

---

## 4. Bloated files and package shape

### 4.1 `pkg/botcli/commands_impl.go` is becoming a “god file”

**Problem:** `pkg/botcli/commands_impl.go` now owns command builder logic, command structs, help/list execution, run execution, public factory selection, and command registration orchestration.

**Where to look:**
- `pkg/botcli/commands_impl.go:1-276`

**Why it matters:**
- cross-cutting edits will continue to accumulate here,
- code review becomes expensive because behavior, command wiring, and execution live together,
- further growth will make it harder to identify stable seams.

**Cleanup sketch:**

```text
pkg/botcli/
  command_root.go       # NewBotsCommand, NewCommand
  command_list.go       # listBotsCommand
  command_help.go       # helpBotsCommand
  command_run.go        # botRunCommand
  command_register.go   # discovered verb registration orchestration
```

This is a low-risk structural split once the public behavior stabilizes.

---

### 4.2 `internal/botcli` package is now bloated relative to its residual value

**Problem:** `internal/botcli` still has many files even though the public package now owns the public behavior story.

**Where to look:**
- `internal/botcli/command.go` — 173 lines
- `internal/botcli/bootstrap.go` — 201 lines
- `internal/botcli/run_description.go` — 138 lines
- `internal/botcli/command_test.go` — 235 lines

**Why it matters:**
- package size implies ongoing responsibility,
- future contributors may keep adding features there by inertia,
- the extraction never really finishes if the old center of gravity remains in `internal/botcli`.

**Cleanup sketch:**
Mark `internal/botcli` as transition-only in a package comment or collapse it aggressively.

A good test for whether code belongs there:
- if `cmd/discord-bot/root.go` no longer needs it,
- and downstream examples no longer need it,
- then the package should not keep a full independent implementation.

---

## 5. Legacy wrappers / panic-oriented APIs that should be cut cleanly

### 5.1 `NewCommand(...)` panics in both public and internal command layers

**Problem:** Convenience constructors panic on configuration errors.

**Where to look:**
- `pkg/botcli/commands_impl.go:270-275`
- `internal/botcli/command.go:167-173`

**Example:**
```go
func NewCommand(bootstrap Bootstrap, opts ...CommandOption) *cobra.Command {
    cmd, err := NewBotsCommand(bootstrap, opts...)
    if err != nil {
        panic(err)
    }
    return cmd
}
```

**Why it matters:**
- acceptable for tiny internal helpers,
- less acceptable for public package surfaces,
- pushes failure handling into process crashes in downstream apps.

**Cleanup sketch:**
- keep `NewBotsCommand(...) (*cobra.Command, error)` as the canonical API,
- consider deprecating public `NewCommand(...)` or making its panic behavior very explicit in docs,
- if kept, treat it as a strict convenience wrapper only.

This is a strong candidate for deprecation rather than indefinite preservation.

---

### 5.2 Explicit `WithAppName("discord")` in app roots is now boilerplate noise

**Problem:** `cmd/discord-bot/root.go` and examples pass the default app name explicitly.

**Where to look:**
- `cmd/discord-bot/root.go:65-70`
- `examples/framework-combined/main.go:36-43`

**Example:**
```go
botsCmd := publicbotcli.NewCommand(bootstrap, publicbotcli.WithAppName("discord"))
```

**Why it matters:**
- not wrong, but noisy,
- makes the default path look more configurable than it needs to be,
- contributes to “legacy wrapper” feel where option plumbing is shown even when unnecessary.

**Cleanup sketch:**
- remove explicit `WithAppName("discord")` from first-party call sites,
- reserve the option for actual custom prefixes.

This is a clean cut with minimal risk.

---

## 6. Public documentation gap after the runtime-factory addition

### 6.1 `WithRuntimeFactory` exists, but docs/examples lag the code

**Problem:** the public runtime-factory hook has arrived faster than its documentation.

**Where to look:**
- `pkg/botcli/options.go:77-87`
- examples/README surfaces currently emphasize `WithAppName` and runtime module registrars more than the factory path.

**Why it matters:**
- new public API without docs is indistinguishable from experimental internals,
- reviewers cannot easily judge whether it is intended for regular use or last-resort escape hatches.

**Cleanup sketch:**
- add one focused public API section documenting:
  - when `WithRuntimeModuleRegistrars(...)` is enough,
  - when `WithRuntimeFactory(...)` is necessary,
  - why `HostOptionsProvider` exists.

This is the stabilization/docs pass I would do before more extraction work.

---

## 7. Highest-leverage cleanup plan

### Low-risk / do next
1. **Remove explicit default `WithAppName("discord")` calls** from first-party examples/app root.
2. **Document `WithRuntimeFactory(...)`** and the difference from module registrars.
3. **Rename/split `pkg/botcli/commands_impl.go`** into a few clearer files.
4. **Delete placeholder `pkg/botcli/command.go`** or rename files so it is no longer just a comment shell.

### Medium-risk / high value
5. **Replace public aliases in `pkg/botcli/bootstrap.go` with owned public structs.**
6. **Collapse duplicated internal helpers** (`run_description`, discovery/command logic) behind a single canonical implementation.

### Strategic / clean cut
7. **Deprecate or reduce `internal/botcli` to a transitional shim** rather than a parallel implementation.
8. **Reassess `NewCommand(...)` panic behavior** and potentially make `NewBotsCommand(...)` the canonical public constructor.

---

## Bottom line judgment

The extraction is now **design-valid and functionally strong**, but **cleanup debt has shifted from architecture risk to duplication risk**.

The code is no longer blocked on major conceptual gaps. The pressing issue is to avoid entering a long phase where:
- `pkg/botcli` is the public truth,
- `internal/botcli` remains a shadow implementation,
- and future maintenance must update both.

If I were prioritizing strictly for code quality, I would do the next work in this order:

1. docs/stabilization for the public surface,
2. clean-cut reduction of `internal/botcli`,
3. then only additional features if still necessary.
