# JavaScript Discord bot examples

These scripts are loaded by the local `internal/jsdiscord` host layer via:

```js
const { defineBot } = require("discord")
```

## Example

- `ping.js` — defines `ping` and `echo` slash commands, uses embeds/components, demonstrates deferred edit + follow-up responses, and listens for `ready`, `guildCreate`, and `messageCreate` events.

## Intended runtime usage

Set the script path in your environment or via CLI flags:

```bash
export DISCORD_BOT_SCRIPT=./examples/js-bots/ping.js
GOWORK=off go run ./cmd/discord-bot sync-commands
GOWORK=off go run ./cmd/discord-bot run
```

## Runtime notes

- `/ping` returns a richer payload with an embed and a link button.
- `/echo` demonstrates `ctx.defer(...)`, `ctx.edit(...)`, and `ctx.followUp(...)`.
- Sending `!pingjs` in a guild channel exercises the `messageCreate` event path.
