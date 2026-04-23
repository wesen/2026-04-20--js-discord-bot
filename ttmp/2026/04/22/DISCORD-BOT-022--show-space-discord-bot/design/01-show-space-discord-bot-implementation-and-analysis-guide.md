---
Title: Show Space Discord Bot — Implementation and Analysis Guide
Slug: show-space-discord-bot-implementation-and-analysis-guide
DocType: design-doc
Ticket: DISCORD-BOT-022
Status: draft
Topics:
  - discord
  - javascript
  - bots
  - database
  - moderation
  - announcements
  - venues
  - analysis
  - implementation
RelatedFiles:
  - /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/ping/index.js
  - /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/knowledge-base/index.js
  - /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/knowledge-base/lib/store.js
  - /home/manuel/code/wesen/2026-04-20--js-discord-bot/examples/discord-bots/moderation/lib/register-message-moderation-commands.js
  - /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/host_responses.go
  - /home/manuel/code/wesen/2026-04-20--js-discord-bot/internal/jsdiscord/bot_compile.go
  - /home/manuel/code/wesen/2026-04-20--js-discord-bot/pkg/doc/topics/discord-js-bot-api-reference.md
---

# Show Space Discord Bot — Implementation and Analysis Guide

## Executive summary

The proposed bot is a venue-operations tool for a physical artist/show space. In Phase 1 it should help staff list upcoming shows, announce a show into `#upcoming-shows`, and pin or unpin those announcements. In Phase 2 it should persist shows in a real datastore so staff can manage them by ID, archive old announcements automatically, and build toward a booking pipeline later.

For this repository, the correct implementation target is the embedded JavaScript bot runtime under `examples/discord-bots/`, not Discord.js or discord.py. The repo already has the command, event, component, modal, and outbound Discord APIs needed for the first pass, plus a database module and message pinning helpers. The main gap is domain logic: show modeling, date normalization, permissions, and a clean persistence boundary.

## What the spec asks for

The developer spec defines two early phases:

- Phase 1: bot only, no database, using a `shows.json` seed file or inline data
- Phase 2: bot plus persistent database, with show IDs, cancellation, archiving, and a daily cleanup job

It also defines the intended staff model:

- `@admin` can do everything
- `@booker` can manage shows but not delete/archive everything
- everyone else is read-only

The bot must operate primarily in these channels:

- `#upcoming-shows` for public show announcements and pins
- `#announcements` for broader public notices
- `#staff` for internal operational logs
- `#booking-requests` later, outside the Phase 1/2 scope

## Current repo fit

The repository already contains useful precedents:

- `examples/discord-bots/ping/index.js` shows the basic bot wiring pattern: `defineBot(...)`, commands, events, components, modals, autocomplete, deferred replies, and follow-ups.
- `examples/discord-bots/knowledge-base/index.js` shows a richer JS bot with `configure({ run: ... })` fields and a dedicated persistence wrapper.
- `examples/discord-bots/knowledge-base/lib/store.js` shows how the current runtime uses `require("database")` behind a small JS module boundary.
- `examples/discord-bots/moderation/lib/register-message-moderation-commands.js` already demonstrates pin/unpin/list-pins flows and permission-sensitive moderation commands.
- `internal/jsdiscord/host_responses.go` already distinguishes update-in-place interactions from follow-up messages.
- `internal/jsdiscord/bot_compile.go` exposes `ctx.member` into command dispatch, so role-gating can be done in JS without inventing a new host API.
- `pkg/doc/topics/discord-js-bot-api-reference.md` already documents `ctx.discord.messages.pin`, `unpin`, and `listPinned`.

That means the bot can be implemented without changes to the host runtime for Phase 1. Phase 2 mostly needs domain code and a persistence choice.

## Gap analysis

### Already available

- JS command registration via `defineBot(...)`
- Slash commands, events, component handlers, modal handlers, and autocomplete
- Runtime config fields via `configure({ run: ... })`
- `ctx.member` and `ctx.discord.roles.*` / `ctx.discord.members.*` access
- `ctx.discord.channels.send(...)` for public announcements
- `ctx.discord.messages.pin(...)`, `unpin(...)`, `listPinned(...)` for show pin management
- `ctx.discord.messages.list(...)` for history-based cleanup or audit helpers
- `require("database")` as a durable data surface in JavaScript bots

### Missing or still domain-specific

