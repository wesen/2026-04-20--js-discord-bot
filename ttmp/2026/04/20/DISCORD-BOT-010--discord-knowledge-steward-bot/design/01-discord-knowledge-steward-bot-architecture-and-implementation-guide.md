---
Title: Discord Knowledge Steward Bot Architecture and Implementation Guide
Ticket: DISCORD-BOT-010
Status: active
Topics:
    - backend
    - chat
    - javascript
    - goja
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: examples/discord-bots/README.md
      Note: Named-bot repository flow and runtime notes inform the implementation plan
    - Path: examples/discord-bots/knowledge-base/index.js
      Note: |-
        Current knowledge-base example is the closest starting point for a record-enabled bot
        MVP bot entrypoint with SQLite-backed capture
    - Path: examples/discord-bots/knowledge-base/lib/capture.js
      Note: Passive capture heuristics for messageCreate and teach modal payloads
    - Path: examples/discord-bots/knowledge-base/lib/render.js
      Note: Discord card and queue rendering for knowledge entries
    - Path: examples/discord-bots/knowledge-base/lib/store.js
      Note: SQLite-backed knowledge store implementation and schema
    - Path: internal/jsdiscord/bot.go
      Note: DispatchRequest defines the data and response helpers available to knowledge-capture handlers
    - Path: internal/jsdiscord/host_ops_messages.go
      Note: Message lookup and moderation helpers support source citation
    - Path: internal/jsdiscord/knowledge_base_runtime_test.go
      Note: Runtime coverage for capture
    - Path: internal/jsdiscord/runtime.go
      Note: JS bot definition DSL and available runtime hooks shape the proposed command/event surface
ExternalSources: []
Summary: |
    Architecture and implementation guide for a transparent, community-first Discord bot that listens to chat, records candidate knowledge, and routes it through human review before it becomes canonical.
LastUpdated: 2026-04-20T22:40:00-04:00
WhatFor: Explain how to build a shared knowledge steward bot on the current JS Discord runtime.
WhenToUse: Use when designing or implementing the community knowledge capture and curation bot.
---



# Discord Knowledge Steward Bot Architecture and Implementation Guide

## Executive Summary

The goal of this ticket is to build a Discord bot that behaves like a modern IRC knowledge helper: it listens to public chat, records candidate knowledge, and helps a group refine that information into a durable shared memory.

The bot is not meant to be a private assistant that quietly answers one user's question. It is meant to be a **community knowledge steward**. Its core responsibilities are to observe, capture, curate, and resurface knowledge in a way that is visible to the channel and easy for humans to correct.

The current runtime already exposes the main JS extension points needed for this project: commands, events, components, modals, autocomplete, and a reasonably rich Discord host surface. What is still missing is a persistence model for knowledge entries, a capture/review workflow, and a retrieval experience that treats channel knowledge as a shared artifact rather than a one-off response.

## Problem Statement and Scope

Discord channels accumulate useful answers in a very human way: someone posts a fix, another person shares a link, a third person clarifies the detail, and the thread ends with the channel having learned something important. Without a steward, that knowledge is easy to lose.

This bot should solve that problem by:

1. listening for useful information in chat,
2. recording candidate knowledge entries with source attribution,
3. letting people verify, edit, merge, or mark entries stale,
4. surfacing canonical answers later with citations, and
5. keeping the process transparent so the channel understands what the bot captured and why.

### In scope

- Passive capture from chat events, especially `messageCreate`.
- Explicit capture via commands and modals.
- Review and curation actions from buttons or slash commands.
- Search, lookup, and answer retrieval from the curated corpus.
- Source attribution and entry history.
- Opt-in channel or guild scoping.

### Out of scope for the first implementation

- Multi-tenant SaaS hosting.
- Automatic truth decisions with no human review.
- Advanced semantic embeddings or external vector databases.
- Direct DM-first workflows.
- Moderation actions that are unrelated to knowledge capture.

The first milestone should stay small enough to ship as a named bot under `examples/discord-bots/`, using the current JS bot runtime and the named-bot runner flow already documented in the repository.

## Current-State Analysis

This project should be built on top of the runtime that already exists, not against it.

### JS bot definition and dispatch are already in place

