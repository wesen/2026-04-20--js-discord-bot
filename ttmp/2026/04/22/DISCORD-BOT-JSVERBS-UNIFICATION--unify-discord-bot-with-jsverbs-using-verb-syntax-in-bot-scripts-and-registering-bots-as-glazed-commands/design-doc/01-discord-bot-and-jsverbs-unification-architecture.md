---
Title: Discord-Bot and Jsverbs Unification Architecture
Ticket: DISCORD-BOT-JSVERBS-UNIFICATION
Status: active
Topics:
    - discord-bot
    - jsverbs
    - glazed
    - cli
    - bot-registration
    - command-discovery
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/botcli/bootstrap.go
      Note: Bot discovery from --bot-repository directories
    - Path: internal/botcli/command.go
      Note: Raw cobra commands for bots list
    - Path: internal/botcli/run_dynamic_schema.go
      Note: Temporary cobra parser hack for RunSchema
    - Path: internal/botcli/run_static_args.go
      Note: Manual flag parsing that Glazed would replace
    - Path: internal/jsdiscord/bot_compile.go
      Note: BotHandle dispatch methods and botDraft capture
    - Path: internal/jsdiscord/descriptor.go
      Note: BotDescriptor and RunSchema parsing
    - Path: internal/jsdiscord/runtime.go
      Note: Registers discord module and defineBot — needs __verb__ polyfill
ExternalSources: []
Summary: 'Comprehensive analysis of the discord-bot architecture, the jsverbs architecture from go-go-goja, and a detailed design for unifying them so that Discord bot scripts can use __verb__ syntax while remaining fully runnable as bots. Target audience: new interns.'
LastUpdated: 2026-04-22T18:00:00-04:00
WhatFor: Onboarding guide and architecture reference for engineers working on discord-bot + jsverbs integration. Explains every subsystem in detail with file references, pseudocode, and implementation guidance.
WhenToUse: When you need to understand how the discord-bot discovers bots, how jsverbs discovers commands, how Glazed registration works, or how to bridge the two systems.
---








# Discord-Bot and Jsverbs Unification Architecture

## Executive Summary

The `discord-bot` project and the `go-go-goja` project both let you write Go-backed CLI tools whose behavior is defined in JavaScript. However, they use **completely different** JS APIs and discovery mechanisms:

- **jsverbs** (in `go-go-goja`) scans `.js` files for `__verb__`, `__section__`, and `__package__` metadata declarations, then registers each discovered function as a **Glazed command** with structured output, rich help, and Cobra integration.
- **discord-bot** (in `2026-04-20--js-discord-bot`) uses a **runtime `defineBot` API**: JS scripts call `require("discord")`, use `defineBot(({ command, event, configure }) => { ... })` to build a bot object, and export it. The Go host then loads this object, calls `describe()` to get metadata, and calls `dispatchCommand()`, `dispatchEvent()`, etc. at runtime.

This document explains both architectures in exhaustive detail, then asks and answers the central question: **Can a single JavaScript file use `__verb__` syntax (for CLI usage) AND `defineBot` syntax (for Discord bot usage) at the same time?**

The answer is **yes, in principle**, but it requires understanding the fundamental differences in how the two systems treat JavaScript source code:

| Aspect | jsverbs | discord-bot |
|--------|---------|-------------|
| Discovery | Static scan (Tree-sitter parses AST) | Runtime load (script executes in Goja VM) |
| JS API | `__verb__(name, meta)`, `__section__`, `__package__` | `defineBot(({ command, event, configure }) => ...)` |
| Output | Glazed rows (`--output json/csv/table`) | Discord interaction responses |
| Execution | One-shot: parse args → call JS function → emit rows | Long-running: connect to gateway → dispatch events → reply |
| Command registration | `registry.Commands()` → `cli.AddCommandsToRootCommand` | Manual cobra commands in `internal/botcli/command.go` |
| Help system | Glazed rich markdown help | Plain text via `printBotHelp` |
| Config schema | `__verb__` fields + `__section__` | `configure({ run: { fields: {...} } })` |

The document then proposes a **phased unification plan**:

1. **Phase 1**: Convert `bots list`, `bots help`, and `bots run` to use Glazed commands (immediate UX improvement).
2. **Phase 2**: Teach the discord-bot scanner to ALSO recognize `__verb__` metadata in bot scripts, so bot scripts can expose CLI verbs alongside their Discord behavior.
3. **Phase 3**: Create a unified runtime where the same JS function can be invoked both as a Glazed command (via `registry.invoke`) and as a Discord command handler (via `dispatchCommand`).

---

## Part 1: The Discord-Bot Architecture

### 1.1 The big picture

The discord-bot is a Go application that loads JavaScript files, treats them as "bot implementations," and either:

- **Runs them as a live Discord bot** (`discord-bot run` or `discord-bot bots run <bot>`) — connects to the Discord gateway, registers slash commands, and dispatches events.
- **Lists them** (`discord-bot bots list`) — shows discovered bots.
- **Shows help for one** (`discord-bot bots help <bot>`) — shows metadata, commands, events, and run config.

The architecture has three layers:

```
┌─────────────────────────────────────────────────────────────┐
│  CLI Layer (cmd/discord-bot/)                                │
│  - root.go: root cobra command, logging, help system         │
│  - commands.go: run, validate-config, sync-commands           │
│  - botcli (internal/botcli/): bots list, help, run           │
├─────────────────────────────────────────────────────────────┤
│  Discovery Layer (internal/botcli/bootstrap.go)              │
│  - Walk --bot-repository directories                         │
│  - Find .js files with defineBot + require("discord")        │
│  - Inspect via jsdiscord.InspectScript                       │
├─────────────────────────────────────────────────────────────┤
│  Runtime Layer (internal/jsdiscord/)                         │
│  - host.go: creates Goja runtime, loads script               │
│  - runtime.go: registers "discord" native module             │
│  - bot_compile.go: CompileBot, BotHandle, botDraft           │
│  - descriptor.go: BotDescriptor, RunSchema, parsing          │
│  - dispatch: host_dispatch.go, runtime_dispatch_*.go         │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 How a bot script works (JavaScript side)

A minimal bot script looks like this:

```js
// examples/discord-bots/support/index.js
const { defineBot } = require("discord");