- A show-space bot example and folder structure
- A normalized show date parser/formatter
- Permission helpers for `@admin` and `@booker`
- A venue-specific render layer for public announcements and internal staff notices
- A stable show repository abstraction that can back commands by ID
- A migration path from `shows.json` to persistent storage
- A daily archive/cleanup strategy that fits the actual deployment model

### Important constraint: Postgres vs current runtime

The spec prefers PostgreSQL so the bot can later share a database with a web dashboard. The current repository runtime does **not** advertise a Postgres adapter; the documented `require("database")` example uses SQLite and the knowledge-base bot is built around that shape.

So for this repository there are two practical options:

1. Implement Phase 2 on top of the current database module first, most likely with SQLite, and keep the bot’s data access behind one JS store module so a future Postgres adapter can be swapped in later.
2. Open a separate runtime/host ticket to add a Postgres-capable database adapter before implementing the bot’s Phase 2 persistence.

This guide recommends option 1 for immediate implementation speed, with a clearly isolated store module so the persistence backend is swappable later.

## Recommended repository shape

A good layout would be:

```text
examples/discord-bots/show-space/
  index.js
  shows.json                # Phase 1 seed data
  lib/
    dates.js
    permissions.js
    render.js
    shows.js                # Phase 1 list/filter helper
    store.js                # Phase 2 persistence boundary
    migrate.js              # one-time seed/import helper
```

### Responsibilities by module

- `index.js`
  - define the bot
  - register commands
  - read runtime config
  - keep command handlers thin

- `lib/dates.js`
  - parse `Apr 25`, `2025-04-25`, and similar inputs
  - normalize to a display string like `Fri Apr 25, 2025`
  - return `Date` or ISO strings in a predictable timezone

- `lib/permissions.js`
  - determine whether the actor has an admin/booker role
  - hide role lookup details behind one helper

- `lib/render.js`
  - build clean announcement embeds
  - build upcoming-show text output
  - build cancellation, archive, and staff-log messages

- `lib/shows.js`
  - read and filter Phase 1 show data
  - convert seeded JSON into a uniform in-memory model

- `lib/store.js`
  - wrap database access in one API
  - keep SQL out of command handlers
  - expose operations like `listUpcoming`, `getShow`, `createShow`, `cancelShow`, `archiveShow`

- `lib/migrate.js`
  - copy `shows.json` into the database once
  - preserve original dates and optional notes

## Command design

### Phase 1 commands

#### `/upcoming`

- anyone can run it
- list upcoming shows from `shows.json`
- format the response cleanly for Discord and screenshots
- decide whether the response is public or ephemeral and keep that decision consistent

#### `/announce`

- admin/booker only
- accepts plain text args for artist, date, doors time, age restriction, price, and notes
- normalizes the date for display
- posts an embed to `#upcoming-shows`
- pins the announcement
- replies with a success message

#### `/unpin-old`

- admin only
- lists pinned messages in `#upcoming-shows`
- unpins any show whose date has passed
- returns the count of removed pins

### Phase 2 commands

#### `/add-show`

- admin/booker only
- creates a DB row first
- posts and pins the Discord announcement
- stores the Discord message/channel IDs on the row
- returns the show ID

#### `/show <id>`

- anyone or staff-only, depending on the venue’s comfort level
- returns the full row for a show by ID
- useful for quick checks and staff coordination

#### `/cancel-show <id>`

- admin/booker only
- marks the show as cancelled
- unpins the original announcement
- posts a cancellation notice in `#upcoming-shows`

#### `/archive-show <id>`

- admin only
- marks the show archived
- unpins the original announcement if needed
- keeps the row available for future lookup

#### `/past-shows`

- return the last 5 archived or past shows
- the access level should be decided up front and documented

#### `/unpin-old`

- remains available in Phase 2
- now also marks affected rows as archived
- should share logic with the scheduled cleanup path

## Permission gating

The spec’s role model is simple enough that the bot can enforce it entirely in JS:

- read commands run for everyone
- write commands check the actor’s member roles
- if the user lacks access, reply ephemerally with:

```text
❌ You don't have permission to use this command.
```

Because `ctx.member` is already available in command dispatch, the helper can be written as a normal JS utility that reads the configured role IDs and compares them to the member role list.

## Presentation and announcement style

This bot is part operations tool, part public-facing promo surface. The announcement and upcoming-show rendering should therefore be visually clean:

- a short title line
- clear date and time
- age restriction and price on separate fields or lines
- optional notes at the bottom
- enough whitespace to look good in Discord and in screenshots