The runtime's `defineBot` entry point already exposes the core JS bot building blocks: `command`, `event`, `component`, `modal`, `autocomplete`, and `configure` (`internal/jsdiscord/runtime.go:124-143`). That means the knowledge steward bot can be assembled as a pure JS bot without adding a new top-level bot abstraction.

The dispatched context is also already broad enough for the first version of this bot. `DispatchRequest` already carries message, guild, channel, member, reaction, metadata, config, and Discord host references, plus response helpers such as `reply`, `followUp`, `edit`, `defer`, and `showModal` (`internal/jsdiscord/bot.go:86-111`). In other words, the JS side already receives the information needed to inspect chat content and respond with structured knowledge cards.

### The host surface already includes the operations the bot will need

The Discord host now exposes message, member, role, guild, and channel helpers. Of particular importance for a knowledge steward bot are the message operations in `internal/jsdiscord/host_ops_messages.go` and the member/role lookup operations in `internal/jsdiscord/host_ops_members.go` and `internal/jsdiscord/host_ops_roles.go`:

- fetch/list messages
- pin/unpin/list pinned messages
- bulk delete messages when needed
- fetch/list members
- fetch/list roles
- fetch and update channels
- add/remove member roles and timeout/kick/ban/unban flows

Those APIs matter because a knowledge bot often needs to cite source messages, inspect who said what, and optionally pin canonical references or clean up noisy duplicates. The bot does not need all moderation functions on day one, but the host already supports the surrounding workflows.

### The repository already has a knowledge-oriented bot example, but it is read-only

The existing `examples/discord-bots/knowledge-base/index.js` script is a useful starting point, but it is fundamentally a search bot. It configures a `knowledge-base` bot, exposes `kb-search` and `kb-article`, and listens to `messageCreate` only to react to the `!kb` trigger (`examples/discord-bots/knowledge-base/index.js:1-79`). It demonstrates reading from a docs index, but not recording community knowledge back into a store.

That distinction matters: this ticket is not just a rename or a UX polish pass. It is a write-enabled evolution of the current read-only example.

### The CLI help system is already wired for documentation

`cmd/discord-bot/root.go` already loads embedded docs into the Glazed help system (`cmd/discord-bot/root.go:30-35`). That means the final implementation can and should be documented in embedded help pages and ticket docs without inventing a second documentation system.

### The named-bot runner model is already established

The repository's example bot layout and bot CLI tests show that named bots are the intended operational model. The example repository documents `bots list`, `bots help`, and `bots run` against `examples/discord-bots`, and the bot CLI tests assert the same named-bot behavior (`examples/discord-bots/README.md:1-47`, `internal/botcli/command_test.go:13-59`, `internal/botcli/command_test.go:76-128`). The knowledge steward should follow that same pattern.

## Gap Analysis

The current runtime is a good foundation, but the knowledge steward bot still needs several missing pieces.

### 1. No persistence layer for knowledge entries

There is no built-in knowledge store yet. The runtime provides event and host access, but no durable entry store, version history, or queue system. A knowledge steward needs a persistence layer that can store candidates, review decisions, aliases, citations, and status transitions.

### 2. No capture pipeline

The existing bot examples answer commands or log events, but they do not turn chat content into structured records. The bot needs heuristics that can detect likely knowledge candidates from ordinary messages.

### 3. No curation workflow

The bot needs a human-in-the-loop review loop. People should be able to verify, edit, merge, reject, or mark stale knowledge entries without editing raw storage by hand.

### 4. No retrieval layer with citations

A search box is not enough. The bot should answer with source-backed knowledge cards and show where each answer came from. That means every entry needs source metadata and a representation that is easy to render in Discord.

### 5. No transparency model

Because the bot listens to chat, it must be explicit about what it records. Capturing knowledge silently would be a trust problem. The design must include visible audit breadcrumbs and clear scope boundaries.

## Proposed Architecture and APIs

### High-level architecture

The bot should be split into four logical pieces:

1. **Capture layer** — listens to Discord events and decides whether a message is a knowledge candidate.
2. **Store layer** — persists entries, versions, sources, and review state.
3. **Curation layer** — exposes commands, buttons, and modals for human review.
4. **Retrieval layer** — searches the corpus and returns answers with citations.

