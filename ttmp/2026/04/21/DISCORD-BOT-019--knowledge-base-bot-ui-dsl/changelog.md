# Changelog

## 2026-04-22

- Built the `ui-showcase` example bot implementing the recommended DSL approach from the design brainstorm.
- Created `lib/ui/primitives.js` with generic builder helpers: `message()`, `embed()`, `button()`, `select()`, `form()`, `card()`, `pager()`, `actions()`, `confirm()`, `ok()`, `error()`, `emptyResults()`.
- Created `lib/ui/screen.js` with `flow()` stateful screen helper, `alias()`, and `aliasAutocomplete()`.
- Created `lib/demo-store.js` with sample articles, products, and tasks.
- Built 10 showcase commands covering builders, forms, stateful search, review queues, confirmations, pagination, card galleries, all select types, and alias registration.
- Updated `examples/discord-bots/README.md` with the new bot.
- Fixed button builder to support chain methods on the returned object.
- Fixed `message().embed()` to auto-build builder objects passed as arguments.

## 2026-04-21

- Initial workspace created
- Added a detailed design brainstorm analyzing the current knowledge-base bot UI composition style and proposing several UI DSL directions.
- Added a companion reference doc with concrete DSL sketches for teach forms, search results, review queues, source sheets, and alias registration.
- Added a working diary documenting the analysis flow and the files inspected.

