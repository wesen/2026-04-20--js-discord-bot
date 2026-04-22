---
Title: "Discord JavaScript Bot API Reference"
Slug: "discord-js-bot-api-reference"
Short: "Reference for the JavaScript bot DSL, handler contexts, payload shapes, and outbound Discord operations."
Topics:
- discord
- javascript
- bots
- runtime
- components
- modals
- autocomplete
- commands
Commands:
- bots help
- bots list
- bots run
Flags:
- bot-repository
- bot-token
- application-id
- guild-id
- sync-on-start
- print-parsed-values
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

## What this API is for

This repository lets you write Discord bots in JavaScript while the Go host handles Discord connectivity, slash-command sync, event dispatch, and outbound operations. The JavaScript side stays small and expressive:

- you declare one bot with `defineBot(...)`
- you register commands, events, components, modals, and autocomplete handlers
- you use the provided context object to reply, defer, edit, log, persist small state, and call Discord operations

The main idea is simple: the bot repository owns the process, but the bot behavior lives in JavaScript.

> ⚠️ **Runtime Environment**
> Bot scripts run inside a Goja JavaScript engine embedded in Go, **not Node.js**.
> - **Available modules:** `require("discord")`, `require("timer")`, `require("database")`, `require("ui")`
> - **Unavailable:** `fs`, `path`, `http`, `fetch`, `process`, npm packages, or any Node.js standard library
> - **No file system access from JS.** Deliver generated content as Discord file attachments via `ctx.discord.channels.send()` with `files: [...]`

## The runtime model

A bot repository is a directory tree full of bot scripts. Each script exports one bot definition through `require("discord")`.

A typical repository looks like this:

```text
examples/discord-bots/
  ping/index.js
  poker/index.js
  knowledge-base/index.js
  support/index.js
  moderation/index.js
```

The CLI discovers bots by scanning the repository, loading each script, and reading its metadata. That is why the bot name, description, commands, components, modals, and autocomplete registrations all matter: they are not just documentation, they are the contract the host uses to route interactions.

## Quick API summary

| Helper | Purpose |
| --- | --- |
| `defineBot(builderFn)` | Create one bot from a builder callback |
| `configure(options)` | Set bot metadata and runtime config fields |
| `command(name, spec?, handler)` | Register a slash command |
| `userCommand(name, handler)` | Register a user context menu command |
| `messageCommand(name, handler)` | Register a message context menu command |
| `subcommand(rootName, name, spec?, handler)` | Register a subcommand handler |
| `event(name, handler)` | Register a gateway/event handler |
| `component(customId, handler)` | Handle button or select-menu interactions |
| `modal(customId, handler)` | Handle modal submissions |
| `autocomplete(commandName, optionName, handler)` | Return autocomplete choices |
| `require("ui")` | Go-side fluent builders for Discord messages, embeds, components, modals, and screen helpers |

## `require("ui")` — Go-side UI DSL builders

`require("ui")` exposes a native builder DSL implemented in Go and surfaced through Goja Proxy traps. This module exists for cases where you want JavaScript to keep the fluent authoring experience, but you want the host to own validation, type checks, and response-shape control.

The most important design rule is that `ui` returns **typed builders**, not plain JS objects. Call `.build()` at the end of a chain to get a typed Discord payload or the host's `normalizedResponse` fast path.

### Available builders and helpers

| Helper | Purpose |
| --- | --- |
| `ui.message()` | Build a message / interaction response payload |
| `ui.embed(title?)` | Build an embed payload |
| `ui.card(title?)` | Embed-style helper with `.meta()` for inline key/value fields |
| `ui.button(customId, label, style?)` | Build a button component |
| `ui.select(customId)` | Build a string select menu |
| `ui.userSelect(customId)` | Build a user select menu |
| `ui.roleSelect(customId)` | Build a role select menu |
| `ui.channelSelect(customId)` | Build a channel select menu |
| `ui.mentionableSelect(customId)` | Build a mentionable select menu |
| `ui.form(customId, title)` | Build a modal form payload |
| `ui.row(...components)` | Wrap components in an action row |
| `ui.pager(prevId, nextId, controls?)` | Build previous/next pager buttons |
| `ui.actions(definitions)` | Build a button row from `{id, label, style}` definitions |
| `ui.confirm(confirmId, cancelId, options?)` | Build a confirmation response with buttons |
| `ui.ok(content)` | Convenience ephemeral success response |
| `ui.error(content)` | Convenience ephemeral error response |
| `ui.emptyResults(query?)` | Convenience empty-results response |
| `ui.flow(namespace, options?)` | Generate stable component IDs and flow helpers |

