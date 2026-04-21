---
Title: Glazed Command Framework Migration and JSVerb Integration Design
Ticket: CODEQUAL-2026-0421
Status: active
Topics:
    - code-quality
    - architecture
    - refactoring
    - glazed
    - jsverbs
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../corporate-headquarters/go-go-goja/pkg/jsverbs/command.go
      Note: jsverbs Registry.Commands() builds Glazed commands from JS annotations
    - Path: ../../../../../../../corporate-headquarters/loupedeck/cmd/loupedeck/cmds/verbs/command.go
      Note: Loupedeck jsverb Glazed command wrapper (reference implementation)
    - Path: cmd/discord-bot/commands.go
      Note: Existing Glazed host commands (reference pattern)
    - Path: internal/botcli/command.go
      Note: Current pure-Cobra bots commands that need Glazed migration
    - Path: internal/botcli/run_schema.go
      Note: Manual flag parser and runtime config schema builder
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---






# Glazed Command Framework Migration and JSVerb Integration Design

**Ticket:** CODEQUAL-2026-0421  
**Date:** 2026-04-21  
**Scope:** Migrate all commands to Glazed; evaluate jsverbs for JS bot command declarations  
**Target audience:** New engineering interns and maintainers  

---

## 1. Executive Summary

The `js-discord-bot` CLI currently lives in a **mixed state**: the top-level host commands (`run`, `validate-config`, `sync-commands`) are proper Glazed commands, but the `bots` subcommands (`list`, `help`, `run`) are hand-rolled Cobra commands with manual flag parsing and no Glazed integration. This inconsistency means:

- The `bots` commands do not support Glazed output formats (`--output json`, `--output yaml`, `--fields`).
- The `bots run` command reimplements flag parsing that Cobra already does.
- There is no declarative schema for bot runtime config; it is built imperatively in `run_schema.go`.
- The JavaScript bots themselves cannot declare their commands in a way that the Go CLI can introspect without loading the full JS runtime.

This document proposes:
1. **Migrate the `bots` subcommands to Glazed** using the standard `cmds.CommandDescription` + `RunIntoGlazeProcessor` pattern.
2. **Introduce jsverbs** (as used in `loupedeck`) so that JavaScript bots can annotate their handlers with declarative metadata that the Go CLI can scan statically, without executing the JS runtime.
3. **Unify the command hierarchy** so every command in the tree is a first-class Glazed citizen.

The loupedeck project already solved this exact problem. We can model our solution on its `cmd/loupedeck/cmds/verbs` package.

---

## 2. Current Command Landscape

### 2.1 What is Glazed?

Glazed is a Go framework for building structured CLIs. Every Glazed command has:
- A **schema** with sections and fields (flags, arguments, choices, defaults).
- A **`RunIntoGlazeProcessor`** method that emits rows instead of printing directly.
- Automatic support for `--output`, `--fields`, `--stream`, `--print-schema`, etc.
- Embedded help integration.

### 2.2 Current command inventory

```text
discord-bot
├── run              ✅ Glazed command (cmd/discord-bot/commands.go)
├── validate-config  ✅ Glazed command (cmd/discord-bot/commands.go)
├── sync-commands    ✅ Glazed command (cmd/discord-bot/commands.go)
└── bots             ❌ Pure Cobra (internal/botcli/command.go)
    ├── list         ❌ Pure Cobra
    ├── help         ❌ Pure Cobra
    └── run          ❌ Pure Cobra + manual pre-parser
```

### 2.3 The `bots` commands in detail

**`bots list`** (`internal/botcli/command.go`, lines 17–35)
- Discovers bots from `--bot-repository` directories.
- Prints name, source, and description as tab-separated lines.
- No JSON/YAML output. No field selection.

**`bots help <bot>`** (`internal/botcli/command.go`, lines 37–69)
- Prints bot metadata, command list, event list, and run schema in plain text.
- No structured output. No way to pipe into `jq`.

