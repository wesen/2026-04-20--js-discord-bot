# Diary

## Goal

Track the implementation of the Show Space Discord Bot in small, reviewable steps, with frequent diary updates and commits between task slices.

## Step 1: Create the bot shell and verify repo discovery

I started by creating the smallest real `show-space` bot entrypoint under `examples/discord-bots/show-space/index.js`. The shell only defines bot metadata and a ready log for now, which is enough to confirm the repository can discover the new venue bot without introducing any announcement logic before the surrounding design is in place.

That gave us a clean baseline for the later task slices: commands, permissions, pinning, and persistence can now be added incrementally while the bot already exists as a discoverable named bot.

### Prompt Context

**User prompt (verbatim):** "alright, work task by task, commit at appropriate intervals, and keep a frequent detailed diary"

**Assistant interpretation:** Implement the show-space ticket incrementally, pause to commit after meaningful slices, and maintain a detailed chronological diary of what changed and why.

**Inferred user intent:** Build the new venue bot carefully instead of in one large burst, with review-friendly commits and a work log that can be used to resume or audit the implementation later.

**Commit (code):** `1616557` — "feat(show-space): add bot shell and discovery metadata"

### What I did
- Created `examples/discord-bots/show-space/index.js`
- Added `require("discord")` and `defineBot(...)`
- Registered the bot as `show-space` with category `venues`
- Added a `ready` event that logs the bot username and script path
- Verified the bot is discoverable via `GOWORK=off go run ./cmd/discord-bot bots list --bot-repository ./examples/discord-bots | rg -n "show-space|venues"`
- Ran `go test ./internal/jsdiscord/... -count=1`

### Why
- The ticket asks for task-by-task work and frequent commits
- A bot shell establishes the named-bot repository entry without overcommitting to Phase 1 logic too early
- Discovery verification ensures the new bot is wired into the repo’s bot inventory before more code lands

### What worked
- The bot list command found the new entry immediately:
  - `show-space	show-space/index.js	Venue operations bot for upcoming shows and pinned announcements`
- The existing `internal/jsdiscord` test suite still passed after the new bot file was added
- The bot shell stayed intentionally small, which will make later command and persistence slices easier to review

### What didn't work
- No code failures occurred in this slice
- There were unrelated untracked working-tree files already present at the repo root (`.playwright-mcp/`, `discord-bot`, `update_dispatch.py`, `update_kb_test.py`, `update_tests.py`), but I did not touch them

### What I learned
- The repo’s named-bot discovery path is already enough to surface a new example bot as soon as an `index.js` with `defineBot(...)` exists
- A minimal `ready` event is a good first milestone because it proves the runtime loads the bot without requiring command scaffolding yet

### What was tricky to build
- The main constraint was keeping the first slice small enough to be a true shell while still being useful for validation
- I intentionally avoided adding command logic too early so Phase 1 can be built around a stable entrypoint rather than a half-designed workflow

### What warrants a second pair of eyes
- The bot category choice (`venues`) should be revisited later when the inventory docs are updated, to make sure it fits the repo’s organization conventions

### What should be done in the future
- Add Phase 1 command modules and show rendering helpers
- Decide the first runtime config fields for channel and role IDs
- Update the example bot inventory README once the bot has user-facing commands
- Continue diary updates after each meaningful task slice

### Code review instructions
- Start with `examples/discord-bots/show-space/index.js`
- Confirm the bot metadata, category, and ready logger are the only changes in this slice
- Re-run `GOWORK=off go run ./cmd/discord-bot bots list --bot-repository ./examples/discord-bots | rg -n "show-space|venues"`
- Re-run `go test ./internal/jsdiscord/... -count=1`

### Technical details
- Bot entrypoint:
  - `require("discord")`
  - `defineBot(({ configure, event }) => { ... })`
- Metadata used:
  - `name: "show-space"`
  - `description: "Venue operations bot for upcoming shows and pinned announcements"`
  - `category: "venues"`
- Discovery result:
  - `show-space	show-space/index.js	Venue operations bot for upcoming shows and pinned announcements`

## Step 2: Build phase-1 show announcement flow and test it

I expanded the bot into a working venue workflow: a seeded `shows.json`, helper modules for dates, permissions, rendering, and show catalog state, and the three phase-1 commands (`/upcoming`, `/announce`, and `/unpin-old`). I also added runtime tests that prove the bot loads, lists shows, rejects unauthorized announces, posts/pins announcements, and unpins expired shows.

The only runtime-specific snag was the embedded Goja environment not exposing `Intl`. I fixed that by switching the date formatter to a manual weekday/month renderer, which keeps the code portable inside the bot runtime.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Continue the show-space bot work in small commits, keeping the diary current while implementing the next task slice.

**Inferred user intent:** Keep building the venue bot methodically, with clear checkpoints and enough logging to resume or audit later.

