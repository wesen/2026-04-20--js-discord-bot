const { defineBot } = require("discord")
const { createKnowledgeStore } = require("./lib/store")
const capture = require("./lib/capture")
const render = require("./lib/render")
const review = require("./lib/review")
const search = require("./lib/search")
const reactions = require("./lib/reactions")

const store = createKnowledgeStore()

module.exports = defineBot(({ command, event, component, modal, autocomplete, configure }) => {
  configure({
    name: "knowledge-base",
    description: "Listen to Discord chat, record candidate knowledge, and curate it as a shared memory",
    category: "knowledge",
    run: {
      fields: {
        dbPath: {
          type: "string",
          help: "SQLite path for the knowledge store",
          default: "./examples/discord-bots/knowledge-base/data/knowledge.sqlite",
        },
        captureEnabled: {
          type: "bool",
          help: "Enable passive capture from messageCreate events",
          default: true,
        },
        captureThreshold: {
          type: "number",
          help: "Minimum confidence required to save a passive capture",
          default: 0.65,
        },
        captureChannels: {
          type: "string",
          help: "Optional comma-separated channel IDs to allow for passive capture",
          default: "",
        },
        reviewLimit: {
          type: "integer",
          help: "Number of entries to show in review lists",
          default: 5,
        },
        seedEntries: {
          type: "bool",
          help: "Seed onboarding entries the first time the SQLite store is created",
          default: true,
        },
        reactionPromoteEmojis: {
          type: "string",
          help: "Comma-separated emojis that promote a captured message into the review queue",
          default: "🧠,📌",
        },
        trustedReviewerIds: {
          type: "string",
          help: "Optional comma-separated user IDs allowed to promote candidates with reactions",
          default: "",
        },
        trustedReviewerRoleIds: {
          type: "string",
          help: "Optional comma-separated role IDs allowed to promote candidates with reactions",
          default: "",
        },
      },
    },
  })

  event("ready", async (ctx) => {
    store.ensure(ctx.config)
    const counts = store.listCounts(ctx.config)
    ctx.log.info("knowledge-base bot ready", {
      user: ctx.me && ctx.me.username,
      dbPath: store.configPath(),
      counts,
    })
  })

  event("messageCreate", async (ctx) => {
    const candidate = capture.captureFromMessage(ctx, ctx.config)
    if (!candidate) {
      return
    }
    const saved = store.saveCandidate(ctx.config, candidate)
    if (!saved) {
      return
    }
    await ctx.reply(render.knowledgeAnnouncement(saved, "Captured"))
  })

  command("remember", {
    description: "Open the teach modal to add knowledge",
  }, async (ctx) => {
    await ctx.showModal(buildTeachModal())
  })

  command("teach", {
    description: "Open the teach modal to add knowledge",
  }, async (ctx) => {
    await ctx.showModal(buildTeachModal())
  })

  modal("knowledge:submit", async (ctx) => {
    const entry = capture.captureFromModal(ctx.values, ctx)
    if (!entry) {
      return {
        content: "Please provide at least a title and body before submitting knowledge.",
        ephemeral: true,
      }
    }
    const saved = store.saveManualEntry(ctx.config, entry)
    return render.knowledgeAnnouncement(saved, "Saved")
  })

  command("ask", {
    description: "Search the shared knowledge base",
    options: {
      query: {
        type: "string",
        description: "Search query",
        required: true,
        autocomplete: true,
      },
    },
  }, async (ctx) => {
    store.ensure(ctx.config)
    const query = String((ctx.args || {}).query || "").trim()
    const limit = Number((ctx.config || {}).reviewLimit || 5)
    const results = store.search(ctx.config, query, limit)
    search.stateFromSearchCommand(ctx, query, limit, results)
    return search.buildSearchMessage(search.searchView(ctx, store))
  })

  command("kb-search", {
    description: "Search the shared knowledge base",
    options: {
      query: {
        type: "string",
        description: "Search query",
        required: true,
        autocomplete: true,
      },
    },
  }, async (ctx) => {
    store.ensure(ctx.config)
    const query = String((ctx.args || {}).query || "").trim()
    const limit = Number((ctx.config || {}).reviewLimit || 5)
    const results = store.search(ctx.config, query, limit)
    search.stateFromSearchCommand(ctx, query, limit, results)
    return search.buildSearchMessage(search.searchView(ctx, store))
  })

  command("article", {
    description: "Fetch one knowledge entry",
    options: {
      name: {
        type: "string",
        description: "Entry id or slug",
        required: true,
        autocomplete: true,
      },
    },
  }, async (ctx) => {
    store.ensure(ctx.config)
    const entry = store.getEntry((ctx.args || {}).name)
    if (!entry) {
      return {
        content: `No knowledge entry found for ${(ctx.args || {}).name}.`,
        ephemeral: true,
      }
    }
    return render.knowledgeAnnouncement(entry, "Opened")
  })

  command("kb-article", {
    description: "Fetch one knowledge entry",
    options: {
      name: {
        type: "string",
        description: "Entry id or slug",
        required: true,
        autocomplete: true,
      },
    },
  }, async (ctx) => {
    store.ensure(ctx.config)
    const entry = store.getEntry((ctx.args || {}).name)
    if (!entry) {
      return {
        content: `No knowledge entry found for ${(ctx.args || {}).name}.`,
        ephemeral: true,
      }
    }
    return render.knowledgeAnnouncement(entry, "Opened")
  })

  command("review", {
    description: "List knowledge entries waiting for review",
    options: {
      status: {
        type: "string",
        description: "Entry status to review",
        required: false,
      },
      limit: {
        type: "integer",
        description: "Maximum number of entries to show",
        required: false,
      },
    },
  }, async (ctx) => {
    store.ensure(ctx.config)
    const status = review.normalizeStatus((ctx.args || {}).status)
    const limit = Number((ctx.args || {}).limit || (ctx.config || {}).reviewLimit || 5)
    const entries = store.listByStatus(ctx.config, status, limit)
    review.stateFromQueueCommand(ctx, status, limit, entries)
    return review.buildQueueMessage(entries, { status, limit, selectedId: entries[0] && entries[0].id })
  })

  command("kb-review", {
    description: "List knowledge entries waiting for review",
    options: {
      status: {
        type: "string",
        description: "Entry status to review",
        required: false,
      },
      limit: {
        type: "integer",
        description: "Maximum number of entries to show",
        required: false,
      },
    },
  }, async (ctx) => {
    store.ensure(ctx.config)
    const status = review.normalizeStatus((ctx.args || {}).status)
    const limit = Number((ctx.args || {}).limit || (ctx.config || {}).reviewLimit || 5)
    const entries = store.listByStatus(ctx.config, status, limit)
    review.stateFromQueueCommand(ctx, status, limit, entries)
    return review.buildQueueMessage(entries, { status, limit, selectedId: entries[0] && entries[0].id })
  })

  command("recent", {
    description: "Show the newest knowledge entries",
    options: {
      limit: {
        type: "integer",
        description: "Maximum number of entries to show",
        required: false,
      },
    },
  }, async (ctx) => {
    const entries = store.recent(ctx.config, Number((ctx.args || {}).limit || (ctx.config || {}).reviewLimit || 5))
    return render.recentResults("Recent knowledge entries", entries)
  })

  command("kb-recent", {
    description: "Show the newest knowledge entries",
    options: {
      limit: {
        type: "integer",
        description: "Maximum number of entries to show",
        required: false,
      },
    },
  }, async (ctx) => {
    const entries = store.recent(ctx.config, Number((ctx.args || {}).limit || (ctx.config || {}).reviewLimit || 5))
    return render.recentResults("Recent knowledge entries", entries)
  })

  command("kb-verify", {
    description: "Mark a knowledge entry as verified",
    options: {
      entry: {
        type: "string",
        description: "Entry id or slug",
        required: true,
      },
      note: {
        type: "string",
        description: "Review note",
        required: false,
      },
    },
  }, async (ctx) => {
    store.ensure(ctx.config)
    const entry = store.setStatus(ctx.config, (ctx.args || {}).entry, "verified", actorId(ctx), (ctx.args || {}).note || "Verified from Discord")
    if (!entry) {
      return { content: `No knowledge entry found for ${(ctx.args || {}).entry}.`, ephemeral: true }
    }
    return render.knowledgeAnnouncement(entry, "Verified")
  })

  command("kb-stale", {
    description: "Mark a knowledge entry as stale",
    options: {
      entry: {
        type: "string",
        description: "Entry id or slug",
        required: true,
      },
      note: {
        type: "string",
        description: "Review note",
        required: false,
      },
    },
  }, async (ctx) => {
    store.ensure(ctx.config)
    const entry = store.setStatus(ctx.config, (ctx.args || {}).entry, "stale", actorId(ctx), (ctx.args || {}).note || "Marked stale from Discord")
    if (!entry) {
      return { content: `No knowledge entry found for ${(ctx.args || {}).entry}.`, ephemeral: true }
    }
    return render.knowledgeAnnouncement(entry, "Marked stale")
  })

  command("kb-reject", {
    description: "Reject a knowledge entry",
    options: {
      entry: {
        type: "string",
        description: "Entry id or slug",
        required: true,
      },
      note: {
        type: "string",
        description: "Review note",
        required: false,
      },
    },
  }, async (ctx) => {
    store.ensure(ctx.config)
    const entry = store.setStatus(ctx.config, (ctx.args || {}).entry, "rejected", actorId(ctx), (ctx.args || {}).note || "Rejected from Discord")
    if (!entry) {
      return { content: `No knowledge entry found for ${(ctx.args || {}).entry}.`, ephemeral: true }
    }
    return render.knowledgeAnnouncement(entry, "Rejected")
  })

  component(review.REVIEW_COMPONENTS.select, async (ctx) => {
    store.ensure(ctx.config)
    const selectedId = firstValue(ctx.values)
    if (!selectedId) {
      return { content: "Please choose a knowledge entry from the review dropdown.", ephemeral: true }
    }
    review.setReviewSelection(ctx, selectedId)
    const current = review.currentReviewEntry(ctx, store)
    if (!current) {
      return { content: "No review entry is currently available.", ephemeral: true }
    }
    return review.reviewReply(current, "Selected")
  })

  component(review.REVIEW_COMPONENTS.verify, async (ctx) => {
    store.ensure(ctx.config)
    const entry = review.currentReviewEntry(ctx, store)
    if (!entry) {
      return { content: "No review entry is currently selected.", ephemeral: true }
    }
    const updated = store.setStatus(ctx.config, entry.id, "verified", actorId(ctx), "Verified from review UI")
    return review.reviewReply(updated, "Verified")
  })

  component(review.REVIEW_COMPONENTS.stale, async (ctx) => {
    store.ensure(ctx.config)
    const entry = review.currentReviewEntry(ctx, store)
    if (!entry) {
      return { content: "No review entry is currently selected.", ephemeral: true }
    }
    const updated = store.setStatus(ctx.config, entry.id, "stale", actorId(ctx), "Marked stale from review UI")
    return review.reviewReply(updated, "Marked stale")
  })

  component(review.REVIEW_COMPONENTS.reject, async (ctx) => {
    store.ensure(ctx.config)
    const entry = review.currentReviewEntry(ctx, store)
    if (!entry) {
      return { content: "No review entry is currently selected.", ephemeral: true }
    }
    const updated = store.setStatus(ctx.config, entry.id, "rejected", actorId(ctx), "Rejected from review UI")
    return review.reviewReply(updated, "Rejected")
  })

  component(review.REVIEW_COMPONENTS.source, async (ctx) => {
    store.ensure(ctx.config)
    const entry = review.currentReviewEntry(ctx, store)
    return review.reviewSourceReply(entry)
  })

  component(review.REVIEW_COMPONENTS.edit, async (ctx) => {
    store.ensure(ctx.config)
    const entry = review.currentReviewEntry(ctx, store)
    if (!entry) {
      return { content: "No review entry is currently selected.", ephemeral: true }
    }
    await ctx.showModal(review.buildEntryModal(entry))
  })

  modal(review.REVIEW_COMPONENTS.edit, async (ctx) => {
    store.ensure(ctx.config)
    const entry = review.currentReviewEntry(ctx, store)
    if (!entry) {
      return { content: "No review entry is currently selected.", ephemeral: true }
    }
    const patch = review.entryToEditPatch(ctx.values)
    const updated = store.updateEntry(ctx.config, entry.id, patch)
    if (!updated) {
      return { content: `Could not update knowledge entry ${entry.id}.`, ephemeral: true }
    }
    review.setReviewSelection(ctx, updated.id)
    return render.knowledgeAnnouncement(updated, "Updated")
  })

  component(search.SEARCH_COMPONENTS.select, async (ctx) => {
    store.ensure(ctx.config)
    const selectedId = firstValue(ctx.values)
    if (!selectedId) {
      return { content: "Please choose a knowledge entry from the search dropdown.", ephemeral: true }
    }
    search.setSearchSelection(ctx, selectedId)
    const view = search.searchView(ctx, store)
    if (!view.selectedEntry) {
      return { content: "No search result is currently available.", ephemeral: true }
    }
    return search.buildSearchMessage(view)
  })

  component(search.SEARCH_COMPONENTS.previous, async (ctx) => {
    store.ensure(ctx.config)
    search.shiftSearchPage(ctx, -1)
    const view = search.searchView(ctx, store)
    if (!view.selectedEntry) {
      return { content: "No search result is currently available.", ephemeral: true }
    }
    return search.buildSearchMessage(view)
  })

  component(search.SEARCH_COMPONENTS.next, async (ctx) => {
    store.ensure(ctx.config)
    search.shiftSearchPage(ctx, 1)
    const view = search.searchView(ctx, store)
    if (!view.selectedEntry) {
      return { content: "No search result is currently available.", ephemeral: true }
    }
    return search.buildSearchMessage(view)
  })

  component(search.SEARCH_COMPONENTS.open, async (ctx) => {
    store.ensure(ctx.config)
    const entry = search.currentSearchEntry(ctx, store)
    if (!entry) {
      return { content: "No search result is currently selected.", ephemeral: true }
    }
    return render.knowledgeAnnouncement(entry, "Opened")
  })

  component(search.SEARCH_COMPONENTS.source, async (ctx) => {
    store.ensure(ctx.config)
    const entry = search.currentSearchEntry(ctx, store)
    return search.searchSourceReply(entry)
  })

  component(search.SEARCH_COMPONENTS.export, async (ctx) => {
    store.ensure(ctx.config)
    const view = search.searchView(ctx, store)
    const entry = view.selectedEntry
    if (!entry) {
      return { content: "No search result is currently selected.", ephemeral: true }
    }
    await ctx.defer({ ephemeral: true })
    await ctx.followUp(search.searchExportPayload(entry, view.query))
    await ctx.edit({
      content: `Exported **${entry.title || "knowledge entry"}** into the channel from ${view.query ? `search for ${view.query}` : "search results"}.`,
      embeds: [search.renderSearchResultCard(entry, { query: view.query, total: view.allResults.length, position: view.selectedIndex, page: view.page, pageCount: view.pageCount, relatedEntries: view.relatedEntries })],
    })
  })

  autocomplete("ask", "query", async (ctx) => {
    store.ensure(ctx.config)
    return search.searchAutocompleteSuggestions(ctx, store)
  })

  autocomplete("kb-search", "query", async (ctx) => {
    store.ensure(ctx.config)
    return search.searchAutocompleteSuggestions(ctx, store)
  })

  autocomplete("article", "name", async (ctx) => {
    store.ensure(ctx.config)
    return search.articleAutocompleteSuggestions(ctx, store)
  })

  autocomplete("kb-article", "name", async (ctx) => {
    store.ensure(ctx.config)
    return search.articleAutocompleteSuggestions(ctx, store)
  })

  reactions.registerReactionPromotions({ event }, store, render)
})

