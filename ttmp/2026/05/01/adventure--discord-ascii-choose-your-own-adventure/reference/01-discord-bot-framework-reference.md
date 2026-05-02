---
Title: Discord Bot Framework Reference
Ticket: adventure
Status: active
Topics:
    - discord
    - game
    - adventure
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/bot/bot.go
      Note: Discord session wrapper
    - Path: internal/jsdiscord/bot_compile.go
      Note: Bot draft and handle structures
    - Path: internal/jsdiscord/host.go
      Note: Goja host creation
    - Path: internal/jsdiscord/host_dispatch.go
    - Path: internal/jsdiscord/runtime.go
      Note: JavaScript discord module and defineBot registration API
    - Path: internal/jsdiscord/ui_module.go
      Note: UI builder module registration for messages
    - Path: pkg/botcli/discover.go
      Note: Repository bot discovery rules and selector resolution
    - Path: pkg/botcli/doc.go
      Note: Public repo-driven bot CLI package overview and customization guidance
    - Path: pkg/framework/framework.go
      Note: Public single-bot embedding API for running one explicit JavaScript bot
ExternalSources: []
Summary: Reference for how this repository creates, discovers, embeds, and runs Discord bots.
LastUpdated: 2026-05-01T13:14:35.59249-07:00
WhatFor: Use when designing the adventure bot's runtime shape, command handlers, buttons, modals, and optional Go embedding layer.
WhenToUse: Before implementing a new bot script or deciding whether to use pkg/framework, pkg/botcli, or internal jsdiscord extension points.
---


# Discord Bot Framework Reference

## Goal

Summarize the repository's Discord bot framework so the `adventure` ticket can choose the right implementation path for a YAML/LLM-powered choose-your-own-adventure bot.

## Context

This repo is a Go-hosted Discord runtime with JavaScript-authored bot behavior:

- Go owns Discord sessions, gateway event registration, command sync, credentials, and host operations.
- JavaScript owns bot behavior through `require("discord")` and `defineBot(...)`.
- The framework supports slash commands, user/message commands, events, components/buttons/selects, modals, autocomplete handlers, runtime config, and custom Go-native JS modules.
- There are two public Go layers:
  - `pkg/framework`: run one explicit bot script.
  - `pkg/botcli`: discover named bot scripts from repositories and mount a `bots` CLI subtree.

## Quick Reference

### Main implementation layers

| Layer | Purpose | Key symbols | Files |
|---|---|---|---|
| Public single-bot embedding | Embed one explicit JavaScript bot in a Go app | `framework.New`, `framework.WithScript`, `framework.WithCredentialsFromEnv`, `framework.WithRuntimeConfig`, `framework.WithSyncOnStart`, `(*framework.Bot).Run` | [`pkg/framework/framework.go`](../../../../../../pkg/framework/framework.go) |
| Public repo-driven CLI | Discover bot scripts and expose `bots list/help/run` style commands | `botcli.BuildBootstrap`, `botcli.NewBotsCommand`, `botcli.DiscoverBots`, `botcli.ResolveBot`, `botcli.WithRuntimeModuleRegistrars`, `botcli.WithRuntimeFactory` | [`pkg/botcli/doc.go`](../../../../../../pkg/botcli/doc.go), [`pkg/botcli/bootstrap.go`](../../../../../../pkg/botcli/bootstrap.go), [`pkg/botcli/discover.go`](../../../../../../pkg/botcli/discover.go), [`pkg/botcli/command_root.go`](../../../../../../pkg/botcli/command_root.go) |
| Internal Discord session wrapper | Creates `discordgo.Session`, loads JS host, wires gateway handlers, syncs commands | `bot.NewWithScript`, `(*bot.Bot).SyncCommands`, `(*bot.Bot).Open`, `(*bot.Bot).Close`, `handleInteractionCreate`, `handleMessageCreate` | [`internal/bot/bot.go`](../../../../../../internal/bot/bot.go) |
| JS host/runtime | Builds goja runtime, registers modules, loads script, compiles bot handle | `jsdiscord.NewHost`, `jsdiscord.LoadBot`, `jsdiscord.InspectScript`, `(*Host).ApplicationCommands`, `(*Host).Describe` | [`internal/jsdiscord/host.go`](../../../../../../internal/jsdiscord/host.go), [`internal/jsdiscord/descriptor.go`](../../../../../../internal/jsdiscord/descriptor.go) |
| JavaScript authoring API | Provides `require("discord")`, `defineBot`, and registration functions | `NewRegistrar`, `RuntimeState.Loader`, `RuntimeState.defineBot`, `botDraft.command`, `botDraft.component`, `botDraft.modal`, `botDraft.event`, `botDraft.configure` | [`internal/jsdiscord/runtime.go`](../../../../../../internal/jsdiscord/runtime.go), [`internal/jsdiscord/bot_compile.go`](../../../../../../internal/jsdiscord/bot_compile.go) |
| Dispatch bridge | Converts Discord events/interactions into JS handler calls | `DispatchRequest`, `BotHandle.DispatchCommand`, `DispatchComponent`, `DispatchModal`, `DispatchEvent`, `Host.DispatchInteraction` | [`internal/jsdiscord/bot_compile.go`](../../../../../../internal/jsdiscord/bot_compile.go), [`internal/jsdiscord/bot_dispatch.go`](../../../../../../internal/jsdiscord/bot_dispatch.go), [`internal/jsdiscord/host_dispatch.go`](../../../../../../internal/jsdiscord/host_dispatch.go) |
| UI DSL | JS `require("ui")` builders for messages, embeds, rows, buttons, selects, forms/modals | `UIRegistrar`, `ui.message`, `ui.embed`, `ui.button`, `ui.select`, `ui.form`, `ui.confirm`, `ui.pager` | [`internal/jsdiscord/ui_module.go`](../../../../../../internal/jsdiscord/ui_module.go), [`examples/discord-bots/ui-showcase/index.js`](../../../../../../examples/discord-bots/ui-showcase/index.js) |

