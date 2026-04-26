---
title: "Section 4: Target Architecture"
description: Detailed design for the reusable package and published binary.
doc_type: design-doc
status: active
topics: [packaging, architecture]
ticket: DISCORD-BOT-PUBLISH
---

## 4. Target Architecture

### 4.1 Reusable Go Package (`pkg/`)

The public Go API surface will consist of three packages. This is already mostly in place; the work is about stabilizing naming, documenting contracts, and ensuring importability.

#### Package 1: `pkg/framework/` — Simple single-bot embedding

**Purpose:** Downstream Go apps that want to embed exactly one Discord bot with minimal code.

**Public types:**

```go
// pkg/framework/framework.go

package framework

// Bot is the public single-bot wrapper.
// Callers create it with New(), then call Open()/Run()/Close().
type Bot struct { ... }

// New creates one explicit bot instance without any repository scanning.
// This is the main entrypoint for embedding.
func New(opts ...Option) (*Bot, error)

// Option configures the framework constructor.
type Option func(*Config) error

// --- Functional options ---

func WithScript(path string) Option
func WithCredentials(credentials Credentials) Option
func WithCredentialsFromEnv() Option
func WithRuntimeConfig(runtimeConfig map[string]any) Option
func WithSyncOnStart(enabled bool) Option
func WithRuntimeModuleRegistrars(registrars ...engine.RuntimeModuleRegistrar) Option

// --- Bot methods ---

func (b *Bot) Open() error       // Open gateway session (optionally sync commands first)
func (b *Bot) Run(ctx context.Context) error  // Open + block until context canceled
func (b *Bot) Close() error      // Close session + JS runtime
func (b *Bot) SyncCommands() error  // Manually sync slash commands

// Credentials holds the Discord credentials needed.
type Credentials struct {
    BotToken      string
    ApplicationID string
    GuildID       string
    PublicKey     string
    ClientID      string
    ClientSecret  string
}
```

**Stability note:** This API is already close to final. The main changes needed are:
- Add `WithVersion(version string)` so the embedding app can inject its own version.
- Consider `WithLogger(logger zerolog.Logger)` for custom log routing.
- Document the `Config` struct fields as part of the public contract.

#### Package 2: `pkg/botcli/` — Repo-driven multi-bot CLI

**Purpose:** Downstream Go apps that want the full `bots list / bots help / bots <name> run` CLI experience inside their own Cobra command tree.

**Public types:**

```go
// pkg/botcli/bootstrap.go

// Bootstrap holds resolved bot repositories ready for scanning.
type Bootstrap struct {
    Repositories []Repository
}

// BuildBootstrap resolves bot repositories using CLI > env > default precedence.
func BuildBootstrap(rawArgs []string, opts ...BuildOption) (Bootstrap, error)

// BuildOption customizes bootstrap construction.
type BuildOption func(*buildOptions) error

func WithWorkingDirectory(dir string) BuildOption
func WithEnvironmentVariable(name string) BuildOption
func WithDefaultRepositories(paths ...string) BuildOption
func WithRepositoryFlagName(name string) BuildOption
```

```go
// pkg/botcli/command_root.go

// NewBotsCommand builds the public repo-driven bot command tree from a bootstrap.
func NewBotsCommand(bootstrap Bootstrap, opts ...CommandOption) (*cobra.Command, error)

// CommandOption customizes the bots command tree.
type CommandOption func(*commandOptions) error

func WithAppName(name string) CommandOption
func WithRuntimeModuleRegistrars(registrars ...engine.RuntimeModuleRegistrar) CommandOption
func WithRuntimeFactory(factory RuntimeFactory) CommandOption
```

**Stability note:** This API is already well-shaped. The customization hooks (`WithAppName`, `WithRuntimeModuleRegistrars`, `WithRuntimeFactory`) follow a deliberate "smallest hook first" pattern documented in `doc.go`.

#### Package 3: `pkg/doc/` — Embedded help documentation

**Purpose:** Help pages shipped inside the binary, accessible via `discord-bot help <topic>`.

This package already exists with `//go:embed` and Glazed's help system. No changes needed beyond content updates.

### 4.2 Standalone Published Binary (`cmd/discord-bot/`)

The standalone binary is the CLI that operators install via Homebrew, deb, or rpm. It uses the same `pkg/` packages internally.

#### What the binary does today (already correct)

```text
discord-bot bots list                              # List all discovered bots
discord-bot bots help <name>                       # Show one bot's metadata
discord-bot bots <name> run --bot-token $TOK ...   # Run one bot
discord-bot run --bot-script ./path.js ...         # Run an explicit script
discord-bot validate-config ...                    # Validate credentials
discord-bot sync-commands ...                      # Sync slash commands
discord-bot help <topic>                           # Show embedded help
```

