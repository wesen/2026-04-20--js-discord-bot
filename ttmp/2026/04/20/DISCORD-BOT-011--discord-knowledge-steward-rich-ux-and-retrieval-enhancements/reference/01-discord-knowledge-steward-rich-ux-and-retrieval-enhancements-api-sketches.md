---
Title: Discord Knowledge Steward Rich UX and Retrieval Enhancements API Sketches
Ticket: DISCORD-BOT-011
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
    - Path: examples/discord-bots/knowledge-base/index.js
      Note: Concrete bot wiring for the review and reaction components sketched in the reference
    - Path: examples/discord-bots/knowledge-base/lib/reactions.js
      Note: Trusted emoji and reviewer policy for reaction promotion
    - Path: examples/discord-bots/knowledge-base/lib/review.js
      Note: Queue card
    - Path: internal/jsdiscord/knowledge_base_runtime_test.go
      Note: Runtime test coverage for the review and promotion API sketch
ExternalSources: []
Summary: |
    Practical API sketches for the rich review queue, reaction promotion, rich search cards, citation rendering, and export-to-channel behavior in the knowledge steward bot.
LastUpdated: 2026-04-20T23:15:00-04:00
WhatFor: Provide a concrete starting point for implementing the post-MVP UX and retrieval features.
WhenToUse: Use when wiring the bot UI or validating the command and event contracts for this ticket.
---


# Discord Knowledge Steward Rich UX and Retrieval Enhancements API Sketches

## Purpose

This reference captures the user-facing shapes for the next knowledge steward bot slice. It is intentionally concrete and implementation-oriented: the goal is to make the review queue, reaction promotion, and export flows easy to build and test in the existing JS bot runtime.

## Review Queue Sketch

### Command

```js
command("review", {
  description: "List knowledge entries waiting for review",
  options: {
    status: { type: "string", required: false },
    limit: { type: "integer", required: false },
  },
}, async (ctx) => {
  // return a queue card with select menu + action buttons
})
```

### UI response

- one selected entry card,
- select menu of queued entries,
- buttons for Verify / Edit / Source / Stale / Reject,
- ephemeral acknowledgements for select and button actions.

### State stored per reviewer

```js
{
  status: "draft" | "review" | "verified" | "stale" | "rejected",
  limit: 5,
  selectedId: "kb_..."
}
```

## Review Component Sketch

```js
component("knowledge:review:select", async (ctx) => {
  // ctx.values[0] is the selected entry id
})

component("knowledge:review:verify", async (ctx) => {
  // mark current entry verified
})

component("knowledge:review:edit", async (ctx) => {
  // open the edit modal for the current entry
})

component("knowledge:review:source", async (ctx) => {
  // show source attribution for the current entry
})

component("knowledge:review:stale", async (ctx) => {
  // mark current entry stale
})

component("knowledge:review:reject", async (ctx) => {
  // reject current entry
})
```

### Edit modal

```js
modal("knowledge:review:edit", async (ctx) => {
  // ctx.values.title
  // ctx.values.summary
  // ctx.values.body
  // ctx.values.tags
  // ctx.values.aliases
  // ctx.values.source
})
```

## Reaction Promotion Sketch

### Event

```js
event("reactionAdd", async (ctx) => {
  // trust the reaction if the emoji and user/role policy matches
  // map message -> knowledge entry via source metadata
  // promote status draft -> review, review -> verified
})
```

### Recommended config

```js
configure({
  run: {
    fields: {
      reactionPromoteEmojis: {
        type: "string",
        help: "Comma-separated emojis that promote a candidate",
        default: "🧠,📌",
      },
      trustedReviewerIds: {
        type: "string",
        help: "Optional comma-separated trusted user IDs",
        default: "",
      },
      trustedReviewerRoleIds: {
        type: "string",
        help: "Optional comma-separated trusted role IDs",
        default: "",
      },
    }
  }
})
```

### Promotion response

```js
{
  content: "Promoted <title> from draft to review via 🧠.",
  embeds: [knowledgeEmbed(updatedEntry)]
}
```

## Citation Rendering Sketch

The selected entry card should expose:

- Entry ID,
- title,
- status,
- confidence,
- tags,
- aliases,
- source metadata,
- version number.

For source details, the bot should render:

- guild id,
- channel id,
- message id,
- jump URL,
- source note.

## Rich Search Sketch

### Command

```js
command("ask", {
  description: "Search the shared knowledge base",
  options: {
    query: { type: "string", required: true },
  },
}, async (ctx) => {
  // return richer cards with citations and actions
})
```

### Suggested card shape

- title: query or answer title,
- description: concise answer summary,
- fields: status, confidence, source, related,
- buttons: export to channel, open source, open article.

## Export-to-Channel Sketch

```js
component("knowledge:ask:export", async (ctx) => {
  // post the selected answer to ctx.channel.id
})
```

Suggested behavior:

1. keep the ask response ephemeral while the user decides,
2. post a channel-visible answer when export is confirmed,
3. preserve citations in the posted message,
4. keep the answer identifiable as bot-curated knowledge.

## Minimal Store Operations Needed

The current store already supports most of the required verbs. The next layer should continue to rely on:

- `listByStatus(...)`
- `getEntry(...)`
- `setStatus(...)`
- `updateEntry(...)`
- `search(...)`
- `recent(...)`

If a future feature requires more, prefer extending the store in JS before adding new Go host APIs.

## Validation Checklist

- [ ] Review queue returns select + button UI.
- [ ] Selected entry changes when the select menu changes.
- [ ] Edit modal pre-fills the current entry.
- [ ] Reaction promotion updates entry status.
- [ ] Rich search cards include citations.
- [ ] Export action posts a channel-visible answer.

## References

- `examples/discord-bots/knowledge-base/index.js`
- `examples/discord-bots/knowledge-base/lib/review.js`
- `examples/discord-bots/knowledge-base/lib/reactions.js`
- `examples/discord-bots/knowledge-base/lib/render.js`
- `examples/discord-bots/knowledge-base/lib/store.js`
- `internal/jsdiscord/knowledge_base_runtime_test.go`
