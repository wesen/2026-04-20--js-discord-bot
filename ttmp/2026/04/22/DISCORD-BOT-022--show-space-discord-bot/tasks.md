# Tasks

## ✅ Planning and analysis

- [x] Create ticket workspace from the Show Space Discord Bot spec
- [x] Inspect current JS bot examples for command, outbound-message, pin, and database patterns
- [x] Confirm the current runtime exposes `ctx.discord.messages.pin`, `ctx.discord.messages.unpin`, and `ctx.discord.messages.listPinned`
- [x] Confirm the current runtime already documents `require("database")` and `require("ui")`
- [x] Capture the product/spec text verbatim inside the implementation guide

## Phase 1: JS bot only, no durable database

Goal: ship a working venue bot that can post, list, and pin announcements with simple JSON-backed or inline show data.

- [x] **1.1** Create a new bot example under `examples/discord-bots/show-space/`
  - `index.js` should use `require("discord")`
  - register a bot name like `show-space` or `artist-space`
  - keep all runtime behavior in JavaScript, not Discord.js/discord.py

- [x] **1.2** Add a small helper module layout
  - `lib/dates.js` for parsing and normalizing inputs like `Apr 25` and `2025-04-25`
  - `lib/render.js` for announcement embeds and upcoming-show text output
  - `lib/permissions.js` for `@admin` / `@booker` role checks
  - `lib/shows.js` for loading `shows.json` and filtering upcoming entries

- [x] **1.3** Implement `shows.json` as the Phase 1 data source
  - seed a few upcoming shows for local testing
  - keep the file small and hand-editable
  - normalize date display to `Fri Apr 25, 2025`

- [x] **1.4** Implement `/upcoming`
  - visible to anyone
  - render a clean list of upcoming shows from `shows.json`
  - decide whether the command reply is public or ephemeral, then document that choice in the guide

- [x] **1.5** Implement `/announce`
  - allow only `@admin` and `@booker`
  - accept simple text args: `artist`, `date`, `doors_time`, `age_restriction`, `price`, `notes`
  - post a formatted embed to the configured `#upcoming-shows` channel
  - pin the posted message
  - reply with a success acknowledgement

- [x] **1.6** Implement `/unpin-old`
  - allow only `@admin`
  - list pins in `#upcoming-shows`
  - unpin entries whose dates have passed
  - report the count of removed pins

- [x] **1.7** Add venue configuration fields through `configure({ run: ... })`
  - channel IDs: `upcomingShowsChannelId`, `announcementsChannelId`, `staffChannelId`
  - role IDs: `adminRoleId`, `bookerRoleId`
  - optional timezone / date-display fields if needed

- [x] **1.8** Add command-level permission failures
  - reply ephemerally with: `❌ You don't have permission to use this command.`
  - keep the check reusable across all write commands

- [x] **1.9** Add smoke tests for the Phase 1 workflow
  - list shows
  - announce a show
  - pin the announcement
  - unpin an expired announcement

## Phase 2: Durable database and staff operations

Goal: move show state into a persistent store and make show management safer.

- [ ] **2.1** Decide the Phase 2 persistence strategy for this repository
  - prefer the current JS runtime’s database module for the first implementation
  - if Postgres is still required later, record that as a separate host/runtime follow-up
  - document the tradeoff explicitly in the guide

- [ ] **2.2** Create a database wrapper module
  - `examples/discord-bots/show-space/lib/store.js`
  - keep all SQL or persistence details behind one module boundary
  - provide simple methods like `listUpcoming()`, `getShow(id)`, `createShow()`, `cancelShow()`, `archiveShow()`

- [ ] **2.3** Add a show schema/migration
  - include fields for `artist`, `date`, `doors_time`, `age`, `price`, `notes`, `status`, `discord_message_id`, `discord_channel_id`, and timestamps
  - seed from `shows.json` once, then switch to the database as the source of truth

- [ ] **2.4** Implement `/add-show`
  - save the show first
  - post the announcement to `#upcoming-shows`
  - pin the message
  - store the Discord message/channel IDs on the row
  - reply with the new show ID

- [ ] **2.5** Update `/upcoming`
  - read from the database instead of `shows.json`
  - filter to confirmed shows ordered by date
  - support a sensible default limit

- [ ] **2.6** Implement `/show <id>`
  - return a full detail view for one show
  - make it easy for staff to quick-check a booking or announcement

- [ ] **2.7** Implement `/cancel-show <id>`
  - update the DB status to `cancelled`
  - unpin the original Discord announcement
  - post a cancellation notice in `#upcoming-shows`
  - keep the record for later reference

- [ ] **2.8** Implement `/archive-show <id>`
  - allow `@admin` only
  - archive old or completed shows
  - unpin the message if it is still pinned

- [ ] **2.9** Implement `/past-shows`
  - return the last few archived shows
  - decide whether staff-only or public access is the right default

- [ ] **2.10** Implement the daily auto-archive flow
  - create a reusable `archiveExpiredShows()` helper in JS
  - trigger it from a host-side cron/job wrapper or deployment scheduler
  - avoid relying on an in-process timer unless the hosting model explicitly supports it
  - post a quiet summary to `#staff`

- [ ] **2.11** Add a one-time migration from `shows.json`
  - seed the DB from the Phase 1 file
  - preserve dates and announcement metadata where possible

## Validation and docs

- [ ] **3.1** Add runtime tests for the bot commands
  - command permission failures
  - announce/pin/unpin flows
  - DB lookups by ID
  - archived/past-show listing behavior

- [ ] **3.2** Add a local operator runbook
  - how to start the bot
  - how to configure IDs and role gates
  - how to seed or migrate initial shows
  - how to run the manual unpin/archival command safely

- [x] **3.3** Update `examples/discord-bots/README.md`
  - add the new `show-space` bot to the example inventory
  - document the Phase 1 and Phase 2 command set
  - note any required guild/channel permissions

- [ ] **3.4** Update the JS API reference if the bot reveals missing docs
  - add command examples if needed
  - add any new persistence guidance if the implementation needs it

- [ ] **3.5** Decide whether the bot should live under one of the existing example categories or a new `venues/` category
  - document the final choice in the guide
  - keep the bot name stable once the workspace is in active use