### Bot script shape

Bot scripts normally export the result of `defineBot(...)`:

```js
const { defineBot } = require("discord")
const ui = require("ui")

module.exports = defineBot(({ command, component, modal, event, configure }) => {
  configure({
    name: "adventure",
    description: "ASCII choose-your-own-adventure bot",
    run: {
      llm: {
        title: "LLM",
        fields: {
          model: { type: "string", help: "LLM model name", default: "gpt-4.1-mini" }
        }
      }
    }
  })

  command("adventure-start", {
    description: "Start an adventure session"
  }, async (ctx) => {
    return ui.message()
      .content("```\nTHE GATE WAITS...\n```")
      .row(ui.button("adv:choice:open", "Open the gate", "primary"))
  })

  component("adv:choice:open", async (ctx) => {
    return { content: "```\nThe hinges scream.\n```" }
  })

  modal("adv:freeform", async (ctx) => {
    return { content: `You try: ${ctx.modal && ctx.modal.values && ctx.modal.values.action}` }
  })

  event("ready", async (ctx) => {
    ctx.log.info("adventure ready", { bot: ctx.metadata && ctx.metadata.name })
  })
})
```

### JavaScript registration API

`defineBot` receives an API object assembled in [`RuntimeState.defineBot`](../../../../../../internal/jsdiscord/runtime.go):

| Registration function | Use for adventure bot |
|---|---|
| `command(name, spec, handler)` | Slash commands such as `/adventure-start`, `/adventure-resume`, `/adventure-state`. |
| `subcommand(root, name, spec, handler)` | A grouped `/adventure start`, `/adventure resume`, `/adventure inspect` surface if preferred. |
| `component(customID, handler)` | Buttons/select menus for structured choices. Likely the main turn-advance path. |
| `modal(customID, handler)` | Free-form actions, dialogue, passwords, character names, or “try something else.” |
| `autocomplete(commandName, optionName, handler)` | Adventure IDs, save slots, known inventory names, etc. |
| `event(name, handler)` | `ready`, `messageCreate`, guild/member/message/reaction lifecycle hooks. |
| `configure(options)` | Metadata and run config surfaced to help/CLI. |

### Runtime context exposed to handlers

Go builds a `DispatchRequest` and turns it into JS `ctx`. Important fields for the adventure design are declared in [`internal/jsdiscord/bot_compile.go`](../../../../../../internal/jsdiscord/bot_compile.go):

| Context field | Meaning / likely use |
|---|---|
| `ctx.args` | Slash command options. Use for adventure name, visibility, mode. |
| `ctx.values` | Component/select values. Use for selected choice IDs. |
| `ctx.interaction`, `ctx.user`, `ctx.guild`, `ctx.channel`, `ctx.member` | Discord identity/scope. Use to key sessions by user/channel/thread/guild. |
| `ctx.message` | Message event payload. Useful if supporting free-text channel input. |
| `ctx.component` | Component metadata for buttons/selects. |
| `ctx.modal` | Modal payload/values for free-form actions. |
| `ctx.metadata` | Bot metadata from `configure(...)`. |
| `ctx.config` | Runtime config injected from CLI/framework. Useful for model, database path, feature flags. |
| `ctx.discord` | Host operations for channels, messages, threads, guilds, roles, members. |
| `ctx.reply`, `ctx.followUp`, `ctx.edit`, `ctx.defer`, `ctx.showModal` | Interaction response lifecycle. Use `defer` around LLM calls. |
| `ctx.log` | Structured logging from JS handlers. |

### Discord operations available to JS

`DispatchRequest.Discord` exposes host operations via `ctx.discord`. For adventure, the most relevant families are:

