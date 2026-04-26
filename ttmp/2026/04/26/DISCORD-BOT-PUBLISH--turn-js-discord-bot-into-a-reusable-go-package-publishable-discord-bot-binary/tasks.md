# Tasks

## Phase 1: Rename and Reparent

- [x] Update go.mod module path to `github.com/go-go-golems/discord-bot`
- [x] Replace all import paths in Go source files (use `scripts/02-rename-module-path.sh`)
- [x] Remove or gate the local `replace` directive for go-go-goja
- [x] Run `go mod tidy && go build ./... && go test ./...`
- [x] Create GitHub repository `go-go-golems/discord-bot`
- [x] Push to GitHub

## Phase 2: Extract Public API Surface

- [ ] Audit `pkg/framework/` exported types for naming and docs
- [ ] Audit `pkg/botcli/` exported types for naming and docs
- [ ] Add doc comments to all exported types/functions
- [ ] Verify examples compile (`go build ./examples/...`)

## Phase 3: Infrastructure from go-template

- [ ] Copy `.golangci.yml`, `.golangci-lint-version`, `lefthook.yml`, `LICENSE` (use `scripts/03-copy-infrastructure.sh`)
- [ ] Create `Makefile` adapted for discord-bot
- [ ] Create `.goreleaser.yaml` adapted for discord-bot
- [ ] Copy `.github/workflows/` from go-template
- [ ] Install lefthook and verify hooks work
- [ ] Run `make lint && make test && make build`

## Phase 4: Version Injection and Entry Point Cleanup

- [ ] Add `var version = "dev"` to `cmd/discord-bot/main.go`
- [ ] Wire version into root command
- [ ] Update `.goreleaser.yaml` ldflags
- [ ] Verify version injection with manual build

## Phase 5: CI and Publishing

- [ ] Configure GitHub secrets (GPG, GoReleaser, Homebrew, fury.io)
- [ ] Create first tag (`v0.1.0`)
- [ ] Push tag and watch release workflow
- [ ] Approve release in GitHub UI
- [ ] Verify Homebrew formula, deb/rpm packages

## Phase 6: Polish and Documentation

- [ ] Update README.md with installation instructions and Go API examples
- [ ] Create AGENT.md
- [ ] Review embedded help pages
- [ ] Add Dependabot config
- [ ] Final validation: `make lint && make test && make build`