function buildTeachModal() {
  return {
    customId: "knowledge:submit",
    title: "Teach the knowledge bot",
    components: [
      {
        type: "actionRow",
        components: [
          {
            type: "textInput",
            customId: "title",
            label: "Title",
            style: "short",
            required: true,
            minLength: 3,
            maxLength: 100,
          },
        ],
      },
      {
        type: "actionRow",
        components: [
          {
            type: "textInput",
            customId: "summary",
            label: "Summary",
            style: "paragraph",
            required: true,
            minLength: 10,
            maxLength: 300,
          },
        ],
      },
      {
        type: "actionRow",
        components: [
          {
            type: "textInput",
            customId: "body",
            label: "Body",
            style: "paragraph",
            required: true,
            minLength: 20,
            maxLength: 2000,
          },
        ],
      },
      {
        type: "actionRow",
        components: [
          {
            type: "textInput",
            customId: "tags",
            label: "Tags (comma-separated)",
            style: "short",
            required: false,
            maxLength: 200,
          },
        ],
      },
      {
        type: "actionRow",
        components: [
          {
            type: "textInput",
            customId: "source",
            label: "Source URL or note",
            style: "short",
            required: false,
            maxLength: 300,
          },
        ],
      },
    ],
  }
}

function actorId(ctx) {
  return String((ctx.user && ctx.user.id) || (ctx.me && ctx.me.id) || "").trim()
}

function firstValue(values) {
  if (Array.isArray(values) && values.length > 0) {
    return String(values[0] || "").trim()
  }
  if (values && typeof values === "object") {
    const keys = Object.keys(values)
    if (keys.length > 0) {
      return String(values[keys[0]] || "").trim()
    }
  }
  return ""
}
