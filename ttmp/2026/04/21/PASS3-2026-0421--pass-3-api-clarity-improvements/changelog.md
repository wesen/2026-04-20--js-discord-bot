# Changelog

## 2026-04-21

- Initial workspace created


## 2026-04-21

Pass 3 complete: extracted base dispatch request builder (6e4bf6f), split run_schema.go into 3 files (2e5f9aa), extracted DiscordOps nil-guard wrappers (7a24a6a). Task 1 (typed envelopes) deferred. All tests pass.

### Related Files

- internal/botcli/run_static_args.go — Explicit static parsing phase
- internal/jsdiscord/bot_ops.go — Generic nil-guard wrappers
- internal/jsdiscord/host_dispatch.go — Builder pattern for DispatchRequest

