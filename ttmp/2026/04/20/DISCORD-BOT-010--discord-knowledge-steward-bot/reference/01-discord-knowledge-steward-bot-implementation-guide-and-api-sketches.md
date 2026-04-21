---
Title: Discord Knowledge Steward Bot Implementation Guide and API Sketches
Ticket: DISCORD-BOT-010
Status: active
Topics:
    - backend
    - chat
    - javascript
    - goja
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/discord-bot/root.go
      Note: Embedded help loading is the path for surfacing the implementation guidance in CLI help
    - Path: examples/discord-bots/README.md
      Note: Updated example repository guidance and runtime notes for the knowledge steward bot
    - Path: examples/discord-bots/knowledge-base/index.js
      Note: |-
        Shows the current read-only search/article bot that should be extended into a capture-capable steward
        Concrete JS composition for the MVP bot
    - Path: examples/discord-bots/knowledge-base/lib/store.js
      Note: SQLite schema and store verbs for drafts
    - Path: internal/botcli/command_test.go
      Note: Confirms named-bot list/help/run expectations that the new bot should preserve
    - Path: internal/jsdiscord/host_ops_messages.go
      Note: Message helper surface supports citation
    - Path: internal/jsdiscord/knowledge_base_runtime_test.go
      Note: Runtime test showing the end-to-end capture/search/review path
ExternalSources: []
Summary: |
    Practical implementation notes, command sketches, data model guidance, and rollout steps for the first Discord knowledge steward bot.
LastUpdated: 2026-04-20T22:40:00-04:00
WhatFor: Provide a concrete build plan for the MVP knowledge-capture and curation workflow.
WhenToUse: Use when implementing the bot example or reviewing the bot's command and data contract.
---



# Discord Knowledge Steward Bot Implementation Guide and API Sketches

## Purpose

This reference turns the architecture ticket into a buildable bot plan. It assumes the current JS Discord runtime and the named-bot repository pattern already used by the examples in this repository.

The implementation should start from the existing `examples/discord-bots/knowledge-base` bot and evolve it into a bot that can both **capture** knowledge and **serve** it back later with citations and review state.

## Recommended Example Layout

The first version can stay self-contained under `examples/discord-bots/knowledge-base/`:

```text
examples/discord-bots/knowledge-base/
  index.js
  lib/
    store.js
    capture.js
    render.js
    classify.js
  data/
    knowledge.json
    review-queue.json
```

This layout keeps the example easy to understand and keeps the persistence story local to the bot repository. If the store later moves to SQLite or an external service, the public JS commands and event handlers should stay the same.

## Bot Configuration Sketch

A practical starting point is a small `configure()` block with only the fields needed for scoping and storage:

```js
configure({
  name: "knowledge-base",
  description: "Capture, curate, and search shared knowledge from Discord chat",
  category: "knowledge",
  run: {
    fields: {
      store_path: {
        type: "string",
        help: "Path to the persistent knowledge store",
        default: "./data/knowledge.json"
      },
      review_channel_id: {
        type: "string",
        help: "Optional channel ID for review breadcrumbs",
        default: ""
      },
      capture_channel_ids: {
        type: "string",
        help: "Comma-separated list of channel IDs that may be captured",
        default: ""
      },
      read_only: {
        type: "bool",
        help: "Disable write operations for troubleshooting or dry runs",
        default: false
      }
    }
  }
})
```

Keep the initial configuration intentionally small. The more flags the bot gets before the workflow is proven, the harder it becomes to reason about behavior in a live guild.

## Command and Event Contract

### Core commands

| Command | Purpose |
| --- | --- |
| `remember` | Save a useful message or thread snippet as a knowledge draft |
| `teach` | Add knowledge intentionally, with a short summary and tags |
| `ask` | Search and summarize the curated knowledge base |
| `search` | Direct lookup for knowledge entries |
| `article` | Render one canonical knowledge entry |
| `review` | Open the pending review queue |
| `sources` | Show the source messages for a given entry |
| `recent` | Show the newest entries and drafts |

### Core events

| Event | Purpose |
| --- | --- |
| `messageCreate` | Passive capture of candidate knowledge |
| `messageUpdate` | Refresh a draft when the source message is edited |
| `reactionAdd` | Promotion signal, for example a trusted "📌" or "🧠" reaction |
| `ready` | Log bot startup and show capture scope |

### Core component actions

| Action | Purpose |
| --- | --- |
| `knowledge:verify` | Promote a draft to verified |
| `knowledge:reject` | Remove or hide a low-value draft |
| `knowledge:stale` | Mark an entry stale |
| `knowledge:merge` | Merge duplicate entries |
| `knowledge:source` | Open the source message or source thread |
| `knowledge:edit` | Open the edit modal |

## Storage Contract

The MVP store should preserve the following facts:

- the canonical entry text,
- the source message metadata,
- the current status,
- a version history,
- an alias list,
- a list of related entries,
- a confidence score, and
- any curator notes.

A JSON structure is enough for the first pass if the bot runs in a writable working directory.

```json
{
  "entries": [
    {
      "id": "kb_01J...",
      "title": "How we start the JS bot examples",
      "summary": "Use the named-bot runner with the examples repository.",
      "body": "Run `go run ./cmd/discord-bot bots run ping --bot-repository ./examples/discord-bots ...`",
      "tags": ["discord", "javascript", "runbook"],
      "aliases": ["start bot", "run examples"],
      "status": "verified",
      "confidence": 0.93,
      "source": {
        "guildId": "123",
        "channelId": "456",
        "messageId": "789",
        "authorId": "111",
        "jumpUrl": "https://discord.com/channels/...",
        "capturedAt": "2026-04-20T22:00:00Z"
      },
      "versions": [
        {
          "version": 1,
          "editedBy": "111",
          "editedAt": "2026-04-20T22:00:00Z",
          "note": "Initial draft"
        }
      ],
      "relatedIds": []
    }
  ]
}
```

