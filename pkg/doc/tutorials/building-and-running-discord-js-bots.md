---
Title: "Build and Run Discord JavaScript Bots"
Slug: "build-and-run-discord-js-bots"
Short: "Step-by-step guide for creating a bot, adding interactions, and running it from the named-bot repository."
Topics:
- discord
- javascript
- bots
- tutorial
- playbook
- run
- examples
Commands:
- bots list
- bots help
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
SectionType: Tutorial
---

## What this playbook helps you do

This guide shows the full day-one path for a new bot developer:

1. discover the repository layout
2. inspect existing bots with `bots list` and `bots help`
3. create a new JavaScript bot
4. add commands, events, buttons, modals, and autocomplete
5. add runtime config fields when a bot needs them
6. run the bot through the named-bot CLI path
7. test the bot in Discord without guessing

The goal is practical fluency. By the end, you should be able to build a complete bot from scratch, run it locally, and know where to look when something fails.

> ⚠️ **Runtime Environment**
> Bot scripts run inside a Goja JavaScript engine embedded in Go, **not Node.js**.
> - **Available modules:** `require("discord")`, `require("timer")`, `require("database")`, `require("ui")`
> - **Unavailable:** `fs`, `path`, `http`, `fetch`, `process`, npm packages, or any Node.js standard library
> - **No file system access from JS.** Deliver generated content as Discord file attachments via `ctx.discord.channels.send()` with `files: [...]`

## How this works (three sentences)

1. You write a JavaScript module that calls `defineBot(...)` and registers commands, events, and handlers.
2. The Go host loads your script, syncs slash commands to Discord, opens the gateway, and dispatches events.
3. Your script uses `ctx.discord.*` to call Discord APIs; the host handles authentication, rate limits, and reconnections.

## 1. Understand the repository layout

Bots in this repo are not individual ad hoc scripts. They are named bot implementations under `examples/discord-bots/`.

A bot usually lives at:

```text
examples/discord-bots/<bot-name>/index.js
```

If the bot needs helper code, put that in a nearby `lib/` directory:

```text
examples/discord-bots/<bot-name>/
  index.js
  lib/
    helpers.js
    data.js
```

The existing examples are good starting points:

- `ping/` — the richest API showcase
- `poker/` — game state, help, buttons, and modals
- `knowledge-base/` — runtime config and docs search
- `support/` — deferred replies and follow-ups
- `moderation/` — message-triggered workflows

## 2. Discover what is already there

Start every new bot by looking at the current repository inventory.

```bash
go run ./cmd/discord-bot bots --bot-repository ./examples/discord-bots list
```

This tells you the canonical bot names. Those names are what you pass to `bots help` and `bots run`.

Then inspect the bot you care about:

```bash
go run ./cmd/discord-bot bots --bot-repository ./examples/discord-bots help ping
```

That command shows the bot’s description, slash commands, events, and any runtime config fields.

## 3. Create a minimal bot first

The easiest way to succeed is to start with one command and one event.

```js
const { defineBot } = require("discord")

module.exports = defineBot(({ command, event, configure }) => {
  configure({
    name: "demo",
    description: "A minimal Discord JS bot",
    category: "examples",
  })

  command("ping", {
    description: "Reply with pong",
  }, async () => {
    return { content: "pong" }
  })

  event("ready", async (ctx) => {
    ctx.log.info("demo bot ready", {
      user: ctx.me && ctx.me.username,
    })
  })
})
```

### Why this shape works

- `configure(...)` makes the bot discoverable
- `command(...)` creates a slash command users can run
- `event("ready", ...)` proves the gateway connection is alive

Once this works, you can grow the bot safely.

## 4. Add one slow command the right way

When a command needs to do work after submission, do not block without acknowledging the interaction. Use the defer/edit pattern.

```js
const { sleep } = require("timer")

command("search", {
  description: "Search for a topic",
  options: {
    query: {
      type: "string",
      description: "Topic to search for",
      required: true,
      autocomplete: true,
    },
  },
}, async (ctx) => {
  await ctx.defer({ ephemeral: true })
  await ctx.edit({
    content: `Searching for ${ctx.args.query}...`,
    ephemeral: true,
  })

  // Simulate work or call an API.
  await sleep(2000)

  await ctx.edit({
    content: `Results for ${ctx.args.query}: Architecture, Testing, Runbooks`,
    ephemeral: true,
  })
})
```