### Usage model

```js
const ui = require("ui")

return ui.message()
  .content("Search results")
  .embed(ui.embed("Results").description("Found 3 items"))
  .row(ui.button("search:next", "Next", "primary"))
  .build()
```

A few rules matter here:

- **Wrong-parent calls should fail loudly.** `ui.message().field(...)` is an error, and the error should tell you that you probably meant `ui.embed()`.
- **Raw JS objects are not accepted where builders are expected.** Pass a `ui.embed()` builder to `message.embed(...)`, not a plain object.
- **Component interactions default to in-place updates.** A component click should usually edit the original message instead of creating a new one.
- **Use `.followUp()` when you explicitly want a new message.** The DSL allows the host to distinguish between update-in-place and follow-up responses.

### Modal fields

Modal fields are keyed by `customId`, so the form builder uses a `customId`-first API:

```js
ui.form("feedback:submit", "Feedback")
  .text("title", "Title")
  .textarea("feedback", "Your feedback")
  .build()
```

On submit, those `customId` values become the keys in `ctx.values`.

### Screen helpers and state

The `ui` module also includes helpers for stable screen behavior:

- `ui.flow(namespace)` generates deterministic component IDs like `namespace:next`
- `ui.pager(...)` builds pager button rows
- `ui.actions(...)` turns small definitions into button rows
- `ui.card(...)` is a convenient embed variant for browsable items

In the showcase bot, these helpers drive search screens, card galleries, and review flows that update in place instead of spamming the channel with new messages.

## `defineBot(builderFn)`

`defineBot(...)` is the entrypoint every bot script exports.

```js
const { defineBot } = require("discord")

module.exports = defineBot(({ command, event, configure }) => {
  configure({
    name: "ping",
    description: "A minimal example bot",
  })

  command("ping", { description: "Reply with pong" }, async () => {
    return { content: "pong" }
  })
})
```

The builder callback receives the registration helpers you ask for. If you only destructure `command` and `event`, those are the only hooks you plan to use. Most bots will ask for at least `command`, `event`, and `configure`.

## `configure(options)`

`configure(...)` stores metadata on the bot and can also describe runtime config fields.

```js
configure({
  name: "knowledge-base",
  description: "Search and summarize internal docs from JavaScript",
  category: "knowledge",
  run: {
    fields: {
      index_path: {
        type: "string",
        help: "Optional path label for the active docs index",
        default: "builtin-docs",
      },
      read_only: {
        type: "bool",
        help: "Disable write operations for future mutations",
        default: true,
      },
    },
  },
})
```

### Metadata keys

The host keeps arbitrary metadata keys. The common ones are:

- `name` — the canonical bot name shown by `bots list` and `bots help`
- `description` — a human-readable summary
- `category` — a group label used by your own conventions
- `run` — runtime configuration schema for `bots run`

### Runtime config fields

Each field under `run.fields` becomes a CLI flag when you run the bot through `bots run <bot>`.

Rules to remember:

- field names become the keys inside `ctx.config`
- field names are also converted into kebab-case CLI flags
- the help text from `help` is shown in `bots help <bot>`
- default values are exposed to the CLI and to `ctx.config`

For example, the field `index_path` becomes the flag `--index-path` and is read in JavaScript as `ctx.config.index_path`.

Supported runtime field types are:

- `string`
- `bool` / `boolean`
- `int` / `integer`
- `number` / `float`
- `string_list` / `string-list` / `string[]`

## `command(name, spec?, handler)`

`command(...)` registers a slash command.

```js
command("echo", {
  description: "Echo text back",
  options: {
    text: {
      type: "string",
      description: "Text to echo back",
      required: true,
    },
  },
}, async (ctx) => {
  return {
    content: ctx.args.text,
    ephemeral: true,
  }
})
```

### Command specification

The runtime reads the `spec` object and converts it into a Discord application command.

The keys that matter are:

- `description` — the command description shown by Discord
- `options` — a map or array of option definitions

Use an object map when you want simple, readable declarations. Use an array if you want to preserve an explicit option list.

