# Tasks

## TODO

### 1. Polyfill `__verb__` in Discord runtime
- [ ] Add no-op `__package__`, `__section__`, `__verb__`, `doc` to `internal/jsdiscord/runtime.go`
- [ ] Test: script with `__verb__` loads without crashing

### 2. Scan bot repos with jsverbs
- [ ] Add `go-go-goja` dependency
- [ ] Create `internal/botcli/jsverbs_scan.go` — wraps `jsverbs.ScanDir` for bot repos
- [ ] Test: discovers `__verb__("run")` and `__verb__("status")` in sample scripts

### 3. Build `botRunCommand` (BareCommand)
- [ ] Create `internal/botcli/bot_run_command.go`
  - `Run(ctx, parsedValues)` extracts Discord creds + runtime config
  - `runtimeFieldInternalName()` converts kebab→snake_case
  - Creates bot, calls `SetRuntimeConfig()`, `Open()`, blocks on `<-ctx.Done()`
- [ ] Create `internal/botcli/field_name.go` with `runtimeFieldInternalName()`
- [ ] Test: config values flow from CLI flags → `ctx.config` in JS

### 4. Wire commands into CLI
- [ ] Replace raw cobra `bots list` with `listBotsCommand` (GlazeCommand)
- [ ] Replace raw cobra `bots help` with `helpBotsCommand` (GlazeCommand)
- [ ] Register `botRunCommand` for each `__verb__("run")` via `BuildCobraCommandFromCommand`
- [ ] Register standard jsverbs for non-`run` verbs
- [ ] Delete `run_static_args.go`, `run_dynamic_schema.go`, `run_help.go`

### 5. Example unified bot script
- [ ] Create `examples/discord-bots/unified-demo/index.js`
  - `defineBot` for Discord behavior
  - `__verb__("run")` with fields: `bot-token`, `api-key`, `db-path`
  - `__verb__("status")` for CLI metadata
  - Demonstrates `ctx.config` access

### 6. Test & validate
- [ ] `go test ./internal/botcli/...`
- [ ] `go run ./cmd/discord-bot bots list --output json`
- [ ] `go run ./cmd/discord-bot bots help unified-demo --output json`
- [ ] `go run ./cmd/discord-bot bots run unified-demo --help` (shows config flags)
- [ ] `discord-bot run` / `validate-config` / `sync-commands` still work

### 7. Docs
- [ ] Update `examples/discord-bots/README.md` with `__verb__` syntax
- [ ] Update ticket design doc with any changes

## Done
- [x] Created ticket and architecture analysis
- [x] Documented `BareCommand` approach + `ctx.config` bridge
