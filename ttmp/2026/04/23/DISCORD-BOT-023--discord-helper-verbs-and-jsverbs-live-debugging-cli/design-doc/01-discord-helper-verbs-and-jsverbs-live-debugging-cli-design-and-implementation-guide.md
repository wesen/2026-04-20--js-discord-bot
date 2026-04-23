---
Title: Discord helper verbs and jsverbs live-debugging CLI design and implementation guide
Ticket: DISCORD-BOT-023
Status: active
Topics:
    - discord
    - jsverbs
    - cli
    - tooling
    - diagnostics
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../corporate-headquarters/go-go-goja/pkg/jsverbs/command.go
      Note: Upstream command and schema generation pipeline for annotated JavaScript verbs
    - Path: ../../../../../../../corporate-headquarters/go-go-goja/pkg/jsverbs/model.go
      Note: Upstream jsverbs metadata model used in the proposed helper-verb architecture
    - Path: ../../../../../../../corporate-headquarters/loupedeck/cmd/loupedeck/cmds/verbs/command.go
      Note: Reference implementation for lazy verbs command registration and custom invoker wiring
    - Path: cmd/discord-bot/root.go
      Note: Current root command wiring; shows where a future top-level verbs subtree would fit
    - Path: internal/botcli/bootstrap.go
      Note: Current defineBot-oriented discovery path; evidence that helper verbs need separate repository scanning
    - Path: internal/botcli/command.go
      Note: Current named-bot Cobra subtree; contrast point for a future jsverbs-based helper CLI
    - Path: internal/jsdiscord/host.go
      Note: Current Goja runtime host construction pattern that a helper-verb runtime can mirror
ExternalSources: []
Summary: Design for a new CLI-exposed jsverbs subsystem that supports live Discord helper verbs, bot simulation verbs, and reusable diagnostic tooling without overloading the interactive defineBot runtime.
LastUpdated: 2026-04-23T10:15:00-04:00
WhatFor: Explain how to add a jsverbs-driven helper CLI for Discord live debugging, payload inspection, and bot simulation in a way that fits the current discord-bot architecture.
WhenToUse: Read this when planning or implementing CLI helper verbs, jsverbs discovery, live Discord probes, or bot-simulation tooling in this repository.
---


# Discord helper verbs and jsverbs live-debugging CLI design and implementation guide

## Executive summary

This ticket proposes a new CLI subsystem for **Discord helper verbs**: standalone JavaScript scripts discovered through `jsverbs`, exposed as CLI commands, and executed in a purpose-built runtime that can perform live Discord inspection work and controlled bot simulation work. The core idea is to **keep rich Discord bots in the existing `defineBot(...)` model** while introducing a **second, tooling-oriented execution model** for diagnostics, live probes, and developer workflows.

The current repository already has the two ingredients needed for this design. First, the main `discord-bot` executable already has a mixed CLI: three top-level host commands are built through Glazed, while the named-bot workflow lives under a separate `bots` Cobra subtree (`cmd/discord-bot/root.go:15-64`). Second, prior design work and upstream reference code already demonstrate a stable `jsverbs` pattern in `go-go-goja` and loupedeck, where annotated JavaScript functions are scanned statically, turned into Glazed command descriptions, and mounted lazily into a Cobra tree (`go-go-goja/pkg/jsverbs/command.go:41-99`, `loupedeck/cmd/loupedeck/cmds/verbs/command.go:57-118`).

The recommended outcome is a new top-level command group, tentatively `discord-bot verbs`, that scans the same shared repositories used by the bot subsystem, builds one CLI command per discovered verb, and runs those verbs inside a runtime that exposes a carefully chosen set of modules: a separate live Discord inspection module, optional UI builders, one bot-simulation module, and structured output helpers. This gives developers a way to investigate live guild state, inspect members and roles, preview UI payloads, and simulate bot commands from the CLI without mutating the interactive Discord bot surface every time a new debugging question appears.

## Decision summary

The user resolved the main architecture choices for this ticket as follows:

1. Keep the live Discord inspection API in a **separate module** rather than overloading `require("discord")`.
2. Keep bot simulation helpers in **one module** rather than splitting them across multiple helper modules.
3. Use **one shared repository concept** for both named bots and helper verbs; embedded repositories should always be loaded.
4. Do **not** add a built-in JSON output flag; verbs that want JSON-like output can return serialized JSON as a single row field or text payload.
5. Allow mutating helper verbs, but keep them in a **separate writable directory**, such as `verbs-rw/`.

