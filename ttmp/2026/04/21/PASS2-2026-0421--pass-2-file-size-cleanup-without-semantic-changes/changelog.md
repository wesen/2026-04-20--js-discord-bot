# Changelog

## 2026-04-21

- Initial workspace created


## 2026-04-21

Pass 2 complete: split bot.go into 6 files (f64f16a), split host_payloads.go into 6 files (c5335fa), split runtime_test.go into 9 files (8e75a3d). All tests pass. 21 files changed, 2,658 insertions, 2,535 deletions.

### Related Files

- internal/jsdiscord/bot_compile.go — Primary compile/draft file
- internal/jsdiscord/bot_dispatch.go — Dispatch and promise settlement
- internal/jsdiscord/payload_model.go — Core payload types and helpers
- internal/jsdiscord/runtime_dispatch_test.go — Dispatch behavior tests

