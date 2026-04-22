---
Title: Pass 3 Implementation Guide
Ticket: PASS3-2026-0421
Status: active
Topics:
    - code-quality
    - refactoring
    - api-design
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/botcli/run_static_args.go
      Note: Explicit static parsing phase
    - Path: internal/jsdiscord/host_dispatch.go
      Note: Builder pattern for DispatchRequest
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---



# Pass 3: API Clarity Improvements

## Executive Summary

Pass 3 improves internal APIs without changing external behavior. It introduces typed structs at the Go↔JS boundary, eliminates repetitive dispatch construction, makes CLI parsing phases explicit, and refactors the DiscordOps builder into a registry pattern.

Estimated time: **~8 hours**.

## Origin

Derived from CODEQUAL-2026-0421 findings:
- **5.1.6** — `map[string]any` used too far into interior
- **5.2.1** — 18 repetitions of `DispatchRequest` construction
- **5.4.1** — `preparseRunArgs` naming and structure
- **5.2.2** — `DiscordOps` 4-place edit burden

## Task 1: Introduce typed internal envelopes

**Problem:** `DispatchRequest` and descriptor parsing use `map[string]any` throughout the interior.

**Approach:** Introduce typed structs internally, convert to maps only at the JS boundary.

**New types (in `bot_context.go` or new `envelope.go`):**

```go
type DispatchEnvelope struct {
    Name        string
    Args        map[string]any
    Values      any
    Interaction InteractionSnapshot
    Message     MessageSnapshot
    Before      MessageSnapshot
    User        UserSnapshot
    Guild       GuildSnapshot
    Channel     ChannelSnapshot
    ...
}

func (e DispatchEnvelope) ToJSMap() map[string]any { ... }
```

**For descriptors:**

```go
type RawDescribeSnapshot struct {
    Metadata   map[string]any
    Commands   []map[string]any
    Events     []map[string]any
    Components []map[string]any
    ...
}
```

**Migration strategy:**
1. Add new types alongside existing code.
2. Convert one call site at a time.
3. Delete old map-only paths once all callers are converted.

**Verification:**
```bash
go test ./...
```

**Commit message:**
```
refactor: introduce typed DispatchEnvelope and descriptor structs

Replaces interior map[string]any usage with typed structs that convert
to maps only at the JS boundary. No external behavior changes.

Refs: PASS3-2026-0421
```

---

## Task 2: Refactor `host_dispatch.go` with base request builders

**Problem:** 18 call sites construct `DispatchRequest` with the same base fields.

**Approach:** Add builder methods on `Host`:

```go
func (h *Host) baseDispatchRequest(session *discordgo.Session) DispatchRequest {
    return DispatchRequest{
        Metadata: map[string]any{"scriptPath": h.scriptPath},
        Config:   cloneMap(h.runtimeConfig),
        Discord:  buildDiscordOps(h.scriptPath, session),
        Me:       currentUserMap(session),
    }
}

func (h *Host) eventRequest(session *discordgo.Session, eventName string, payload map[string]any) DispatchRequest {
    req := h.baseDispatchRequest(session)
    req.Event = eventName
    req.Payload = payload
    return req
}
```

**Verification:**
```bash
go test ./...
```

**Commit message:**
```
refactor: extract base dispatch request builder

Eliminates 18 repetitions of DispatchRequest construction in
host_dispatch.go by introducing Host.baseDispatchRequest().

Refs: PASS3-2026-0421
```

---

## Task 3: Make botcli parsing phases explicit

**Problem:** `run_schema.go` combines static flag parsing, dynamic schema building, and help rendering in one 346-line file. The function `preparseRunArgs` does not signal its purpose.

**Approach:**
1. Rename `preparseRunArgs` → `parseStaticRunnerArgs`
2. Split into three files:

```text
internal/botcli/
  run_static_args.go     // parseStaticRunnerArgs + StaticRunnerArgs type
  run_dynamic_schema.go  // buildRunSchema + parseRuntimeConfigArgs
  run_help.go            // printRunSchema + selector-aware help
```

**Verification:**
```bash
go test ./...
go build ./...
```

**Commit message:**
```
refactor: split run_schema.go into explicit parsing phases

Renames preparseRunArgs to parseStaticRunnerArgs and splits the
346-line file into three focused files by concern. No behavior changes.

Refs: PASS3-2026-0421
```

---

## Task 4: Refactor DiscordOps into a registry pattern

**Problem:** Adding a new Discord API operation requires editing 4 places: the `DiscordOps` struct, `buildXOps`, the nil-guard check, and `discordOpsObject`.

**Approach:** Replace the struct-of-functions with a registry:

```go
type DiscordOp struct {
    Name    string
    Build   func(ops *DiscordOps, scriptPath string, session *discordgo.Session)
    JSName  string
}

var discordOpRegistry = []DiscordOp{
    {Name: "ChannelFetch", Build: buildChannelFetchOps, JSName: "channel"},
    {Name: "GuildFetch", Build: buildGuildFetchOps, JSName: "guild"},
    // ...
}

func buildDiscordOps(scriptPath string, session *discordgo.Session) *DiscordOps {
    ops := &DiscordOps{}
    for _, op := range discordOpRegistry {
        op.Build(ops, scriptPath, session)
    }
    return ops
}
```

**Verification:**
```bash
go test ./...
```

**Commit message:**
```
refactor: convert DiscordOps to registry pattern

Replaces the 4-place edit burden with a single registry slice.
Adding a new operation now requires one entry in discordOpRegistry.

Refs: PASS3-2026-0421
```

## Risk Assessment

| Task | Risk | Mitigation |
|------|------|------------|
| Typed envelopes | Medium | Add types alongside existing code; convert gradually |
| Dispatch builders | Low | Pure extraction; existing tests validate |
| botcli split | Low | No logic changes |
| DiscordOps registry | Medium | Careful with nil guards and JS object shape |

## Success Criteria

- [ ] Typed envelopes exist and all interior code uses them
- [ ] `host_dispatch.go` uses `baseDispatchRequest()` / `eventRequest()`
- [ ] `run_schema.go` split into 3 files with clear names
- [ ] `DiscordOps` uses registry pattern
- [ ] `go test ./...` passes after each commit
- [ ] `go build ./...` passes

## Related Tickets

- **PASS2-2026-0421** — Previous pass (file splitting)
- **PASS4-2026-0421** — Next pass (larger changes)
