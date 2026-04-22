---
Title: Tasks
Ticket: PASS4-2026-0421
---

# Tasks

## Task 1: Replace manual flag parsing with native CLI parsing
- [ ] Decide: `--` separator vs UnknownFlags whitelist
- [ ] Prototype the chosen approach
- [ ] Update all examples and docs
- [ ] Add migration note to README
- [ ] Run `go test ./...` and `go build ./...`
- [ ] Commit

## Task 2: Code-generate map converters
- [ ] Decide: worth the effort? (count hand-written converters)
- [ ] Design codegen tool schema (struct tags → map functions)
- [ ] Prototype on 3 types (User, Guild, Channel)
- [ ] Generate remaining converters
- [ ] Add `//go:generate` directive
- [ ] Run `go test ./...` and `go build ./...`
- [ ] Commit

## Task 3: Add structured error types
- [ ] Design error taxonomy with JS consumers
- [ ] Create BotError and related types
- [ ] Replace fmt.Errorf in dispatch path
- [ ] Replace fmt.Errorf in ops path
- [ ] Update JS error handling to use codes
- [ ] Run `go test ./...` and `go build ./...`
- [ ] Commit

## Task 4: Evaluate event-driven promise resolution
- [ ] Research goja promise notification capabilities
- [ ] Document findings
- [ ] If supported, prototype event-driven resolution
- [ ] Benchmark vs polling approach
- [ ] Decide: adopt or keep polling
