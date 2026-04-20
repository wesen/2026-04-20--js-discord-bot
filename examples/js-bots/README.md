# JavaScript Discord bot examples

These scripts are loaded by the local `internal/jsdiscord` host layer via:

```js
const { defineBot } = require("discord")
```

## Example

- `ping.js` — defines `ping` and `echo` slash commands plus a `ready` event handler.

## Intended runtime usage

Set the script path in your environment or via CLI flags:

```bash
export DISCORD_BOT_SCRIPT=./examples/js-bots/ping.js
GOWORK=off go run ./cmd/discord-bot sync-commands
GOWORK=off go run ./cmd/discord-bot run
```
