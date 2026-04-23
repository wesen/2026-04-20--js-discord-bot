# framework-single-bot

Minimal example application showing the new public single-bot embedding path.

It imports `pkg/framework`, selects one explicit bot script, loads Discord credentials from the same environment variables as the CLI, injects runtime config, enables `sync-on-start`, and blocks until you press `Ctrl+C`.

## What it demonstrates

- `framework.New(...)`
- `framework.WithCredentialsFromEnv()`
- `framework.WithScript(...)`
- `framework.WithRuntimeConfig(...)`
- `framework.WithSyncOnStart(true)`
- no repository scanning
- no `botcli` dependency

## Run it

From the repository root:

```bash
export DISCORD_BOT_TOKEN=...
export DISCORD_APPLICATION_ID=...
export DISCORD_GUILD_ID=...

GOWORK=off go run ./examples/framework-single-bot
```

The example uses:

- script: `./examples/discord-bots/unified-demo/index.js`
- runtime config:
  - `db_path=./examples/discord-bots/unified-demo/data/demo.sqlite`
  - `api_key=local-dev-key`

So when the bot becomes ready or when you run `/unified-ping`, you can see that `ctx.config.*` values still reach the JavaScript bot even though the bot was started through the public Go package instead of the repo-driven CLI.

## Adapting it in a downstream app

In a real embedding application you would usually:

- replace the script path with your own bundled bot script
- pass explicit credentials with `framework.WithCredentials(...)` or keep env loading
- replace the demo runtime config with your app's own config values
- attach the bot lifecycle to your app's existing context/shutdown handling
