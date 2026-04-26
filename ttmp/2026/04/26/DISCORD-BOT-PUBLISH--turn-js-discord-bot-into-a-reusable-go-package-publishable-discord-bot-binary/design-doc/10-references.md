---
title: "Section 10: References"
description: File reference index for all key files across the three codebases.
doc_type: design-doc
status: active
topics: [packaging]
ticket: DISCORD-BOT-PUBLISH
---

## 10. References

### 10.1 Key files in js-discord-bot (the prototype)

| File | Lines | Purpose |
|------|-------|---------|
| `go.mod` | ~150 | Module definition, dependencies, local replace |
| `cmd/discord-bot/main.go` | ~30 | CLI entrypoint |
| `cmd/discord-bot/root.go` | ~60 | Root Cobra command wiring |
| `cmd/discord-bot/commands.go` | 246 | Subcommand definitions (bots, run, validate-config, sync-commands) |
| `internal/bot/bot.go` | 313 | Discord session wrapper, event handlers |
| `internal/config/config.go` | 51 | Settings struct, validation |
| `internal/jsdiscord/host.go` | ~100 | JS runtime host lifecycle |
| `internal/jsdiscord/runtime.go` | ~110 | require("discord") module registration |
| `internal/jsdiscord/descriptor.go` | 367 | Bot/command/event/component descriptors |
| `internal/jsdiscord/bot_compile.go` | 728 | JS defineBot() parser |
| `internal/jsdiscord/host_dispatch.go` | 535 | Event dispatch to JS handlers |
| `internal/jsdiscord/payload_model.go` | 445 | Response payload types |
| `pkg/framework/framework.go` | ~200 | Public single-bot embedding API |
| `pkg/botcli/bootstrap.go` | ~180 | Repository resolution (CLI/env/defaults) |
| `pkg/botcli/command_root.go` | ~100 | Cobra command tree builder |
| `pkg/botcli/command_run.go` | ~50 | Bot run command implementation |
| `pkg/botcli/options.go` | ~100 | RuntimeFactory, CommandOption hooks |
| `pkg/botcli/doc.go` | ~20 | Package documentation |
| `README.md` | ~250 | Project documentation |

### 10.2 Key files in go-template (the skeleton)

| File | Purpose |
|------|---------|
| `go.mod` | `module github.com/go-go-golems/XXX` |
| `cmd/XXX/main.go` | Empty main() |
| `.goreleaser.yaml` | Full GoReleaser config |
| `.golangci.yml` | Linter configuration |
| `.golangci-lint-version` | Pinned linter version |
| `Makefile` | All build/lint/test/release targets |
| `lefthook.yml` | Pre-commit and pre-push hooks |
| `.github/workflows/release.yaml` | Split GoReleaser release pipeline |
| `.github/workflows/push.yml` | CI on push/PR |
| `.github/workflows/lint.yml` | Lint workflow |
| `.github/workflows/codeql-analysis.yml` | Security scanning |
| `.github/workflows/secret-scanning.yml` | Secret scanning |
| `.github/workflows/dependency-scanning.yml` | Dependency scanning |
| `AGENT.md` | AI agent instructions |
| `LICENSE` | MIT license |

### 10.3 Key files in pinocchio (the reference)

| File | Purpose |
|------|---------|
| `go.mod` | `module github.com/go-go-golems/pinocchio` |
| `cmd/pinocchio/main.go` | Full CLI entrypoint with version injection, embed FS, help system |
| `cmd/pinocchio/doc/doc.go` | `//go:embed *` for embedded help pages |
| `.goreleaser.yaml` | Full GoReleaser config (linux+darwin, brew, deb, rpm, fury) |
| `Makefile` | Full targets including proto-gen, geppetto-lint, web checks |
| `lefthook.yml` | Pre-commit + pre-push hooks |
| `.golangci.yml` | Custom lint rules |
| `.github/workflows/release.yml` | Split build + merge + sign |
| `pkg/cmds/` | Command loading infrastructure |
| `pkg/doc/` | Embedded help documentation |

### 10.4 External references

- **Go module publishing:** https://go.dev/doc/modules/publishing
- **GoReleaser documentation:** https://goreleaser.com/
- **golangci-lint configuration:** https://golangci-lint.run/usage/configuration/
- **Lefthook documentation:** https://github.com/evilmartians/lefthook/blob/master/docs/full_guide.md
- **svu (semantic versioning utility):** https://github.com/caarlos0/svu
- **discordgo documentation:** https://pkg.go.dev/github.com/bwmarrin/discordgo
- **goja (JavaScript runtime for Go):** https://github.com/dop251/goja
- **Glazed CLI framework:** https://github.com/go-go-golems/glazed