```js
options: {
  query: {
    type: "string",
    description: "Search query",
    required: true,
    autocomplete: true,
  },
  limit: {
    type: "integer",
    description: "Maximum results",
  },
}
```

### Supported option types

The framework supports Discord's application command option types with these JavaScript-friendly names:

| Framework type | Discord API type | Value | What the user sees |
| --- | --- | --- | --- |
| `string` | `STRING` | 3 | Free text input |
| `int`, `integer` | `INTEGER` | 4 | Whole number input |
| `bool`, `boolean` | `BOOLEAN` | 5 | True/false toggle |
| `number`, `float` | `NUMBER` | 10 | Decimal number input |
| `user` | `USER` | 6 | @mention picker for users |
| `channel` | `CHANNEL` | 7 | #channel picker |
| `role` | `ROLE` | 8 | @role picker |
| `mentionable` | `MENTIONABLE` | 9 | User or role picker |
| `sub_command` | `SUB_COMMAND` | 1 | Nested command |
| `sub_command_group` | `SUB_COMMAND_GROUP` | 2 | Nested command group |

**Not yet supported:** `attachment` (Discord type 11, value 11) — file upload as a command option.

**Does not exist in Discord's API:** There are no date, datetime, calendar, time-range, or date-picker option types. If you need time-based selection, use a `string` (ISO date), `integer` (relative hours), or a message ID as an anchor point.

### Supported option fields

| Field | Meaning | Valid for types |
| --- | --- | --- |
| `type` | Option type (see table above) | all |
| `description` | Option description shown in Discord | all |
| `required` | Marks the option as required | all except `sub_command` / `sub_command_group` |
| `autocomplete` | Enables autocomplete for that option | `string`, `integer`, `number` |
| `choices` | Static choices for the option | `string`, `integer`, `number` |
| `minLength` | Minimum length for string options | `string` |
| `maxLength` | Maximum length for string options | `string` |
| `minValue` | Minimum numeric value | `integer`, `number` |
| `maxValue` | Maximum numeric value | `integer`, `number` |
| `channel_types` | Restrict channel picker to specific types | `channel` |

Important rules:

- `autocomplete: true` cannot be combined with static `choices`
- required options should be defined before optional ones in your source data
- the host preserves a sensible order when it syncs commands to Discord

### What the command handler gets

The handler receives a context object. For slash commands, the most important fields are:

- `ctx.args` — parsed option values keyed by option name
- `ctx.options` — alias of `ctx.args`
- `ctx.config` — runtime config values from `configure({ run: ... })`
- `ctx.reply(...)` — send the initial response
- `ctx.defer(...)` — acknowledge the interaction and finish later
- `ctx.edit(...)` — edit the deferred or initial response
- `ctx.followUp(...)` — send an additional follow-up
- `ctx.showModal(...)` — open a modal
- `ctx.discord` — call Discord operations directly
- `ctx.log` — structured logger for this bot context
- `ctx.store` — per-runtime in-memory state

If your command does work that might take longer than Discord likes, call `await ctx.defer({ ephemeral: true })`, do the work, and then `await ctx.edit(...)` with the result.

## `userCommand(name, handler)`

`userCommand(...)` registers a user context menu command. Users see it when they right-click a user and choose Apps.