**`bots run <bot>`** (`internal/botcli/command.go`, lines 80–140)
- Uses `DisableFlagParsing: true`.
- Implements a manual pre-parser (`preparseRunArgs` in `run_schema.go`) that extracts known flags (`--bot-token`, `--sync-on-start`) and passes the rest to a dynamic Glazed schema parser.
- This is the most complex and fragile part of the CLI.

### 2.4 Why the manual parser exists

The `bots run` command needs to accept **dynamic flags** based on the selected bot's runtime config schema. For example, the knowledge-base bot exposes `--db-path`, `--capture-threshold`, etc. Cobra does not support dynamic flags natively, so the current code:
1. Pre-parses to find the bot selector and known flags.
2. Resolves the bot to get its `RunSchema`.
3. Builds a Glazed schema from the `RunSchema`.
4. Re-parses remaining args with that schema.

This is clever but violates the Glazed convention. Loupedeck solves the same problem differently: it uses **lazy command construction**.

---

## 3. The Loupedeck Model: How jsverbs Work

### 3.1 What are jsverbs?

`jsverbs` is a package in `go-go-goja` that provides **static analysis of JavaScript files** to extract annotated function definitions. It uses tree-sitter to parse JS without executing it.

In a JS file, you annotate a function like this:

```js
__verb__("configure", {
  short: "Configure the device",
  parents: ["documented"],
  outputMode: "text",
  fields: {
    title: { type: "string", required: true, help: "Display title" },
    theme: { type: "string", default: "dark", choices: ["dark", "light"] },
  }
});

function configure(ctx) {
  // ... implementation ...
}
```

The `__verb__` annotation is a no-op at runtime (it is defined as an empty function in the JS runtime), but tree-sitter can find it statically.

### 3.2 The jsverbs pipeline

```text
JavaScript files in repository
      |
      v
jsverbs.ScanDir(rootDir, opts)   // tree-sitter parse, no JS execution
      |
      v
Registry (collection of VerbSpec)
      |
      v
Registry.Commands()              // builds Glazed commands from each VerbSpec
      |
      v
[]cmds.Command                   // fully typed Glazed commands with schemas
```

### 3.3 Key jsverbs types

```go
// From go-go-goja/pkg/jsverbs/model.go
type VerbSpec struct {
    FunctionName string
    Name         string
    Short        string
    Long         string
    OutputMode   string        // "glaze" or "text"
    Parents      []string      // command hierarchy: ["group", "subgroup"]
    Tags         []string
    UseSections  []string
    Fields       map[string]*FieldSpec
    File         *FileSpec
    Params       []ParameterSpec
}

type FieldSpec struct {
    Name     string
    Type     string        // "string", "int", "bool", "stringList", ...
    Help     string
    Short    string        // short flag, e.g. "t" for --title
    Bind     string
    Section  string
    Default  interface{}
    Choices  []string
    Required bool
    Argument bool          // positional argument
}
```

### 3.4 How loupedeck wires jsverbs into Cobra

Loupedeck's `cmd/loupedeck/cmds/verbs/command.go` shows the pattern:

```go
// 1. Bootstrap: discover repositories
bootstrap, err := DiscoverBootstrap(args)

// 2. Scan: tree-sitter parse all JS files
repositories, err := scanRepositories(bootstrap)

// 3. Collect: deduplicate verbs across repos
discovered, err := collectDiscoveredVerbs(repositories)

// 4. Build: create Glazed command wrappers
commands, err := buildCommands(discovered, liveSceneInvokerFactory)

// 5. Register: add to Cobra tree
for _, command := range commands {
    description := command.Description()
    parentCmd := findOrCreateParentCommand(root, description.Parents)
    cobraCommand, err := buildRuntimeCobraCommand(command)
    parentCmd.AddCommand(cobraCommand)
}
```

The critical insight: **each JS verb becomes a full Glazed command** with:
- A generated schema (sections, fields, defaults, choices).
- A `RunIntoWriter` or `RunIntoGlazeProcessor` implementation.
- Proper `--help` output showing all flags.

