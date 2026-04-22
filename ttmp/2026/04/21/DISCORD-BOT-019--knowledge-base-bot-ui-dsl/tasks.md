# Tasks

## Documentation

- [x] Create ticket workspace
- [x] Inspect the current knowledge-base bot UI surface
- [x] Write a detailed design brainstorm for a possible UI DSL
- [x] Write concrete example DSL sketches for multiple use cases
- [x] Add and maintain a diary while working

## Analysis goals

- [x] Map the current search/review/form UI architecture
- [x] Identify concrete duplication and awkwardness in the current interaction layer
- [x] Propose multiple DSL design directions with trade-offs
- [x] Recommend one preferred implementation path
- [x] Decide whether to follow this doc with a bot-local DSL prototype ticket — YES, build the showcase bot

## Implementation

- [x] Create `examples/discord-bots/ui-showcase/` example bot skeleton
- [x] Create `lib/ui/` with generic UI builder primitives (message, embed, row, button, select, form)
- [x] Create `lib/ui/screen.js` with stateful screen helper (flow namespace, state key, screen renderer)
- [x] Wire up the showcase bot `index.js` using the UI DSL throughout
- [x] Implement showcase commands: /demo-message (builder patterns), /demo-form (modal DSL), /demo-search (stateful search screen), /demo-review (review queue screen), /demo-confirm (confirmation dialog), /demo-pager (paginated list), /demo-cards (card gallery with select), /demo-selects (all select menu types), /demo-alias (alias registration demo)
- [x] Update README.md to include the ui-showcase bot
- [x] Commit and validate