```js
userCommand("Show Avatar", async (ctx) => {
  const target = ctx.args.target
  if (!target || !target.id) {
    return { content: "Could not resolve user", ephemeral: true }
  }
  const avatarUrl = target.avatar
    ? `https://cdn.discordapp.com/avatars/${target.id}/${target.avatar}.png?size=512`
    : `https://cdn.discordapp.com/embed/avatars/${(parseInt(target.discriminator) || 0) % 5}.png`
  return { content: `${target.username}'s avatar: ${avatarUrl}`, ephemeral: true }
})
```

User commands do not accept a `spec` or options. Discord ignores description and options for context menu commands.

### User command context fields

- `ctx.args.target` — the resolved user object: `{ id, username, avatar, discriminator, bot }`
- All standard context fields (`reply`, `defer`, `edit`, `discord`, `log`, etc.) are also available

## `messageCommand(name, handler)`

`messageCommand(...)` registers a message context menu command. Users see it when they right-click a message and choose Apps.

```js
messageCommand("Quote Message", async (ctx) => {
  const target = ctx.args.target
  const author = target.author && target.author.username || "unknown"
  const content = target.content || "(empty)"
  return { content: `> ${content}\n— **${author}**`, ephemeral: true }
})
```

Message commands do not accept a `spec` or options.

### Message command context fields

- `ctx.args.target` — the resolved message object: `{ id, content, author, guildID, channelID }`
- All standard context fields are also available

## `subcommand(rootName, name, spec?, handler)`

`subcommand(...)` registers a handler for a subcommand under a root slash command. Discord represents subcommands as options of type `SUB_COMMAND` on the root command.

```js
subcommand("fun", "roll", {
  description: "Roll a die"
}, async (ctx) => {
  const sides = ctx.args.sides || 6
  const result = Math.floor(Math.random() * sides) + 1
  return { content: `🎲 You rolled a **${result}** (1-${sides})`, ephemeral: true }
})
```

### How subcommands work with the root command

You must also register the root command with `command(...)` so Discord knows the full command structure. The root command's spec should declare the subcommand options:

```js
command("fun", {
  description: "Fun and games",
  options: {
    roll: {
      type: "sub_command",
      description: "Roll a die",
      options: {
        sides: { type: "integer", description: "Number of sides", default: 6 }
      }
    }
  }
}, async (ctx) => {
  // Fallback handler when /fun is called without a subcommand
  return { content: "Use /fun roll or /fun coin" }
})
```

The `subcommand(...)` registration only creates the handler mapping. The `command(...)` registration tells Discord what the command looks like.

### Subcommand context fields

- `ctx.args` — parsed option values from the subcommand's own options
- `ctx.command.name` — the root command name
- `ctx.command.subName` — the subcommand name
- All standard context fields are also available

## `event(name, handler)`

`event(...)` registers a handler for a gateway event.

The runtime currently exposes these event names:

- `ready`
- `guildCreate`
- `guildMemberAdd`
- `guildMemberUpdate`
- `guildMemberRemove`
- `messageCreate`
- `messageUpdate`
- `messageDelete`
- `reactionAdd`
- `reactionRemove`

Example:

```js
event("messageCreate", async (ctx) => {
  const content = String((ctx.message && ctx.message.content) || "").trim()
  if (content === "!pingjs") {
    await ctx.reply({ content: "pong from messageCreate" })
  }
})
```

### Event context fields

The context depends on the event, but these fields are common:

- `ctx.message` — the current message or message-like payload
- `ctx.before` — the previous value for update/delete style events
- `ctx.user` — the user involved in the event
- `ctx.member` — the guild member involved in the event
- `ctx.guild` — the guild context when available
- `ctx.channel` — the channel context when available
- `ctx.reaction` — reaction payload for reaction events
- `ctx.me` — the bot’s current user record
- `ctx.discord` — outbound Discord operations
- `ctx.reply(...)` — send a channel response when the event supports it

For update/delete events, `ctx.before` is especially useful because it gives you the previous state. For reaction events, `ctx.reaction.emoji.name` is the easiest way to inspect what was added or removed.

### What the event handler may return

If you do not call `ctx.reply(...)`, the runtime can use the handler’s return value as a response for event-style workflows. In practice, many bots use events for side effects and explicit replies rather than relying on return values, because that keeps the flow easier to follow.

## `component(customId, handler)`

`component(...)` handles button clicks and select-menu interactions by `customId`.

```js
component("ping:panel", async () => {
  return {
    content: "Panel button clicked from JavaScript",
    ephemeral: true,
  }
})
```

Use it when your command returns a message with buttons or selects and you want those controls to be interactive.

### Component context fields

The most useful fields are:

- `ctx.component.customId` — the clicked component ID
- `ctx.component.type` — `button`, `select`, `userSelect`, `roleSelect`, `mentionableSelect`, or `channelSelect`
- `ctx.values` — selected values for select menus
- `ctx.reply(...)` — respond to the component interaction
- `ctx.defer(...)` — acknowledge the click when you need more time
- `ctx.edit(...)` — edit the original interaction response
- `ctx.showModal(...)` — open a modal from a component click

Buttons do not carry values. Select menus do. For a single-select menu, `ctx.values` is an array that usually contains one string. For multi-select menus, it can contain more.

## `modal(customId, handler)`

`modal(...)` handles modal submissions by `customId`.

```js
modal("feedback:submit", async (ctx) => {
  const summary = ctx.values && ctx.values.summary || "(empty)"
  const details = ctx.values && ctx.values.details || "(none)"
  return {
    content: `Thanks for the feedback: ${summary}\nDetails: ${details}`,
    ephemeral: true,
  }
})
```

### Modal context fields

The key field is:

- `ctx.values` — an object keyed by text input custom IDs

For text inputs, the runtime collects the submitted values into a plain object so you can read them directly by name.

## `autocomplete(commandName, optionName, handler)`

`autocomplete(...)` supplies suggestions while the user is typing a slash-command option.

```js
autocomplete("search", "query", async (ctx) => {
  const current = String(ctx.focused && ctx.focused.value || "")
  return [
    { name: "Architecture", value: "architecture" },
    { name: "Testing", value: "testing" },
    { name: `Custom: ${current || "query"}`, value: current || "custom" },
  ]
})
```

### Autocomplete context fields

- `ctx.focused` — the currently focused option
- `ctx.args` — the parsed option map for the command
- `ctx.command` — the command metadata
- `ctx.config` — runtime config values
- `ctx.discord` — outbound Discord operations

### Autocomplete return values

Return an array of choices. Each choice should have:

- `name`
- `value`

The runtime will keep at most 25 choices, because that is Discord’s limit.

## Response payloads

Most handlers can return either a string or an object. Returning a string becomes the message content. Returning an object lets you build richer replies.

Common response fields are:

- `content`
- `embeds`
- `components`
- `files`
- `allowedMentions`
- `tts`
- `ephemeral`
- `replyTo`

Example:

```js
return {
  content: "Search results",
  embeds: [
    {
      title: "Architecture",
      description: "Bot wiring, handlers, and runtime layers.",
      color: 0x5865F2,
    },
  ],
  ephemeral: true,
}
```

### Deferred response pattern

For slow work, use the defer/edit pattern:

```js
await ctx.defer({ ephemeral: true })
await ctx.edit({ content: "Searching for arch..." })
await sleep(2000)
await ctx.edit({ content: "Results for arch: ..." })
```

This is the right pattern for commands that need to perform a search, call an API, or otherwise wait before replying.

## `ctx.store`

`ctx.store` is a per-runtime in-memory key/value store.

It is useful for lightweight bot state such as counters, current hand state, or temporary caches.

```js
const hits = ctx.store.get("hits", 0)
ctx.store.set("hits", hits + 1)
```

Available methods:

- `get(key, defaultValue)`
- `set(key, value)`
- `delete(key)`
- `keys(prefix)`
- `namespace(...parts)`

Important behavior:

- state lives only in memory
- it is lost when the bot process restarts
- use `namespace(...)` to keep per-bot or per-channel state separate

## `ctx.log`

`ctx.log` is a structured logger for the current bot context.

```js
ctx.log.info("js discord bot connected", {
  user: ctx.me && ctx.me.username,
  script: ctx.metadata && ctx.metadata.scriptPath,
})
```

Available levels:

- `info`
- `debug`
- `warn`
- `error`

Any field object you pass is merged into the structured log output.

## `ctx.discord`

`ctx.discord` exposes outbound Discord operations for when a command or event needs to do more than answer the original interaction. The same namespace is available from command, event, component, modal, and autocomplete handlers.

### Discord operations by namespace

#### `ctx.discord.guilds`

| Operation | Purpose |
| --- | --- |
| `ctx.discord.guilds.fetch(guildId)` | Fetch a guild snapshot |

#### `ctx.discord.roles`

| Operation | Purpose |
| --- | --- |
| `ctx.discord.roles.list(guildId)` | List roles in a guild |
| `ctx.discord.roles.fetch(guildId, roleId)` | Fetch one role by ID |

#### `ctx.discord.threads`

| Operation | Purpose |
| --- | --- |
| `ctx.discord.threads.fetch(threadId)` | Fetch a thread snapshot |
| `ctx.discord.threads.join(threadId)` | Join a thread |
| `ctx.discord.threads.leave(threadId)` | Leave a thread |
| `ctx.discord.threads.start(channelId, payload)` | Start a thread from a channel or source message |

#### `ctx.discord.channels`

| Operation | Purpose |
| --- | --- |
| `ctx.discord.channels.send(channelId, payload)` | Send a normal message to a channel |
| `ctx.discord.channels.fetch(channelId)` | Fetch a channel snapshot |
| `ctx.discord.channels.setTopic(channelId, topic)` | Update a channel topic |
| `ctx.discord.channels.setSlowmode(channelId, seconds)` | Update channel slowmode |

#### `ctx.discord.messages`

| Operation | Purpose |
| --- | --- |
| `ctx.discord.messages.fetch(channelId, messageId)` | Fetch one channel message |
| `ctx.discord.messages.list(channelId, payload)` | List recent channel messages with `before`, `after`, `around`, and `limit` options |
| `ctx.discord.messages.edit(channelId, messageId, payload)` | Edit an existing channel message |
| `ctx.discord.messages.delete(channelId, messageId)` | Delete a channel message |
| `ctx.discord.messages.react(channelId, messageId, emoji)` | Add a reaction to a channel message |
| `ctx.discord.messages.pin(channelId, messageId)` | Pin a message |
| `ctx.discord.messages.unpin(channelId, messageId)` | Unpin a message |
| `ctx.discord.messages.listPinned(channelId)` | List pinned messages in a channel |
| `ctx.discord.messages.bulkDelete(channelId, messageIds)` | Bulk-delete a list of messages |

#### `ctx.discord.members`

| Operation | Purpose |
| --- | --- |
| `ctx.discord.members.fetch(guildId, userId)` | Fetch one guild member |
| `ctx.discord.members.list(guildId, payload)` | List guild members with pagination options |
| `ctx.discord.members.addRole(guildId, userId, roleId)` | Add a role to a member |
| `ctx.discord.members.removeRole(guildId, userId, roleId)` | Remove a role from a member |
| `ctx.discord.members.timeout(guildId, userId, payload)` | Set or clear a member timeout |
| `ctx.discord.members.kick(guildId, userId, payload)` | Kick a member |
| `ctx.discord.members.ban(guildId, userId, payload)` | Ban a member |
| `ctx.discord.members.unban(guildId, userId)` | Remove a ban |

### List and action payloads

The list helpers accept plain objects so you can pass the same shape from a command, a modal, or a helper function.

- `ctx.discord.messages.list(channelId, { before, after, around, limit })`
  - choose at most one of `before`, `after`, or `around`
  - `limit` defaults to 25 and is capped at 100
- `ctx.discord.members.list(guildId, { after, limit })`
  - `limit` defaults to 25 and is capped at 1000
- `ctx.discord.messages.bulkDelete(channelId, messageIds)`
  - accepts an array of IDs or an object like `{ messageIds: [...] }`
  - duplicate IDs are ignored
- `ctx.discord.members.timeout(guildId, userId, payload)`
  - accepts `{ durationSeconds }`, `{ until }`, or `{ clear: true }`
- `ctx.discord.members.ban(guildId, userId, payload)`
  - accepts a string reason or an object like `{ reason, deleteMessageDays }`

### Message history pagination example

To fetch more messages than a single call allows, paginate backwards with `before`:

```js
async function fetchAllMessages(ctx, channelId, maxMessages) {
  const allMessages = []
  let lastMessageId = null
  const pageSize = 100 // Discord max per request

  while (true) {
    const options = { limit: pageSize }
    if (lastMessageId) {
      options.before = lastMessageId
    }

    const batch = await ctx.discord.messages.list(channelId, options)
    if (!batch || batch.length === 0) {
      break
    }

    allMessages.push(...batch)
    lastMessageId = batch[batch.length - 1].id

    if (maxMessages && allMessages.length >= maxMessages) {
      allMessages.splice(maxMessages)
      break
    }
  }

  // Discord returns newest first; reverse to chronological order
  return allMessages.reverse()
}
```

Use this pattern when building archive, export, or backfill commands that need the full history of a channel or thread.

### What the returned snapshots look like

The fetch/list helpers return plain JavaScript objects and arrays of objects. They are intentionally easy to inspect and serialize. Common fields include:

- guild snapshots: `id`, `name`, `ownerID`, `description`, `memberCount`, `features`
- role snapshots: `id`, `guildID`, `name`, `position`, `permissions`, `mentionable`, `hoist`
- thread snapshots: `id`, `parentID`, `name`, `type`, `archived`, `locked`, `autoArchiveDuration`
- channel snapshots: `id`, `guildID`, `name`, `type`, `topic`, `position`, `rateLimitPerUser`
- message snapshots: `id`, `content`, `guildID`, `channelID`, `author`
- member snapshots: `id`, `guildId`, `nick`, `roles`, `pending`, `joinedAt`

### Common payload shapes

Channel send and interaction response payloads support a shared shape:

- `content` — message text
- `embeds` — one or more embeds
- `components` — buttons and select menus
- `files` — attachments like `{ name, content, contentType }`
- `allowedMentions` — mention policy
- `tts` — text-to-speech flag
- `replyTo` — message reference for channel messages

For `ctx.discord.channels.send(...)`, the payload is a normal message payload. For interaction responses, you can also use `ephemeral: true` to keep the response private.

### Files and replies

A file attachment looks like this:

```js
files: [
  {
    name: "report.txt",
    content: "This report was created inside the JS bot runtime.",
  },
]
```

A reply reference looks like this:

```js
replyTo: {
  messageId: "orig-1",
  channelId: "chan-1",
}
```

## Supporting utility modules

The Discord bot runtime also exposes a small `timer` module that is useful for demos and deferred workflows.

```js
const { sleep } = require("timer")
await sleep(2000)
```

Use this for simulated search, background work, or tests that need a visible pause. It is not a Discord API, but it is part of the runtime environment that the example bots use.

### `require("database")`

`require("database")` is the runtime module for durable data access from JavaScript bots. Use it when the state must survive restarts or when the bot owns real application data such as knowledge entries, reviews, or long-lived records.

The knowledge-base bot is the best example in this repository. It initializes SQLite through the database module and then uses SQL queries from JavaScript to search, insert, update, and review entries.

A minimal shape looks like this:

```js
const database = require("database")