### 3.5 Lazy command construction

Because scanning JS files is not instant, loupedeck provides a `NewLazyCommand` that defers the full scan until the user actually runs `loupedeck verbs ...`:

```go
func NewLazyCommand() *cobra.Command {
    return &cobra.Command{
        Use:                "verbs",
        DisableFlagParsing: true,
        Args:               cobra.ArbitraryArgs,
        RunE: func(cmd *cobra.Command, args []string) error {
            bootstrap, err := DiscoverBootstrapFromCommand(cmd)
            resolvedCmd, err := NewCommand(bootstrap)  // <-- scan happens here
            resolvedCmd.SetArgs(args)
            return resolvedCmd.ExecuteContext(cmd.Context())
        },
    }
}
```

This means `loupedeck --help` is fast (no JS scanning). `loupedeck verbs --help` triggers the scan once.

---

## 4. Proposed Design for js-discord-bot

### 4.1 High-level vision

Every command in `discord-bot` should be a Glazed command. The `bots` subcommands should be reimplemented using Glazed patterns. In the long term, JavaScript bots should optionally use `__verb__` annotations so their handlers are discoverable as static Glazed commands.

```text
discord-bot                           (Glazed root)
├── run                               (existing Glazed host command)
├── validate-config                   (existing Glazed host command)
├── sync-commands                     (existing Glazed host command)
└── bots                              (Glazed group — lazy-loaded)
    ├── list                          (Glazed command — emits rows)
    ├── help                          (Glazed command — emits rows)
    └── run                           (Glazed command — with dynamic schema)
```

### 4.2 Phase 1: Migrate `bots list` to Glazed

**Current implementation:** `internal/botcli/command.go`, lines 17–35.

**Current behavior:**
```go
for _, bot := range bots {
    line := bot.Name() + "\t" + bot.SourceLabel()
    if desc := bot.Description(); desc != "" {
        line += "\t" + desc
    }
    fmt.Fprintln(cmd.OutOrStdout(), line)
}
```

**Proposed Glazed implementation:**

```go
// internal/botcli/commands_list.go
package botcli

import (
    "context"
    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/fields"
    "github.com/go-go-golems/glazed/pkg/cmds/values"
    "github.com/go-go-golems/glazed/pkg/middlewares"
    "github.com/go-go-golems/glazed/pkg/settings"
    "github.com/go-go-golems/glazed/pkg/types"
)

type BotsListCommand struct {
    *cmds.CommandDescription
}

func NewBotsListCommand() (*BotsListCommand, error) {
    glazedSection, err := settings.NewGlazedSchema()
    if err != nil { return nil, err }

    cmdDesc := cmds.NewCommandDescription(
        "list",
        cmds.WithShort("List discovered bot implementations"),
        cmds.WithLong(`List all JavaScript bot implementations found in the configured repositories.`),
        cmds.WithFlags(
            fields.New("bot-repository", fields.TypeString,
                fields.WithHelp("Directory containing bot implementations (repeatable)"),
            ),
        ),
        cmds.WithSections(glazedSection),
    )
    return &BotsListCommand{CommandDescription: cmdDesc}, nil
}

func (c *BotsListCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    vals *values.Values,
    gp middlewares.Processor,
) error {
    // Decode repositories from vals or use defaults
    bootstrap, err := bootstrapFromValues(vals)
    if err != nil { return err }

    bots, err := DiscoverBots(ctx, bootstrap)
    if err != nil { return err }

    for _, bot := range bots {
        row := types.NewRow(
            types.MRP("name", bot.Name()),
            types.MRP("source", bot.SourceLabel()),
            types.MRP("description", bot.Description()),
            types.MRP("script_path", bot.ScriptPath()),
            types.MRP("commands", len(bot.Descriptor.Commands)),
            types.MRP("events", len(bot.Descriptor.Events)),
        )
        if err := gp.AddRow(ctx, row); err != nil {
            return err
        }
    }
    return nil
}
```