#### What needs to change

1. **Version injection.** Add `var version = "dev"` to `main.go`. GoReleaser injects the real version via ldflags.

2. **Module path in help.** The root command `Use` field should say `discord-bot`, not the current development path.

3. **GOWORK=off removal.** Once the module is properly published, `GOWORK=off` is no longer needed (but keeping it is harmless for local development).

#### Version injection pattern (from pinocchio)

```go
// cmd/discord-bot/main.go

package main

var version = "dev"  // Injected by GoReleaser: -ldflags -X main.version=<tag>

func main() {
    rootCmd.Version = version
    // ... rest of initialization
}
```

GoReleaser config handles the injection automatically:

```yaml
# .goreleaser.yaml (implicit via build flags)
builds:
  - flags:
      - -trimpath
    ldflags:
      - -X main.version={{.Version}}
```

### 4.3 JavaScript Bot Authoring API

The JavaScript API (`require("discord")`) is internal to the runtime. It does **not** need to change for packaging. However, its stability is important because downstream bots depend on it.

**Current API surface (stable, do not break):**

```javascript
// Registration
defineBot(({ command, event, component, modal, autocomplete, configure }) => { ... })

// Response helpers (on ctx)
ctx.reply({ content, embeds, components, files, ephemeral })
ctx.defer({ ephemeral })
ctx.edit({ content, embeds, components })
ctx.followUp({ content, embeds, components, files, ephemeral })
ctx.showModal({ customId, title, components })

// Logging (on ctx)
ctx.log.info(msg, data)
ctx.log.warn(msg, data)
ctx.log.error(msg, data)
ctx.log.debug(msg, data)

// Store (on ctx)
ctx.store.get(key)
ctx.store.set(key, value)
ctx.store.delete(key)
ctx.store.keys()
ctx.store.has(key)

// Discord operations (on ctx.discord)
ctx.discord.channels.send(channelId, { content, embeds })
ctx.discord.messages.fetch(channelId, messageId)
ctx.discord.messages.edit(channelId, messageId, { content })
ctx.discord.messages.delete(channelId, messageId)
ctx.discord.members.addRole(guildId, userId, roleId)
ctx.discord.members.removeRole(guildId, userId, roleId)
ctx.discord.guilds.fetch(guildId)
ctx.discord.roles.list(guildId)
ctx.discord.threads.start(channelId, { name, message })
ctx.discord.threads.archive(channelId, threadId)

// Snapshots (read-only Discord objects)
ctx.me         // Current bot user
ctx.guild      // Guild snapshot (in guild context)
ctx.channel    // Channel snapshot
ctx.member     // Member snapshot (in member context)
```

### 4.4 Repository Layout and Module Naming

#### Target module path

```
github.com/go-go-golems/discord-bot
```

This follows the go-go-golems convention: `github.com/go-go-golems/<binary-name>`.

#### Final repository layout

```text
discord-bot/
  go.mod                        # module github.com/go-go-golems/discord-bot
  go.sum
  .goreleaser.yaml              # From go-template, adapted
  .golangci.yml                 # From go-template
  .golangci-lint-version        # From go-template
  Makefile                      # From go-template, adapted
  lefthook.yml                  # From go-template
  LICENSE                       # MIT
  README.md                     # Public-facing README
  AGENT.md                      # AI agent instructions

  cmd/
    discord-bot/
      main.go                   # CLI entrypoint with version injection
      root.go                   # Root command wiring
      commands.go               # Subcommand definitions

  internal/
    bot/                        # Discord session wrapper
    config/                     # Host config
    jsdiscord/                  # Embedded JS runtime engine

  pkg/
    framework/                  # Public: simple single-bot embedding
    botcli/                     # Public: repo-driven multi-bot CLI
    doc/                        # Public: embedded help pages

  examples/
    discord-bots/               # Named JS bot implementations
    framework-single-bot/       # Embedding example
    framework-custom-module/    # Custom module example
    framework-combined/         # Combined example

  .github/
    workflows/
      release.yaml              # Split GoReleaser
      push.yml                  # CI on push
      lint.yml                  # Lint
      codeql-analysis.yml       # Security
      secret-scanning.yml       # Secrets
      dependency-scanning.yml   # Dependencies
```

#### What changes from the current layout

Almost nothing structurally changes. The repo layout is already correct. The changes are:

1. `go.mod` module path changes from `github.com/manuel/wesen/2026-04-20--js-discord-bot` to `github.com/go-go-golems/discord-bot`.
2. All import paths in every `.go` file update accordingly.
3. The `replace` directive for `go-go-goja` is removed (or kept temporarily during transition).
4. Infrastructure files (Makefile, .goreleaser.yaml, etc.) are added.
5. `cmd/discord-bot/main.go` gets version injection.
