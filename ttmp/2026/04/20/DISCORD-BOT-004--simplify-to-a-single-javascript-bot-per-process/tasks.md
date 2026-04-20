# Tasks

## Design and decision package

- [x] Create the ticket workspace
- [x] Write a detailed architecture and implementation guide for the single-bot-per-process model
- [x] Write a reference/migration note describing the operator-facing changes from the multi-bot model
- [x] Validate the ticket workspace and upload the document bundle to reMarkable

## Planned implementation tasks

### 1. Runtime simplification
- [x] Remove the multi-bot runtime path from the main host flow
- [x] Return `internal/bot/bot.go` to loading exactly one selected bot script at a time
- [x] Remove or archive `internal/jsdiscord/multihost.go` from the live path
- [x] Keep event dispatch and command routing inside one selected bot runtime only

### 2. CLI simplification
- [x] Change `discord-bot bots run <bot...>` back to `discord-bot bots run <bot>`
- [x] Remove multi-bot selection and duplicate-command collision handling from the main CLI path
- [x] Keep `bots list` and `bots help <bot>` discovery behavior
- [x] Preserve `--print-parsed-values` for the single selected bot runner

### 3. Bot-run flag architecture
- [x] Define a bot-level run-schema contract under `configure(...)`
- [x] Reuse jsverbs-style field/section definitions for bot startup config
- [x] Reuse Glazed/Cobra schema building for single-bot dynamic runtime flags
- [x] Inject parsed runtime config into the JS context as `ctx.config`
- [x] Decide how `--print-parsed-values` should work once Glazed parsing is used for bot startup fields

### 4. Example and migration cleanup
- [ ] Decide whether the example repository should stay multi-package for discovery or be reduced to clearer single-bot examples
- [ ] Add one canonical example bot that demonstrates in-bot composition instead of multi-process composition
- [ ] Update docs so operators do not expect `bots run bot-a bot-b`
- [x] Delete the now-obsolete multi-bot composition code and tests

### 5. Observability and runtime diagnostics
- [x] Add richer interaction lifecycle debug logging for defer/reply/edit/follow-up/modal flows
- [x] Add debug logging for request-scoped Discord host operations such as channel sends and message edits
- [x] Improve operator-facing error context so failures identify the script and interaction being dispatched