The embed format should be stable, because the venue may later want to reuse those announcement screenshots on Instagram or the site.

## Phase 1 implementation plan

### 1. Create the bot shell

Start with a named bot under `examples/discord-bots/show-space/index.js`:

```js
const { defineBot } = require("discord")

module.exports = defineBot(({ command, event, configure }) => {
  configure({
    name: "show-space",
    description: "Venue operations bot for announcements and show management",
    category: "venues",
    run: {
      fields: {
        upcomingShowsChannelId: { type: "string", help: "Public show channel ID" },
        announcementsChannelId: { type: "string", help: "Public announcements channel ID" },
        staffChannelId: { type: "string", help: "Private staff log channel ID" },
        adminRoleId: { type: "string", help: "Admin role ID" },
        bookerRoleId: { type: "string", help: "Booker role ID" },
      },
    },
  })

  event("ready", async (ctx) => {
    ctx.log.info("show-space bot ready", { user: ctx.me && ctx.me.username })
  })

  // commands go here
})
```

### 2. Build the data helpers

Use a `shows.json` file and a small loader to:

- parse dates into a stable form
- filter out past shows
- sort by date ascending
- map the raw JSON into a shape the renderer can use

### 3. Build announcement rendering

Make one `renderAnnouncement(show)` helper that can be reused by both `/announce` and `/add-show`.

### 4. Wire the Discord operations

The bot already has host operations for:

- sending channel messages
- pinning messages
- unpinning messages
- listing pinned messages

So the command handlers only need to supply channel IDs and payloads.

### 5. Add role checks

Use one helper like:

```js
function hasShowWriteAccess(ctx) {
  const roles = (ctx.member && ctx.member.roles) || []
  return roles.includes(ctx.config.adminRoleId) || roles.includes(ctx.config.bookerRoleId)
}
```

That keeps role gating consistent across `/announce`, `/add-show`, `/cancel-show`, `/archive-show`, and `/unpin-old`.

## Phase 2 implementation plan

### 1. Hide persistence behind one store

The store module should expose a very small domain API and nothing else:

```js
const store = {
  listUpcoming(limit = 10) {},
  getShow(id) {},
  createShow(show) {},
  cancelShow(id) {},
  archiveShow(id) {},
  listPast(limit = 5) {},
  setDiscordMessage(id, channelId, messageId) {},
  findExpiredPinnnedShows(today) {},
}
```

Command handlers should not know whether that store is SQLite today, Postgres later, or a future service behind the same JS interface.

### 2. Add schema and migration code

The spec’s schema is a good starting point. If the implementation remains inside this repository, keep the schema behind the store layer and add a one-time import from `shows.json`.

### 3. Add staff-safe command flows

The most important new flows are:

- `/add-show` saves first, then posts the Discord message
- `/cancel-show` updates the DB, then unpins and posts the cancellation notice
- `/archive-show` keeps records without cluttering the active queue

This order matters because the Discord message ID must be recorded and the DB should remain the source of truth.

### 4. Make cleanup idempotent

The archive job should be safe to run twice.

That means:

- only archive rows whose dates are actually past
- skip rows already cancelled or archived
- ignore missing or already-unpinned messages
- log the count of rows affected

### 5. Prefer an external scheduler for daily cleanup

The runtime docs only advertise `sleep` in the timer module, not a full cron or interval API. For that reason, the safest plan is to keep the cleanup logic in a reusable JS function and invoke it through:

- a host-side cron job,
- a deployment scheduler,
- or an explicit daily command runner

rather than assuming the bot process should self-schedule forever.

## Testing strategy

### Unit tests

Add pure JS tests for:

- date parsing and display normalization
- role gating
- show filtering/sorting
- embed rendering
- JSON-to-model mapping

### Runtime tests

Add bot runtime tests for:

- `/upcoming` with seeded data
- `/announce` posting and pinning
- `/unpin-old` removing expired pins
- `/add-show` storing the Discord message ID
- `/cancel-show` unpinning and posting a cancellation notice
- `/archive-show` respecting admin-only access

### Operational checks

Document the following operator steps:

- how to set channel and role IDs
- how to seed the initial `shows.json`
- how to verify the bot has pin/unpin permissions
- how to run the daily archive routine safely

## Risks and open questions

### 1. Database backend choice