**Benefits:**
- `discord-bot bots list --output json` now works.
- `discord-bot bots list --fields name,commands` now works.
- Consistent with other Glazed commands.

---

### 4.3 Phase 2: Migrate `bots help <bot>` to Glazed

**Current implementation:** `internal/botcli/command.go`, lines 37–69.

**Current behavior:** Plain text output. No structured data.

**Proposed Glazed implementation:**

```go
// internal/botcli/commands_help.go
func (c *BotsHelpCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    vals *values.Values,
    gp middlewares.Processor,
) error {
    bootstrap, err := bootstrapFromValues(vals)
    if err != nil { return err }

    bots, err := DiscoverBots(ctx, bootstrap)
    if err != nil { return err }

    selector := "" // decoded from vals positional arg
    bot, err := ResolveBot(selector, bots)
    if err != nil { return err }

    // Emit one row for the bot metadata
    if err := gp.AddRow(ctx, types.NewRow(
        types.MRP("kind", "bot"),
        types.MRP("name", bot.Name()),
        types.MRP("description", bot.Description()),
        types.MRP("script_path", bot.ScriptPath()),
        types.MRP("source", bot.SourceLabel()),
    )); err != nil { return err }

    // Emit one row per command
    for _, cmd := range bot.Descriptor.Commands {
        if err := gp.AddRow(ctx, types.NewRow(
            types.MRP("kind", "command"),
            types.MRP("name", cmd.Name),
            types.MRP("description", cmd.Description),
            types.MRP("type", cmd.Type),
        )); err != nil { return err }
    }

    // Emit one row per event
    for _, ev := range bot.Descriptor.Events {
        if err := gp.AddRow(ctx, types.NewRow(
            types.MRP("kind", "event"),
            types.MRP("name", ev.Name),
        )); err != nil { return err }
    }

    // Emit one row per run config field
    if bot.Descriptor.RunSchema != nil {
        for _, section := range bot.Descriptor.RunSchema.Sections {
            for _, field := range section.Fields {
                if err := gp.AddRow(ctx, types.NewRow(
                    types.MRP("kind", "config_field"),
                    types.MRP("section", section.Slug),
                    types.MRP("name", field.Name),
                    types.MRP("type", field.Type),
                    types.MRP("required", field.Required),
                    types.MRP("default", field.Default),
                )); err != nil { return err }
            }
        }
    }
    return nil
}
```

**Benefits:**
- `discord-bot bots help ping --output yaml` now works.
- External tools can parse the output reliably.

---

### 4.4 Phase 3: Migrate `bots <bot>` to Glazed

This is the hardest migration because of the dynamic schema problem.

#### 4.4.1 The problem

The `bots run` command's schema depends on which bot is selected. Glazed commands normally have static schemas. We need a pattern for dynamic schema construction.

#### 4.4.2 The loupedeck solution

Loupedeck uses **lazy command construction**: the `verbs` Cobra command is created empty, and when the user runs `loupedeck verbs ...`, it scans JS files, builds Glazed commands dynamically, and re-executes with the real command tree.

We adapt this for `discord-bot bots <bot>`.

#### 4.4.3 Proposed design

**Decision:** The new canonical UX is `discord-bot bots <bot>` (flat), not `discord-bot bots run <bot>` (nested). This is a breaking change accepted by the team.

**Step 1: Static `bots` group command**

The `bots` command itself is a lazy placeholder:

```go
// internal/botcli/root.go
func NewBotsGroupCommand() *cobra.Command {
    return &cobra.Command{
        Use:   "bots",
        Short: "List, inspect, and run named JavaScript bot implementations",
        PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
            return logging.InitLoggerFromCobra(cmd)
        },
    }
}
```

**Step 2: Static Glazed subcommands (`list`, `help`)**

`bots list` and `bots help` are regular Glazed commands registered directly.

**Step 3: Lazy bot subcommands (`bots <bot>`)**