## Problem statement and scope

### The problem

The repository currently has a strong runtime model for interactive Discord bots, but it does not yet have an equally strong model for **standalone JavaScript-powered developer tooling**. This gap shows up whenever developers want to ask questions such as:

1. What roles does this member actually have right now?
2. What payload would this UI builder return before I post it to Discord?
3. What would this bot command return if I invoked it as a specific user in a specific guild?
4. How can I run reusable live probes from the CLI instead of adding temporary slash commands?

Today, answering those questions tends to push work into one of three awkward places:

- temporary bot commands inside `defineBot(...)` scripts,
- Go unit tests that simulate dispatch but are slow to write for exploratory work,
- ad hoc shell scripts or one-off Go snippets with poor reuse.

### Why the current bot model is not enough

The current `bots` flow is built around discovering JavaScript files that look like named Discord bot implementations. Discovery only accepts scripts that contain both `defineBot` and `require("discord")` (`internal/botcli/bootstrap.go:184-200`). That is exactly the right filter for interactive bot scripts, but it is the wrong filter for standalone helper tools. A helper verb should be able to exist as an ordinary JavaScript file without pretending to be a bot.

### Scope of this ticket

This design ticket covers:

1. CLI discovery and command registration for helper verbs.
2. Runtime architecture for executing helper verbs.
3. Proposed module surface for live Discord inspection and bot simulation.
4. Safety model for read-only vs mutating operations.
5. Phased implementation plan and validation strategy.

This ticket does **not** implement the subsystem itself. It is a design and onboarding deliverable.

## Current-state architecture (evidence-based)

### 1. The main CLI is already split between Glazed host commands and a Cobra bot subtree

The current root command wires three host-level commands (`run`, `validate-config`, `sync-commands`) and then separately mounts the named-bot command tree from `internal/botcli` (`cmd/discord-bot/root.go:36-64`). In practice, that means the executable already tolerates multiple command systems:

```text
discord-bot
├── run              # Glazed-backed host command
├── validate-config  # Glazed-backed host command
├── sync-commands    # Glazed-backed host command
└── bots             # Cobra subtree for named JS bot workflows
```

This matters because adding a top-level `verbs` subtree is not architecturally disruptive. The command tree is already mixed.

### 2. The existing `bots` subtree is oriented around `defineBot(...)` scripts, not arbitrary JS tools

`internal/botcli/command.go` shows that the current `bots` command only supports three workflows: `list`, `help`, and `run` (`internal/botcli/command.go:12-21`). The underlying discovery path walks repositories, looks for candidate `.js` files, and then filters them through `looksLikeBotScript(...)`, which requires both `defineBot` and `require("discord")` (`internal/botcli/bootstrap.go:142-200`).

This is a strong signal that the current `bots` implementation should remain focused on interactive bot scripts. Helper verbs should not be squeezed into that discovery path.

### 3. The current JS runtime host is already capable of owning a runtime per script

`internal/jsdiscord/host.go` constructs a fresh Goja runtime from a specific script path, registers runtime modules, requires the script, and compiles it into a bot handle (`internal/jsdiscord/host.go:21-52`). The important architectural point is that the repository already uses a **runtime-factory pattern** in practice, even if it is not named that way yet.

This makes a helper-verb runtime feasible. The new subsystem does not need to invent runtime ownership from scratch; it needs to create a runtime variant that is optimized for helper verbs rather than `defineBot(...)` dispatch.

### 4. The upstream `jsverbs` package already solves command discovery and schema generation

The upstream `go-go-goja/pkg/jsverbs` package exposes three particularly relevant layers:

1. a static registry of discovered verbs (`Registry`, `VerbSpec`, `FieldSpec`) (`go-go-goja/pkg/jsverbs/model.go:74-157`),
2. a command-building path from verbs to Glazed command descriptions (`go-go-goja/pkg/jsverbs/command.go:41-99`),
3. a schema builder that turns field metadata and shared sections into Glazed sections (`go-go-goja/pkg/jsverbs/command.go:102-210`).

This is the technical foundation we should reuse rather than rebuilding a bespoke JS-metadata scanner in this repo.

