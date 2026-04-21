function registerKnowledgeBot({ command, event, modal }, store, capture, render) {
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
      },
    },
  }, async (ctx) => {
    const results = store.search(ctx.config, (ctx.args || {}).query, 5)
    return render.searchResults((ctx.args || {}).query, results)
  })

  command("kb-search", {
    description: "Search the shared knowledge base",
    options: {
      query: {
        type: "string",
        description: "Search query",
        required: true,
      },
    },
  }, async (ctx) => {
    const results = store.search(ctx.config, (ctx.args || {}).query, 5)
    return render.searchResults((ctx.args || {}).query, results)
  })

  command("article", {
    description: "Fetch one knowledge entry",
    options: {
      name: {
        type: "string",
        description: "Entry id or slug",
        required: true,
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
    const status = normalizeStatusFilter((ctx.args || {}).status)
    const limit = Number((ctx.args || {}).limit || (ctx.config || {}).reviewLimit || 5)
    const entries = store.listByStatus(ctx.config, status, limit)
    return render.reviewQueue(entries, status)
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
    const status = normalizeStatusFilter((ctx.args || {}).status)
    const limit = Number((ctx.args || {}).limit || (ctx.config || {}).reviewLimit || 5)
    const entries = store.listByStatus(ctx.config, status, limit)
    return render.reviewQueue(entries, status)
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
}

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

function normalizeStatusFilter(value) {
  const status = String(value || "draft").trim().toLowerCase()
  switch (status) {
    case "draft":
    case "review":
    case "verified":
    case "stale":
    case "rejected":
      return status
    default:
      return "draft"
  }
}

function actorId(ctx) {
  return String((ctx.user && ctx.user.id) || (ctx.me && ctx.me.id) || "").trim()
}

module.exports = registerKnowledgeBot