module.exports = defineBot(({ command, event, configure }) => {
  configure({
    name: "support",
    description: "Support workflows",
    run: {
      fields: {
        api_key: { type: "string", help: "External API key", required: true }
      }
    }
  });

  command("support-ticket", {
    description: "Create a support ticket",
    options: { topic: { type: "string", required: true } }
  }, async (ctx) => {
    await ctx.defer({ ephemeral: true });
    await ctx.edit({ content: `Ticket: ${ctx.args.topic}` });
  });

  event("guildCreate", async (ctx) => {
    ctx.log.info("Joined guild", { guild: ctx.guild.name });
  });
});
```

Key JavaScript APIs provided by the `discord` module:

- **`defineBot(builderFn)`**: The entry point. `builderFn` receives an API object and must return a bot configuration. The builder is called **at module load time** inside the Goja VM.
- **`configure(options)`**: Sets bot metadata (name, description) and runtime config schema (`run.fields`).
- **`command(name, [spec], handler)`**: Registers a slash command. `handler` receives a `ctx` object with `args`, `discord`, `reply`, `edit`, `defer`, etc.
- **`event(name, handler)`**: Registers a Discord gateway event handler (e.g., `guildCreate`, `messageCreate`).
- **`subcommand(rootName, name, [spec], handler)`**: Registers a subcommand.
- **`component(customId, handler)`**: Registers a component interaction handler (buttons, select menus).
- **`modal(customId, handler)`**: Registers a modal submit handler.
- **`autocomplete(commandName, optionName, handler)`**: Registers autocomplete suggestions.

### 1.3 How the Go host loads a bot (the `Host` lifecycle)

**File**: `internal/jsdiscord/host.go`

The `Host` struct encapsulates everything needed to run one bot script:

```go
type Host struct {
    scriptPath    string      // absolute path to the .js file
    runtime       *engine.Runtime  // go-go-goja engine runtime
    handle        *BotHandle       // compiled bot with dispatch methods
    runtimeConfig map[string]any   // config values from CLI
}
```

**Step 1: Build the engine** (`NewHost`)

```go
factory, err := engine.NewBuilder(
    engine.WithModuleRootsFromScript(absScript, engine.DefaultModuleRootsOptions()),
).WithModules(engine.DefaultRegistryModules()).
    WithRuntimeModuleRegistrars(NewRegistrar(Config{})).
    WithRequireOptions(require.WithGlobalFolders(
        filepath.Dir(absScript),
        filepath.Join(filepath.Dir(absScript), "node_modules"),
    )).Build()
```

This creates a Goja runtime with:
- The standard go-go-goja module registry (`fs`, `exec`, `timer`, `database`, etc.)
- A custom native module called `"discord"` (registered via `NewRegistrar`)
- Node.js-style `require()` that can find modules relative to the script

**Step 2: Load the script**

```go
rt, err := factory.NewRuntime(ctx)
value, err := rt.Require.Require(absScript)
```

When `require(absScript)` executes the JS file, the script calls `defineBot(...)`, which invokes `RuntimeState.defineBot()` in Go. That function:

1. Creates a `botDraft`
2. Builds an API object with `command`, `event`, `configure`, etc.
3. Calls the user's builder function with that API
4. Calls `draft.finalize(vm)` which returns a bot object with `describe`, `dispatchCommand`, `dispatchEvent`, etc.

**Step 3: Compile the bot**

```go
handle, err := CompileBot(rt.VM, value)
```

`CompileBot` extracts the callable methods from the bot object:
- `describe`
- `dispatchCommand`
- `dispatchSubcommand`
- `dispatchEvent`
- `dispatchComponent`
- `dispatchModal`
- `dispatchAutocomplete`

**Step 4: Describe**

```go
desc, err := host.Describe(ctx)
```

This calls the JS `describe()` function via `runtimebridge.Owner.Call()`, which runs inside the runtime owner's safe execution context. The result is a `map[string]any` that gets parsed into a `BotDescriptor` by `descriptorFromDescribe()`.

### 1.4 The BotDescriptor

**File**: `internal/jsdiscord/descriptor.go`

`BotDescriptor` is the Go-side representation of everything the bot declared:

```go
type BotDescriptor struct {
    Name          string
    Description   string
    ScriptPath    string
    Metadata      map[string]any
    Commands      []CommandDescriptor       // slash commands
    Subcommands   []SubcommandDescriptor    // subcommands
    Events        []EventDescriptor         // gateway events
    Components    []ComponentDescriptor     // button/select handlers
    Modals        []ModalDescriptor         // modal submit handlers
    Autocompletes []AutocompleteDescriptor  // autocomplete handlers
    RunSchema     *RunSchemaDescriptor      // CLI config schema
}
```

The `RunSchemaDescriptor` is particularly important for CLI usage:

```go
type RunSchemaDescriptor struct {
    Sections []RunSectionDescriptor
}

type RunSectionDescriptor struct {
    Slug   string
    Title  string
    Fields []RunFieldDescriptor
}