### 5. Loupedeck demonstrates the exact lazy-Cobra + jsverbs integration pattern we need

The strongest reference implementation is loupedeck’s `verbs` command. It does all of the following:

1. creates a lazy `verbs` root command that defers scan cost until the user actually invokes the subtree (`loupedeck/.../verbs/command.go:61-80`),
2. scans repositories and collects discovered verbs (`.../command.go:97-118`),
3. converts each discovered verb into a runtime command wrapper (`.../command.go:121-143`),
4. creates Cobra commands from the generated Glazed descriptions (`.../command.go:146-192`),
5. executes the wrapped command through a custom invoker (`.../command.go:134-141`).

This is not just similar to what we need. It is structurally the same problem with a different runtime payload.

## Gap analysis

### Gap 1: there is no helper-verb discovery path inside a shared repository model

The current discovery flow only knows how to find named bots. There is no equivalent of:

- one shared repository abstraction consumed by both `bots` and `verbs`,
- `scanRepositories(...)` for helper scripts within that shared repository list,
- one-command-per-verb command generation.

### Gap 2: there is no helper-verb runtime contract

The current JS runtime host is bot-oriented. It expects a `defineBot(...)` script and produces a compiled bot handle. There is no helper runtime that says:

- this is a standalone JS verb,
- here are its flags and context,
- here are the runtime modules it can call,
- here is how it returns data to the CLI.

### Gap 3: there is no explicit bridge between live Discord inspection and bot simulation

Today, live Discord inspection and bot simulation are conceptually separate.

- The interactive runtime can dispatch bot commands with a synthetic `DispatchRequest`.
- The Discord host can talk to the real gateway and APIs.

What is missing is a tooling layer that can do both in one CLI workflow, for example:

1. fetch the live member from Discord,
2. build a synthetic request using that member snapshot,
3. run a bot command locally,
4. print the normalized response.

### Gap 4: there is no safety model for tool verbs yet

A helper-verb system that can call real Discord APIs can accidentally become a second mutation surface. We need a clear model for:

- read-only defaults,
- explicit opt-in write behavior,
- dry-run support,
- auditability in command output.

## Proposed architecture

## 1. Keep two execution models on purpose

The most important design decision is to keep **two intentionally separate JavaScript execution models**.

### Model A: `defineBot(...)` for interactive Discord bots

Use this for:

- slash commands,
- components,
- modals,
- events,
- long-lived runtime state,
- Discord-facing user workflows.

### Model B: `jsverbs` for standalone CLI helper tools

Use this for:

- live diagnostics,
- guild/member/role inspection,
- payload previews,
- bot-simulation workflows,
- maintenance tools,
- reproducible live debugging.

This separation is critical because it lets helper tools stay small and composable without forcing every diagnostic script to pretend it is a bot.

## 2. Add a new top-level `verbs` subtree

The recommended CLI shape is:

```text
discord-bot
├── run
├── validate-config
├── sync-commands
├── bots
└── verbs
```

The `verbs` subtree should be a lazy command, similar to loupedeck:

```text
NewLazyVerbsCommand()
    -> discover helper-verb repositories from CLI/env/config
    -> scan JS files with jsverbs.ScanDir / ScanFS
    -> convert VerbSpec -> Glazed command descriptions
    -> mount generated commands under the verbs subtree
    -> re-execute with the real generated tree
```

### Why top-level `verbs` instead of nesting under `bots`

Because helper verbs are not named bots. They are tools. Nesting them under `bots` would blur the conceptual contract and create discovery confusion.

## 3. Use one shared repository concept for both bots and helper verbs

The user explicitly chose a **single repository input** rather than separate `--bot-repository` and `--verbs-repository` flags. The long-term design should therefore move toward one repeatable repository flag such as `--repository`, with both named-bot discovery and helper-verb discovery reading from the same repository set.

### Repository rules

1. The same repository list feeds both named-bot discovery and helper-verb discovery.
2. Embedded repositories are always loaded in addition to filesystem repositories.
3. The `bots` subtree still handles named bot execution specially.
4. The `verbs` subtree still handles jsverb discovery and invocation specially.

Recommended layout inside one shared repository root:

```text
<repository-root>/
  examples/
    discord-bots/
      show-space/
      ui-showcase/
  verbs/
    discord/
      inspect-member.js
      inspect-guild-roles.js
      simulate-bot-command.js
      simulate-bot-component.js
      build-ui-payload.js
  verbs-rw/
    discord/
      send-message.js
      pin-message.js
      add-role.js
```

