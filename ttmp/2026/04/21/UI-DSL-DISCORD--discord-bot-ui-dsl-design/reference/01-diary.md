---
Title: Diary
Ticket: UI-DSL-DISCORD
Status: active
Topics:
    - discord-bots
    - ui-dsl
    - framework-design
    - js-goja
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Step-by-step narrative of the UI DSL analysis and design process"
LastUpdated: 2026-04-21T08:04:14.256013672-04:00
---

# Diary

## Goal

Analyze the `examples/discord-bots/knowledge-base/` bot, identify repetitive UI construction patterns, and design a UI DSL that makes Discord bot code more elegant, declarative, and maintainable. Keep a chronological record of the investigation, design decisions, and deliverables.

---

## Step 1: Code Archaeology — Understanding the Current Patterns

### Prompt Context

**User prompt (verbatim):** "Analyze the @examples/discord-bots/knowledge-base/ bot and make some proposals for a UI DSL to make the code more elegant. new docmgr ticket for the UI DSL, keep a diary. Create a detail design brainstorm and different UI DSL examples for different usecases upload ui dsl doc to remarkable when done."

**Assistant interpretation:** The user wants a deep analysis of the knowledge-base Discord bot example, identification of UI boilerplate and repetitive patterns, and concrete proposals for a declarative UI DSL that could be integrated into the discord-js framework. This should be tracked in a docmgr ticket with a diary, and the final deliverable uploaded to reMarkable.

**Inferred user intent:** The knowledge-base bot has grown organically and now shows significant boilerplate in message construction, component wiring, state management, and modal building. The user wants a DSL that elevates the abstraction level so bot authors can express UI intent declaratively rather than imperatively assembling raw Discord API payloads.

### What I did
- Read the entire knowledge-base bot codebase:
  - `index.js` — main bot definition with commands, components, modals, autocompletes
  - `lib/store.js` — SQLite persistence layer
  - `lib/render.js` — embed and message rendering functions
  - `lib/review.js` — review queue state management + component building
  - `lib/search.js` — search state management + component building
  - `lib/capture.js` — message capture logic
  - `lib/reactions.js` — reaction promotion logic
  - `lib/register-knowledge-bot.js` — simplified registration variant
- Read the framework runtime (`internal/jsdiscord/bot.go`, `runtime.go`, `descriptor.go`) to understand available APIs
- Read other example bots (`poker`, `interaction-types`) for comparative patterns
- Created docmgr ticket `UI-DSL-DISCORD`
- Created this diary document

### What I found (key patterns)

**Pattern A: Raw payload construction everywhere**
Every response is built by hand-crafting `{ content, embeds: [{ title, description, color, fields: [...] }], components: [...] }`. This is verbose and error-prone.

**Pattern B: Component ID strings scattered across files**
Search uses `SEARCH_COMPONENTS = { select, previous, next, open, source, export }`. Review uses `REVIEW_COMPONENTS = { select, verify, edit, stale, reject, source }`. These are defined in one file, referenced in another, and wired in a third (`index.js`).

**Pattern C: State management boilerplate**
Both search and review implement nearly identical state key construction, load/save/normalize cycles, pagination, and selection tracking. ~100 lines each of cookie-cutter state logic.

**Pattern D: Modal building is deeply nested**
`buildTeachModal()` and `buildEntryModal()` contain 5 action rows each, each with a single textInput. The nesting is 4+ levels deep.

**Pattern E: Ephemeral/reply branching**
Handlers frequently branch between ephemeral error responses and public success responses with nearly identical embeds.

**Pattern F: Duplicate command aliases**
`ask`/`kb-search`, `article`/`kb-article`, `review`/`kb-review`, `recent`/`kb-recent` — pairs of commands with identical bodies.

### What was tricky to build
Nothing tricky yet — this is pure analysis. The challenge will be designing a DSL that:
- Feels native to the existing `defineBot` API
- Doesn't require Go-side changes (or minimizes them)
- Can be implemented as a JS module (e.g., `require("discord/ui")`)
- Handles Discord's constraints (5 action rows max, 25 select options max, etc.)
- Still allows escape hatches to raw payloads

### What should be done in the future
- Prototype the DSL as a JS module
- Rewrite `knowledge-base` bot using the DSL to validate
- Consider framework-level integration (auto-generating component IDs, implicit state scoping)

---

## Step 2: Design Brainstorm — DSL Concepts

### What I did
- Brainstormed 5 different DSL styles/flavors
- Analyzed trade-offs for each
- Selected a preferred direction: **"Discord UI Builder" — a fluent, chainable API with declarative helpers**

### DSL flavors considered

1. **React-like JSX (declarative, hierarchical)**
   - Pros: Very readable for complex layouts
   - Cons: Requires transpilation or runtime JSX transform; heavy for simple bots

2. **Template-string tagged literals**
   - Pros: Compact, good for text-heavy UIs
   - Cons: Weak typing, hard to compose components

3. **Fluent builder chain (`ui.embed().title(...).field(...).build()`)**
   - Pros: IDE-friendly, composable, no transpilation
   - Cons: Verbose for simple cases

4. **Declarative object schema with helpers**
   - Pros: Close to existing Discord API, easy to adopt
   - Cons: Doesn't solve the verbosity problem

5. **Hybrid: Declarative layouts + fluent embeds + state helpers**
   - Pros: Best of both worlds; addresses all identified patterns
   - Cons: Slightly larger API surface

**Selected:** Option 5, implemented as a `discord-ui` JS module that bot authors `require()`.

### What should be done in the future
- Write detailed DSL specification with examples
- Create use-case-specific examples (search UI, review queue, modal forms, announcements)
- Compare before/after code size and readability

---

## Step 3: Detailed DSL Design + Examples

### What I did
- Wrote comprehensive design documents:
  - `design/01-ui-dsl-proposal.md` — Core DSL specification
  - `design/02-dsl-use-cases.md` — Before/after examples for 5 use cases
  - `design/03-implementation-sketch.md` — How to build it as a JS module
- Related source files to the ticket docs

### Key design decisions
- **Module name:** `discord-ui` (loaded via `require("discord-ui")` or bundled)
- **Core abstractions:** `View`, `Embed`, `ComponentRow`, `ModalForm`, `State`
- **Auto-ID generation:** Components get scoped IDs automatically (`bot:command:action`)
- **State helper:** `ui.state(ctx, "search")` returns a namespaced key-value store
- **Escape hatch:** All builders expose `.raw()` to get the Discord payload

### Code review instructions
- Review `design/01-ui-dsl-proposal.md` for API ergonomics
- Review `design/02-dsl-use-cases.md` for real-world applicability
- Check that every "before" snippet is traceable to actual knowledge-base code

---

## Step 4: Upload to reMarkable

### What I did
- Bundled all design documents into a single PDF with ToC
- Uploaded to `/ai/2026/04/21/UI-DSL-DISCORD/` on reMarkable
- Verified upload with `remarquee cloud ls`

### What worked
- `remarquee upload bundle` successfully created a single PDF with clickable ToC

### What didn't work
- N/A

---

## Summary

| Step | Deliverable | Status |
|------|-------------|--------|
| 1 | Code analysis + pattern identification | ✅ Done |
| 2 | DSL brainstorm + direction selection | ✅ Done |
| 3 | Detailed design docs + use cases | ✅ Done |
| 4 | reMarkable upload | ✅ Done |
