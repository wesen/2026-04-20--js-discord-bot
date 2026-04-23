# framework-combined

Example downstream-style application showing both public layers together:

- one explicit built-in bot started through `pkg/framework`
- one repo-driven `bots` subtree mounted through `pkg/botcli`

## What it demonstrates

- `framework.New(...)` for the simplest built-in bot path
- `botcli.BuildBootstrap(...)` for raw-argv repository selection
- `botcli.NewBotsCommand(...)` for mounting repo-driven bots into an existing Cobra root
- the current public `botcli` customization surface:
  - `botcli.WithAppName(...)`
  - `botcli.WithRuntimeModuleRegistrars(...)`
  - `botcli.WithRuntimeFactory(...)`
- the recommended split:
  - core framework = one explicit bot is easy
  - optional `botcli` = repository-driven multi-bot workflows are easy

## File layout

- `main.go` — downstream app root command
- `builtin-bot/index.js` — explicit built-in bot script used by `run-builtin`

## Commands

### Run the built-in bot

From the repository root:

```bash
export DISCORD_BOT_TOKEN=...
export DISCORD_APPLICATION_ID=...
export DISCORD_GUILD_ID=...

GOWORK=off go run ./examples/framework-combined run-builtin
```

This path does **not** use repository discovery. It starts:

- `./examples/framework-combined/builtin-bot/index.js`

through `pkg/framework` directly.

### Use repo-driven bots through the same downstream app

```bash
GOWORK=off go run ./examples/framework-combined bots list --output json
GOWORK=off go run ./examples/framework-combined bots help unified-demo --output json
GOWORK=off go run ./examples/framework-combined bots knowledge-base run --help
```

By default the example uses:

- `./examples/discord-bots`

as the repo-driven bot repository, but you can still override it with the public bootstrap flag:

```bash
GOWORK=off go run ./examples/framework-combined \
  --bot-repository ./pkg/botcli/testdata/scanner-repo \
  bots demo-bot status --output json
```

## Why this example exists

The earlier examples each covered one public layer in isolation:

- `framework-single-bot` — one explicit bot via `pkg/framework`
- `framework-custom-module` — one explicit bot plus custom runtime module injection

This example shows the next level up: a downstream application can combine both public packages in one process and choose between the simple built-in bot path and the optional repo-driven `botcli` path.

If you need deeper runtime control on the repo-driven side, choose the smallest hook that fits:
- `botcli.WithAppName(...)` when only the dynamic env prefix should change
- `botcli.WithRuntimeModuleRegistrars(...)` when extra Go-native `require()` modules are enough and the default runtime construction is still correct
- `botcli.WithRuntimeFactory(...)` only when runtime creation itself must change, for example custom module roots, require behavior, builder/runtime setup, or runtime lifecycle details
- if that same customization should also affect discovery and host-managed runs, make the runtime factory implement `botcli.HostOptionsProvider`