The core requirement is that helper-verb discovery scans the shared repository roots through `jsverbs`, while named-bot discovery continues to filter for `defineBot(...)` scripts.

## 4. Build a helper-verb runtime factory

Introduce a runtime factory dedicated to helper verbs.

### Responsibilities

For each verb invocation, the runtime factory should:

1. create a fresh Goja runtime,
2. register the helper modules,
3. load the JS file that defines the verb,
4. invoke the selected verb function with parsed values,
5. normalize the result for CLI output.

### Required modules

The helper runtime should expose at least:

1. **A separate live Discord inspection module**
   - read-only guild/role/member/channel/message fetch helpers,
2. **UI module**
   - so tooling verbs can build and inspect payloads,
3. **One bot simulation module**
   - to load a `defineBot(...)` script and dispatch synthetic requests,
4. **Optional database module**
   - for tools that need to inspect bot-side persisted state,
5. **Structured output helpers**
   - text/glaze-friendly output, with JSON blobs returned explicitly when a verb wants them.

### Pseudocode

```go
func NewDiscordVerbRuntime(ctx context.Context, registry *jsverbs.Registry, repo Repository) (*engine.Runtime, error) {
    builder := engine.NewBuilder(...)
        .WithModules(engine.DefaultRegistryModules())
        .WithRuntimeModuleRegistrars(
            jsdiscord.NewRegistrar(jsdiscord.Config{}),     // low-level Discord JS module(s)
            &jsdiscord.UIRegistrar{},                      // UI builders
            &VerbProbeRegistrar{},                         // bot simulation + helper utilities
        )

    return builder.Build().NewRuntime(ctx)
}
```

## 5. Use two explicit helper modules instead of stuffing everything into `require("discord")`

The helper-verb runtime needs both a live inspection surface and a bot-simulation surface, but the user chose to keep those concerns explicit.

### Chosen module split

1. **Live inspection module**
   - working name: `require("discord-cli")`
2. **Simulation module**
   - working name: `require("discord-probe")`

### Why not overload `require("discord")`

Because `require("discord")` already means “the embedded bot runtime DSL” in this repository. Overloading it with probe-only helpers would blur the line between bot authoring and tooling.

### Recommended API surface

```js
const probe = require("discord-probe")

const bot = await probe.loadBot("./examples/discord-bots/show-space/index.js")

const result = await probe.dispatchCommand({
  bot,
  name: "announce",
  args: { artist: "Test", date: "2026-05-22" },
  user: { id: "123" },
  member: { id: "123", roles: ["admin-role"] },
  guild: { id: "456", name: "Venue" },
  channel: { id: "789" },
  config: { adminRoleId: "admin-role" },
})
```

### Proposed probe functions

1. `loadBot(path)`
2. `dispatchCommand(request)`
3. `dispatchComponent(request)`
4. `dispatchModal(request)`
5. `normalizeResponse(value)`
6. `makeMemberSnapshot(liveDiscordMember)`

## 6. Separate pure tooling verbs from live Discord verbs

Not every helper verb should need live credentials.

### Class A: pure local tooling verbs

Examples:

- build a UI payload,
- normalize an embed,
- inspect a bot descriptor,
- simulate a command from static inputs.

These can run entirely against local files and synthetic inputs.

### Class B: live Discord probe verbs

Examples:

- list guild roles,
- fetch a member by user ID,
- fetch a channel,
- inspect pinned messages.

These need real Discord credentials and should say so explicitly in the schema.

### Class C: hybrid verbs

Examples:

- fetch a live member,
- convert it to a `MemberSnapshot`,
- simulate a specific bot command locally,
- print the resulting response.

These are the most valuable tools for troubleshooting permission bugs and payload issues.

## 7. Separate read-only and writable helper verbs physically

This system should assume that helper verbs are **inspection tools first**, but the user explicitly wants writable helper verbs to exist as well.

### Safety rules

1. Read-only helper verbs live under a normal directory such as `verbs/`.
2. Mutating helper verbs live under a separate directory such as `verbs-rw/`.
3. Mutating verbs still require explicit opt-in flags such as `--allow-write`.
4. Mutating verbs should support `--dry-run` where possible.
5. Output should clearly state whether the verb performed writes.