type RunFieldDescriptor struct {
    Name         string  // external name (e.g., "db-path")
    InternalName string  // JS property name (e.g., "db_path")
    Type         string  // "string", "integer", "bool", etc.
    Help         string
    Required     bool
    Default      any
}
```

When you run `discord-bot bots run mybot --db-path ./data.db`, the CLI:
1. Parses static args (`--bot-token`, `--application-id`, etc.) via `parseStaticRunnerArgs`
2. Builds a dynamic schema from `bot.Descriptor.RunSchema`
3. Creates a throwaway `cobra.Command` + `CobraParser` to parse the remaining args
4. Maps parsed values back to `map[string]any` using `runtimeConfigFromParsedValues`
5. Passes that map to the bot at runtime via `Host.SetRuntimeConfig()`
6. The bot accesses config via `ctx.config.dbPath` in JS

### 1.5 How commands are dispatched at runtime

**File**: `internal/jsdiscord/bot_compile.go`

When Discord sends an interaction (e.g., a user types `/support-ticket`), the Go host builds a `DispatchRequest` and calls `BotHandle.dispatchCommand`:

```go
// Go side (simplified)
req := DispatchRequest{
    Name:    "support-ticket",
    Args:    map[string]any{"topic": "hello"},
    User:    userSnapshot,
    Guild:   guildSnapshot,
    Channel: channelSnapshot,
    Reply:   func(ctx context.Context, msg any) error { ... },
    // ... many more fields
}
result, err := handle.dispatchCommand(goja.Undefined(), vm.ToValue(req))
```

On the JS side, `dispatchCommand` looks up the handler by name and calls it:

```js
// Inside the bot object (generated by botDraft.finalize)
dispatchCommand: function(input) {
    const command = findCommand(commands, input.name);
    const ctx = buildContext(store, input, "command", name, metadata);
    return command.handler(undefined, ctx);
}
```

The `ctx` object passed to the handler contains:
- `ctx.args` — parsed command arguments
- `ctx.config` — runtime config values
- `ctx.discord` — Discord API proxy (guilds, channels, messages, members, etc.)
- `ctx.reply(msg)`, `ctx.edit(msg)`, `ctx.followUp(msg)`, `ctx.defer(opts)` — response helpers
- `ctx.user`, `ctx.guild`, `ctx.channel`, `ctx.member` — Discord entity snapshots

### 1.6 The current CLI registration (bots subcommand)

**File**: `internal/botcli/command.go`

The `bots` subcommand tree is registered manually with raw `cobra.Command` objects:

```go
func NewCommand() *cobra.Command {
    root := &cobra.Command{Use: "bots", Short: "..."}
    root.AddCommand(newListCommand())    // plain text
    root.AddCommand(newHelpCommand())    // plain text
    root.AddCommand(newRunCommand())     // manual flag parsing
    return root
}
```

**`bots list`** prints tab-separated plain text:
```go
for _, bot := range bots {
    line := bot.Name() + "\t" + bot.SourceLabel()
    fmt.Fprintln(cmd.OutOrStdout(), line)
}
```

**`bots help <bot>`** prints plain text via `printBotHelp`:
```go
fmt.Fprintf(out, "Bot: %s\n", bot.Name())
fmt.Fprintf(out, "Source: %s\n", bot.ScriptPath())
// ... lists commands, events, run config fields
```

**`bots run <bot>`** is the most complex:
- Uses `DisableFlagParsing: true` to bypass Cobra entirely
- Manually parses ~10 static flags via `parseStaticRunnerArgs`
- Then manually parses dynamic flags via a throwaway `cobra.Command`
- No Glazed output pipeline, no `--output json`, no `--fields`

---

## Part 2: The Jsverbs Architecture

### 2.1 The big picture

jsverbs (in `go-go-goja/pkg/jsverbs`) is a system for exposing JavaScript functions as Glazed CLI commands. The architecture has four layers:

```
┌─────────────────────────────────────────────────────────────┐
│  Host Application (cmd/jsverbs-example/main.go)              │
│  - ScanDir(dir) → Registry                                   │
│  - registry.Commands() → []cmds.Command                      │
│  - cli.AddCommandsToRootCommand(root, commands)              │
├─────────────────────────────────────────────────────────────┤
│  Command Layer (pkg/jsverbs/command.go)                      │
│  - buildDescription() → CommandDescription                   │
│  - CommandForVerb() → Command (GlazeCommand)                 │
│  - WriterCommand for text output                             │
├─────────────────────────────────────────────────────────────┤
│  Binding Layer (pkg/jsverbs/binding.go)                      │
│  - buildVerbBindingPlan() → VerbBindingPlan                  │
│  - Maps JS params to CLI flag/section bindings               │
├─────────────────────────────────────────────────────────────┤
│  Scan Layer (pkg/jsverbs/scan.go)                            │
│  - Tree-sitter parses .js/.cjs files                         │
│  - Extracts __verb__, __section__, __package__               │
│  - Builds Registry with VerbSpec, SectionSpec, FileSpec      │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 The JavaScript API

A jsverb script looks like this:

```js
// testdata/jsverbs/basics.js
__section__("filters", {
  title: "Filters",
  description: "Shared filter flags",
  fields: {
    state: { type: "choice", choices: ["open", "closed"], default: "open", help: "Issue state" },
    labels: { type: "stringList", help: "Labels to filter on" }
  }
});

function greet(name, excited) {
  return { greeting: excited ? `Hello, ${name}!` : `Hello, ${name}` };
}

__verb__("greet", {
  short: "Greet one person",
  fields: {
    name: { argument: true, help: "Person name" },
    excited: { type: "bool", short: "e", help: "Add excitement" }
  }
});
```

Key JS APIs:

