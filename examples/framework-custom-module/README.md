# framework-custom-module

Example application showing the next Track A embedding seam: one explicit bot plus a custom Go-native module exposed to JavaScript through `require()`.

## What it demonstrates

- `framework.New(...)`
- `framework.WithScript(...)`
- `framework.WithCredentialsFromEnv()`
- `framework.WithRuntimeModuleRegistrars(...)`
- one explicit bot script with no repository scanning
- a custom runtime-scoped native module named `app`

## File layout

- `main.go` — embedding app that registers the custom Go module
- `bot/index.js` — explicit JavaScript bot script that calls `require("app")`

## Run it

From the repository root:

```bash
export DISCORD_BOT_TOKEN=...
export DISCORD_APPLICATION_ID=...
export DISCORD_GUILD_ID=...

GOWORK=off go run ./examples/framework-custom-module
```

The example loads `./examples/framework-custom-module/bot/index.js`, injects a Go-native `app` module, syncs commands on startup, and then opens the gateway.

Once the bot is connected:

- the ready event logs metadata provided by `require("app")`
- `/app-info` replies using the value returned from the custom Go module

## Why this example exists

The minimal `framework-single-bot` example proves the public package can run one explicit bot script. This example proves the next thing Track A needs: a downstream Go app can still extend the runtime with its own module surface without adopting repo-driven `botcli` discovery.