### Why this matters

A separate `verbs-rw/` directory creates a visible boundary in the repository. It helps reviewers and operators distinguish safe diagnostic tools from scripts that can send messages, pin posts, or mutate members.

## 8. Reuse `jsverbs` metadata, but do not make it Discord-specific

This repository already has prior design guidance that `jsverbs` should remain generic rather than growing Discord-specific annotations. That is the right long-term choice.

### What stays in `jsverbs`

- command names,
- descriptions,
- flags,
- sections,
- positional arguments,
- output mode,
- command hierarchy.

### What stays in the runtime layer

- Discord session wiring,
- bot loading,
- member/guild/channel fetch helpers,
- request simulation,
- read-only/write-mode enforcement.

This keeps `jsverbs` reusable and keeps Discord-specific behavior in this repository’s runtime modules.

## Proposed CLI and file layout

## 1. CLI shape

```text
discord-bot verbs list

discord-bot verbs inspect guild-roles \
  --guild-id 586274407350272042

discord-bot verbs inspect member \
  --guild-id 586274407350272042 \
  --user-id 363877777977376768

discord-bot verbs simulate bot-command \
  --bot ./examples/discord-bots/show-space/index.js \
  --command announce \
  --guild-id 586274407350272042 \
  --user-id 363877777977376768
```

## 2. Suggested source layout

```text
cmd/discord-bot/
  root.go
  verbs.go                    # lazy verbs command registration

internal/verbcli/
  command.go                  # lazy command builder
  bootstrap.go                # repository discovery for helper verbs
  runtime_factory.go          # runtime creation per verb
  invoker.go                  # shared invoker implementation
  output.go                   # glaze/text output handling plus optional JSON-as-string row helpers

internal/jsdiscord/
  verb_discord_cli_module.go  # require("discord-cli") for live Discord inspection
  verb_probe_module.go        # require("discord-probe") for simulation helpers
  verb_probe_runtime.go       # loadBot + dispatch helpers

verbs/
  discord/
    inspect-member.js
    inspect-guild-roles.js
    simulate-bot-command.js
    simulate-bot-component.js
    build-ui-payload.js

verbs-rw/
  discord/
    send-message.js
    pin-message.js
    add-role.js
```

## API sketches

## 1. Example helper verb annotation

```js
__verb__("inspect-member", {
  short: "Fetch a guild member and print roles",
  parents: ["inspect"],
  outputMode: "glaze",
  fields: {
    guildId: { type: "string", required: true, help: "Discord guild ID" },
    userId: { type: "string", required: true, help: "Discord user ID" },
  },
})

async function inspectMember(ctx) {
  const discord = require("discord-cli")
  const member = await discord.members.fetch(ctx.parameters.guildId, ctx.parameters.userId)
  return {
    userId: member.id,
    roleCount: (member.roles || []).length,
    roles: member.roles || [],
  }
}
```

## 2. Example hybrid simulation verb

```js
__verb__("simulate-bot-command", {
  short: "Run a bot command locally using live member data",
  parents: ["simulate"],
  outputMode: "glaze",
  fields: {
    botScript: { type: "string", required: true },
    command: { type: "string", required: true },
    guildId: { type: "string", required: true },
    userId: { type: "string", required: true },
    configJson: { type: "string", required: false },
  },
})

async function simulateBotCommand(ctx) {
  const discord = require("discord-cli")
  const probe = require("discord-probe")

  const member = await discord.members.fetch(ctx.parameters.guildId, ctx.parameters.userId)
  const bot = await probe.loadBot(ctx.parameters.botScript)

  return await probe.dispatchCommand({
    bot,
    name: ctx.parameters.command,
    user: { id: ctx.parameters.userId },
    member: probe.makeMemberSnapshot(member),
    guild: { id: ctx.parameters.guildId },
    config: JSON.parse(ctx.parameters.configJson || "{}"),
  })
}
```

## Pseudocode for the Go integration

## 1. Lazy root command