- **`__package__({ name, short, long, parents, tags })`**: Sets package metadata. Affects how the verb is parented in the command tree.
- **`__section__(slug, { title, description, fields })`**: Declares a reusable section schema. Can be referenced by verbs via `sections: ["slug"]`.
- **`__verb__(functionName, { short, long, output, parents, fields, sections })`**: Declares that a function should be exposed as a CLI command.
- **`doc\`...\``**: Doc template with YAML frontmatter for rich help.

### 2.3 The scan process

**File**: `pkg/jsverbs/scan.go`

`jsverbs.ScanDir(root)` walks the directory and parses each `.js` file with Tree-sitter:

1. **Parse the AST** using `tree-sitter-javascript`
2. **Walk top-level nodes** looking for:
   - `__package__({...})` calls
   - `__section__("slug", {...})` calls
   - `__verb__("name", {...})` calls
   - `doc\`...\`` template literals
   - Function declarations and arrow functions
3. **Build a `FileSpec`** with all extracted metadata
4. **Call `finalizeVerbs()`** to resolve defaults (names, parents, output modes)
5. **Return a `Registry`** containing all discovered verbs

### 2.4 From verb to Glazed command

**File**: `pkg/jsverbs/command.go`

`Registry.Commands()` converts each `VerbSpec` into a `cmds.Command`:

```go
func (r *Registry) Commands() ([]cmds.Command, error) {
    commands := make([]cmds.Command, 0, len(r.verbs))
    for _, verb := range r.verbs {
        cmd, err := r.CommandForVerb(verb)
        if err != nil { return nil, err }
        commands = append(commands, cmd)
    }
    return commands, nil
}
```

`CommandForVerb`:
1. Calls `buildDescription(verb)` to create a `CommandDescription` with schema sections
2. Checks `verb.OutputMode`:
   - `OutputModeGlaze` → returns `&Command{...}` (implements `GlazeCommand`)
   - `OutputModeText` → returns `&WriterCommand{...}` (implements `WriterCommand`)

### 2.5 The binding plan

**File**: `pkg/jsverbs/binding.go`

Before executing a verb, jsverbs builds a `VerbBindingPlan` that maps each JS function parameter to a source of parsed CLI values:

```go
type VerbBindingPlan struct {
    Verb               *VerbSpec
    Parameters         []ParameterBinding   // one per JS param
    ExtraFields        []ExtraFieldBinding  // declared fields with no param
    ReferencedSections []string
}
```

Binding modes:

- **`BindingModePositional`** (default): The param receives a single field value from the default section.
- **`BindingModeSection`**: The param receives an entire section object (e.g., `bind: "filters"`).
- **`BindingModeAll`**: The param receives the flat map of all parsed values (e.g., `bind: "all"`).
- **`BindingModeContext`**: The param receives host metadata (e.g., `bind: "context"`).

### 2.6 Runtime execution

**File**: `pkg/jsverbs/runtime.go`

When a jsverb command runs:

1. A new Goja runtime is created via `engine.NewBuilder()`
2. The script source is loaded via `runtime.Require.Require(modulePath)`
3. An **overlay** is injected that registers captured functions:
   ```js
   globalThis.__glazedVerbRegistry["/basics.js"] = {
     greet: typeof greet === "function" ? greet : undefined,
     echo: typeof echo === "function" ? echo : undefined,
     // ...
   };
   ```
4. The JS function is looked up in `__glazedVerbRegistry` and called with parsed arguments
5. The result is converted to Glazed rows (for `GlazeCommand`) or text (for `WriterCommand`)

---

## Part 3: Comparing the Two Systems

### 3.1 Discovery: static vs runtime

**jsverbs** uses **static discovery**. It parses the AST without executing the script. This means:
- ✅ Fast: no VM startup needed to discover commands
- ✅ Safe: malicious code in the script body can't run during discovery
- ❌ Limited: can only extract metadata from literal expressions (no computed values)

**discord-bot** uses **runtime discovery**. It loads the script into a Goja VM and calls `describe()`:
- ✅ Flexible: the script can compute metadata dynamically
- ✅ Rich: the bot object has methods (dispatch, describe) that the host can call
- ❌ Slower: requires VM creation per script
- ❌ Riskier: script executes during discovery

### 3.2 Command model: one-shot vs long-running

**jsverbs** commands are **one-shot**:
- Parse CLI args → call JS function → emit output → exit
- The JS function returns a value (object, array, string)
- No persistent state between invocations

**discord-bot** commands are **event-driven**:
- Load bot → connect to Discord gateway → wait for events
- Each event triggers a handler that receives a rich `ctx` object
- The bot maintains state via `ctx.store` (a MemoryStore)
- Runs until interrupted

### 3.3 Output model: structured rows vs Discord responses

**jsverbs** outputs through Glazed's pipeline:
- `RunIntoGlazeProcessor` feeds `types.Row` objects into a `middlewares.Processor`
- The processor handles `--output json`, `--output csv`, `--fields`, `--filter`, etc.