This is the right pattern when the user should get immediate feedback, then a delayed result.

### Autocomplete can still work

The autocomplete handler is separate from the command handler. Add it alongside the command:

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

That gives the user suggestions while typing, then a deferred result after submission.

## 5. Add buttons and select menus

Buttons and select menus use the `component(...)` registration helper.

```js
command("ping", {
  description: "Reply with a rich message",
}, async () => {
  return {
    content: "pong",
    components: [
      {
        type: "actionRow",
        components: [
          {
            type: "button",
            style: "primary",
            label: "Open panel",
            customId: "ping:panel",
          },
          {
            type: "select",
            customId: "ping:topic",
            placeholder: "Choose a topic",
            options: [
              { label: "Architecture", value: "architecture" },
              { label: "Testing", value: "testing" },
            ],
          },
        ],
      },
    ],
  }
})

component("ping:panel", async () => {
  return {
    content: "Panel button clicked from JavaScript",
    ephemeral: true,
  }
})

component("ping:topic", async (ctx) => {
  const selected = Array.isArray(ctx.values) && ctx.values.length > 0 ? ctx.values[0] : "(none)"
  return {
    content: `Selected topic: ${selected}`,
    ephemeral: true,
  }
})
```

### When to use component handlers

Use them when you want a message to stay interactive after the initial slash command response.

A good mental model is:

- slash command starts the workflow
- command response renders the UI
- component handlers respond to clicks or selections

## 6. Add a modal workflow

Modals are great when you need more than a single slash-command field.

```js
command("feedback", {
  description: "Open a feedback modal",
}, async (ctx) => {
  await ctx.showModal({
    customId: "feedback:submit",
    title: "Feedback",
    components: [
      {
        type: "actionRow",
        components: [
          {
            type: "textInput",
            customId: "summary",
            label: "Summary",
            style: "short",
            required: true,
            minLength: 5,
            maxLength: 100,
          },
        ],
      },
      {
        type: "actionRow",
        components: [
          {
            type: "textInput",
            customId: "details",
            label: "Details",
            style: "paragraph",
            maxLength: 500,
          },
        ],
      },
    ],
  })
})

modal("feedback:submit", async (ctx) => {
  return {
    content: `Thanks for the feedback: ${ctx.values.summary}\nDetails: ${ctx.values.details || "(none)"}`,
    ephemeral: true,
  }
})
```

### When to use a modal

Use a modal when you want a structured form with multiple text inputs. It is much better than stuffing long text into a single slash-command option.

## 7. Add runtime config when the bot needs operator input

Some bots need a few values at startup, but those values are not Discord command arguments. Use `configure({ run: { fields: ... }})`.

```js
configure({
  name: "knowledge-base",
  description: "Search and summarize internal docs from JavaScript",
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

Then read those values in JavaScript with `ctx.config`:

```js
const indexPath = ctx.config && ctx.config.index_path || "builtin-docs"
```

### How runtime config becomes CLI flags

For each field:

- the field name becomes the JavaScript config key
- the CLI gets a kebab-case flag
- the help output shows the flag and the field description

For example:

- `index_path` becomes `--index-path`
- `read_only` becomes `--read-only`

## 8. Run the bot through the named-bot CLI path

The normal workflow in this repository is:

```bash
go run ./cmd/discord-bot bots --bot-repository ./examples/discord-bots run ping --bot-token "$DISCORD_BOT_TOKEN" --application-id "$DISCORD_APPLICATION_ID" --guild-id "$DISCORD_GUILD_ID" --sync-on-start
```

### What each part means

- `bots` — the named-bot subcommand group
- `--bot-repository ./examples/discord-bots` — where the CLI should discover bot scripts
- `run ping` — run the bot named `ping`
- `--bot-token` — the Discord bot token
- `--application-id` — the Discord application/client ID
- `--guild-id` — optional fast sync target for development
- `--sync-on-start` — replace the bot’s slash commands before opening the gateway

If your bot has runtime config fields, add them after the selector:

```bash
go run ./cmd/discord-bot bots --bot-repository ./examples/discord-bots run knowledge-base \
  --bot-token "$DISCORD_BOT_TOKEN" \
  --application-id "$DISCORD_APPLICATION_ID" \
  --guild-id "$DISCORD_GUILD_ID" \
  --index-path ./docs/local-index \
  --read-only \
  --sync-on-start