```go
func NewLazyVerbsCommand() *cobra.Command {
    return &cobra.Command{
        Use:                "verbs",
        Short:              "Run annotated Discord helper verbs",
        DisableFlagParsing: true,
        Args:               cobra.ArbitraryArgs,
        RunE: func(cmd *cobra.Command, args []string) error {
            bootstrap, err := verbcli.DiscoverBootstrapFromCommand(cmd)
            if err != nil {
                return err
            }
            resolvedCmd, err := verbcli.NewCommand(bootstrap)
            if err != nil {
                return err
            }
            adoptHelpAndOutput(cmd, resolvedCmd)
            resolvedCmd.SetArgs(args)
            return resolvedCmd.ExecuteContext(cmd.Context())
        },
    }
}
```

## 2. Runtime factory

```go
func NewVerbRuntime(ctx context.Context, repo Repository, registry *jsverbs.Registry) (*engine.Runtime, error) {
    builder := engine.NewBuilder(...)
        .WithModules(engine.DefaultRegistryModules())
        .WithRuntimeModuleRegistrars(
            &DiscordCLIRegistrar{SessionFactory: ...},
            &jsdiscord.UIRegistrar{},
            &DiscordProbeRegistrar{},
        )
    factory, err := builder.Build()
    if err != nil {
        return nil, err
    }
    return factory.NewRuntime(ctx)
}
```

## 3. Invoker flow

```text
CLI invocation
  -> parse flags through Glazed schema built from VerbSpec
  -> create runtime via runtime factory
  -> load script file that owns the selected verb
  -> call annotated function with parsed values
  -> normalize return value
  -> print as glaze rows or text
```

## Phased implementation plan

## Phase 1 — create the CLI skeleton

### Goal

Add a top-level lazy `verbs` subtree without yet implementing live Discord helpers.

### Files

- `cmd/discord-bot/root.go`
- `cmd/discord-bot/verbs.go` (new)
- `internal/verbcli/command.go` (new)
- `internal/verbcli/bootstrap.go` (new)

### Deliverables

1. `discord-bot verbs` exists.
2. Helper-verb discovery reads from the same shared repository list as the named-bot subsystem.
3. Embedded repositories are always included.
4. A dry “hello world” jsverb can be listed and run.

## Phase 2 — add a runtime factory for helper verbs

### Goal

Execute helper verbs in a runtime that can load JavaScript files and expose controlled modules.

### Files

- `internal/verbcli/runtime_factory.go` (new)
- `internal/verbcli/invoker.go` (new)

### Deliverables

1. helper verbs run through a custom invoker rather than the default runtime-owning path,
2. text and glaze output modes both work,
3. runtime lifetime is per invocation.

## Phase 3 — add low-level live Discord inspection verbs

### Goal

Support read-only live probes against guilds, roles, members, channels, and messages through the separate `discord-cli` module.

### Files

- `internal/jsdiscord/verb_discord_cli_module.go` (new)
- `verbs/discord/inspect-*.js` (new)

### Deliverables

1. `inspect-guild-roles`
2. `inspect-member`
3. `inspect-channel`
4. `inspect-pins`

## Phase 4 — add bot simulation helpers

### Goal

Load `defineBot(...)` scripts and dispatch synthetic requests from helper verbs.

### Files

- `internal/jsdiscord/verb_probe_module.go` (new)
- `internal/jsdiscord/verb_probe_runtime.go` (new)
- `verbs/discord/simulate-*.js` (new)

### Deliverables

1. `simulate-bot-command`
2. `simulate-bot-component`
3. `simulate-bot-modal`

## Phase 5 — add writable helper verbs under `verbs-rw/` and harden safety/docs

### Goal

Make the subsystem safe and self-explanatory for daily developer use.

### Files

- help docs under `pkg/doc/...`
- example verb repo docs
- test suites for live-probe stubs and simulation paths

### Deliverables

1. read-only defaults for `verbs/`,
2. writable-verb segregation under `verbs-rw/`,
3. explicit write-mode flags,
4. better dry-run output,
5. clear usage docs,
6. example workflows for debugging real bots.

## Testing and validation strategy

## 1. Unit tests

Add focused tests for:

- helper-verb discovery,
- Glazed schema generation from annotated verbs,
- runtime factory creation,
- probe module request normalization,
- command hierarchy generation.

## 2. Golden/help tests

Add help-output tests similar to the loupedeck pattern to ensure:

- lazy `verbs` command stays stable,
- generated verb help shows the expected flags,
- parent grouping is preserved.

## 3. Runtime tests with stubs

For live-probe modules, prefer session stubs so tests can validate:

- request construction,
- normalized outputs,
- read-only enforcement,
- error formatting.

