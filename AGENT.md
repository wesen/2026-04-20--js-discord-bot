# AGENT.md

## Project Overview

discord-bot is a Go-hosted Discord bot runtime with a JavaScript authoring API. Go owns the Discord gateway/session (via discordgo) and embeds a JavaScript runtime (goja) that runs bot scripts via `require("discord")`.

## Module

`github.com/go-go-golems/discord-bot`

## Key Commands

- `make lint` — run golangci-lint
- `make test` — run all tests
- `make build` — build the binary
- `make goreleaser` — snapshot release (local, no publish)
- `make tag-patch && git push upstream --tags` — create and push a new release tag

## Architecture

- `cmd/discord-bot/` — CLI entrypoint (Cobra + Glazed)
- `internal/bot/` — Discordgo session wrapper, event handlers
- `internal/config/` — Host config (credentials, validation)
- `internal/jsdiscord/` — Embedded JS runtime engine, `require("discord")` module, dispatch, payloads
- `pkg/framework/` — Public simple single-bot embedding API
- `pkg/botcli/` — Public repo-driven multi-bot CLI layer
- `pkg/doc/` — Embedded help pages (`//go:embed`)
- `examples/discord-bots/` — Named JS bot implementations (also serve as integration tests)

## Key Patterns

- One JS bot per process (intentional design choice)
- JS API is request-scoped (`ctx.reply`, `ctx.discord.*`, `ctx.store.*`)
- Public API uses functional options (`framework.WithScript(...)`, `botcli.WithAppName(...)`)
- Bot discovery: CLI flag > env var > default path

## Testing

```bash
make test                    # All packages
go test ./internal/jsdiscord # Runtime engine tests
go test ./pkg/botcli         # CLI layer tests
go test ./pkg/framework      # Embedding tests
```

## Version

Version is injected via ldflags: `var version = "dev"` in `cmd/discord-bot/main.go`, set by GoReleaser.

## Known Issues

- ~30 pre-existing lint issues in `internal/jsdiscord/` (unused functions, exhaustive switches). These are tracked separately.
- Lint hooks are non-blocking (`|| true` in lefthook.yml) until the lint debt is cleaned up.