- `ctx.discord.channels.send/fetch/setTopic/setSlowmode`
- `ctx.discord.messages.fetch/list/edit/delete/react/pin/unpin/listPinned`
- `ctx.discord.threads.fetch/start/join/leave`
- `ctx.discord.guilds.fetch`, `ctx.discord.roles.*`, `ctx.discord.members.*`

The Go operation surface is represented by `DiscordOps` in [`internal/jsdiscord/bot_compile.go`](../../../../../../internal/jsdiscord/bot_compile.go), with implementations split across `internal/jsdiscord/host_ops_*.go`.

### Runtime/config flow

1. CLI or embedding app obtains credentials and runtime config.
2. `framework.New(...)` or `botcli` run command calls `internal/bot.NewWithScript(...)`.
3. `NewWithScript` creates a Discord session, loads the JS bot with `jsdiscord.LoadBot`, sets runtime config, and registers gateway handlers.
4. `jsdiscord.NewHost` creates a goja runtime, registers `discord` and `ui` modules, loads the bot script, and compiles the exported bot object.
5. `defineBot` records registered commands/events/components/modals/autocompletes in a `botDraft` and finalizes into a dispatchable bot object.
6. `SyncCommands` calls `Host.ApplicationCommands` to convert JS command descriptors into Discord application commands.
7. Discord interactions/events enter `internal/bot` handlers and are dispatched through `Host.Dispatch*` / `BotHandle.Dispatch*` into JS handlers.

### Discovery rules for repo-driven bots

`botcli.DiscoverBots` scans configured repositories for JavaScript files that look like bot scripts:

- files with `.js` extension;
- either at repository root or named `index.js`;
- source contains `defineBot`;
- source contains `require("discord")` or `require('discord')`.

See [`discoverScriptCandidates`](../../../../../../pkg/botcli/discover.go) and [`looksLikeBotScript`](../../../../../../pkg/botcli/discover.go).

### Examples to copy from

| Example | Why it matters | Files |
|---|---|---|
| Single bot embedding | Minimal Go app using `pkg/framework` directly | [`examples/framework-single-bot/main.go`](../../../../../../examples/framework-single-bot/main.go), [`examples/framework-single-bot/README.md`](../../../../../../examples/framework-single-bot/README.md) |
| Custom native module | How to add `require("app")` from Go | [`examples/framework-custom-module/main.go`](../../../../../../examples/framework-custom-module/main.go), [`examples/framework-custom-module/bot/index.js`](../../../../../../examples/framework-custom-module/bot/index.js) |
| Combined framework + botcli | Downstream app with both built-in explicit bot and repo-driven bots | [`examples/framework-combined/main.go`](../../../../../../examples/framework-combined/main.go), [`examples/framework-combined/README.md`](../../../../../../examples/framework-combined/README.md) |
| UI showcase | Buttons, selects, modals/forms, pagers, cards, aliases | [`examples/discord-bots/ui-showcase/index.js`](../../../../../../examples/discord-bots/ui-showcase/index.js) |
| Show-space bot | Larger practical bot with commands, config, permissions, persistence-style modules, message operations | [`examples/discord-bots/show-space/index.js`](../../../../../../examples/discord-bots/show-space/index.js) |
| Archive helper | Message/thread operations and message commands | [`examples/discord-bots/archive-helper/index.js`](../../../../../../examples/discord-bots/archive-helper/index.js) |

## Usage Examples

### Recommended adventure implementation path

For a first version, implement `examples/discord-bots/adventure/index.js` as a repo-discovered JS bot:

- it matches existing CLI workflows;
- it can be listed/helped/run with `discord-bot bots adventure ...`;
- it can use `ctx.config` for LLM/database settings;
- it can use `component` handlers for choices and `modal` handlers for free-form action entry;
- later, if Go-native LLM/YAML/session services are preferred, inject them as a custom module through `framework.WithRuntimeModuleRegistrars` or `botcli.WithRuntimeModuleRegistrars`.

### Adventure-specific mapping

| Adventure need | Framework feature |
|---|---|
| `/adventure start` | Slash command or subcommand. |
| Choice buttons | `component(customID, handler)` and UI button builders. |
| Free-form action | Button opens modal via `ctx.showModal`; modal handler interprets text. |
| LLM latency | `ctx.defer(...)`, then `ctx.edit(...)` or `ctx.followUp(...)`. |
| Session scope | Use `ctx.user.id`, `ctx.channel.id`, `ctx.guild.id`, and optionally `ctx.discord.threads.start`. |
| YAML state persistence | Start with JS module/file/db storage; move to Go-native `require("adventure")` module if stronger validation is needed. |
| ASCII scene rendering | Send fenced code blocks in message content or embed descriptions. |
| Audit trail | Store validated YAML scene patches and player inputs keyed by session/turn. |

## Related

- Design brainstorm: [`../design/01-initial-brainstorm.md`](../design/01-initial-brainstorm.md)
- Project README: [`README.md`](../../../../../../README.md)