Each discovered bot becomes a direct subcommand of `bots`:

```go
// Proposed new UX:
discord-bot bots ping --bot-token ... --db-path ./data.sqlite
```

This matches the loupedeck pattern (`loupedeck verbs <verb>`) and eliminates one level of nesting.

**The loupedeck approach applied:**

Loupedeck does not try to make one command with a dynamic schema. Instead, it makes **one command per verb**. Each verb has a static schema derived from its `__verb__` annotation.

For discord-bot, this means:

```text
discord-bot bots                    (lazy group)
├── list                            (Glazed)
├── help                            (Glazed)
├── ping                            (Glazed command for ping bot)
├── knowledge-base                  (Glazed command for KB bot)
├── moderation                      (Glazed command for moderation bot)
└── ...
```

Each bot-specific command has a schema that includes:
- The static host flags (`--bot-token`, `--application-id`, `--guild-id`, `--sync-on-start`).
- The bot's runtime config fields (`--db-path`, `--capture-threshold`, etc.).

**Implementation sketch:**

```go
// internal/botcli/commands_bots_lazy.go

func NewBotsLazyGroup() *cobra.Command {
    return &cobra.Command{
        Use:                "bots",
        Short:              "List, inspect, and run named JavaScript bot implementations",
        DisableFlagParsing: true,
        RunE: func(cmd *cobra.Command, args []string) error {
            // Step 1: Pre-parse just enough to find repositories and the bot selector
            parsed, err := preparseRunArgs(args, defaultPreParsedRunArgs())
            if err != nil { return err }

            // Step 2: Discover bots
            bootstrap, err := bootstrapFromRepositories(parsed.BotRepositories)
            bots, err := DiscoverBots(cmd.Context(), bootstrap)
            selected, err := ResolveBot(parsed.Selector, bots)

            // Step 3: Build a Glazed command for this specific bot
            botCmd, err := buildBotCommand(selected, parsed)

            // Step 4: Re-execute with the real command
            botCmd.SetArgs(args)
            return botCmd.ExecuteContext(cmd.Context())
        },
    }
}

func buildBotCommand(bot DiscoveredBot, preArgs preParsedRunArgs) (*cobra.Command, error) {
    // Build schema: static host flags + dynamic bot config fields
    schema_, nameMap, err := buildRunSchema(bot)
    // Add host flags to schema
    // ...

    cmdDesc := cmds.NewCommandDescription(
        bot.Name(),
        cmds.WithShort("Run "+bot.Name()),
        cmds.WithSections(schema_),
    )

    // Create a Glazed command wrapper
    glazedCmd := &botRunGlazedCommand{desc: cmdDesc, bot: bot, nameMap: nameMap}

    // Build Cobra command from Glazed command
    cobraCmd, err := cli.BuildCobraCommandFromCommand(glazedCmd)
    return cobraCmd, err
}
```

**Key difference from current code:**
- The current code builds a schema and parses args in one shot inside the `RunE` function.
- The proposed code builds a **real Glazed command** with a real schema, then lets Glazed handle parsing and execution.
- The output goes through `RunIntoGlazeProcessor`, so `--output json` works for the bot command too (it can emit a "running" row).

---

### 4.5 Phase 4: Introduce jsverbs for bot handlers

This is the long-term, more ambitious part. Currently, JavaScript bots declare their handlers imperatively:

```js
const { defineBot } = require("discord")
module.exports = defineBot(({ command, event, configure }) => {
  configure({ name: "ping", description: "Ping bot" })
  command("ping", { description: "Reply with pong" }, async (ctx) => {
    return { content: "pong" }
  })
})
```

The Go host learns about these commands by:
1. Loading the JS into goja.
2. Calling `defineBot`.
3. Calling `describe()` on the returned bot object.
4. Parsing the returned JSON into `BotDescriptor`.

This requires executing the JS runtime. With jsverbs, we could discover commands statically:

```js
// Proposed jsverb-style annotation
__verb__("ping", {
  short: "Reply with pong",
  parents: ["discord-bot", "ping"],
  outputMode: "glaze",
  fields: {
    text: { type: "string", help: "Text to echo", argument: true }
  }
});

function ping(ctx) {
  return { content: "pong: " + ctx.args.text }
}
```

**Wait — this does not match the discord-bot model.**

In loupedeck, each `__verb__` maps to one CLI command. In discord-bot, the JS file defines a **bot** that contains multiple commands, events, and components. The `__verb__` model is designed for standalone CLI verbs, not for Discord bot definitions.

**Revised proposal: Use `__verb__` for standalone bot scripts, not for `defineBot` bots.**

For simple bots that are essentially one command, a jsverb annotation makes sense:

```js
// examples/discord-bots/simple-echo/index.js
__verb__("echo", {
  short: "Echo a message back",
  parents: ["discord-bot", "bots"],
  outputMode: "text",
  fields: {
    message: { type: "string", required: true, help: "Message to echo" }
  }
});

function echo(ctx) {
  return ctx.args.message;
}
```

But for rich bots with multiple commands and events, the `defineBot` model is still appropriate. The host needs to load and execute the JS to register handlers with Discordgo.

**Hybrid model:**

```text
Simple scripts (one command):    use __verb__ annotations
Rich bots (multiple commands):   use defineBot (existing model)
```

For the scope of this document, **Phase 4 is optional and exploratory**. The immediate value is in Phases 1–3.

---

## 5. File Layout

After migration, the `botcli` package should look like this:

```text
internal/botcli/
  model.go              // DiscoveredBot, Repository, Bootstrap (unchanged)
  bootstrap.go          // Discovery logic (unchanged)
  resolve.go            // ResolveBot (unchanged)
  run_schema.go         // Runtime schema building (refactored, simplified)
  runtime.go            // Run loop (unchanged)
  root.go               // Bots group command
  commands_list.go      // Glazed "list" command
  commands_help.go      // Glazed "help" command
  commands_bots_lazy.go // Lazy bot-specific command builder
  command.go            // OLD — delete after migration
  command_test.go       // Update for Glazed patterns
```

The `cmd/discord-bot/` tree should look like this:

```text
cmd/discord-bot/
  main.go
  root.go               // Add bots group registration
  commands.go           // Host commands (run, validate, sync)
  commands_bots.go      // Import and register botcli Glazed commands
```

---

## 6. Implementation Plan

### Phase 1: Migrate `bots list` (1–2 days)
1. Create `internal/botcli/commands_list.go` with `BotsListCommand`.
2. Implement `RunIntoGlazeProcessor` that emits rows.
3. Wire into `cmd/discord-bot/root.go` via `cli.BuildCobraCommandFromCommand`.
4. Delete the old `newListCommand` from `internal/botcli/command.go`.
5. Update tests.

### Phase 2: Migrate `bots help` (1–2 days)
1. Create `internal/botcli/commands_help.go` with `BotsHelpCommand`.
2. Implement `RunIntoGlazeProcessor` that emits bot metadata, commands, events, and config fields as rows.
3. Wire into root.
4. Delete the old `newHelpCommand`.
5. Update tests.

### Phase 3: Migrate `bots <bot>` (3–5 days)
1. Refactor `buildRunSchema` and `parseRuntimeConfigArgs` to be reusable.
2. Create `internal/botcli/commands_bots_lazy.go` that builds bot-specific Glazed commands on demand.
3. Ensure the lazy builder composes host flags + bot config fields into one schema.
4. Wire the lazy builder into the `bots` group command so each bot becomes a direct subcommand.
5. Delete the old `newRunCommand` and `preparseRunArgs`.
6. Update tests.

### Phase 4: Evaluate jsverbs (exploratory, 2–3 days)
1. Prototype a simple bot using `__verb__` annotations.
2. Use `jsverbs.ScanDir` to discover it.
3. Build a Glazed command from the `VerbSpec`.
4. Compare complexity with the `defineBot` model.
5. Document findings and decide whether to adopt.