**Commit (code):** `5900aff` — "feat(show-space): implement phase 1 announcement flow and runtime coverage"

### What I did
- Added `examples/discord-bots/show-space/shows.json` with seeded upcoming shows
- Added `examples/discord-bots/show-space/lib/dates.js`
- Added `examples/discord-bots/show-space/lib/permissions.js`
- Added `examples/discord-bots/show-space/lib/render.js`
- Added `examples/discord-bots/show-space/lib/shows.js`
- Expanded `examples/discord-bots/show-space/index.js` with:
  - runtime config fields for channel IDs, role IDs, and timezone
  - `/upcoming`
  - `/announce`
  - `/unpin-old`
- Added `internal/jsdiscord/show_space_runtime_test.go`
- Updated `examples/discord-bots/README.md` to list `show-space`
- Verified `GOWORK=off go run ./cmd/discord-bot bots list --bot-repository ./examples/discord-bots | rg -n "show-space|venues"`
- Verified `go test ./internal/jsdiscord/... -count=1`

### Why
- Phase 1 needed a real working surface before moving into DB-backed persistence
- A seeded JSON file plus an in-memory catalog is enough to validate the venue announcement flow without introducing database complexity too early
- Runtime tests provide confidence that the bot behaves correctly under the current Goja-backed JS host

### What worked
- The bot list command continued to show the new venue bot after the phase-1 expansion
- The runtime tests covered the important behavior slices:
  - seeded upcoming shows
  - permission denial for non-admin/non-booker users
  - announcement posting + pinning
  - expired pin cleanup
- The bot code remained readable enough to keep the command handlers thin

### What didn't work
- The first version of the date formatter used `Intl.DateTimeFormat`, which is not available in this Goja runtime
- The failure looked like this:
  - `ReferenceError: Intl is not defined at formatDisplayDate (/home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/show-space/lib/dates.js:50:25(15))`
- I fixed it by replacing the `Intl` usage with manual weekday/month name formatting

### What I learned
- This embedded JS runtime does not assume browser-like or Node-like internationalization APIs
- A manual formatter is safer for bot-runtime portability when only the output shape matters
- It is still practical to model show data in JS before introducing persistence

### What was tricky to build
- The `/announce` flow needed to post a Discord message, find the exact posted message, and pin it without a direct `send()` return value containing the message ID
- I solved that by looking up the recent messages in the target channel and matching the rendered embed title
- The `/unpin-old` flow had to parse the date back out of the rendered pinned title so the cleanup behavior could work without a database

### What warrants a second pair of eyes
- The current announce flow depends on a title-matching lookup after posting; if the channel is very active, that matching strategy may need a tighter discriminator later
- The phase-1 in-memory catalog is intentionally simple and will be replaced by durable storage in phase 2

### What should be done in the future
- Add the database-backed store and phase-2 commands
- Decide whether `/announce` should remain as a quick command or become a thin alias over `/add-show`
- Add a staff-facing detail command and persisted past-show listing
- Add a daily archive helper that can be triggered by an external scheduler

### Code review instructions
- Start with `examples/discord-bots/show-space/index.js`
- Review `examples/discord-bots/show-space/lib/dates.js` for the no-Intl formatter
- Review `examples/discord-bots/show-space/lib/shows.js` for the in-memory catalog shape
- Review `internal/jsdiscord/show_space_runtime_test.go` for the runtime coverage of permission gating, pinning, and cleanup
- Re-run:
  - `go test ./internal/jsdiscord -run TestShowSpace -count=1`
  - `go test ./internal/jsdiscord/... -count=1`
  - `GOWORK=off go run ./cmd/discord-bot bots list --bot-repository ./examples/discord-bots | rg -n "show-space|venues"`

### Technical details
- Seed data file:
  - `examples/discord-bots/show-space/shows.json`
- Helper modules:
  - `dates.js` — date parsing and display formatting
  - `permissions.js` — role checks and permission-denied payload
  - `render.js` — announcement and upcoming-show payload formatting
  - `shows.js` — in-memory catalog for phase 1
- Runtime command behavior:
  - `/upcoming` returns an ephemeral formatted list
  - `/announce` posts and pins in the configured `upcomingShowsChannelId`
  - `/unpin-old` parses the rendered date title to find expired pins
- Validation notes:
  - the bot shell still loads via `bots list`
  - the internal JS runtime suite remains green after the new example and tests

## Step 3: Add the database-backed show store and phase-2 commands

I extended the bot from the phase-1 in-memory catalog into a database-backed venue tool. The command layer now supports both modes: it still works with seeded JSON-only data, but when `dbPath` is configured it uses a SQLite store, seeds from `shows.json` once, and tracks Discord message IDs so staff can manage shows by ID.

This step also added the maintenance helper for expiring shows, plus a more complete runbook and docs update so the operator path is documented alongside the implementation.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Continue the show-space bot work through the DB-backed phase, keep committing in coherent slices, and preserve the diary/doc trail.