## 4. Manual validation workflow

Recommended manual validation sequence for developers:

```bash
# list helper verbs
GOWORK=off go run ./cmd/discord-bot verbs --help

# run a pure local helper verb
GOWORK=off go run ./cmd/discord-bot verbs build-ui-payload --script ./verbs/discord/build-ui-payload.js

# run a live inspection verb
GOWORK=off go run ./cmd/discord-bot verbs inspect-member \
  --guild-id "$DISCORD_GUILD_ID" \
  --user-id "$DISCORD_USER_ID"

# run a bot simulation verb
GOWORK=off go run ./cmd/discord-bot verbs simulate-bot-command \
  --bot-script ./examples/discord-bots/show-space/index.js \
  --command debug \
  --guild-id "$DISCORD_GUILD_ID" \
  --user-id "$DISCORD_USER_ID"
```

## Risks, tradeoffs, and alternatives

## Risk 1: helper verbs become a shadow mutation surface

If the runtime exposes too many write-capable APIs too early, developers may accidentally rely on helper verbs for production actions instead of debugging. The mitigation is read-only defaults, explicit write flags, and dry-run support.

## Risk 2: tool sprawl

If helper verbs are added without strong conventions, the repo could fill with one-off scripts that duplicate each other. The mitigation is a dedicated helper-verb repository, documented naming conventions, and a small curated starter set.

## Risk 3: conflating bot authoring with tooling authoring

If the project tries to make every helper verb look like a bot or every bot look like a verb, both models get worse. The mitigation is to keep `defineBot(...)` and `jsverbs` intentionally separate.

## Alternative 1: keep adding slash debug commands to bots

Rejected as the primary solution. Slash debug commands are useful for some operator-facing diagnostics, but they are too heavyweight and too coupled to the bot surface for broad developer tooling.

## Alternative 2: build all helper tools in Go only

Rejected as the primary solution. Pure Go tooling is possible, but it loses the main ergonomic benefit here: reusing the same JavaScript runtime concepts, helper modules, and payload-building APIs that bot authors already use.

## Alternative 3: put helper verbs under the existing `bots` repository model

Rejected. The current `bots` discovery path intentionally looks for `defineBot(...)` scripts (`internal/botcli/bootstrap.go:184-200`). Forcing helper verbs into that model would create confusing discovery rules and poor separation of concerns.

## Resolved decisions

1. **Separate live inspection module**: keep live Discord inspection in its own helper module rather than overloading `require("discord")`. The guide uses `require("discord-cli")` as the working name.
2. **One simulation module**: keep bot simulation helpers together in a single module, `require("discord-probe")`.
3. **One shared repository flag/concept**: use one repository input for both named bots and helper verbs. Embedded repositories should always be loaded.
4. **No built-in JSON mode flag**: helper verbs that want JSON-like output can emit a serialized JSON blob as text or as a single row field.
5. **Writable verbs are allowed**: keep them in a separate repository subtree such as `verbs-rw/` and continue to gate them with explicit write flags.

## References

### Current repository files

- `cmd/discord-bot/root.go:15-64` — current root command wiring; shows where a top-level `verbs` subtree would fit.
- `internal/botcli/command.go:12-215` — current named-bot CLI subtree using Cobra.
- `internal/botcli/bootstrap.go:18-200` — named-bot repository discovery and the `defineBot`/`require("discord")` filter.
- `internal/jsdiscord/host.go:21-52` — current runtime creation path for `defineBot(...)` scripts.

### Upstream/reference implementations

- `/home/manuel/code/wesen/corporate-headquarters/go-go-goja/pkg/jsverbs/command.go:41-210` — jsverbs command/schema generation.
- `/home/manuel/code/wesen/corporate-headquarters/go-go-goja/pkg/jsverbs/model.go:74-157` — core jsverbs registry and verb metadata model.
- `/home/manuel/code/wesen/corporate-headquarters/loupedeck/cmd/loupedeck/cmds/verbs/command.go:57-218` — lazy verbs subtree, runtime wrappers, and custom invoker pattern.

### Prior design analysis in this repository

- `ttmp/2026/04/21/CODEQUAL-2026-0421--code-quality-review-js-discord-bot/design-doc/02-glazed-command-framework-migration-and-jsverb-integration-design.md` — prior analysis of Glazed migration and jsverbs integration tradeoffs.