---

## 7. API References

### 7.1 Glazed types used

```go
// Command definition
type GlazeCommand interface {
    RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error
}

// Command description constructor
cmds.NewCommandDescription(name, ...Option)

// Field definition
fields.New(name, fields.Type, ...Option)

// Cobra builder
cli.BuildCobraCommandFromCommand(cmd cmds.Command, ...BuildOption)
```

### 7.2 jsverbs types used

```go
// Scanning
jsverbs.ScanDir(rootDir, opts) (*Registry, error)
jsverbs.ScanFS(fsys, root, opts) (*Registry, error)

// Command generation
registry.Commands() ([]cmds.Command, error)
registry.CommandDescriptionForVerb(verb *VerbSpec) (*cmds.CommandDescription, error)

// VerbSpec key fields
VerbSpec.Name         // command name
VerbSpec.Parents      // command hierarchy
VerbSpec.OutputMode   // "glaze" or "text"
VerbSpec.Fields       // map of field specs
```

### 7.3 Key files in loupedeck (reference implementation)

| File | Purpose |
|------|---------|
| `cmd/loupedeck/cmds/verbs/bootstrap.go` | Repository discovery, scanning, deduplication |
| `cmd/loupedeck/cmds/verbs/command.go` | Glazed command wrapper, lazy construction, Cobra wiring |
| `cmd/loupedeck/cmds/verbs/command_test.go` | Tests for help output, custom invokers, result routing |
| `pkg/scriptmeta/scriptmeta.go` | Higher-level metadata extraction using jsverbs |

---

## 8. Risks and Alternatives

### 8.1 Risk: Lazy command construction breaks `--help` autocomplete

**Problem:** Shell autocomplete tools (bash, zsh, fish) need to know the full command tree statically. If `bots` subcommands are built lazily, autocomplete cannot list bot names.

**Mitigation:** Provide a `discord-bot bots list` command that emits bot names. Users can pipe this into scripts. Alternatively, generate a static command tree at build time by scanning known repositories.

### 8.2 Risk: Dynamic schema breaks `print-schema`

**Problem:** `--print-schema` should show the full schema. If the schema is built lazily, `--print-schema` on the placeholder command shows nothing useful.

**Mitigation:** The lazy builder should handle `--print-schema` by building the real command first, then delegating.

### 8.3 Alternative: Keep `bots run <bot>` as a single command with sub-flags

Instead of making each bot a subcommand, keep the current UX but use Glazed's `RunIntoGlazeProcessor`:

```bash
discord-bot bots run ping --db-path ./data.sqlite
```

The schema would include ALL possible runtime config fields from ALL bots, but only the selected bot's fields would be validated. This is simpler but messier.

**Verdict:** Rejected. The team decided to move to `bots <bot>` flat UX.

### 8.4 Alternative: Use Cobra dynamic flags

Cobra supports adding flags at runtime via `cmd.Flags().AddFlag(...)`. We could build the schema, convert it to Cobra flags, and attach them to the command before parsing.

**Verdict:** Rejected. It bypasses Glazed entirely and reintroduces the manual parsing problem we are trying to solve.

---

## 9. Decisions

1. **UX: `bots <bot>` (flat)** ✅ DECIDED
   - The new canonical UX is `discord-bot bots <bot>`, not `discord-bot bots run <bot>`.
   - This is a breaking change. Update scripts and documentation accordingly.

2. **jsverbs: no Discord-specific metadata** ✅ DECIDED
   - jsverbs remains generic. Discord metadata (command types, permissions, etc.) stays in the `defineBot` model.
   - Phase 4 (jsverbs evaluation) is for simple standalone scripts only.

3. **Bot config validation**
   - Currently, `cfg.Validate()` checks `BotToken` and `ApplicationID`.
   - In the Glazed model, these should be `fields.WithRequired(true)` on the schema.

---

*End of document.*
