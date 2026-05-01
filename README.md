# discord-bot

A Go-hosted Discord bot runtime with a local JavaScript bot API.

[![Go Reference](https://pkg.go.dev/badge/github.com/go-go-golems/discord-bot.svg)](https://pkg.go.dev/github.com/go-go-golems/discord-bot)

Go owns the Discord gateway/session and embeds a JavaScript runtime (goja). JavaScript owns the bot behavior through `require("discord")`. The current model is **one selected JavaScript bot per process**.

---

## Install

**Homebrew (macOS / Linux):**

```bash
brew install go-go-golems/tap/discord-bot
```

**deb / rpm (Linux):**

See [releases](https://github.com/go-go-golems/discord-bot/releases) for `.deb` and `.rpm` packages, or use the [fury.io apt repo](https://push.fury.io/go-go-golems/).

**From source:**

```bash
go install github.com/go-go-golems/discord-bot/cmd/discord-bot@latest
```

---

## Quick start

### 1. List available bots

```bash
discord-bot bots list --bot-repository ./examples/discord-bots
```

### 2. Inspect one bot

```bash
discord-bot bots help ping --bot-repository ./examples/discord-bots
```

### 3. Run a bot

```bash
discord-bot bots ping run \
  --bot-repository ./examples/discord-bots \
  --bot-token "$DISCORD_BOT_TOKEN" \
  --application-id "$DISCORD_APPLICATION_ID" \
  --guild-id "$DISCORD_GUILD_ID" \
  --sync-on-start
```

### Environment variables

| Variable | Required | Description |
|----------|----------|-------------|
| `DISCORD_BOT_TOKEN` | Yes | Discord bot token |
| `DISCORD_APPLICATION_ID` | Yes | Discord application ID |
| `DISCORD_GUILD_ID` | No | Scope commands to a specific guild |
| `DISCORD_PUBLIC_KEY` | No | For HTTP interaction verification |
| `DISCORD_CLIENT_ID` | No | For OAuth flows |
| `DISCORD_CLIENT_SECRET` | No | For OAuth flows |

---

## Using discord-bot with an LLM agent

To use discord-bot as the runtime for an AI agent or LLM-powered bot:

1. Install: `brew install go-go-golems/tap/discord-bot` (or see [Install](#install))
2. Point the LLM to this project and tell it to run:
   ```bash
   discord-bot help --all
   ```
   to get started.
3. The LLM can then use the CLI to explore, configure, and run bot scripts. Example workflow for the LLM:
   ```bash
   # Discover available bots
   discord-bot bots list --bot-repository ./examples/discord-bots
   
   # Inspect a bot's help
   discord-bot bots help knowledge-base --bot-repository ./examples/discord-bots
   
   # Run a bot
   discord-bot bots knowledge-base run --bot-repository ./examples/discord-bots \
     --bot-token "$DISCORD_BOT_TOKEN" \
     --application-id "$DISCORD_APPLICATION_ID"
   ```

4. For documentation, tell the LLM to run:
   ```bash
   discord-bot help build-and-run-discord-js-bots   # Starter guide
   discord-bot help discord-js-bot-api-reference    # API reference
   discord-bot help go-side-ui-dsl-for-discord-bots # UI DSL tutorial
   ```

5. For embedding in Go: see [Go API — embedding discord-bot in your application](#go-api--embedding-discord-bot-in-your-application)

---


## Go API — embedding discord-bot in your application

### Simple single-bot embedding (`pkg/framework`)

```go
package main

import (
    "context"
    "github.com/go-go-golems/discord-bot/pkg/framework"
)

func main() {
    bot, err := framework.New(
        framework.WithCredentialsFromEnv(),
        framework.WithScript("./my-bot/index.js"),
        framework.WithSyncOnStart(true),
    )
    if err != nil {
        panic(err)
    }
    bot.Run(context.Background())
}
```

### Repo-driven multi-bot CLI (`pkg/botcli`)

For the full `bots list / bots help / bots <name> run` experience inside your own Cobra command tree:

```go
bootstrap, _ := botcli.BuildBootstrap(
    os.Args[1:],
    botcli.WithDefaultRepositories("./bots"),
)
botsCmd, _ := botcli.NewBotsCommand(bootstrap)
rootCmd.AddCommand(botsCmd)
```

### Custom native modules

Add Go-native `require()` modules that your JS bot scripts can use:

```go
botsCmd, _ := botcli.NewBotsCommand(
    bootstrap,
    botcli.WithRuntimeModuleRegistrars(&MyAppModule{}),
)
```

See `examples/framework-custom-module/` for a complete example.

### Embedding examples

| Example | Description |
|---------|-------------|
| `examples/framework-single-bot/` | Minimal single-bot embedding |
| `examples/framework-custom-module/` | Custom `require("app")` module |
| `examples/framework-combined/` | Built-in bot + repo-driven discovery |

---

## JavaScript bot authoring API

### Minimal bot

```js
const { defineBot } = require("discord")

module.exports = defineBot(({ command, event, configure }) => {
  configure({
    name: "demo",
    description: "A minimal Discord JS bot",
  })

  event("ready", async (ctx) => {
    ctx.log.info("demo bot ready", { user: ctx.me && ctx.me.username })
  })

  command("ping", { description: "Reply with pong" }, async () => {
    return { content: "pong" }
  })
})
```

### Handler context (`ctx`)

- **Responses:** `ctx.reply(...)`, `ctx.defer(...)`, `ctx.edit(...)`, `ctx.followUp(...)`, `ctx.showModal(...)`
- **Logging:** `ctx.log.info(...)`, `ctx.log.warn(...)`, `ctx.log.error(...)`, `ctx.log.debug(...)`
- **Store:** `ctx.store.get(key)`, `ctx.store.set(key, value)`, `ctx.store.delete(key)`, `ctx.store.keys()`
- **Discord ops:** `ctx.discord.channels.*`, `ctx.discord.messages.*`, `ctx.discord.members.*`, `ctx.discord.guilds.*`, `ctx.discord.roles.*`, `ctx.discord.threads.*`
- **Config:** `ctx.config` — runtime config from CLI flags

### Registration functions

- `command(name, spec, handler)` — slash commands, user commands, message commands
- `event(name, handler)` — ready, messageCreate, messageUpdate, guildMemberAdd, ...
- `component(customId, handler)` — button and select menu interactions
- `modal(customId, handler)` — modal submit interactions
- `autocomplete(commandName, handler)` — autocomplete for command options
- `configure(spec)` — bot metadata and runtime config fields

### Runtime config

Bots can declare CLI flags that become `ctx.config` values:

```js
configure({
  run: {
    fields: {
      dbPath: { type: "string", default: "./data.sqlite" },
    },
  },
})
```

Those fields appear as `--db-path` on `bots <name> run`.

---

## Documentation

Start with the built-in help:

```bash
discord-bot help build-and-run-discord-js-bots   # Step-by-step tutorial
discord-bot help discord-js-bot-api-reference    # API reference
discord-bot help --all                           # List all topics
```


| Topic | Description |
|-------|-------------|
| `build-and-run-discord-js-bots` | Full tutorial: create bots, add interactions, use runtime config, use the database module |
| `discord-js-bot-api-reference` | Complete API reference: `defineBot`, `command`, `event`, `component`, `modal`, `autocomplete`, `ctx.config`, `ctx.discord.*`, `ctx.store`, and `require("database")` |
| `go-side-ui-dsl-for-discord-bots` | Tutorial for the Go-backed `require("ui")` DSL for fluent message, embed, and component builders |

---


## Example bots

| Bot | Description |
|-----|-------------|
| `ping/` | API showcase: buttons, modals, autocomplete, deferred replies, outbound ops |
| `knowledge-base/` | SQLite-backed knowledge steward: capture, search, review — the canonical reference for `require("database")` and runtime config |
| `support/` | Deferred/edit/follow-up flows and thread helpers |
| `moderation/` | Event-heavy admin/moderation helper |
| `poker/` | Stateful game logic example |
| `interaction-types/` | Slash, subcommands, user commands, message commands |
| `show-space/` | Show/space management with date parsing |
| `unified-demo/` | Combined demo of all features |

See `examples/discord-bots/` for the full set.

---

## Architecture

```text
discord-bot binary (cmd/discord-bot)
    │
    ├── internal/bot/        Discordgo session wrapper
    ├── internal/config/     Host config (credentials, validation)
    ├── internal/jsdiscord/  Embedded JS runtime + require("discord")
    │
    ├── pkg/framework/       Public: simple single-bot embedding
    ├── pkg/botcli/          Public: repo-driven multi-bot CLI
    └── pkg/doc/             Embedded help pages
```

Data flow at runtime:

```text
Discord gateway → discordgo session → jsdiscord.Host
  → dispatch event → find JS handler → call with ctx
  → normalize return payload → send response via discordgo
```

---

## Development

```bash
make lint          # Run golangci-lint
make test          # Run all tests
make build         # Build binary
make goreleaser    # Snapshot release (local)
```

### Where to start reading

1. This README
2. `cmd/discord-bot/root.go` — CLI wiring
3. `internal/bot/bot.go` — Session lifecycle
4. `internal/jsdiscord/host.go` — JS runtime host
5. `pkg/doc/tutorials/building-and-running-discord-js-bots.md`

---

## License

MIT