database.configure("sqlite3", "./data/bot.sqlite")
database.exec(`CREATE TABLE IF NOT EXISTS notes (id TEXT PRIMARY KEY, body TEXT)`)
const rows = database.query(`SELECT id, body FROM notes ORDER BY id LIMIT 10`)
```

Use `ctx.store` for ephemeral per-interaction or per-session screen state. Use `require("database")` for durable state that you want to query later.

## Common mistakes and how to avoid them

### 1. Using static choices and autocomplete together

You can use one or the other, not both. If you want typed suggestions, keep `autocomplete: true` and remove `choices`.

### 2. Forgetting to `defer()` for slow commands

If a command does real work after submission, call `ctx.defer(...)` and then edit the response later. That keeps Discord happy and gives the user immediate feedback.

### 3. Reusing the same `customId` in multiple bots

Custom IDs must be unique across loaded bots. The host will reject duplicate component IDs, modal IDs, and autocomplete pairs.

### 4. Assuming `ctx.store` survives restarts

It does not. If you need durability, store state in a real database or file on the host side.

### 5. Expecting `ctx.discord.channels.send(...)` to behave like an ephemeral interaction reply

It is a normal channel message. Use interaction replies when you want ephemeral/private behavior.

## Troubleshooting

| Problem | Likely cause | Fix |
| --- | --- | --- |
| `bot selector is required` when trying to run a bot | The named bot was omitted from `bots run` | Use `discord-bot bots run <bot>` |
| `javascript bot script is required` on direct `run` or `sync-commands` | No explicit `--bot-script` or `DISCORD_BOT_SCRIPT` was provided | Prefer `bots run <bot>` or pass `--bot-script` explicitly |
| A slash command never appears in Discord | Commands were not synced after editing the bot | Run with `--sync-on-start` or use the sync command path |
| A component or modal interaction says no handler exists | The `customId` does not match the registered `component(...)` or `modal(...)` key | Keep `customId` values stable and unique |
| `ctx.config` is missing expected fields | The bot did not declare `configure({ run: ... })` fields or the flags were omitted | Add the run schema and pass the generated flags to `bots run` |
| A message/channel moderation helper fails | The bot lacks the required channel permissions | Check message-management, channel-management, and read/history permissions for the target channel |

## See Also

- `build-and-run-discord-js-bots` — step-by-step tutorial for creating and running bots
- `go-side-ui-dsl-for-discord-bots` — tutorial for the Go-backed `require("ui")` DSL and in-place interaction updates
- `examples/discord-bots/ping/index.js` — button, select, modal, autocomplete, and outbound ops showcase
- `examples/discord-bots/poker/index.js` — a richer bot with game state and action advice
- `examples/discord-bots/knowledge-base/index.js` — runtime config and docs-search example
- `examples/discord-bots/README.md` — repository-level usage notes and command examples
- `examples/discord-bots/interaction-types/index.js` — demo of all command and interaction types
