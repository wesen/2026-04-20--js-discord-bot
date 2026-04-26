---
title: "Section 6: Detailed Implementation Phases"
description: Step-by-step phased implementation plan.
doc_type: design-doc
status: active
topics: [packaging, implementation]
ticket: DISCORD-BOT-PUBLISH
---

## 6. Detailed Implementation Phases

### Phase 1: Rename and Reparent

**Goal:** Change the module path from the local development name to the go-go-golems canonical path.

**Why first:** Everything else depends on having a stable, importable module path. Until this changes, no other tooling (GoReleaser, CI, Homebrew) can work correctly.

**Steps:**

1. **Update `go.mod`:**
   ```bash
   # Old
   module github.com/manuel/wesen/2026-04-20--js-discord-bot
   # New
   module github.com/go-go-golems/discord-bot
   ```

2. **Update all import paths in every .go file.** There are approximately 30+ Go files that import from the old module path. Use a mechanical find-and-replace:

   ```bash
   # Dry run first
   find . -name '*.go' -not -path './.git/*' -not -path './ttmp/*' \
     | xargs grep -l 'github.com/manuel/wesen/2026-04-20--js-discord-bot'

   # Replace (use sed or manual editing)
   find . -name '*.go' -not -path './.git/*' -not -path './ttmp/*' \
     | xargs sed -i 's|github.com/manuel/wesen/2026-04-20--js-discord-bot|github.com/go-go-golems/discord-bot|g'
   ```

   Files that will be affected (evidence from grep):
   - `cmd/discord-bot/root.go`
   - `cmd/discord-bot/commands.go`
   - `internal/bot/bot.go`
   - `internal/config/config.go`
   - `internal/jsdiscord/*.go` (all files)
   - `pkg/framework/framework.go`
   - `pkg/botcli/*.go` (all files)
   - `examples/framework-*/main.go` (all three examples)
   - All test files

3. **Remove or gate the local `replace` directive:**
   ```bash
   # In go.mod, remove:
   replace github.com/go-go-golems/go-go-goja => /home/manuel/code/wesen/corporate-headquarters/go-go-goja

   # Then:
   go mod tidy
   ```

   **Prerequisite:** go-go-goja must already be published with the discord registrar support. If it is not yet published, keep the replace directive temporarily and add a TODO comment.

4. **Verify:**
   ```bash
   go build ./...
   go test ./...
   ```

5. **Create the GitHub repository:**
   ```bash
   gh repo create go-go-golems/discord-bot --public
   git remote add origin git@github.com:go-go-golems/discord-bot.git
   git push -u origin main
   ```

6. **Commit:**
   ```bash
   git add -A
   git commit -m "chore: rename module to github.com/go-go-golems/discord-bot"
   ```

**Estimated effort:** 1-2 hours (mechanical, but verify carefully).

### Phase 2: Extract Public API Surface

**Goal:** Review and stabilize `pkg/framework/` and `pkg/botcli/` as the public Go API.

**Why:** Before publishing, the public API needs to be reviewed for naming consistency, documentation, and completeness.

**Steps:**

1. **Audit `pkg/framework/`:**
   - Read every exported type and function.
   - Ensure doc comments follow Go conventions (start with the name being documented).
   - Verify that the `Option` pattern is consistent with other go-go-golems packages.
   - Add missing doc comments.

2. **Audit `pkg/botcli/`:**
   - Read every exported type and function.
   - Verify `Bootstrap`, `Repository`, `DiscoveredBot` types have stable fields.
   - Ensure `CommandOption` hooks are well-documented.
   - Verify the "smallest hook first" guidance in `doc.go` is accurate.

3. **Add `pkg/doc.go` for the root package:**
   ```go
   // Package discordbot provides a Go-hosted Discord bot runtime with
   // a JavaScript authoring API.
   //
   // The main entrypoints are:
   //   - pkg/framework: Simple single-bot embedding
   //   - pkg/botcli: Repo-driven multi-bot CLI
   //
   // See the README for a quick start guide.
   package discordbot
   ```

4. **Review the examples:**
   - Verify `examples/framework-single-bot/main.go` compiles and is minimal.
   - Verify `examples/framework-custom-module/main.go` shows the custom module pattern clearly.
   - Verify `examples/framework-combined/main.go` shows both embedding paths.

5. **Verify:**
   ```bash
   go vet ./pkg/...
   go test ./pkg/...
   # Test that examples compile
   go build ./examples/...
   ```

**Estimated effort:** 2-3 hours.

### Phase 3: Infrastructure from go-template

**Goal:** Add all infrastructure files adapted from go-template.

**Steps:**

1. **Copy these files from go-template:**
   ```bash
   SRC=~/code/wesen/corporate-headquarters/go-template

   cp "$SRC/.golangci.yml" .
   cp "$SRC/.golangci-lint-version" .
   cp "$SRC/lefthook.yml" .
   cp "$SRC/LICENSE" .
   cp "$SRC/.github/workflows/release.yaml" .github/workflows/
   cp "$SRC/.github/workflows/push.yml" .github/workflows/
   cp "$SRC/.github/workflows/lint.yml" .github/workflows/
   cp "$SRC/.github/workflows/codeql-analysis.yml" .github/workflows/
   cp "$SRC/.github/workflows/secret-scanning.yml" .github/workflows/
   cp "$SRC/.github/workflows/dependency-scanning.yml" .github/workflows/
   ```

