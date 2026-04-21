# Design and Implementation Guide: Discord Interaction Types Demo Bot

## Goal

This document explains how to extend the `js-discord-bot` framework to support all Discord application command interaction types — slash commands (chat input), user context menu commands, message context menu commands, and subcommands — and how to build a demo bot that exercises each one. It is written for a new intern who needs to understand the system end-to-end.

## Table of Contents

1. [What are Discord Application Commands?](#what-are-discord-application-commands)
2. [Current Framework Architecture](#current-framework-architecture)
3. [Gap Analysis: What is Missing](#gap-analysis-what-is-missing)
4. [Design: How We Will Add the Missing Types](#design-how-we-will-add-the-missing-types)
5. [File-by-File Implementation Plan](#file-by-file-implementation-plan)
6. [The Demo Bot](#the-demo-bot)
7. [Testing Strategy](#testing-strategy)
8. [Reference Material](#reference-material)

---

## What are Discord Application Commands?

Discord application commands are the primary way users interact with bots in modern Discord. There are three top-level command types:

| Type | Discord Name | Value | How Users See It | Use Case |
|------|-------------|-------|------------------|----------|
| Chat Input | `CHAT_INPUT` | 1 | `/command-name` typed into the message box | Rich commands with options, subcommands, autocomplete |
| User | `USER` | 2 | Right-click a user → Apps → "Command Name" | Actions targeted at a specific user (e.g., "Show Avatar", "Kick") |
| Message | `MESSAGE` | 3 | Right-click a message → Apps → "Command Name" | Actions targeted at a specific message (e.g., "Quote", "Pin") |

In addition, Chat Input commands support **subcommands** and **subcommand groups**, which let you nest related operations under a single root command. For example, `/admin kick @user` and `/admin ban @user` are two subcommands under the root `/admin` command.

### Subcommands vs Subcommand Groups

A subcommand is an option of type `SUB_COMMAND` (value 1). A subcommand group is an option of type `SUB_COMMAND_GROUP` (value 2) that contains further subcommands.

```
/admin                    ← root command
  /kick                   ← subcommand (option type 1)
    user: @user           ← string option
    reason: "spam"        ← string option
  /ban                    ← subcommand (option type 1)
    user: @user           ← string option
    duration: 7           ← integer option
```

The Discord API docs we downloaded via defuddle live in:
- `ttmp/2026/04/21/DISCORD-BOT-020--.../sources/discord-application-commands-docs.md`

---

## Current Framework Architecture

Our framework is a Go application that embeds a JavaScript runtime (via `goja`) and exposes a `discord` module to bot authors. The architecture has four layers:

### Layer 1: The JavaScript Bot DSL

Bot authors write `.js` files that call `defineBot()` and register handlers:

```js
const { defineBot } = require("discord")

module.exports = defineBot(({ command, event, configure }) => {
  configure({ name: "my-bot", description: "..." })

  command("ping", { description: "Reply with pong" }, async () => {
    return { content: "pong" }
  })
})
```

The available registration helpers today are:
- `command(name, spec?, handler)` — slash commands only
- `event(name, handler)` — gateway events
- `component(customId, handler)` — button/select clicks
- `modal(customId, handler)` — modal submissions
- `autocomplete(commandName, optionName, handler)` — typed suggestions
- `configure(options)` — bot metadata and runtime config

**Source file:** `internal/jsdiscord/runtime.go`

### Layer 2: The Bot Draft

When `defineBot()` runs, it builds a `botDraft` in Go memory. The draft collects all registered handlers into typed slices:

```go
type botDraft struct {
  commands      []*commandDraft
  events        []*eventDraft
  components    []*componentDraft
  modals        []*modalDraft
  autocompletes []*autocompleteDraft
}
```

**Source file:** `internal/jsdiscord/bot.go`

### Layer 3: The Descriptor

After the draft is finalized, the framework extracts a `BotDescriptor` — a plain Go struct that describes everything the bot exposes. This descriptor is used for:
- Listing commands in `bots help`
- Syncing commands to Discord
- Validating that handlers exist before dispatching

**Source file:** `internal/jsdiscord/descriptor.go`

### Layer 4: Discord Sync and Dispatch

The host layer does two things:

1. **Command Sync:** Converts `BotDescriptor.Commands` into `discordgo.ApplicationCommand` objects and uploads them to Discord via the REST API.
   - **Source file:** `internal/jsdiscord/host_commands.go`

2. **Interaction Dispatch:** When Discord sends an interaction over the gateway, the host inspects it and calls the right JS handler.
   - **Source file:** `internal/jsdiscord/host_dispatch.go`

### Data Flow Diagram

```
Bot Author writes JS
       │
       ▼
+----------------------------------+
|  JS Runtime (goja)               |
|  defineBot() → botDraft          |
+----------------------------------+
       │
       ▼
+----------------------------------+
|  Descriptor extraction           |
|  BotDescriptor                   |
+----------------------------------+
       │
       ├──► bots help (CLI output)
       │
       ├──► Discord REST API
       │    (ApplicationCommand sync)
       │
       └──► Gateway websocket
            (Interaction dispatch)
```

---

## Gap Analysis: What is Missing

The framework currently only supports `CHAT_INPUT` slash commands. Here is the complete gap:

### 1. No User Context Menu Commands

Discord users right-click on a user and see "Apps". We cannot register handlers for these.

**Gap in `bot.go`:** No `userCommandDraft` or `userCommands` slice.  
**Gap in `runtime.go`:** No `userCommand()` API exposed to JS.  
**Gap in `descriptor.go`:** No way to distinguish user commands from chat commands.  
**Gap in `host_commands.go`:** `applicationCommandFromSnapshot` always creates `Type: 1` (ChatInput).  
**Gap in `host_dispatch.go`:** When `data.Type == UserApplicationCommand`, we do not populate the resolved user in the dispatch request.

### 2. No Message Context Menu Commands

Discord users right-click on a message and see "Apps". We cannot register handlers for these.

**Same gaps as user commands**, but for messages.

### 3. No Subcommands

We can register a root command like `/admin`, but we cannot register separate handlers for `/admin kick` and `/admin ban`. The framework treats the entire command as one handler.

**Gap in `bot.go`:** No `subcommandDraft` or `subcommands` slice.  
**Gap in `runtime.go`:** No `subcommand()` API exposed to JS.  
**Gap in `descriptor.go`:** No parsing of subcommand descriptors.  
**Gap in `host_commands.go`:** `optionTypeFromSpec` does not support `"sub_command"` or `"sub_command_group"`.  
**Gap in `host_dispatch.go`:** When `data.Options[0].Type == SubCommand`, we do not look for a subcommand handler; we just pass all options to the root handler.

---

## Design: How We Will Add the Missing Types

### Principle: Minimal API Surface

We want bot authors to write intuitive code. The additions are:

```js
// User context menu command (right-click user)
userCommand("Show Avatar", async (ctx) => {
  const targetUser = ctx.args.target
  return { content: `Avatar for ${targetUser.username}` }
})

// Message context menu command (right-click message)
messageCommand("Quote Message", async (ctx) => {
  const targetMessage = ctx.args.target
  return { content: `Quoted: ${targetMessage.content}` }
})

// Subcommand under a root
subcommand("admin", "kick", {
  description: "Kick a user",
  options: {
    user: { type: "user", required: true },
    reason: { type: "string" }
  }
}, async (ctx) => {
  return { content: `Kicked ${ctx.args.user} for ${ctx.args.reason}` }
})
```

### Design Decision: Unified Command List with Type Annotation

Instead of keeping separate descriptor lists for each command type, we will:

1. Add a `Type` field to `CommandDescriptor` (empty string defaults to `"chat_input"`)
2. Keep user and message commands in the same `Commands` slice
3. Add a new `Subcommands` slice to `BotDescriptor`

This keeps the descriptor simple while making it easy for `host_commands.go` to build the right Discord structures.

### Design Decision: How Subcommands are Dispatched

When Discord sends an interaction for `/admin kick user:@bob reason:spam`:

1. `data.Name` is `"admin"` (the root)
2. `data.Options` has one entry: `{ Name: "kick", Type: SubCommand, Options: [...] }`
3. The Go dispatch layer will:
   - Check if a subcommand handler exists for `("admin", "kick")`
   - If yes, call `DispatchSubcommand` with `rootName="admin"`, `subName="kick"`, and args flattened from the inner options
   - If no, fall back to the root command handler with the raw options

This lets bot authors choose:
- Use `subcommand("admin", "kick", ...)` for clean separation
- Or just use `command("admin", ...)` and branch on `ctx.args` manually

### Design Decision: Context Menu Target Access

For user and message commands, Discord includes a `TargetID` and `Resolved` data structure. We will extract the resolved target and place it in `ctx.args.target` so the JS handler can access it uniformly:

```js
userCommand("Show Avatar", async (ctx) => {
  // ctx.args.target is the full user object
  const user = ctx.args.target
  return { content: `${user.username}'s avatar: ...` }
})
```

---

## File-by-File Implementation Plan

### File 1: `internal/jsdiscord/bot.go`

**What to change:**

1. Add `commandType string` to `commandDraft` so we know if a command is chat_input, user, or message.
2. Add `subcommandDraft` struct with `rootName`, `name`, `spec`, `handler`.
3. Expand `botDraft` to hold:
   - `userCommands []*commandDraft`
   - `messageCommands []*commandDraft`
   - `subcommands []*subcommandDraft`
4. Add methods:
   - `userCommand(vm, call)` — validates args, appends to `userCommands`
   - `messageCommand(vm, call)` — validates args, appends to `messageCommands`
   - `subcommand(vm, call)` — validates args, appends to `subcommands`
5. Modify `command()` to read `spec["type"]` and route to the right slice.
6. Modify `finalize()` to:
   - Include all command types in the `commands` snapshot
   - Add a `dispatchSubcommand` callable to the bot object
7. Add `findSubcommand()` helper.
8. Update `BotHandle` struct and `CompileBot()` to extract `dispatchSubcommand`.
9. Add `DispatchSubcommand()` method on `BotHandle`.

**Pseudocode for `finalize().dispatchSubcommand`:**

```go
bot.Set("dispatchSubcommand", func(call goja.FunctionCall) goja.Value {
  input := objectFromValue(vm, call.Arguments[0])
  rootName := input.Get("rootName").String()
  subName := input.Get("subName").String()
  sub := findSubcommand(subcommands, rootName, subName)
  ctx := buildContext(vm, store, input, "subcommand", rootName+"/"+subName, metadata)
  return sub.handler(goja.Undefined(), ctx)
})
```

### File 2: `internal/jsdiscord/runtime.go`

**What to change:**

In `defineBot()`, add three new API methods to the builder object:

```go
_ = api.Set("userCommand", func(call goja.FunctionCall) goja.Value { return draft.userCommand(vm, call) })
_ = api.Set("messageCommand", func(call goja.FunctionCall) goja.Value { return draft.messageCommand(vm, call) })
_ = api.Set("subcommand", func(call goja.FunctionCall) goja.Value { return draft.subcommand(vm, call) })
```

### File 3: `internal/jsdiscord/descriptor.go`

**What to change:**

1. Add `Type string` to `CommandDescriptor`.
2. Add `SubcommandDescriptor` struct.
3. Add `Subcommands []SubcommandDescriptor` to `BotDescriptor`.
4. In `descriptorFromDescribe()`, parse `desc["subcommands"]`.
5. In `parseCommandDescriptors()`, read `spec["type"]` into `CommandDescriptor.Type`.
6. Add `parseSubcommandDescriptors()` function.

### File 4: `internal/jsdiscord/host_commands.go`

**What to change:**

1. In `applicationCommandFromSnapshot()`, read the command type:
   - `"user"` → `discordgo.UserApplicationCommand`
   - `"message"` → `discordgo.MessageApplicationCommand`
   - default → `discordgo.ChatApplicationCommand`
2. For user/message commands, do not include `Description` or `Options` in the Discord payload (Discord ignores them for context menu commands).
3. In `optionTypeFromSpec()`, add:
   - `"sub_command"` → `discordgo.ApplicationCommandOptionSubCommand`
   - `"sub_command_group"` → `discordgo.ApplicationCommandOptionSubCommandGroup`

### File 5: `internal/jsdiscord/host_dispatch.go`

**What to change:**

In `DispatchInteraction()`, case `discordgo.InteractionApplicationCommand`:

1. Check `data.Type`:
   - `ChatApplicationCommand`: existing behavior, but check for subcommands
   - `UserApplicationCommand`: find user command handler, populate `args.target` from `data.Resolved.Users[data.TargetID]`
   - `MessageApplicationCommand`: find message command handler, populate `args.target` from `data.Resolved.Messages[data.TargetID]`

2. For ChatApplicationCommand with subcommands:
   - If `len(data.Options) > 0 && data.Options[0].Type == SubCommand`:
     - `subName := data.Options[0].Name`
     - `args := optionMap(data.Options[0].Options)` (flatten only the subcommand's options)
     - Call `h.handle.DispatchSubcommand(ctx, DispatchRequest{...})` with `Name: rootName + "/" + subName`
   - Else: existing behavior

**Pseudocode for subcommand dispatch:**

```go
data := interaction.ApplicationCommandData()
switch data.Type {
case discordgo.ChatApplicationCommand:
  if len(data.Options) > 0 && data.Options[0].Type == discordgo.ApplicationCommandOptionSubCommand {
    subName := data.Options[0].Name
    args := optionMap(data.Options[0].Options)
    result, err := h.handle.DispatchSubcommand(ctx, DispatchRequest{
      Name: data.Name + "/" + subName,
      Args: args,
      // ... other fields
    })
  } else {
    // existing dispatch
  }
case discordgo.UserApplicationCommand:
  targetUser := data.Resolved.Users[data.TargetID]
  args := map[string]any{"target": userMap(targetUser)}
  result, err := h.handle.DispatchCommand(ctx, DispatchRequest{
    Name: data.Name,
    Args: args,
    // ... other fields
  })
case discordgo.MessageApplicationCommand:
  targetMessage := data.Resolved.Messages[data.TargetID]
  args := map[string]any{"target": messageMap(targetMessage)}
  result, err := h.handle.DispatchCommand(ctx, DispatchRequest{
    Name: data.Name,
    Args: args,
    // ... other fields
  })
}
```

---

## The Demo Bot

### Location

`examples/discord-bots/interaction-types/index.js`

### What it Demonstrates

The bot will register one example of each interaction type:

| Interaction Type | Command Name | What It Does |
|------------------|-------------|--------------|
| Simple slash command | `/hello` | Replies with a greeting |
| Slash command with options | `/echo` | Echoes back the provided text |
| Subcommand | `/admin kick` | Shows a kick confirmation |
| Subcommand | `/admin ban` | Shows a ban confirmation |
| User context menu | "Show Avatar" | Returns the target user's avatar URL |
| Message context menu | "Quote Message" | Quotes the target message |

### Bot Structure

```js
const { defineBot } = require("discord")

module.exports = defineBot(({ command, userCommand, messageCommand, subcommand, event, configure }) => {
  configure({
    name: "interaction-types",
    description: "Demo of all Discord application command interaction types",
    category: "examples"
  })

  // Simple slash command
  command("hello", {
    description: "Say hello"
  }, async () => {
    return { content: "Hello from the interaction types bot!" }
  })

  // Slash command with options
  command("echo", {
    description: "Echo text back",
    options: {
      text: { type: "string", description: "Text to echo", required: true }
    }
  }, async (ctx) => {
    return { content: ctx.args.text }
  })

  // Subcommands under /admin
  subcommand("admin", "kick", {
    description: "Kick a user",
    options: {
      user: { type: "user", description: "User to kick", required: true },
      reason: { type: "string", description: "Reason for kick" }
    }
  }, async (ctx) => {
    return { content: `Would kick ${ctx.args.user} for: ${ctx.args.reason || "no reason"}` }
  })

  subcommand("admin", "ban", {
    description: "Ban a user",
    options: {
      user: { type: "user", description: "User to ban", required: true },
      duration: { type: "integer", description: "Ban duration in days" }
    }
  }, async (ctx) => {
    return { content: `Would ban ${ctx.args.user} for ${ctx.args.duration || 0} days` }
  })

  // Also register the root /admin command so Discord knows it exists
  command("admin", {
    description: "Administration commands",
    options: {
      kick: { type: "sub_command", options: { ... } },
      ban: { type: "sub_command", options: { ... } }
    }
  }, async (ctx) => {
    // Fallback handler if someone invokes /admin without a subcommand
    return { content: "Please use /admin kick or /admin ban" }
  })

  // User context menu command
  userCommand("Show Avatar", async (ctx) => {
    const user = ctx.args.target
    return { content: `${user.username}'s avatar: https://cdn.discordapp.com/avatars/${user.id}/${user.avatar}.png` }
  })

  // Message context menu command
  messageCommand("Quote Message", async (ctx) => {
    const msg = ctx.args.target
    return { content: `> ${msg.content}\n— ${msg.author.username}` }
  })

  event("ready", async (ctx) => {
    ctx.log.info("interaction-types bot ready")
  })
})
```

### Note on Subcommand Registration

The current framework requires that the root command (`/admin`) also be registered via `command()` so that Discord receives the full application command definition with subcommand options. The `subcommand()` registrations only create handler mappings on the Go side; they do not by themselves tell Discord what the command structure looks like. Therefore, the bot author must:

1. Register the root command with `command("admin", { options: { kick: { type: "sub_command", ... } } })`
2. Register handlers with `subcommand("admin", "kick", ..., handler)`

This is a deliberate design choice: it keeps the JS API simple and matches how Discord represents commands internally.

---

## Testing Strategy

### Unit Tests

Add tests in `internal/jsdiscord/runtime_test.go` or a new test file:

1. **Descriptor parsing:** Load a script that uses `userCommand`, `messageCommand`, and `subcommand`, then assert the descriptor has the right types and names.
2. **Command sync:** Verify `applicationCommandFromSnapshot` produces the correct `discordgo.ApplicationCommand` with `Type` set properly.
3. **Dispatch routing:** Mock a `discordgo.InteractionCreate` for each command type and verify the right JS handler is called with the right args.

### Manual Testing

1. Run the bot with `--sync-on-start` and `--guild-id`:
   ```bash
   go run ./cmd/discord-bot bots run interaction-types \
     --bot-repository ./examples/discord-bots \
     --bot-token "$DISCORD_BOT_TOKEN" \
     --application-id "$DISCORD_APPLICATION_ID" \
     --guild-id "$DISCORD_GUILD_ID" \
     --sync-on-start
   ```
2. In Discord:
   - Type `/hello` → expect greeting
   - Type `/echo text:hi` → expect "hi"
   - Type `/admin kick user:@someone reason:test` → expect kick confirmation
   - Type `/admin ban user:@someone duration:7` → expect ban confirmation
   - Right-click a user → Apps → "Show Avatar" → expect avatar URL
   - Right-click a message → Apps → "Quote Message" → expect quoted text

---

## Reference Material

### Key Source Files

| File | Purpose |
|------|---------|
| `internal/jsdiscord/bot.go` | JS draft types, handler dispatch, context building |
| `internal/jsdiscord/runtime.go` | `defineBot()` and API registration |
| `internal/jsdiscord/descriptor.go` | Descriptor parsing from JS describe output |
| `internal/jsdiscord/host_commands.go` | Discord ApplicationCommand building |
| `internal/jsdiscord/host_dispatch.go` | Gateway interaction routing |
| `internal/jsdiscord/host_maps.go` | Discord object → plain map conversion |
| `pkg/doc/topics/discord-js-bot-api-reference.md` | Existing API docs |
| `pkg/doc/tutorials/building-and-running-discord-js-bots.md` | Existing tutorial |

### Discord API Reference (Downloaded)

Stored at:
`ttmp/2026/04/21/DISCORD-BOT-020--.../sources/discord-application-commands-docs.md`

Key facts from that document:
- Application commands have `name`, `description`, `type`, `options`
- `type` is 1 (Chat Input), 2 (User), or 3 (Message)
- User and Message commands do not support `description` or `options`
- Subcommands are options of type `1` (`SUB_COMMAND`)
- Subcommand groups are options of type `2` (`SUB_COMMAND_GROUP`)

### Existing Example Bots

| Bot | What to Learn From It |
|-----|----------------------|
| `examples/discord-bots/ping/index.js` | Buttons, selects, modals, autocomplete, deferred replies |
| `examples/discord-bots/moderation/index.js` | Splitting a large bot into `lib/` modules |
| `examples/discord-bots/knowledge-base/index.js` | Runtime config via `configure({ run: ... })` |

---

## Appendix: Data Shapes

### JS Context for User Commands

```js
ctx.args.target = {
  id: "123456789",
  username: "alice",
  avatar: "abc123...",
  // ... other user fields
}
```

### JS Context for Message Commands

```js
ctx.args.target = {
  id: "987654321",
  content: "Hello world",
  author: { id: "123456789", username: "alice" },
  channelID: "...",
  guildID: "..."
}
```

### JS Context for Subcommands

```js
// For /admin kick user:@bob reason:spam
ctx.args = {
  user: "123456789",   // resolved user ID
  reason: "spam"
}
```

Note: Discord resolves `user` type options to user IDs as strings. If you need the full user object, you can fetch it via `ctx.discord.members.fetch(ctx.guild.id, ctx.args.user)`.

---

## Implementation Log

This section records what was actually built, how it differed from the plan, and what lessons we learned.

### What was built

1. **JS runtime API additions** (`internal/jsdiscord/bot.go`, `internal/jsdiscord/runtime.go`)
   - `userCommand(name, handler)` — 2-argument registration, no spec
   - `messageCommand(name, handler)` — 2-argument registration, no spec
   - `subcommand(rootName, name, spec?, handler)` — 3 or 4 argument registration
   - `command(name, spec?, handler)` — now reads `spec.type` to route to user/message/chat slices
   - `dispatchSubcommand` callable added to the compiled bot object

2. **Descriptor parsing** (`internal/jsdiscord/descriptor.go`)
   - `CommandDescriptor.Type` field added
   - `SubcommandDescriptor` struct added
   - `BotDescriptor.Subcommands` slice added
   - `parseSubcommandDescriptors()` function added

3. **Discord command sync** (`internal/jsdiscord/host_commands.go`)
   - `applicationCommandFromSnapshot()` reads `"user"` / `"message"` type and returns the correct `discordgo.ApplicationCommandType`
   - User/message commands omit `Description` and `Options` (Discord ignores them)
   - `optionTypeFromSpec()` supports `"sub_command"` and `"sub_command_group"`

4. **Interaction dispatch** (`internal/jsdiscord/host_dispatch.go`)
   - `InteractionApplicationCommand` case split into three sub-cases by `data.CommandType`
   - User commands: extract resolved user from `data.Resolved.Users[data.TargetID]` into `args.target`
   - Message commands: extract resolved message from `data.Resolved.Messages[data.TargetID]` into `args.target`
   - Chat input with subcommands: detect `SubCommand` option type, flatten inner options, call `DispatchSubcommand()`
   - Chat input without subcommands: existing behavior

5. **Demo bot** (`examples/discord-bots/interaction-types/index.js`)
   - `/hello` — simple slash command
   - `/echo text:` — slash command with string option
   - `/admin kick user: reason:` and `/admin ban user: duration:` — subcommands
   - "Show Avatar" user context menu command
   - "Quote Message" message context menu command
   - `ready` event for logging

6. **Documentation updates**
   - `examples/discord-bots/README.md` — lists `interaction-types/` bot
   - `pkg/doc/topics/discord-js-bot-api-reference.md` — documents `userCommand`, `messageCommand`, `subcommand`

### Deviations from the plan

| Planned | Actual | Reason |
|---------|--------|--------|
| Store subcommands in a separate descriptor list only | Also include them in the JS `describe()` output alongside commands | Discord sync only needs the root command, but the descriptor keeps subcommands for validation and future `bots help` support |
| Pass `rootName`/`subName` via custom JS object fields | Added `RootName` and `SubName` to `DispatchRequest` and `buildDispatchInput()` | Cleaner integration with existing dispatch pipeline |
| `data.Type` for command type check | `data.CommandType` | `ApplicationCommandInteractionData` has a `Type()` method (returns `InteractionType`), not a `Type` field. The actual field is `CommandType` |

### Lessons learned

- **Verify compilation after every file change.** My first edit to `bot.go` struct fields failed silently because the `oldText` didn't match exactly. Running `go build` immediately would have caught it.
- **DiscordGo struct fields vs methods matter.** `data.Type` resolved to a method returning `InteractionType`, while `data.CommandType` is the actual `ApplicationCommandType` field. The compiler error was confusing because it mentioned `func() InteractionType`.
- **Context menu commands need resolved data.** The target user/message is NOT in the top-level interaction; it's in `data.Resolved`. Must nil-check before accessing.
- **Root command + subcommand handler pattern requires both registrations.** Bot authors must register the root command with `command(...)` AND the handler with `subcommand(...)`. This is deliberate — it matches Discord's representation and keeps the API simple.

### Commits

| Hash | Message |
|------|---------|
| `08b92cb` | DISCORD-BOT-020: Add userCommand, messageCommand, and subcommand to JS bot API runtime |
| `09fc236` | DISCORD-BOT-020: Update command sync and interaction dispatch for user/message/subcommand types |
| `e16bda2` | DISCORD-BOT-020: Create interaction-types demo bot with all command variants |
| `bfdc56d` | DISCORD-BOT-020: Update docs for userCommand, messageCommand, and subcommand APIs |