In the first implementation, these layers can live mostly in JS under `examples/discord-bots/knowledge-base/`, backed by a small helper module or local store. The key design goal is not the storage technology itself; it is the ability to keep a consistent knowledge lifecycle.

### Suggested knowledge entry model

A knowledge entry should be explicit about status and provenance.

```ts
type KnowledgeEntry = {
  id: string
  title: string
  summary: string
  body: string
  tags: string[]
  aliases: string[]
  status: "draft" | "review" | "verified" | "stale" | "rejected"
  source: {
    guildId: string
    channelId: string
    messageId: string
    authorId: string
    jumpUrl?: string
    capturedAt: string
  }
  versions: Array<{
    version: number
    editedBy: string
    editedAt: string
    note?: string
  }>
  relatedIds: string[]
  confidence: number
}
```

This model is intentionally simple. It preserves the minimum facts needed to explain where knowledge came from and how it changed over time.

### Suggested runtime surface

The bot can be configured with the current JS DSL:

```js
module.exports = defineBot(({ command, event, component, modal, autocomplete, configure }) => {
  configure({
    name: "knowledge-base",
    description: "Capture, curate, and surface shared knowledge from chat",
    category: "knowledge"
  })

  event("messageCreate", async (ctx) => { /* capture candidates */ })
  command("remember", { /* explicit capture */ }, async (ctx) => { /* open modal */ })
  command("ask", { /* retrieval */ }, async (ctx) => { /* search corpus */ })
  component("knowledge:verify", async (ctx) => { /* promote draft */ })
  modal("knowledge:submit", async (ctx) => { /* finalize entry */ })
  autocomplete("ask", "query", async (ctx) => { /* topic and alias suggestions */ })
})
```

The first version should not require a new host API unless a truly shared persistence service becomes necessary. The runtime already gives the bot enough hooks to express the full workflow.

### Visibility and transparency rules

The bot should follow three simple rules:

1. **Opt in by scope** — only record in approved channels or guilds.
2. **Leave breadcrumbs** — when a message becomes a knowledge candidate, post a visible note or review-card reference so the channel can see it happened.
3. **Require human confirmation** — candidates do not become canonical until a person verifies them.

Those rules keep the bot trustworthy and reduce the risk that it becomes a hidden archival system.

## Pseudocode and Key Flows

### Capture flow

```js
event("messageCreate", async (ctx) => {
  if (!isCaptureEnabled(ctx.guild, ctx.channel)) return
  if (!looksLikeKnowledge(ctx.message.content)) return
  if (isDuplicateOrLowValue(ctx.message)) return

  const candidate = await buildCandidate(ctx)
  await store.appendDraft(candidate)

  if (candidate.confidence >= 0.7) {
    await ctx.reply({
      content: `Saved a draft knowledge entry: **${candidate.title}**`,
      embeds: [renderKnowledgeCard(candidate)],
      components: [reviewActions(candidate.id)]
    })
  }
})
```

The heuristics should be transparent and easy to debug. Start with simple signals such as code blocks, direct instruction language, repeated answers, and reaction-based promotion from trusted users. Do not begin with a black-box classifier.

### Explicit capture flow

```js
command("remember", { description: "Save a useful message as knowledge" }, async (ctx) => {
  await ctx.showModal({
    customId: `knowledge:submit:${ctx.message?.id || "new"}`,
    title: "Save knowledge",
    components: [/* title, summary, tags, source fields */]
  })
})
```

A modal is the right UX for explicit capture because it lets the user add a short title, summary, tags, and source hint without forcing a long command line.

### Review flow

```js
component("knowledge:verify", async (ctx) => {
  await store.markVerified(ctx.values.entryId, ctx.user.id)
  await ctx.edit({
    content: "Knowledge entry verified.",
    embeds: [renderKnowledgeCard(await store.get(ctx.values.entryId))]
  })
})
```

Review actions should always preserve history. Verification should add a new version or state transition, not overwrite the original source snapshot.

### Retrieval flow