```

## 9. Inspect parsed values before you run

If something behaves strangely, print the resolved bot and runtime config before opening Discord:

```bash
go run ./cmd/discord-bot bots --bot-repository ./examples/discord-bots run ping \
  --bot-token "$DISCORD_BOT_TOKEN" \
  --application-id "$DISCORD_APPLICATION_ID" \
  --print-parsed-values
```

This is useful when you want to confirm:

- which bot was selected
- which runtime config flags were parsed
- which secrets are present
- whether your bot help text is surfacing the fields you expect

## 10. Test in Discord with a small checklist

Once the bot is running and synced, test the behavior in this order:

1. Run `/ping` and confirm the command responds.
2. Click any buttons and confirm the `component(...)` handler fires.
3. Open any modal and confirm the `modal(...)` handler receives submitted values.
4. Type into any autocomplete field and confirm suggestions appear.
5. Run a slower command such as `/search` and confirm you see an immediate private acknowledgement followed by the final result.
6. Try the message trigger, if the bot has one, such as `!pingjs`.

For the sample bots in this repo, the most useful smoke tests are:

- `ping` — slash command, button, select menu, modal, autocomplete, and outbound ops
- `poker` — help flow, game state, rank evaluation, and action advice
- `knowledge-base` — runtime config plus docs search

## 11. Organize a real bot beyond the first commit

Once the minimal bot works, split it into clear pieces.

Recommended structure:

```text
examples/discord-bots/my-bot/
  index.js
  lib/
    helpers.js
    search.js
    ui.js
```

A good bot file usually contains:

- metadata in `configure(...)`
- one or more slash commands
- a small number of event handlers
- component/modals/autocomplete registrations if needed
- a single place where help text and example commands live

## 12. Troubleshoot the common failures

| Problem | Likely cause | Fix |
| --- | --- | --- |
| `bot selector is required` | `bots run` was called without a bot name | Add `run <bot-name>` |
| `bot "x" not found` | The bot name does not match `bots list` | Use the exact name from `bots list` |
| Slash command sync fails with option ordering errors | Required options were not ordered first in the source data | Declare required options first, then sync again |
| Autocomplete never appears | The option is missing `autocomplete: true` | Add autocomplete and remove static choices |
| A button click says no handler exists | The `customId` does not match any `component(...)` registration | Make the IDs match exactly and keep them unique |
| Modal submit fails | The modal was not registered or the `customId` changed | Keep the modal `customId` stable |
| `ctx.config` is empty | No runtime config fields were declared or the flags were omitted | Add `configure({ run: ... })` and pass the matching flags |
| `ctx.defer` does nothing useful | The handler deferred but never edited or followed up | Call `ctx.edit(...)` or `ctx.followUp(...)` after the work finishes |
| Discord permission errors | The bot token lacks permission in the guild or channel | Check bot permissions and channel access |

## 13. Use the examples as living templates

The best starting points in this repository are:

- `examples/discord-bots/ping/index.js` — richest API showcase
- `examples/discord-bots/poker/index.js` — a complete command set with help and modals
- `examples/discord-bots/knowledge-base/index.js` — runtime config and docs search

Treat them as copyable templates, not just demos.

## See Also

- `discord-js-bot-api-reference` — API reference for the builder, contexts, payloads, and operations
- `examples/discord-bots/README.md` — repository command examples and runtime notes
- `examples/discord-bots/ping/index.js` — full JS showcase bot
- `examples/discord-bots/poker/index.js` — help-oriented bot with game-state commands
