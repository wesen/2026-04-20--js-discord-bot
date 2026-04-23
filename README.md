# js-discord-bot

A Go-hosted Discord bot runtime with a local JavaScript bot API.

This repository lets you build Discord bots in JavaScript while keeping the outer Discord session, command sync, process lifecycle, and runtime embedding in Go.

At a high level:

- Go owns the Discord gateway/session and embeds the JavaScript runtime.
- JavaScript owns the bot behavior through `require("discord")`.
- Bots are discovered as named implementations from a bot repository such as `examples/discord-bots/`.
- The current runtime model is **one selected JavaScript bot per process**.

---

## What this project is for

This project is useful when you want:

- a real Discord bot process with Go-level control over runtime and deployment,
- a JavaScript authoring experience for bot logic,
- local example bots that double as executable documentation,
- a host API that exposes Discord interactions, events, components, modals, autocomplete, and request-scoped outbound operations.

It is **not** a generic Node Discord bot template. The JavaScript runs inside an embedded goja runtime hosted by Go.

---

## Current model in one diagram

```text
operator / CLI
    ↓
cmd/discord-bot
    ↓
internal/bot          (Discordgo session host)
    ↓
internal/jsdiscord    (embedded JS runtime + require("discord"))
    ↓
examples/discord-bots/<bot>/index.js
```

More concretely:

```text
bots list / bots help / bots run
    ↓
discover one named bot implementation
    ↓
load selected script into goja runtime
    ↓
expose defineBot(...) via require("discord")
    ↓
forward Discord events/interactions into JS handlers
    ↓
normalize JS return payloads back into Discordgo responses
```

---

## Main features

### JavaScript bot authoring API
From JavaScript, bots currently use:

- `defineBot(...)`
- `command(...)`
- `event(...)`
- `component(...)`
- `modal(...)`
- `autocomplete(...)`
- `configure(...)`
- request-scoped context helpers like:
  - `ctx.reply(...)`
  - `ctx.defer(...)`
  - `ctx.edit(...)`
  - `ctx.followUp(...)`
  - `ctx.showModal(...)`
  - `ctx.log.*(...)`
  - `ctx.store.*(...)`
  - `ctx.discord.*(...)`

### Discord interaction support
The host supports a growing set of Discord features, including:

- slash commands
- subcommands
- user commands
- message commands
- buttons and select menus
- modals and text inputs
- autocomplete
- message, reaction, and guild-member events
- request-scoped outbound Discord operations
- message moderation utilities
- guild/role/member lookup helpers
- message history helpers
- thread helpers

### Bot-level runtime config
Bots can describe startup/runtime config with:

```js
configure({
  run: {
    fields: {
      dbPath: { type: "string", default: "./data.sqlite" },
    },
  },
})
```

Those fields become CLI flags on `bots run ...` and are exposed to handlers as `ctx.config`.

---

## Quick start

### 1. Inspect the available bots

```bash
cd /home/manuel/code/wesen/2026-04-20--js-discord-bot
GOWORK=off go run ./cmd/discord-bot bots list --bot-repository ./examples/discord-bots
```

### 2. Inspect one bot

```bash
GOWORK=off go run ./cmd/discord-bot bots help ping --bot-repository ./examples/discord-bots
```

### 3. Run one selected bot

```bash
GOWORK=off go run ./cmd/discord-bot bots run ping \
  --bot-repository ./examples/discord-bots \
  --bot-token "$DISCORD_BOT_TOKEN" \
  --application-id "$DISCORD_APPLICATION_ID" \
  --guild-id "$DISCORD_GUILD_ID" \
  --sync-on-start
```

### 4. Run a bot with runtime config

Example with the knowledge-base bot:

```bash
GOWORK=off go run ./cmd/discord-bot bots run knowledge-base \
  --bot-repository ./examples/discord-bots \
  --bot-token "$DISCORD_BOT_TOKEN" \
  --application-id "$DISCORD_APPLICATION_ID" \
  --guild-id "$DISCORD_GUILD_ID" \
  --db-path ./examples/discord-bots/knowledge-base/data/knowledge.sqlite \
  --sync-on-start
```

---

## Operator-facing commands

### Named bot runner
The main recommended UX is:

```bash
discord-bot bots list
discord-bot bots help <bot>
discord-bot bots run <bot>
```

In practice:

```bash
GOWORK=off go run ./cmd/discord-bot bots --bot-repository ./examples/discord-bots list
GOWORK=off go run ./cmd/discord-bot bots --bot-repository ./examples/discord-bots help knowledge-base
GOWORK=off go run ./cmd/discord-bot bots --bot-repository ./examples/discord-bots run moderation --sync-on-start
```

### Direct host commands
There are also direct host-level commands:

- `run`
- `validate-config`
- `sync-commands`

Use them when you want to point the host directly at one explicit script path:

```bash
GOWORK=off go run ./cmd/discord-bot run \
  --bot-script ./examples/discord-bots/ping/index.js \
  --bot-token "$DISCORD_BOT_TOKEN" \
  --application-id "$DISCORD_APPLICATION_ID" \
  --guild-id "$DISCORD_GUILD_ID"
```

For everyday use, `bots run <bot>` is usually the clearer path.

### Public single-bot embedding path
There is now also a public Go package for the simple one-bot case:

```go
bot, err := framework.New(
    framework.WithCredentialsFromEnv(),
    framework.WithScript("./examples/discord-bots/unified-demo/index.js"),
    framework.WithRuntimeConfig(map[string]any{
        "db_path": "./examples/discord-bots/unified-demo/data/demo.sqlite",
        "api_key": "local-dev-key",
    }),
    framework.WithSyncOnStart(true),
)
```

