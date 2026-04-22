# Changelog

## 2026-04-22

- Decided to move UI DSL from JS-side builders to Go-side Goja Proxy-based builders.
- Wrote detailed implementation guide: `design/02-goja-proxy-ui-dsl-implementation-guide.md`.
- Proved Goja Proxy works with `scripts/01-proxy-poc.go`.
- Updated tasks: deprecated JS implementation tasks, added comprehensive Go implementation tasks.
- Fixed JS-side chain-object-leaking bug (auto-build in row()).
- Added Go integration tests for all 9 showcase commands.

## 2026-04-21

- Initial workspace created
- Added a detailed design brainstorm analyzing the current knowledge-base bot UI composition style and proposing several UI DSL directions.
- Added a companion reference doc with concrete DSL sketches for teach forms, search results, review queues, source sheets, and alias registration.
- Added a working diary documenting the analysis flow and the files inspected.

