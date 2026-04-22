---
Title: Pass 4 Implementation Guide
Ticket: PASS4-2026-0421
Status: active
Topics:
    - code-quality
    - refactoring
    - architecture
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
---

# Pass 4: Larger Architectural Changes

## Executive Summary

Pass 4 contains changes that alter UX, require new tooling, or depend on external capabilities. These should be discussed and approved before implementation. None of these are urgent — they are backlog items for when the team has bandwidth.

## Origin

Derived from CODEQUAL-2026-0421 findings:
- **5.4.1** — Manual flag parsing could be replaced with native parsing
- **5.2.3** — `host_maps.go` could be code-generated
- **5.5.2** — `fmt.Errorf` everywhere prevents structured error handling
- **5.5.3** — Promise polling could become event-driven

## Task 1: Replace manual flag parsing with native CLI parsing

**Current:** `bots run` disables normal flag parsing and implements its own pre-parser.

**Proposed:** Use standard flag parsing. Collect dynamic args after `--`:

```bash
discord-bot bots run ping --bot-repository ./examples -- --db-path ./data.sqlite
```

**Alternative:** Use `cmd.FParseErrWhitelist.UnknownFlags = true` and post-process unknown flags.

**Impact:** Changes CLI UX. Requires updating docs, examples, and user habits.

**Effort:** ~3 hours.

**Decision needed:** Is the `--` separator acceptable? If not, which alternative?

---

## Task 2: Code-generate map converters

**Current:** `host_maps.go` contains hand-written `currentUserMap`, `guildMap`, `channelMap`, etc.

**Proposed:** A small codegen tool that reads discordgo struct tags and generates map converters.

**Example input:**
```go
//go:generate go run ./cmd/gendiscordmaps
```

**Effort:** ~1 day (tool) + ongoing maintenance savings.

**Decision needed:** Is the API surface large enough to justify a codegen tool?

---

## Task 3: Add structured error types

**Current:** All errors are `fmt.Errorf` strings.

**Proposed:** Domain-specific error types:

```go
type BotError struct {
    Code    string // "dispatch_failed", "validation_failed", etc.
    BotName string
    Script  string
    Cause   error
}

func (e *BotError) Error() string { ... }
func (e *BotError) Unwrap() error  { return e.Cause }
```

**Benefit:** JS side can switch on error codes instead of parsing strings.

**Effort:** ~1 day.

**Decision needed:** What error taxonomy does the JS side need?

---

## Task 4: Evaluate event-driven promise resolution

**Current:** `waitForPromise` polls every 5ms.

**Proposed:** Use goja's promise notification mechanism if one exists.

**Investigation:**
```bash
# Check goja source for promise hooks, channels, or callbacks
go doc github.com/dop251/goja Promise
```

**Effort:** Unknown (depends on goja capabilities).

**Decision needed:** Does goja support event-driven promise resolution?

## Risk Assessment

| Task | Risk | Mitigation |
|------|------|------------|
| Flag parsing | High (UX break) | Discuss with users; keep old syntax as alias |
| Codegen | Medium | Prototype on 3 types first |
| Error types | Medium | Design taxonomy with JS consumers |
| Promise polling | Low | Investigation only; no code yet |

## Success Criteria

- [ ] Task decisions documented and approved
- [ ] Any implemented task has tests and docs updates
- [ ] `go test ./...` passes
- [ ] README updated if UX changes

## Related Tickets

- **PASS3-2026-0421** — Previous pass (API clarity)
- **CODEQUAL-2026-0421** — Parent review