See:
- `pkg/framework/` — public single-bot API
- `examples/framework-single-bot/` — minimal embeddable app example

---

## Embedded docs

The CLI ships with embedded help pages.

Show them with:

```bash
GOWORK=off go run ./cmd/discord-bot help discord-js-bot-api-reference
GOWORK=off go run ./cmd/discord-bot help build-and-run-discord-js-bots
```

The source files live at:

- `pkg/doc/topics/discord-js-bot-api-reference.md`
- `pkg/doc/tutorials/building-and-running-discord-js-bots.md`

If you want the full handler/payload/API details, start there after reading this README.

---

## Example bots

Current example bot repository:

- `examples/discord-bots/ping/`
  - API showcase for buttons, modals, autocomplete, deferred replies, and outbound operations
- `examples/discord-bots/knowledge-base/`
  - SQLite-backed knowledge steward with capture, teach/remember, search, article, review, and source workflows
- `examples/discord-bots/support/`
  - deferred/edit/follow-up flows and thread helpers
- `examples/discord-bots/moderation/`
  - event-heavy admin/moderation helper bot exercising member/message/channel/guild/role utilities
- `examples/discord-bots/poker/`
  - richer stateful example with game logic and action advice
- `examples/discord-bots/interaction-types/`
  - demo of slash commands, subcommands, user commands, and message commands
- `examples/discord-bots/announcements.js`
  - root-level script to exercise direct-file discovery

Repository-level notes and examples:

- `examples/discord-bots/README.md`

---

## Project layout

```text
cmd/
  discord-bot/             CLI entrypoint

internal/
  bot/                     live Discordgo session wrapper
  botcli/                  named bot repository discovery and runner
  config/                  host config decoding and validation
  jsdiscord/               embedded JS runtime, defineBot API, dispatch, payload normalization

examples/
  discord-bots/            named JS bot implementations

pkg/
  doc/                     embedded help pages

ttmp/
  ...                      ticket-based project documentation and diaries
```

---

## Important runtime concepts

### One selected JS bot per process
This repo intentionally uses a **single selected JavaScript bot** in each running process.

That means:

- `discord-bot bots run <bot>` selects one named implementation,
- startup/runtime config applies to that one bot,
- composition should happen inside the selected bot rather than via host-side multi-bot composition.

### The JS API is request-scoped
Outbound Discord operations live under `ctx.discord`, for example:

- `ctx.discord.channels.send(...)`
- `ctx.discord.messages.edit(...)`
- `ctx.discord.messages.fetch(...)`
- `ctx.discord.members.addRole(...)`
- `ctx.discord.guilds.fetch(...)`
- `ctx.discord.roles.list(...)`
- `ctx.discord.threads.start(...)`

This is intentional: the bot API is tied to a live request/session context, not a global singleton client.

### The example bots are part of the product story
The example bots are not just fixtures. They are the main demonstrations of the authoring API and should be treated as executable docs.

---

## Minimal JavaScript bot example

```js
const { defineBot } = require("discord")

module.exports = defineBot(({ command, event, configure }) => {
  configure({
    name: "demo",
    description: "A minimal Discord JS bot",
  })

  event("ready", async (ctx) => {
    ctx.log.info("demo bot ready", {
      user: ctx.me && ctx.me.username,
    })
  })

  command("ping", {
    description: "Reply with pong",
  }, async () => {
    return { content: "pong" }
  })
})
```

---

## Development notes

### Validation
The repo has been worked on in a way where `GOWORK=off` is the safest default for tests and local runs.

Common validation commands:

```bash
GOWORK=off go test ./internal/jsdiscord ./internal/bot ./cmd/discord-bot
GOWORK=off go test ./internal/botcli ./internal/jsdiscord ./internal/bot ./cmd/discord-bot
GOWORK=off go test ./...
```

### Local dependency note
This repo currently uses a local replace for `go-go-goja`:

- see `go.mod`
- current replace target: `/home/manuel/code/wesen/corporate-headquarters/go-go-goja`

That means local development expects that dependency to exist on this machine or be adjusted in `go.mod` for your environment.

### Environment variables
Common runtime variables are:

- `DISCORD_BOT_TOKEN`
- `DISCORD_APPLICATION_ID`
- `DISCORD_GUILD_ID`
- optionally `DISCORD_PUBLIC_KEY`
- optionally `DISCORD_CLIENT_ID`
- optionally `DISCORD_CLIENT_SECRET`

Example environment scaffolding:

- `.envrc.example`

---

## Where to read next

If you are new to the project, this is the recommended reading order:

1. `README.md`
2. `cmd/discord-bot/root.go`
3. `cmd/discord-bot/commands.go`
4. `internal/config/config.go`
5. `internal/bot/bot.go`
6. `internal/jsdiscord/runtime.go`
7. `internal/jsdiscord/bot.go`
8. `internal/jsdiscord/host_dispatch.go`
9. `internal/jsdiscord/host_payloads.go`
10. `examples/discord-bots/ping/index.js`
11. `examples/discord-bots/knowledge-base/index.js`
12. `pkg/doc/tutorials/building-and-running-discord-js-bots.md`

---

## Current status in one sentence

This is a working Go-hosted Discord bot platform with a real local JavaScript bot API, a named bot repository runner, embedded help docs, and a growing set of example bots that exercise the runtime in realistic ways.