```js
command("ask", { description: "Search the shared knowledge base" }, async (ctx) => {
  const matches = await store.search(ctx.args.query)
  const answer = synthesizeAnswer(matches)
  return {
    content: answer.summary,
    embeds: answer.citations.map(renderCitationEmbed),
    ephemeral: true
  }
})
```

The retrieval layer should rank verified entries ahead of drafts and stale entries, then show citations prominently. If confidence is low, the bot should say so and offer a path to create a new entry instead of pretending certainty.

## Implementation Phases

### Phase 1 — foundation and capture

- Choose the MVP storage backend.
- Create the bot-local data model and store helpers.
- Add `messageCreate` capture heuristics.
- Record source metadata for every candidate.
- Add an audit breadcrumb or review-card reply when a candidate is created.

### Phase 2 — curation and human review

- Add `/remember` and `/teach` commands.
- Add modal-based entry editing.
- Add verify, reject, stale, and merge actions.
- Preserve version history.
- Add duplicate detection before entry creation.

### Phase 3 — retrieval and synthesis

- Add `/ask`, `/search`, `/article`, and `/recent`.
- Render knowledge cards with citations and status.
- Add autocomplete for aliases, tags, and article titles.
- Tune result ranking so verified entries win over drafts.

### Phase 4 — maintenance and rollout

- Add staleness detection.
- Add export support for the corpus.
- Add tests for capture, curation, and retrieval.
- Document the required Discord intents and permissions.
- Add a small operator playbook for troubleshooting capture and review flows.

## Test Strategy

The test strategy should match the bot lifecycle.

1. **Runtime unit tests** — verify the capture heuristics, store transitions, and render helpers.
2. **Bot help tests** — verify the named bot appears in `bots list` and `bots help` output.
3. **Interaction flow tests** — verify commands, components, and modals for explicit capture and review.
4. **Smoke tests** — start the bot with the named-bot runner and confirm the capture and retrieval commands work in a live guild.
5. **Negative tests** — ensure the bot ignores disabled channels, low-confidence noise, malformed payloads, and duplicate entries.

The most important test is not just whether the bot can answer a question; it is whether the bot can explain where the answer came from and preserve a review trail.

## Risks, Alternatives, and Open Questions

### Risks

- A passive capture bot can become noisy if it records too aggressively.
- Knowledge quality can degrade if verified and draft entries are not visually distinct.
- A local file store is simple, but it may not scale across multiple processes or deployments.
- If the bot ever auto-replies too often, the channel may perceive it as spam rather than a steward.

### Alternatives considered

- **Write-only wiki style** — simpler, but less conversational and less useful in the flow of chat.
- **Search-only bot** — easier to build, but misses the main value of capturing tribal knowledge in the first place.
- **External persistence from day one** — more scalable, but adds operational overhead before the interaction model is proven.

### Open questions

- What is the first storage backend: JSON, SQLite, or an external service?
- Which channels are opted into passive capture by default?
- Should candidates be announced in-channel or only in a review channel?
- Should reactions from trusted users trigger promotion automatically, or only queue review?
- How much answer synthesis belongs in the bot versus in human-curated summaries?

## References

### Repository files

- `internal/jsdiscord/runtime.go` — JS bot definition entry point and runtime DSL.
- `internal/jsdiscord/bot.go` — dispatch context, event data, and reply helpers.
- `internal/jsdiscord/host_ops_messages.go` — message fetch/list/pin/unpin/bulk delete helpers.
- `internal/jsdiscord/host_ops_members.go` — member lookup and moderation helpers.
- `internal/jsdiscord/host_ops_roles.go` — role lookup helpers.
- `examples/discord-bots/knowledge-base/index.js` — current read-only knowledge bot example.
- `examples/discord-bots/README.md` — named-bot flow and example repository guidance.
- `cmd/discord-bot/root.go` — embedded help system wiring.
- `internal/botcli/command_test.go` — named-bot list/help/run behavior.

### Related docs

- `ttmp/2026/04/20/DISCORD-BOT-010--discord-knowledge-steward-bot/reference/01-discord-knowledge-steward-bot-implementation-guide-and-api-sketches.md`
- `ttmp/2026/04/20/DISCORD-BOT-010--discord-knowledge-steward-bot/reference/02-diary.md`