**Inferred user intent:** Finish the venue bot through the database-backed show-management stage and keep enough narrative context to review or continue the work later.

**Commit (code):** `d86dc43` — "feat(show-space): add phase-2 show store and database-backed commands"

### What I did
- Added `examples/discord-bots/show-space/lib/store.js` as a SQLite-backed persistence layer
- Extended `examples/discord-bots/show-space/index.js` with:
  - `dbPath` and `seedFromJson` runtime fields
  - `/add-show`
  - `/show`
  - `/cancel-show`
  - `/archive-show`
  - `/past-shows`
  - `/archive-expired`
- Kept `/upcoming`, `/announce`, and `/unpin-old` working against either the in-memory catalog or the database store
- Extended `examples/discord-bots/show-space/lib/render.js` with show detail and past-show formatting
- Extended `examples/discord-bots/show-space/lib/shows.js` with pinned-message archive helpers for the catalog mode
- Added runtime tests for database-backed add/show/cancel/archive/past-shows and archive-expired behavior
- Updated `examples/discord-bots/README.md` to document the phase-2 command set and DB-backed mode
- Added the operator runbook at `ttmp/2026/04/22/DISCORD-BOT-022--show-space-discord-bot/reference/02-operator-runbook.md`
- Updated the JS API reference to link the new `show-space` example

### Why
- The spec wanted persistent shows by ID and a path toward a future web dashboard
- Using the repository’s current `require("database")` module keeps the implementation aligned with the existing JS runtime patterns
- A seed-on-empty store gives the bot an easy migration path from `shows.json` into SQLite without a separate bootstrap script
- The maintenance helper is needed so expired shows can be archived and unpinned without relying on an always-on timer inside the bot process

### What worked
- The database-backed store seeded cleanly from `shows.json` when empty
- The bot could add, show, cancel, archive, and list past shows under a database-backed configuration
- The archive-expired path correctly archived past shows, unpinned the Discord message, and posted a quiet staff summary
- The runtime suite stayed green after the phase-2 expansion

### What didn't work
- No new runtime error surfaced in this slice after the phase-1 `Intl` issue was fixed
- The DB store needed one config-bool correction (`seedFromJson`) so a false value would not accidentally fall back to the default; that was corrected before validation

### What I learned
- The same command surface can support both a simple seed-only mode and a persistent DB mode if the repository layer is abstracted carefully
- Recording the Discord message ID at announcement time is the key to making later cancel/archive flows deterministic
- For this runtime, the maintenance helper is best treated as a callable JS function or command hook that an external scheduler can invoke, rather than an in-process interval loop

### What was tricky to build
- The announce flow still has to post first, discover the message ID, and then pin the message, because the channel-send helper does not return a direct message object
- The date/title parsing needs to stay consistent between the embed renderer and the cleanup/archive paths so the bot can recognize its own announcements later
- The DB store had to preserve the same show shape as the in-memory catalog so the command handlers did not need a second code path for presentation

### What warrants a second pair of eyes
- The new `archive-expired` maintenance command is useful as a hook, but the long-term scheduling mechanism should be reviewed when the bot is deployed
- The store currently uses SQLite because that is what the repository runtime exposes today; if a future web dashboard needs Postgres, the store boundary should be the seam for that migration

### What should be done in the future
- Keep the operator runbook in sync if the deployment model changes
- Decide whether `/announce` should eventually become a thin alias over `/add-show` in phase 2+
- If a staff dashboard is added later, reuse the store boundary rather than reaching into bot internals

### Code review instructions
- Start with `examples/discord-bots/show-space/lib/store.js`
- Review the phase-2 command handlers in `examples/discord-bots/show-space/index.js`
- Review the new runtime tests in `internal/jsdiscord/show_space_runtime_test.go`
- Confirm the operator runbook at `ttmp/2026/04/22/DISCORD-BOT-022--show-space-discord-bot/reference/02-operator-runbook.md`
- Re-run:
  - `go test ./internal/jsdiscord -run TestShowSpace -count=1`
  - `go test ./internal/jsdiscord/... -count=1`
  - `GOWORK=off go run ./cmd/discord-bot bots list --bot-repository ./examples/discord-bots | rg -n "show-space|venues"`

### Technical details
- Database fields used by the store:
  - `artist`, `date`, `doors_time`, `age`, `price`, `notes`, `status`, `discord_message_id`, `discord_channel_id`, `created_at`, `updated_at`
- Commands now supported in DB mode:
  - `/add-show`
  - `/show`
  - `/cancel-show`
  - `/archive-show`
  - `/past-shows`
  - `/archive-expired`
- Maintenance flow:
  - archive expired shows in the store
  - unpin their Discord messages
  - optionally send a quiet `#staff` summary
- Runtime state:
  - the bot can still run phase-1-style with only `shows.json`
  - `dbPath` turns on persistent storage and seeded migration from the JSON file