**discord-bot** outputs via Discord interactions:
- Handlers call `ctx.reply()`, `ctx.edit()`, `ctx.followUp()`
- These send HTTP requests to Discord's API
- No local structured output (except the `run` command's status row)

### 3.4 Config model: fields/sections vs run schema

**jsverbs** config is declarative:
```js
__verb__("greet", {
  fields: { name: { argument: true }, excited: { type: "bool" } }
});
```

**discord-bot** config is also declarative but nested differently:
```js
configure({
  run: {
    fields: { api_key: { type: "string", required: true } }
  }
});
```

Both map to Glazed `fields.Definition` + `schema.Section` in the end, but through different code paths.

---

## Part 4: The Unification Vision

### 4.1 Can a bot script use `__verb__` syntax?

**Yes.** A single `.js` file can contain both `__verb__` calls and `defineBot` calls because:

1. **Tree-sitter parsing** (jsverbs) looks for specific call expressions (`__verb__`, `__section__`, `__package__`). It ignores everything else, including `defineBot`.
2. **Goja execution** (discord-bot) executes the script. `__verb__` is not defined by default, but we could provide a no-op polyfill so the script doesn't crash.

Example of a unified script:

```js
// my-bot.js — works as both a Discord bot AND a CLI command

// ===== Discord bot API =====
const { defineBot } = require("discord");

module.exports = defineBot(({ command, event, configure }) => {
  configure({ name: "my-bot", description: "Does things" });

  command("ping", async (ctx) => {
    return { content: "Pong!" };
  });

  event("messageCreate", async (ctx) => {
    // ...
  });
});

// ===== jsverbs API (for CLI usage) =====
function status() {
  return { status: "ok", uptime: 42 };
}

__verb__("status", {
  short: "Check bot status",
  output: "glaze"
});

function sendMessage(channelId, content) {
  // Imagine this uses a Discord API module
  return { sent: true, channelId, content };
}

__verb__("send-message", {
  short: "Send a message to a channel",
  fields: {
    channel_id: { argument: true, help: "Channel ID" },
    content: { argument: true, help: "Message content" }
  },
  output: "glaze"
});
```

When scanned by jsverbs:
- `__verb__("status", ...)` and `__verb__("send-message", ...)` are discovered
- `defineBot` is ignored (Tree-sitter doesn't care about undefined functions)

When loaded by discord-bot:
- `defineBot` executes normally
- `__verb__` would need to be defined (or ignored). We can provide a no-op:
  ```js
  globalThis.__verb__ = globalThis.__verb__ || function() {};
  globalThis.__section__ = globalThis.__section__ || function() {};
  globalThis.__package__ = globalThis.__package__ || function() {};
  ```

### 4.2 What does unification actually mean?

There are **three levels** of unification, from simplest to most ambitious:

| Level | What changes | Effort | Value |
|-------|-------------|--------|-------|
| **A** | Convert `bots list`, `bots help`, `bots run` to Glazed commands | Low | High — immediate UX improvement |
| **B** | Teach jsverbs scanner to recognize bot scripts + teach bot runtime to ignore `__verb__` | Medium | High — single file for CLI + bot |
| **C** | Shared runtime: the same JS function can be called by both `registry.invoke()` and `dispatchCommand()` | High | Medium — code reuse between CLI and bot handlers |

### 4.3 Level A: Glazed commands for the bot CLI

This is the easiest win. Instead of raw cobra commands, implement:

**`bots list`** as a `GlazeCommand`:
```go
type listBotsCommand struct {
    *cmds.CommandDescription
    bootstrap botcli.Bootstrap
}

func (c *listBotsCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
    bots, err := botcli.DiscoverBots(ctx, c.bootstrap)
    for _, bot := range bots {
        gp.AddRow(ctx, types.NewRow(
            types.MRP("name", bot.Name()),
            types.MRP("source", bot.SourceLabel()),
            types.MRP("description", bot.Description()),
            types.MRP("commands", len(bot.Descriptor.Commands)),
            types.MRP("events", len(bot.Descriptor.Events)),
        ))
    }
    return nil
}
```

**`bots help <bot>`** as a `GlazeCommand` that emits structured metadata:
```go
// Instead of printBotHelp (plain text), emit rows for commands, events, config fields
// Then users can do: discord-bot bots help mybot --output json
```

**`bots run <bot>`** as a `GlazeCommand` with a proper schema:
```go
// Build a CommandDescription that includes:
// - Discord credential section (bot-token, application-id, etc.)
// - The bot's RunSchema sections
// - Standard glazed/command-settings sections
// Then use BuildCobraCommandFromCommand instead of DisableFlagParsing
```

### 4.4 Level B: Coexistence of `__verb__` and `defineBot`

This requires two changes:

**Change 1: Polyfill `__verb__` in the Discord runtime**

In `internal/jsdiscord/runtime.go`, extend the `Loader` to define no-ops:

```go
func (s *RuntimeState) Loader(vm *goja.Runtime, moduleObj *goja.Object) {
    exports := moduleObj.Get("exports").(*goja.Object)
    _ = exports.Set("defineBot", func(call goja.FunctionCall) goja.Value {
        return s.defineBot(vm, call)
    })

    // NEW: Define no-ops for jsverbs metadata so bot scripts don't crash
    _ = vm.Set("__package__", func(call goja.FunctionCall) goja.Value {
        return goja.Undefined()
    })
    _ = vm.Set("__section__", func(call goja.FunctionCall) goja.Value {
        return goja.Undefined()
    })
    _ = vm.Set("__verb__", func(call goja.FunctionCall) goja.Value {
        return goja.Undefined()
    })
}
```

**Change 2: Teach jsverbs to scan bot directories**

In the discord-bot's `main.go` (or a new `cmd`):

```go
// Scan the same bot repositories with jsverbs
registry, err := jsverbs.ScanDir(botRepoDir)
// The scanner will find __verb__ metadata in bot scripts
// and register them as glazed commands under "discord-bot bots <verb>"
```

### 4.5 Level C: Shared handler invocation

The most ambitious level: make it so the **same JS function** can be invoked both:
- As a Glazed command (via `registry.invoke()`) — one-shot, returns rows
- As a Discord command handler (via `dispatchCommand()`) — receives ctx, calls `ctx.reply()`

This requires a **unified handler signature** or an adapter pattern:

```js
// Unified handler: works for both CLI and Discord
function status(ctx) {
  // In Discord: ctx has ctx.reply, ctx.discord, etc.
  // In CLI: ctx has ctx.output (or we return a value)

  if (ctx.reply) {
    // Discord mode
    ctx.reply({ content: `Status: ${getStatus()}` });
  }
  // Always return a value for CLI mode
  return { status: getStatus() };
}
```

However, this is often more complexity than it's worth. The Discord `ctx` object is radically different from jsverbs' parsed argument binding. **Level B (coexistence)** is the sweet spot: the same file contains both APIs, but they invoke different functions optimized for their respective contexts.

---

## Part 5: Detailed Implementation Guide

### 5.1 Phase 1: Convert `bots list` to Glazed

**Goal**: `discord-bot bots list --output json` should work.

**Steps**:

1. In `internal/botcli/command.go`, replace `newListCommand()`:

```go
// OLD: raw cobra command
func newListCommand() *cobra.Command {
    return &cobra.Command{...}
}

// NEW: return a glazed command description
func newListGlazedCommand(bootstrap botcli.Bootstrap) cmds.GlazeCommand {
    desc := cmds.NewCommandDescription(
        "list",
        cmds.WithShort("List discovered bot implementations"),
        cmds.WithLong("Emit all discovered bots as structured rows."),
    )
    return &listBotsCommand{
        CommandDescription: desc,
        bootstrap:          bootstrap,
    }
}

type listBotsCommand struct {
    *cmds.CommandDescription
    bootstrap botcli.Bootstrap
}

func (c *listBotsCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
    bots, err := botcli.DiscoverBots(ctx, c.bootstrap)
    if err != nil { return err }
    for _, bot := range bots {
        row := types.NewRow(
            types.MRP("name", bot.Name()),
            types.MRP("source", bot.SourceLabel()),
            types.MRP("description", bot.Description()),
            types.MRP("commands", len(bot.Descriptor.Commands)),
            types.MRP("events", len(bot.Descriptor.Events)),
            types.MRP("components", len(bot.Descriptor.Components)),
            types.MRP("modals", len(bot.Descriptor.Modals)),
        )
        if err := gp.AddRow(ctx, row); err != nil {
            return err
        }
    }
    return nil
}
```

2. In `internal/botcli/command.go`, change `NewCommand()` to build the glazed command via `cli.BuildCobraCommandFromCommand`:

```go
func NewCommand(bootstrap botcli.Bootstrap) *cobra.Command {
    root := &cobra.Command{Use: "bots", Short: "..."}
    root.PersistentFlags().StringArray(BotRepositoryFlag, nil, "...")

    listCmd := newListGlazedCommand(bootstrap)
    listCobra, err := cli.BuildCobraCommandFromCommand(listCmd,
        cli.WithParserConfig(cli.CobraParserConfig{
            ShortHelpSections: []string{schema.DefaultSlug, schema.GlobalDefaultSlug},
        }),
    )
    if err != nil { panic(err) }
    root.AddCommand(listCobra)
    // ... same for help and run
    return root
}
```

3. Update `cmd/discord-bot/root.go` to pass bootstrap to `botcli.NewCommand()`.

### 5.2 Phase 2: Convert `bots run` to a `BareCommand` with `__verb__` fields as `ctx.config`

**Goal**: When a bot script declares `__verb__("run", { fields: {...} })`, running that verb should:
1. Parse the CLI flags via Glazed (no manual flag parsing)
2. Load the bot script into a Goja VM
3. Pass the parsed values into the bot as `ctx.config`
4. Connect to Discord and block until interrupted

**Why `BareCommand`**: A bot `run` verb doesn't emit Glazed rows and doesn't write to stdout. It just orchestrates the Discord lifecycle. `BareCommand` is the simplest Glazed interface for this:

```go
// glazed/pkg/cmds/cmds.go
type BareCommand interface {
    Command
    Run(ctx context.Context, parsedValues *values.Values) error
}
```

**The JS side**: The bot script declares a `run` verb with its runtime config fields:

```js
const { defineBot } = require("discord");

module.exports = defineBot(({ command, configure }) => {
  configure({ name: "knowledge-base", description: "Knowledge base bot" });

  command("ask", async (ctx) => {
    // Bot handlers access config from ctx.config
    const dbPath = ctx.config.db_path;
    const apiKey = ctx.config.api_key;
    // ...
  });
});

// ===== CLI verb that runs the bot =====
function run() {
  return { status: "placeholder" };  // never actually called
}

__verb__("run", {
  short: "Run the knowledge-base bot",
  output: "text",
  fields: {
    "bot-token":      { type: "string", required: true, help: "Discord bot token" },
    "application-id": { type: "string", help: "Application/client ID" },
    "guild-id":       { type: "string", help: "Optional guild ID for dev sync" },
    "db-path": {
      type: "string",
      default: "./data/knowledge.sqlite",
      help: "SQLite database path"
    },
    "api-key": {
      type: "string",
      required: true,
      help: "External API key"
    },
    "batch-size": {
      type: "integer",
      default: 100,
      help: "Items per batch"
    }
  }
});
```

**The Go side**: After jsverbs scans the directory, the host app recognizes `__verb__("run")` verbs and creates a `botRunCommand` instead of the standard `jsverbs.Command`:

```go
func buildBotRunCommand(verb *jsverbs.VerbSpec, registry *jsverbs.Registry) cmds.Command {
    desc, err := registry.CommandDescriptionForVerb(verb)
    if err != nil { panic(err) }

    return &botRunCommand{
        CommandDescription: desc,
        scriptPath:         verb.File.AbsPath,
    }
}

type botRunCommand struct {
    *cmds.CommandDescription
    scriptPath string
}

func (c *botRunCommand) Run(ctx context.Context, parsedValues *values.Values) error {
    // 1. Extract Discord credentials from parsed CLI values
    cfg := appconfig.Settings{
        BotToken:      getString(parsedValues, "bot-token"),
        ApplicationID: getString(parsedValues, "application-id"),
        GuildID:       getString(parsedValues, "guild-id"),
    }
    if err := cfg.Validate(); err != nil {
        return err
    }

    // 2. Build runtime config map from ALL parsed values (not just credentials)
    runtimeConfig := map[string]any{}
    parsedValues.ForEach(func(slug string, sectionVals *values.SectionValues) {
        sectionVals.Fields.ForEach(func(fieldName string, fv *fields.FieldValue) {
            if fv == nil || fv.Definition == nil {
                return
            }
            // Convert kebab-case CLI flag name to snake_case JS property name
            configKey := runtimeFieldInternalName(fieldName)
            runtimeConfig[configKey] = fv.Value
        })
    })

    // 3. Create the bot runtime (loads JS file into Goja VM)
    bot, err := botapp.New(cfg, botapp.WithScriptPath(c.scriptPath))
    if err != nil {
        return err
    }
    defer func() { _ = bot.Close() }()

    // 4. Inject config BEFORE opening — makes ctx.config available in all handlers
    bot.Host().SetRuntimeConfig(runtimeConfig)

    // 5. Connect to Discord gateway
    if err := bot.Open(); err != nil {
        return err
    }

    // 6. Block until Ctrl-C (Cobra's context is signal.NotifyContext)
    <-ctx.Done()
    return nil
}

func getString(parsed *values.Values, name string) string {
    if v, ok := parsed.GetDataMap()[name]; ok {
        if s, ok := v.(string); ok {
            return s
        }
    }
    return ""
}

// runtimeFieldInternalName converts kebab-case CLI names to snake_case JS names.
// Same logic as jsdiscord.runtimeFieldInternalName.
func runtimeFieldInternalName(name string) string {
    name = strings.TrimSpace(name)
    if name == "" {
        return ""
    }
    var out []rune
    for i, r := range name {
        switch {
        case r == '-':
            out = append(out, '_')
        case r >= 'A' && r <= 'Z':
            if i > 0 && len(out) > 0 && out[len(out)-1] != '_' {
                out = append(out, '_')
            }
            out = append(out, r+'a'-'A')
        default:
            out = append(out, r)
        }
    }
    return strings.Trim(strings.TrimSpace(string(out)), "_")
}
```

**The wiring in main.go**: After scanning with jsverbs, the host inspects each verb and creates the right command type:

```go
registry, err := jsverbs.ScanDir(botRepoDir)

var allCommands []cmds.Command

for _, verb := range registry.Verbs() {
    if verb.Name == "run" && isBotScript(verb.File.AbsPath) {
        // Host-managed run verb: BareCommand, no JS function call
        cmd := buildBotRunCommand(verb, registry)
        allCommands = append(allCommands, cmd)
    } else {
        // Standard jsverb: one-shot JS function call
        cmd, err := registry.CommandForVerb(verb)
        if err != nil { return nil, err }
        allCommands = append(allCommands, cmd)
    }
}

cli.AddCommandsToRootCommand(root, allCommands, nil, opts...)
```

**What this replaces**: The old `configure({ run: { fields: {...} } })` + `parseStaticRunnerArgs` + `parseRuntimeConfigArgs` stack is entirely replaced by the `__verb__` metadata + Glazed parsing. The JS bot doesn't care where `ctx.config` comes from — it just works.

| Old approach (discord-bot) | New approach (unified) |
|---------------------------|------------------------|
| `configure({ run: { fields: { dbPath: {...} } } })` | `__verb__("run", { fields: { "db-path": {...} } })` |
| `parseStaticRunnerArgs` (200 lines manual flag parsing) | Glazed `CobraParser` (automatic) |
| `parseRuntimeConfigArgs` (throwaway cobra command) | `values.Values` from Glazed parsing |
| `runtimeConfigFromParsedValues` (manual mapping) | `parsedValues.ForEach` → `runtimeConfig` map |
| Plain text help | Glazed markdown help with sections |

### 5.3 Phase 3: Add `__verb__` polyfill to Discord runtime

**File**: `internal/jsdiscord/runtime.go`

When the bot script is loaded for Discord execution (not jsverbs scanning), `__verb__` is undefined and would crash. Add no-op polyfills in the `Loader`:

```go
func (s *RuntimeState) Loader(vm *goja.Runtime, moduleObj *goja.Object) {
    // ... existing defineBot export ...

    // Polyfill jsverbs metadata functions so bot scripts can coexist
    for _, name := range []string{"__package__", "__section__", "__verb__", "doc"} {
        if vm.Get(name) == nil || goja.IsUndefined(vm.Get(name)) {
            _ = vm.Set(name, func(goja.FunctionCall) goja.Value { return goja.Undefined() })
        }
    }
}
```

This means a single JS file can contain:
- `defineBot(...)` — executed when loaded by Discord runtime
- `__verb__("run", ...)` — parsed by jsverbs scanner, ignored by Discord runtime
- `__verb__("status", ...)` — same: discovered by scanner, ignored by runtime

### 5.4 Phase 4: Scan bot repos with jsverbs

In the discord-bot's `main.go` (or a new subcommand):

```go
func scanBotReposForVerbs(repositories []botcli.Repository) ([]cmds.Command, error) {
    allCommands := []cmds.Command{}
    for _, repo := range repositories {
        registry, err := jsverbs.ScanDir(repo.RootDir)
        if err != nil { return nil, err }
        for _, verb := range registry.Verbs() {
            if verb.Name == "run" && isBotScript(verb.File.AbsPath) {
                cmd := buildBotRunCommand(verb, registry)
                allCommands = append(allCommands, cmd)
            } else {
                cmd, err := registry.CommandForVerb(verb)
                if err != nil { return nil, err }
                allCommands = append(allCommands, cmd)
            }
        }
    }
    return allCommands, nil
}
```

These commands get registered under `discord-bot bots <verb-path>` just like any other Glazed command.

---

## Part 6: File Reference for Onboarding

### Discord-bot files (read in this order)

| # | File | What to learn |
|---|------|---------------|
| 1 | `examples/discord-bots/support/index.js` | What a real bot script looks like |
| 2 | `internal/jsdiscord/runtime.go` | How `defineBot` is registered as a native module |
| 3 | `internal/jsdiscord/bot_compile.go` | How `botDraft` captures commands/events, how `CompileBot` works |
| 4 | `internal/jsdiscord/descriptor.go` | How `BotDescriptor` and `RunSchema` are parsed |
| 5 | `internal/jsdiscord/host.go` | How `Host` creates the Goja runtime and loads scripts |
| 6 | `internal/botcli/bootstrap.go` | How bot discovery works (walk dirs, `looksLikeBotScript`) |
| 7 | `internal/botcli/command.go` | How `bots list`, `bots help`, `bots run` are registered |
| 8 | `internal/botcli/run_dynamic_schema.go` | How `RunSchema` is converted to a temporary Cobra parser |
| 9 | `internal/botcli/run_static_args.go` | How static flags are manually parsed |
| 10 | `cmd/discord-bot/root.go` | Root command setup, help system, logging |
| 11 | `cmd/discord-bot/commands.go` | How the main `run`/`validate-config`/`sync-commands` are built as Glazed commands |

### Jsverbs files (read in this order)

| # | File | What to learn |
|---|------|---------------|
| 1 | `testdata/jsverbs/basics.js` | What a real jsverb script looks like |
| 2 | `pkg/jsverbs/scan.go` | Tree-sitter scanning, `__verb__`/`__section__` extraction |
| 3 | `pkg/jsverbs/model.go` | `Registry`, `VerbSpec`, `SectionSpec`, `FieldSpec` types |
| 4 | `pkg/jsverbs/binding.go` | `VerbBindingPlan`, parameter binding modes |
| 5 | `pkg/jsverbs/command.go` | `buildDescription`, `CommandForVerb`, `Command`/`WriterCommand` |
| 6 | `pkg/jsverbs/runtime.go` | Goja runtime creation, `__glazedVerbRegistry` overlay, invocation |
| 7 | `cmd/jsverbs-example/main.go` | Host app that scans and registers commands |

### Glazed files (common to both)

| # | File | What to learn |
|---|------|---------------|
| 1 | `pkg/cli/cobra.go` | `AddCommandsToRootCommand`, `BuildCobraCommandFromCommand` |
| 2 | `pkg/cli/cobra-parser.go` | `CobraParser`, `CobraParserConfig`, `ShortHelpSections` |
| 3 | `pkg/help/cmd/cobra.go` | Help rendering, `shortHelpSections` filtering |
| 4 | `pkg/cmds/schema/cobra_flag_groups.go` | Flag group computation, `DefaultSlug`, `GlobalDefaultSlug` |
| 5 | `pkg/cmds/cmds.go` | `Command`, `GlazeCommand`, `WriterCommand` interfaces |

---

## Part 7: Testing Strategy

1. **Unit test `listBotsCommand`**: Create a mock `Bootstrap` with fake repositories, assert JSON output.
2. **Unit test `buildBotRunDescription`**: Create a `DiscoveredBot` with a `RunSchema`, assert the generated `CommandDescription` has the right sections.
3. **Integration test**: Place a bot script with `__verb__` metadata in `testdata/discord-bots/`, run `discord-bot bots list`, verify it doesn't crash and the `__verb__` metadata is ignored by the bot runtime.
4. **Integration test**: Run `jsverbs.ScanDir` on the discord-bot examples directory, verify it discovers any `__verb__` metadata if present.
5. **E2E test**: `go run ./cmd/discord-bot bots list --output json` produces valid JSON.

---

## Part 8: Open Questions

1. **Should `bots run` dynamically register subcommands per-bot, or should it stay as `bots run <bot>` with a single command that builds its schema at runtime?**
   - Per-bot subcommands: better UX (`discord-bot bots run mybot --help` shows exact flags), but requires rescanning at startup
   - Single dynamic command: simpler, but `--help` can't show bot-specific flags

2. **Should the `__verb__` polyfill capture metadata so bot scripts can also be self-describing for CLI usage?**
   - No-op polyfill: simplest, but `__verb__` metadata is only visible to jsverbs scanner
   - Capturing polyfill: could build a secondary registry inside the VM, but overlaps with `describe()`

3. **How do we handle the fact that discord-bot scripts use `module.exports = defineBot(...)` while jsverbs scripts are scanned, not executed?**
   - Tree-sitter scanning ignores `module.exports` — it only cares about top-level `__verb__` calls
   - As long as `__verb__` calls are at the top level, they'll be found regardless of `module.exports`

4. ~~Should we extract the `RunSchema` fields from `configure({ run: ... })` and also expose them as `__section__` for jsverbs?~~
   - **Resolved**: The `__verb__("run", { fields: {...} })` declaration replaces `configure({ run: { fields: ... } })` entirely. The `__verb__` metadata becomes the single source of truth for the bot's CLI config schema. The Go host converts `*values.Values` → `map[string]any` via `runtimeFieldInternalName()` and injects it via `Host.SetRuntimeConfig()`. The JS bot accesses it as `ctx.config.db_path` just like before.