2. **Create the Makefile** (adapted from go-template, as specified in Section 5.1).

3. **Create .goreleaser.yaml** (as specified in Section 5.2).

4. **Create .github/workflows/ directory** if it doesn't exist:
   ```bash
   mkdir -p .github/workflows
   ```

5. **Install lefthook:**
   ```bash
   go install github.com/evilmartians/lefthook@latest
   lefthook install
   ```

6. **Verify everything locally:**
   ```bash
   make lint
   make test
   make build
   make goreleaser  # snapshot build
   ```

7. **Commit:**
   ```bash
   git add Makefile .goreleaser.yaml .golangci.yml .golangci-lint-version \
           lefthook.yml LICENSE .github/
   git commit -m "chore: add infrastructure from go-template (Makefile, GoReleaser, CI, lint)"
   ```

**Estimated effort:** 1-2 hours.

### Phase 4: Version Injection and Entry Point Cleanup

**Goal:** Wire version injection into the binary entry point.

**Steps:**

1. **Add version variable to cmd/discord-bot/main.go:**
   ```go
   var version = "dev"
   ```

2. **Wire it into the root command:**
   ```go
   func init() {
       rootCmd.Version = version
   }
   ```

3. **Update .goreleaser.yaml** to inject the version via ldflags (already specified in Section 5.2):
   ```yaml
   builds:
     - ldflags:
         - -X main.version={{.Version}}
   ```

4. **Verify locally:**
   ```bash
   go build -ldflags "-X main.version=test-0.0.1" ./cmd/discord-bot
   ./discord-bot --version
   # Should print: test-0.0.1
   ```

5. **Commit:**
   ```bash
   git commit -m "feat: add version injection for GoReleaser"
   ```

**Estimated effort:** 30 minutes.

### Phase 5: CI and Publishing

**Goal:** Push to GitHub, set up secrets, run the first release.

**Steps:**

1. **Push the repository:**
   ```bash
   git push origin main
   ```

2. **Configure GitHub secrets** (in the repository settings):

   | Secret | Purpose | Where to get it |
   |--------|---------|-----------------|
   | `GITHUB_TOKEN` | Automatic (GitHub provides) | N/A |
   | `GORELEASER_KEY` | GoReleaser Pro license | From existing org secrets |
   | `HOMEBREW_TAP_TOKEN` | Push to homebrew-go-go-go | From existing org secrets |
   | `GO_GO_GOLEMS_SIGN_KEY` | GPG signing key | From existing org secrets |
   | `GO_GO_GOLEMS_SIGN_PASSPHRASE` | GPG passphrase | From existing org secrets |
   | `FURY_TOKEN` | Push to fury.io | From existing org secrets |

   **Note:** Most of these secrets likely already exist at the org level and will be inherited automatically.

3. **Create the first tag:**
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```

4. **Watch the release workflow:**
   ```bash
   gh run list --workflow=release.yaml
   gh run watch
   ```

5. **Approve the release** in the GitHub UI (the `environment: release` gate requires manual approval).

6. **Verify the release:**
   ```bash
   # Check GitHub releases
   gh release view v0.1.0

   # Check Homebrew (after a few minutes)
   brew info go-go-golems/tap/discord-bot

   # Install and test
   brew install go-go-golems/tap/discord-bot
   discord-bot --version
   ```

**Estimated effort:** 1-2 hours (mostly waiting for CI).

### Phase 6: Polish and Documentation

**Goal:** Final polish on README, AGENT.md, and public documentation.

**Steps:**

1. **Update README.md:**
   - Change all `GOWORK=off go run ./cmd/discord-bot` examples to just `discord-bot`.
   - Add installation instructions (Homebrew, deb, rpm).
   - Add a "Go API" section showing the embedding examples.
   - Update the module path in any code examples.
   - Add a badge or two (build status, go report).

2. **Create AGENT.md** (from go-template, adapted):
   ```markdown
   # AGENT.md

   ## Project Overview
   discord-bot is a Go-hosted Discord bot runtime with a JavaScript authoring API.

   ## Key Commands
   - `make lint` — run golangci-lint
   - `make test` — run tests
   - `make build` — build the binary
   - `make goreleaser` — snapshot release
   - `make tag-patch && git push origin --tags` — create a release

   ## Architecture
   See design docs in ttmp/.
   ```

3. **Review pkg/doc/ help pages:**
   - Ensure `discord-bot help discord-js-bot-api-reference` still works.
   - Ensure `discord-bot help build-and-run-discord-js-bots` still works.

4. **Final verification:**
   ```bash
   make lint && make test && make build
   ./dist/discord-bot --version
   ./dist/discord-bot bots list --bot-repository ./examples/discord-bots
   ```

**Estimated effort:** 2-3 hours.