The biggest unresolved issue is the spec’s Postgres preference versus the current runtime’s SQLite-oriented `require("database")` module. This ticket should either:

- commit to SQLite first and keep the store abstraction narrow, or
- split out a host/runtime ticket for a Postgres adapter.

### 2. Public vs ephemeral `/upcoming`

The spec leaves this open. For a public venue bot, a public reply is probably the default, but staff may prefer ephemeral if the output gets noisy.

### 3. Public vs staff-only `/past-shows`

This should be decided early, because it affects whether the command belongs in the public command list or a staff-only path.

### 4. Timezone handling

The spec asks for natural input like `Apr 25` and normalized display like `Fri Apr 25, 2025`. That means the bot needs a fixed venue timezone, otherwise date boundaries will drift for staff members in different regions.

### 5. Scheduling model

If the bot is deployed as a long-lived process, an internal interval may be fine. If it is deployed as a managed bot process, external scheduling will be easier to reason about and test.

## Recommended initial delivery order

1. implement `show-space` as a new example bot
2. finish Phase 1 with `shows.json`
3. introduce a JS store wrapper
4. migrate to persistent storage
5. add archival automation and staff logging
6. update the bot inventory docs and reference pages

That sequence gets staff value early while leaving enough architectural room for the future web dashboard and booking pipeline.

## Appendix A — attached spec, verbatim

````text
# Show Space Discord Bot — Developer Spec
## Phases 1 & 2

**Project:** Internal tooling for an artist/show space. A Discord bot backed by a database to help staff manage bookings, post and pin show announcements, and keep the server organized.

**Scope of this document:** Phase 1 (bot only, no DB) and Phase 2 (bot + persistent database). No public webpage yet.

---

## Context

We run a physical artist/show space. Staff manage bookings, announce shows, and coordinate via a Discord server. Right now everything is manual — we want a bot to handle show announcements, auto-pinning, and eventually a full booking pipeline.

The end state (beyond this doc) includes a public webpage and a staff web dashboard, but for now everything lives in Discord.

**Discord server has these relevant channels:**
- `#upcoming-shows` — public, where show announcements get posted and pinned
- `#announcements` — public, broader announcements
- `#staff` — private, internal coordination
- `#booking-requests` — private, where new artist submissions appear (Phase 3, not in scope here)

**Roles:**
- `@admin` — full access to all bot commands
- `@booker` — can manage shows (add, cancel, announce)
- Everyone else — read only, no bot commands

---

## Phase 1 — Bot Only, No Database

**Goal:** Get a working bot in the server that can post and pin formatted show announcements. No persistence — data can live in a JSON file or be passed inline via command. Ship fast, validate the workflow.

### Tech
- Discord.js (Node.js) or discord.py — developer's choice
- No database
- Hosted anywhere (even a cheap VPS or fly.io free tier)
- A simple `shows.json` file for seeding initial upcoming shows

---

### Commands

#### `/upcoming`
- **Who:** Anyone
- **What:** Bot replies with a formatted list of upcoming shows pulled from `shows.json`
- **Output format:**
  ```
  📅 Upcoming Shows

  🎵 Artist Name — Fri Apr 25
  Doors 7pm | 18+ | $10 suggested

  🎵 Another Artist — Sat May 3
  Doors 8pm | All ages | Free
  ```

#### `/announce`
- **Who:** `@admin`, `@booker` only
- **Args:** `artist`, `date`, `doors_time`, `age_restriction`, `price` (all text, keep it simple)
- **What:** Bot posts a formatted show announcement embed to `#upcoming-shows` and pins it
- **Embed fields:** Artist, Date, Doors, Age, Price, optional Notes
- **On success:** Bot replies "✅ Posted and pinned in #upcoming-shows"

#### `/unpin-old`
- **Who:** `@admin` only
- **What:** Bot scans pins in `#upcoming-shows`, unpins any whose date has passed
- **On completion:** Bot replies with a count of how many pins were removed

### Role Gating
All write commands (`/announce`, `/unpin-old`) check that the user has `@admin` or `@booker` role. If not, bot replies with an ephemeral "❌ You don't have permission to use this command."

### Notes / Constraints
- Keep the embed design clean — we may want to screenshot these for Instagram
- `/unpin-old` is manual for now; auto-scheduling comes in Phase 2
- Date format: accept natural input like `Apr 25` or `2025-04-25`, normalize to display as `Fri Apr 25, 2025`

---

## Phase 2 — Bot + Database