The store API can stay simple at first:

```ts
load(): Promise<KnowledgeStore>
appendDraft(entry: KnowledgeEntry): Promise<void>
listDrafts(): Promise<KnowledgeEntry[]>
get(id: string): Promise<KnowledgeEntry | null>
search(query: string): Promise<KnowledgeEntry[]>
update(id: string, patch: Partial<KnowledgeEntry>): Promise<void>
markVerified(id: string, actorId: string): Promise<void>
markStale(id: string, actorId: string): Promise<void>
merge(primaryId: string, duplicateId: string): Promise<void>
```

## Capture and Review Workflow

### 1. Passive capture

The passive capture path should be conservative. It should prefer precision over recall so the bot does not flood the review queue.

Suggested heuristics:

- short instructions or fixes with explicit verbs like "use", "run", "set", "patch", or "remember",
- messages with code blocks or command snippets,
- messages from trusted users in opted-in channels,
- reaction signals from trusted reviewers,
- thread replies that clarify an earlier answer.

The first version should not hide these decisions. If a message becomes a candidate, the bot should make that visible through a review breadcrumb or an entry card in a designated review channel.

### 2. Explicit capture

A user can capture a message intentionally with `/remember` or `/teach`.

Suggested flow:

1. user runs the command or clicks a button on a message,
2. bot opens a modal with `title`, `summary`, `tags`, and `source` fields,
3. bot saves a draft entry,
4. bot replies with a compact knowledge card and review buttons.

### 3. Review

The review queue is where the bot becomes collaborative. Any trusted reviewer should be able to verify, edit, merge, or reject a candidate.

A verified entry should keep the original source message and add a version row, not overwrite the initial evidence.

### 4. Retrieval

Retrieval should be simple and honest:

- verified entries sort first,
- draft entries appear if they are the only matches,
- stale entries are de-emphasized,
- every answer should show citations.

If the bot is not confident, it should say so and suggest creating a new draft instead of inventing confidence.

## Example JS Skeleton

```js
module.exports = defineBot(({ command, event, component, modal, autocomplete, configure }) => {
  configure({
    name: "knowledge-base",
    description: "Capture and curate shared knowledge from Discord chat",
    category: "knowledge",
  })

  event("messageCreate", async (ctx) => {
    const candidate = await classifyMessage(ctx)
    if (!candidate) return
    await store.appendDraft(candidate)
    await postReviewBreadcrumb(ctx, candidate)
  })

  command("remember", { description: "Save the current message as knowledge" }, async (ctx) => {
    await ctx.showModal(renderEntryModal(ctx))
  })

  command("ask", { description: "Search the knowledge base" }, async (ctx) => {
    const matches = await store.search(ctx.args.query)
    return renderSearchResponse(matches)
  })

  component("knowledge:verify", async (ctx) => {
    await store.markVerified(ctx.values.entryId, ctx.user.id)
    return renderUpdatedCard(await store.get(ctx.values.entryId))
  })
})
```

## Data and Presentation Rules

### Entry presentation

A knowledge card should make status obvious:

- **draft** — useful but not yet reviewed,
- **review** — waiting for curator attention,
- **verified** — canonical until proven stale,
- **stale** — still useful context, but not authoritative,
- **rejected** — not fit for the knowledge base.

### Citation presentation

Every answer should include at least one of:

- a Discord message link,
- a source channel and message ID,
- a related canonical entry,
- a short curated explanation of why the answer is trusted.

### History rules

Do not overwrite the only copy of the original source. If an entry changes, create a new version record and keep the old one available for review.

## Rollout Plan

### Milestone 1 — make recording visible

- Build the store.
- Add passive capture.
- Add a visible review breadcrumb.
- Add a compact card renderer.

### Milestone 2 — make curation easy

- Add `/remember` and `/teach`.
- Add edit and verify buttons.
- Add duplicate detection.
- Add the review queue.

### Milestone 3 — make retrieval useful

- Add `/ask` and `/search`.
- Add citations and answer ranking.
- Add autocomplete for tags and aliases.
- Add a `recent` view.

### Milestone 4 — make maintenance safe

- Add staleness detection.
- Add export support.
- Add tests and smoke checks.
- Document the operational envelope.

## Validation Checklist

- [ ] The bot captures only in opted-in scopes.
- [ ] A candidate entry shows source metadata.
- [ ] A reviewer can verify or reject a candidate from Discord.
- [ ] An answer response shows at least one citation.
- [ ] The bot can run with the named-bot runner.
- [ ] The implementation has tests for the store and the capture path.

## Implementation Risks

- **Too much noise** — solve by keeping heuristics conservative and visible.
- **Too much manual review** — solve with explicit commands and buttons.
- **Duplicate knowledge** — solve with aliasing and merge actions.
- **Unclear ownership** — solve by preserving author and reviewer metadata.
- **Store brittleness** — solve by isolating persistence behind a small helper module.

## References

- `examples/discord-bots/knowledge-base/index.js` — current read-only knowledge example.
- `examples/discord-bots/README.md` — named-bot repository flow.
- `internal/jsdiscord/runtime.go` — JS DSL entry point.
- `internal/jsdiscord/bot.go` — dispatch context and host injection.
- `internal/jsdiscord/host_ops_messages.go` — message lookup and moderation helpers.
- `cmd/discord-bot/root.go` — embedded help system wiring.
- `internal/botcli/command_test.go` — bot runner behavior.
