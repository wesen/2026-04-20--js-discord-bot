---
title: "Section 5: Infrastructure and CI/CD"
description: Makefile, GoReleaser, GitHub Actions, linting setup.
doc_type: design-doc
status: active
topics: [packaging, ci, infrastructure]
ticket: DISCORD-BOT-PUBLISH
---

## 5. Infrastructure and CI/CD

This section specifies every infrastructure file that needs to be created or adapted. For each file, we show: where to copy it from, what to change, and what the final result looks like.

### 5.1 Makefile

**Source:** Copy from `go-template/Makefile`, then adapt.

**Changes from go-template:**

1. Replace `XXX` with `discord-bot` everywhere.
2. Replace `./cmd/XXX` with `./cmd/discord-bot`.
3. Remove `GOWORK=off` from all targets (or keep it — it's harmless and useful during transition).
4. Add the `GOLANGCI_LINT_ARGS` to include `./internal/...` and `./pkg/...`.
5. Remove the `gifs` target (not applicable).
6. Keep the `install` target but point it at `discord-bot`.

**Final Makefile (key sections):**

```makefile
.PHONY: all test build lint lintmax golangci-lint-install gosec govulncheck \
        goreleaser tag-major tag-minor tag-patch release bump-glazed install

all: test build

VERSION=v0.0.1
GORELEASER_ARGS ?= --skip=sign --snapshot --clean
GORELEASER_TARGET ?= --single-target
GOLANGCI_LINT_VERSION ?= $(shell cat .golangci-lint-version)
GOLANGCI_LINT_BIN ?= $(CURDIR)/.bin/golangci-lint
GOLANGCI_LINT_ARGS ?= --timeout=5m ./cmd/... ./pkg/... ./internal/...

golangci-lint-install:
	mkdir -p $(dir $(GOLANGCI_LINT_BIN))
	GOBIN=$(dir $(GOLANGCI_LINT_BIN)) go install \
	  github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

lint: golangci-lint-install
	$(GOLANGCI_LINT_BIN) config verify
	$(GOLANGCI_LINT_BIN) run -v $(GOLANGCI_LINT_ARGS)

lintmax: golangci-lint-install
	$(GOLANGCI_LINT_BIN) run -v --max-same-issues=100 $(GOLANGCI_LINT_ARGS)

test:
	go test ./...

build:
	go generate ./...
	go build ./...

goreleaser:
	goreleaser release $(GORELEASER_ARGS) $(GORELEASER_TARGET)

tag-major:
	git tag $(shell svu major)
tag-minor:
	git tag $(shell svu minor)
tag-patch:
	git tag $(shell svu patch)

release:
	git push origin --tags
	GOPROXY=proxy.golang.org go list -m \
	  github.com/go-go-golems/discord-bot@$(shell svu current)

discord-bot_BINARY=$(shell which discord-bot)
install:
	go build -o ./dist/discord-bot ./cmd/discord-bot && \
	  cp ./dist/discord-bot $(discord-bot_BINARY)
```

### 5.2 GoReleaser

**Source:** Copy from `go-template/.goreleaser.yaml`, then adapt.

**Changes from go-template:**

1. Replace `XXX` with `discord-bot`.
2. Replace `./cmd/XXX` with `./cmd/discord-bot`.
3. Update description and homepage URLs.
4. Keep the same build matrix (linux amd64/arm64, darwin amd64/arm64).
5. Keep CGO enabled (goja and go-sqlite3 require CGO).

**Final .goreleaser.yaml (key sections):**

```yaml
version: 2
project_name: discord-bot

before:
  hooks:
    - go mod tidy
    - go generate ./...

builds:
  - id: discord-bot-linux
    env:
      - CGO_ENABLED=1
      - CC_linux_amd64=gcc
      - CXX_linux_amd64=g++
      - CC_linux_arm64=aarch64-linux-gnu-gcc
      - CXX_linux_arm64=aarch64-linux-gnu-g++
    main: ./cmd/discord-bot
    binary: discord-bot
    goos:
      - linux
    goarch:
      - amd64
      - arm64
    flags:
      - -trimpath
    ldflags:
      - -X main.version={{.Version}}

  - id: discord-bot-darwin
    env:
      - CGO_ENABLED=1
    main: ./cmd/discord-bot
    binary: discord-bot
    goos:
      - darwin
    goarch:
      - amd64
      - arm64
    flags:
      - -trimpath
    ldflags:
      - -X main.version={{.Version}}

checksum:
  name_template: 'checksums.txt'

snapshot:
  name_template: "{{ incpatch .Version }}-next"

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'

signs:
  - artifacts: checksum
    args:
      [
        "--batch",
        "-u",
        "{{ .Env.GPG_FINGERPRINT }}",
        "--output",
        "${signature}",
        "--detach-sign",
        "${artifact}",
      ]

brews:
  - name: discord-bot
    description: "discord-bot is a Go-hosted Discord bot runtime with JavaScript bot authoring"
    homepage: "https://github.com/go-go-golems/discord-bot"
    repository:
      owner: go-go-golems
      name: homebrew-go-go-go
      token: "{{ .Env.TAP_GITHUB_TOKEN }}"

nfpms:
  - id: packages
    vendor: GO GO GOLEMS
    homepage: https://github.com/go-go-golems/
    maintainer: Manuel Odendahl <wesen@ruinwesen.com>
    description: |
      discord-bot is a Go-hosted Discord bot runtime with a local
      JavaScript bot API. Bots are authored in JS and run inside an
      embedded goja runtime hosted by Go.
    license: MIT
    formats:
      - deb
      - rpm
    release: "1"
    section: default
    priority: extra
    deb:
      lintian_overrides:
        - statically-linked-binary
        - changelog-file-missing-in-native-package

publishers:
  - name: fury.io
    ids:
      - packages
    dir: "{{ dir .ArtifactPath }}"
    cmd: curl -F package=@{{ .ArtifactName }} https://{{ .Env.FURY_TOKEN }}@push.fury.io/go-go-golems/
```

### 5.3 GitHub Actions Workflows

**Source:** Copy all `.github/workflows/` from `go-template`, adapt naming.

#### release.yaml

Copy from go-template. Changes:
- Replace `XXX` references with `discord-bot`.
- Everything else (split build, merge, GPG sign, brew, fury) stays identical.

```yaml
# Key pattern: split build + merge
on:
  workflow_dispatch:
  push:
    tags: ['*']

jobs:
  goreleaser-linux:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
        with: { fetch-depth: 0 }
      - run: git fetch --force --tags
      - uses: actions/setup-go@v6
        with: { go-version-file: go.mod, cache: true }
      - run: |
          sudo apt-get update
          sudo apt-get install -y gcc-aarch64-linux-gnu g++-aarch64-linux-gnu
      - uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser-pro
          version: "~> v2"
          args: release --clean --split
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GORELEASER_KEY: ${{ secrets.GORELEASER_KEY }}
          GGOOS: linux
      - uses: actions/upload-artifact@v4
        with: { name: dist-linux, path: dist }

  goreleaser-darwin:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v6
        with: { fetch-depth: 0 }
      - run: git fetch --force --tags
      - uses: actions/setup-go@v6
        with: { go-version-file: go.mod, cache: true }
      - uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser-pro
          version: "~> v2"
          args: release --clean --split
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GORELEASER_KEY: ${{ secrets.GORELEASER_KEY }}
          GGOOS: darwin
      - uses: actions/upload-artifact@v4
        with: { name: dist-darwin, path: dist }

  goreleaser-merge:
    runs-on: ubuntu-latest
    environment: release
    needs: [goreleaser-linux, goreleaser-darwin]
    steps:
      - uses: actions/checkout@v6
        with: { fetch-depth: 0 }
      - run: git fetch --force --tags
      - uses: actions/setup-go@v6
        with: { go-version-file: go.mod, cache: true }
      - uses: actions/download-artifact@v4
        with: { path: dist-parts }
      - run: |
          rm -rf dist && mkdir -p dist
          for part in dist-linux dist-darwin; do
            [ -d "dist-parts/${part}/dist" ] && cp -a "dist-parts/${part}/dist/." dist/ || cp -a "dist-parts/${part}/." dist/
          done
      - uses: crazy-max/ghaction-import-gpg@v6
        id: import_gpg
        with:
          gpg_private_key: ${{ secrets.GO_GO_GOLEMS_SIGN_KEY }}
          passphrase: ${{ secrets.GO_GO_GOLEMS_SIGN_PASSPHRASE }}
          fingerprint: "6EBE1DF0BDF48A1BBA381B5B79983EF218C6ED7E"
      - uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser-pro
          version: "~> v2"
          args: continue --merge
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GORELEASER_KEY: ${{ secrets.GORELEASER_KEY }}
          TAP_GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}
          GPG_FINGERPRINT: ${{ steps.import_gpg.outputs.fingerprint }}
          FURY_TOKEN: ${{ secrets.FURY_TOKEN }}
```

#### push.yml

```yaml
name: golang-pipeline
on:
  push:
    branches: ['main']
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
        with: { fetch-depth: 0 }
      - run: git fetch --force --tags
      - uses: actions/setup-go@v6
        with: { go-version-file: go.mod, cache: true }
      - name: generate assets
        run: go generate ./...
      - name: run unit tests
        run: go test ./...
```

#### lint.yml, codeql-analysis.yml, secret-scanning.yml, dependency-scanning.yml

Copy directly from go-template without changes. These are generic and project-agnostic.

### 5.4 Linting and Hooks

#### .golangci.yml

Copy from go-template. No changes needed — the linter configuration is generic:

```yaml
version: "2"
linters:
  default: none
  enable:
    - errcheck
    - govet
    - ineffassign
    - staticcheck
    - unused
    - exhaustive
    - nonamedreturns
    - predeclared
  exclusions:
    rules:
      - linters: [staticcheck]
        text: 'SA1019: cli.CreateProcessorLegacy'
  settings:
    errcheck:
      exclude-functions:
        - (io.Closer).Close
        - fmt.Fprintf
        - fmt.Fprintln
formatters:
  enable:
    - gofmt
```

#### lefthook.yml

Copy from go-template. No changes needed:

```yaml
pre-commit:
  commands:
    lint:
      glob: '*.go'
      run: make lint
    test:
      glob: '*.go'
      run: make test
  parallel: true

pre-push:
  commands:
    release:
      run: make goreleaser
    lint:
      run: make lint
    test:
      run: make test
  parallel: true
```

#### .golangci-lint-version

Copy from go-template. Pin the golangci-lint version:

```
v2.1.0
```

(Check what version go-template currently uses and match it.)