**Goal:** Shows persist in a real database. Staff can manage shows by ID, the bot auto-cleans old pins daily, and we have a foundation to build booking and web features on top of.

### Tech
- Same bot stack as Phase 1
- **PostgreSQL** (preferred) — chosen for compatibility with the planned web dashboard (Next.js / SvelteKit + Prisma or similar)
- ORM: Prisma (Node) or SQLAlchemy (Python) — whichever matches the bot stack
- Bot should expose a simple internal API or module for DB access (we'll reuse this in Phase 4 when we add a web layer)

---

### Database Schema

```sql
-- Shows table
CREATE TABLE shows (
  id          SERIAL PRIMARY KEY,
  artist      TEXT NOT NULL,
  date        DATE NOT NULL,
  doors_time  TEXT,               -- e.g. "7:00 PM"
  age         TEXT,               -- e.g. "18+", "All Ages"
  price       TEXT,               -- e.g. "$10 suggested", "Free"
  notes       TEXT,
  status      TEXT DEFAULT 'confirmed',  -- 'confirmed' | 'cancelled' | 'archived'
  discord_message_id  TEXT,       -- ID of the pinned Discord message, for unpinning
  discord_channel_id  TEXT,       -- Channel where it was posted
  created_at  TIMESTAMP DEFAULT NOW(),
  updated_at  TIMESTAMP DEFAULT NOW()
);
```

---

### Updated / New Commands

#### `/add-show`
- **Who:** `@admin`, `@booker`
- **Args:** `artist`, `date`, `doors_time`, `age`, `price`, `notes` (optional)
- **What:** Saves show to DB, posts embed to `#upcoming-shows`, pins it, stores `discord_message_id`
- **Reply:** "✅ Show added — ID #42. Posted and pinned."

#### `/upcoming`
- **Who:** Anyone
- **What:** Now pulls from DB instead of JSON. Returns next N confirmed shows ordered by date.

#### `/show <id>`
- **Who:** Anyone
- **What:** Returns full details for a specific show by ID. Good for staff to quick-check.

#### `/cancel-show <id>`
- **Who:** `@admin`, `@booker`
- **What:**
  1. Sets `status = 'cancelled'` in DB
  2. Unpins the original Discord message (using stored `discord_message_id`)
  3. Posts a brief cancellation notice in `#upcoming-shows`: "⚠️ [Artist] on [Date] has been cancelled."
- **Reply:** "✅ Show #42 cancelled and unpinned."

#### `/archive-show <id>`
- **Who:** `@admin` only
- **What:** Sets `status = 'archived'`, unpins message. For shows that happened — keeps the record without cluttering upcoming.

#### `/past-shows`
- **Who:** Anyone (or staff-only — your call)
- **What:** Returns last 5 archived/past shows. Useful for reference.

#### `/unpin-old`
- **Who:** Still available as manual command
- **What:** Now also marks affected shows as `archived` in DB

### Automation

#### Daily auto-archive job
- Runs once per day (cron or setInterval)
- Finds shows where `date < today` and `status = 'confirmed'`
- Sets them to `archived`, unpins their Discord messages
- Posts a quiet log to `#staff`: "📦 Auto-archived 2 past shows."

---

### Migration from Phase 1

If Phase 1 populated a `shows.json`, write a one-time migration script to seed the DB from that file. Keep it simple — just insert rows.

---

## Out of Scope for Phases 1 & 2

These are coming but not now:
- Artist booking submission flow (`/book-request`, DM-based intake)
- Public webpage or calendar
- Staff web dashboard
- Volunteer/shift scheduling
- Attendance or revenue logging
- Email/newsletter integration

---

## Open Questions for Developer

1. Node (discord.js) or Python (discord.py)? Either works — just be consistent so Phase 4 web layer can share the DB layer.
2. Where are we hosting this? Needs to stay alive 24/7 for the daily cron job in Phase 2.
3. Do we want the `/upcoming` output as an ephemeral reply (only visible to the requester) or posted publicly in the channel?
4. Should `/past-shows` be public or staff-only?

Create a detailed implementation and analysis guide, store the attached spec in the guide too (verbatim), and add tasks to implement it in js.
````

## Appendix B — practical implementation notes

If this ticket is executed in code later, the first commit should create the bot shell and Phase 1 commands. The second commit should add the persistence wrapper and the DB-backed show management commands. The third commit should add the cleanup path, migration, docs, and operator runbook.
