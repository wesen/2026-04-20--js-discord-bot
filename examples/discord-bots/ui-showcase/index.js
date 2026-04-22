// UI Showcase Bot — demonstrates the UI DSL primitives, forms, stateful
// screens, pagers, card galleries, confirmations, and alias registration.
//
// Commands:
//   /demo-message    — builder patterns: message, embed, row
//   /demo-form       — modal form DSL
//   /demo-search     — stateful search screen with pager
//   /demo-review     — review queue with select + action buttons
//   /demo-confirm    — inline confirmation dialog
//   /demo-pager      — paginated list with previous/next
//   /demo-cards      — card gallery with select navigation
//   /demo-selects    — all select menu types (string, user, role, channel, mentionable)
//   /demo-alias-*    — alias registration demo (two names, same handler)
//   /find            — alias of demo-search
//   /browse          — alias of demo-cards

const { defineBot } = require("discord")
const { sleep } = require("timer")
const ui = require("./lib/ui")
const store = require("./lib/demo-store")

// ── Flow definitions ─────────────────────────────────────────────────────────

const searchFlow = ui.flow("showcase.search", {
  init: { query: "", page: 1, selectedId: "", limit: 3 },
})

const reviewFlow = ui.flow("showcase.review", {
  init: { status: "draft", selectedId: "" },
})

const pagerFlow = ui.flow("showcase.pager", {
  init: { page: 1, pageSize: 3 },
})

const cardFlow = ui.flow("showcase.cards", {
  init: { selectedId: store.PRODUCTS[0] && store.PRODUCTS[0].id || "" },
})

// ── Bot definition ───────────────────────────────────────────────────────────

module.exports = defineBot(({ command, event, component, modal, autocomplete, configure }) => {
  configure({
    name: "ui-showcase",
    description: "Demonstrates the UI DSL: builders, forms, stateful screens, pagers, cards, confirmations, selects, and aliases",
    category: "examples",
  })

  event("ready", async (ctx) => {
    ctx.log.info("ui-showcase bot ready", {
      user: ctx.me && ctx.me.username,
    })
  })

  event("messageCreate", async (ctx) => {
    const content = String((ctx.message && ctx.message.content) || "").trim()
    if (content === "!showcase") {
      await ctx.reply({ content: "UI Showcase bot is online. Try /demo-message, /demo-form, /demo-search, /demo-review, /demo-confirm, /demo-pager, /demo-cards, /demo-selects, /find, or /browse." })
    }
  })

  // ══════════════════════════════════════════════════════════════════════════
  //  /demo-message — builder patterns
  // ══════════════════════════════════════════════════════════════════════════

  command("demo-message", {
    description: "Showcase message, embed, and button builders",
  }, async () => {
    return ui.message()
      .content("This message was built entirely with the UI DSL builder pattern.")
      .ephemeral()
      .embed(
        ui.embed("Builder Showcase")
          .color(0x5865F2)
          .description("Every part of this message — content, embed, buttons, and select — was composed using `ui.message().embed().row()...`")
          .field("Builders used", "message, embed, button, select", true)
          .field("Pattern", "Fluent builder chain", true)
          .footer("ui-showcase bot · /demo-message")
      )
      .row(
        ui.button("showcase:msg:primary", "Primary", "primary"),
        ui.button("showcase:msg:success", "Success", "success"),
        ui.button("showcase:msg:danger", "Danger", "danger"),
      )
      .row(
        ui.select("showcase:msg:select")
          .placeholder("Choose a color theme")
          .option("Discord Blurple", "blurple", "The classic Discord blue")
          .option("Success Green", "green", "Positive actions")
          .option("Warning Yellow", "yellow", "Caution advised")
          .option("Error Red", "red", "Danger zone")
      )
      .build()
  })

  component("showcase:msg:primary", async () => ui.ok("You clicked **Primary** (blurple style)."))
  component("showcase:msg:success", async () => ui.ok("You clicked **Success** (green style)."))
  component("showcase:msg:danger", async () => ui.ok("You clicked **Danger** (red style)."))

  component("showcase:msg:select", async (ctx) => {
    const value = firstValue(ctx.values) || "none"
    const colorMap = { blurple: 0x5865F2, green: 0x57F287, yellow: 0xFEE75C, red: 0xED4245 }
    const color = colorMap[value] || 0x95A5A6
    return ui.message()
      .ephemeral()
      .content(`You selected **${value}**.`)
      .embed(ui.embed("Color selected").color(color).description(`The embed color now matches the ${value} theme.`))
      .build()
  })

  // ══════════════════════════════════════════════════════════════════════════
  //  /demo-form — modal form DSL
  // ══════════════════════════════════════════════════════════════════════════

  command("demo-form", {
    description: "Showcase the modal/form builder DSL",
  }, async (ctx) => {
    await ctx.showModal(
      ui.form("showcase:form:submit", "Feedback Form")
        .text("title", "Title").required().min(3).max(100)
        .textarea("feedback", "Your feedback").required().min(10).max(500)
        .text("rating", "Rating (1–5)")
        .text("tags", "Tags (comma-separated)")
        .build()
    )
  })

  modal("showcase:form:submit", async (ctx) => {
    const title = String((ctx.values || {}).title || "").trim()
    const feedback = String((ctx.values || {}).feedback || "").trim()
    const rating = String((ctx.values || {}).rating || "unrated").trim()
    const tags = String((ctx.values || {}).tags || "").trim()

    return ui.message()
      .ephemeral()
      .content("Thanks for your feedback!")
      .embed(
        ui.embed(title || "Feedback")
          .color(0x57F287)
          .description(feedback || "(no content)")
          .field("Rating", rating, true)
          .field("Tags", tags || "(none)", true)
          .footer("Built with ui.form()")
      )
      .build()
  })

  // ══════════════════════════════════════════════════════════════════════════
  //  /demo-search — stateful search screen with pager
  // ══════════════════════════════════════════════════════════════════════════

  command("demo-search", {
    description: "Search articles with a stateful screen, pager, and select",
    options: {
      query: {
        type: "string",
        description: "Search query",
        required: true,
        autocomplete: true,
      },
    },
  }, async (ctx) => {
    return runSearch(ctx)
  })

  // Alias: /find does the same thing as /demo-search
  command("find", {
    description: "Alias for /demo-search — search articles",
    options: {
      query: {
        type: "string",
        description: "Search query",
        required: true,
        autocomplete: true,
      },
    },
  }, async (ctx) => {
    return runSearch(ctx)
  })

  function runSearch(ctx) {
    const query = String((ctx.args || {}).query || "").trim()
    const limit = 3
    const results = store.searchArticles(query, 50)
    if (results.length === 0) {
      return ui.emptyResults(query)
    }
    const state = { query, page: 1, selectedId: results[0].id, limit }
    searchFlow.save(ctx, state)
    return renderSearchScreen(ctx, results, state)
  }

  function renderSearchScreen(ctx, allResults, state) {
    const page = Number(state.page || 1)
    const limit = Number(state.limit || 3)
    const pageCount = Math.max(1, Math.ceil(allResults.length / limit))
    const currentPage = Math.min(page, pageCount)
    const start = (currentPage - 1) * limit
    const pageEntries = allResults.slice(start, start + limit)

    const selectedId = String(state.selectedId || "")
    const selected = pageEntries.find((a) => a.id === selectedId) || pageEntries[0]

    if (!selected) {
      return ui.emptyResults(state.query)
    }

    // Sync selectedId if it changed
    if (selected.id !== selectedId) {
      searchFlow.save(ctx, { ...state, selectedId: selected.id, page: currentPage })
    }

    const hasPrevious = currentPage > 1
    const hasNext = currentPage < pageCount

    return ui.message()
      .ephemeral()
      .content(`Found **${allResults.length}** article${allResults.length === 1 ? "" : "s"} for **${state.query}**.`)
      .embed(
        ui.card(selected.title)
          .description(selected.summary)
          .color(store.statusColor(selected.status))
          .meta("Status", selected.status, true)
          .meta("Confidence", `${Math.round(selected.confidence * 100)}%`, true)
          .meta("Category", selected.category)
          .meta("Tags", selected.tags.map((t) => `#${t}`).join(" "))
          .footer(`Page ${currentPage}/${pageCount} · Result ${start + pageEntries.indexOf(selected) + 1}/${allResults.length}`)
      )
      .row(
        ui.select(searchFlow.id("select"))
          .placeholder("Choose an article")
          .optionEntries(pageEntries.map((a) => ({
            id: a.id,
            label: a.title,
            description: `${a.status} · ${a.category}`,
          })), selected.id)
      )
      .row(ui.pager(searchFlow.id("previous"), searchFlow.id("next"), { hasPrevious, hasNext }))
      .row(
        ui.button(searchFlow.id("open"), "Open", "primary"),
        ui.button(searchFlow.id("source"), "Details", "secondary"),
        ui.button(searchFlow.id("export"), "Export", "success"),
      )
      .build()
  }

  component(searchFlow.id("select"), async (ctx) => {
    const selectedId = firstValue(ctx.values)
    if (!selectedId) return ui.error("No article selected.")
    const state = searchFlow.load(ctx)
    searchFlow.save(ctx, { ...state, selectedId })
    const results = store.searchArticles(state.query, 50)
    return renderSearchScreen(ctx, results, { ...state, selectedId })
  })

  component(searchFlow.id("previous"), async (ctx) => {
    const state = searchFlow.load(ctx)
    const newPage = Math.max(1, Number(state.page || 1) - 1)
    searchFlow.save(ctx, { ...state, page: newPage, selectedId: "" })
    const results = store.searchArticles(state.query, 50)
    return renderSearchScreen(ctx, results, { ...state, page: newPage, selectedId: "" })
  })

  component(searchFlow.id("next"), async (ctx) => {
    const state = searchFlow.load(ctx)
    const newPage = Number(state.page || 1) + 1
    searchFlow.save(ctx, { ...state, page: newPage, selectedId: "" })
    const results = store.searchArticles(state.query, 50)
    return renderSearchScreen(ctx, results, { ...state, page: newPage, selectedId: "" })
  })

  component(searchFlow.id("open"), async (ctx) => {
    const state = searchFlow.load(ctx)
    const article = store.getArticle(state.selectedId)
    if (!article) return ui.error("No article selected.")
    return ui.message()
      .ephemeral()
      .content(`Opened **${article.title}**`)
      .embed(
        ui.embed(article.title)
          .color(store.statusColor(article.status))
          .description(article.summary)
          .field("ID", article.id, true)
          .field("Status", article.status, true)
          .field("Category", article.category, true)
          .field("Confidence", `${Math.round(article.confidence * 100)}%`, true)
          .field("Author", article.author, true)
          .field("Tags", article.tags.join(", "))
      )
      .build()
  })

  component(searchFlow.id("source"), async (ctx) => {
    const state = searchFlow.load(ctx)
    const article = store.getArticle(state.selectedId)
    if (!article) return ui.error("No article selected.")
    return ui.message()
      .ephemeral()
      .content(`Details for **${article.title}**`)
      .embed(
        ui.embed("Article details")
          .color(0x5865F2)
          .description(`Internal metadata for article **${article.title}**.`)
          .field("ID", article.id)
          .field("Category", article.category, true)
          .field("Status", article.status, true)
          .field("Confidence", `${Math.round(article.confidence * 100)}%`, true)
          .field("Author", article.author, true)
      )
      .build()
  })

  component(searchFlow.id("export"), async (ctx) => {
    const state = searchFlow.load(ctx)
    const article = store.getArticle(state.selectedId)
    if (!article) return ui.error("No article selected.")
    await ctx.defer({ ephemeral: true })
    const channelId = ctx.channel && ctx.channel.id
    if (channelId) {
      await ctx.discord.channels.send(channelId, {
        content: `📚 **${article.title}** (${article.status})\n${article.summary}\nTags: ${article.tags.join(", ")}`,
      })
    }
    await ctx.edit({
      content: `Exported **${article.title}** to the channel.`,
      embeds: [ui.embed("Exported").description(article.title).color(0x57F287).build()],
    })
  })

  autocomplete("demo-search", "query", async (ctx) => {
    return store.articleSuggestions(ctx.focused && ctx.focused.value)
  })

  autocomplete("find", "query", async (ctx) => {
    return store.articleSuggestions(ctx.focused && ctx.focused.value)
  })

  // ══════════════════════════════════════════════════════════════════════════
  //  /demo-review — review queue with select + action buttons
  // ══════════════════════════════════════════════════════════════════════════

  command("demo-review", {
    description: "Review queue for articles — select, verify, reject",
    options: {
      status: {
        type: "string",
        description: "Filter by status (draft, review, verified, stale)",
        required: false,
      },
    },
  }, async (ctx) => {
    const status = String((ctx.args || {}).status || "draft").trim().toLowerCase()
    const entries = store.ARTICLES.filter((a) => a.status === status)
    const selectedId = entries[0] ? entries[0].id : ""
    reviewFlow.save(ctx, { status, selectedId })
    return renderReviewScreen(ctx, entries, { status, selectedId })
  })

  function renderReviewScreen(ctx, entries, state) {
    const selected = entries.find((a) => a.id === state.selectedId) || entries[0]

    if (entries.length === 0) {
      return ui.message()
        .ephemeral()
        .content(`No **${state.status}** articles found.`)
        .build()
    }

    const selectedId = selected ? selected.id : ""

    return ui.message()
      .ephemeral()
      .content(`Review queue: **${state.status}** (${entries.length} article${entries.length === 1 ? "" : "s"})`)
      .embed(
        ui.card(selected ? selected.title : "No selection")
          .description(selected ? selected.summary : "Select an article from the dropdown.")
          .color(selected ? store.statusColor(selected.status) : 0x95A5A6)
          .meta("Status", selected ? selected.status : "—", true)
          .meta("Category", selected ? selected.category : "—", true)
          .meta("Confidence", selected ? `${Math.round(selected.confidence * 100)}%` : "—", true)
          .meta("Tags", selected ? selected.tags.join(", ") : "—")
      )
      .row(
        ui.select(reviewFlow.id("select"))
          .placeholder("Choose an article to review")
          .optionEntries(entries.map((a) => ({
            id: a.id,
            label: a.title,
            description: `${a.status} · ${a.category}`,
          })), selectedId)
      )
      .row(
        ui.button(reviewFlow.id("verify"), "✓ Verify", "success"),
        ui.button(reviewFlow.id("stale"), "⊘ Stale", "secondary"),
        ui.button(reviewFlow.id("reject"), "✗ Reject", "danger"),
        ui.button(reviewFlow.id("edit"), "✎ Edit", "primary"),
        ui.button(reviewFlow.id("details"), "Details", "secondary"),
      )
      .build()
  }

  component(reviewFlow.id("select"), async (ctx) => {
    const selectedId = firstValue(ctx.values)
    if (!selectedId) return ui.error("No article selected.")
    const state = reviewFlow.load(ctx)
    reviewFlow.save(ctx, { ...state, selectedId })
    const entries = store.ARTICLES.filter((a) => a.status === state.status)
    return renderReviewScreen(ctx, entries, { ...state, selectedId })
  })

  component(reviewFlow.id("verify"), async (ctx) => {
    return mutateArticleStatus(ctx, "verified")
  })

  component(reviewFlow.id("stale"), async (ctx) => {
    return mutateArticleStatus(ctx, "stale")
  })

  component(reviewFlow.id("reject"), async (ctx) => {
    return mutateArticleStatus(ctx, "rejected")
  })

  function mutateArticleStatus(ctx, newStatus) {
    const state = reviewFlow.load(ctx)
    const article = store.getArticle(state.selectedId)
    if (!article) return ui.error("No article selected.")
    const oldStatus = article.status
    article.status = newStatus
    const entries = store.ARTICLES.filter((a) => a.status === oldStatus)
    const nextSelected = entries[0] ? entries[0].id : ""
    reviewFlow.save(ctx, { status: oldStatus, selectedId: nextSelected })
    return ui.message()
      .ephemeral()
      .content(`${newStatus.charAt(0).toUpperCase() + newStatus.slice(1)} **${article.title}** (${oldStatus} → ${newStatus}).`)
      .embed(
        ui.embed(article.title)
          .color(store.statusColor(newStatus))
          .description(article.summary)
          .field("Old status", oldStatus, true)
          .field("New status", newStatus, true)
      )
      .build()
  }

  component(reviewFlow.id("edit"), async (ctx) => {
    const state = reviewFlow.load(ctx)
    const article = store.getArticle(state.selectedId)
    if (!article) return ui.error("No article selected.")
    await ctx.showModal(
      ui.form(reviewFlow.id("edit"), `Edit: ${article.title}`)
        .text("title", "Title").required().value(article.title).min(3).max(100)
        .textarea("summary", "Summary").required().value(article.summary).min(10).max(500)
        .text("tags", "Tags").value(article.tags.join(", "))
        .build()
    )
  })

  modal(reviewFlow.id("edit"), async (ctx) => {
    const state = reviewFlow.load(ctx)
    const article = store.getArticle(state.selectedId)
    if (!article) return ui.error("No article selected.")

    const newTitle = String((ctx.values || {}).title || "").trim()
    const newSummary = String((ctx.values || {}).summary || "").trim()
    const newTags = String((ctx.values || {}).tags || "").trim()

    article.title = newTitle || article.title
    article.summary = newSummary || article.summary
    article.tags = newTags ? newTags.split(",").map((t) => t.trim()).filter(Boolean) : article.tags

    return ui.message()
      .ephemeral()
      .content(`Updated **${article.title}**.`)
      .embed(
        ui.embed("Article updated")
          .color(0x57F287)
          .description(article.summary)
          .field("Title", article.title)
          .field("Tags", article.tags.join(", "))
      )
      .build()
  })

  component(reviewFlow.id("details"), async (ctx) => {
    const state = reviewFlow.load(ctx)
    const article = store.getArticle(state.selectedId)
    if (!article) return ui.error("No article selected.")
    return ui.message()
      .ephemeral()
      .content(`Details for **${article.title}**`)
      .embed(
        ui.embed("Article details")
          .color(0x5865F2)
          .field("ID", article.id)
          .field("Title", article.title)
          .field("Summary", article.summary)
          .field("Status", article.status, true)
          .field("Category", article.category, true)
          .field("Confidence", `${Math.round(article.confidence * 100)}%`, true)
          .field("Author", article.author, true)
          .field("Tags", article.tags.join(", "))
      )
      .build()
  })

  // ══════════════════════════════════════════════════════════════════════════
  //  /demo-confirm — inline confirmation dialog
  // ══════════════════════════════════════════════════════════════════════════

  command("demo-confirm", {
    description: "Showcase an inline confirmation dialog",
    options: {
      action: {
        type: "string",
        description: "The action to confirm",
        required: false,
      },
    },
  }, async (ctx) => {
    const action = String((ctx.args || {}).action || "reset the demo data").trim()
    return ui.confirm("showcase:confirm:yes", "showcase:confirm:no", {
      title: "Confirm action",
      body: `Are you sure you want to **${action}**? This is a demo, so nothing will actually happen.`,
      confirmLabel: "Yes, do it",
      cancelLabel: "Never mind",
      confirmStyle: "danger",
    })
  })

  component("showcase:confirm:yes", async (ctx) => {
    return ui.message()
      .ephemeral()
      .content("✅ **Confirmed!** The action was approved (this is just a demo).")
      .embed(
        ui.embed("Action confirmed")
          .color(0x57F287)
          .description("The user clicked Confirm. In a real bot, you would execute the action here.")
      )
      .build()
  })

  component("showcase:confirm:no", async (ctx) => {
    return ui.ok("**Cancelled.** No action was taken.")
  })

  // ══════════════════════════════════════════════════════════════════════════
  //  /demo-pager — paginated list with previous/next
  // ══════════════════════════════════════════════════════════════════════════

  command("demo-pager", {
    description: "Showcase paginated article list",
  }, async (ctx) => {
    pagerFlow.save(ctx, { page: 1, pageSize: 3 })
    return renderPagerScreen(ctx, 1, 3)
  })

  function renderPagerScreen(ctx, page, pageSize) {
    const items = store.ARTICLES
    const totalPages = Math.max(1, Math.ceil(items.length / pageSize))
    const currentPage = Math.min(page, totalPages)
    const start = (currentPage - 1) * pageSize
    const pageItems = items.slice(start, start + pageSize)

    const list = pageItems.map((a, i) =>
      `**${start + i + 1}.** ${a.title} — ${a.status} · ${a.category}`
    ).join("\n")

    return ui.message()
      .ephemeral()
      .content(`All articles (page ${currentPage}/${totalPages}):`)
      .embed(
        ui.embed(`Article list — Page ${currentPage}`)
          .color(0x5865F2)
          .description(list || "No items.")
          .footer(`Showing ${pageItems.length} of ${items.length} articles`)
      )
      .row(ui.pager(pagerFlow.id("previous"), pagerFlow.id("next"), {
        hasPrevious: currentPage > 1,
        hasNext: currentPage < totalPages,
      }))
      .build()
  }

  component(pagerFlow.id("previous"), async (ctx) => {
    const state = pagerFlow.load(ctx)
    const newPage = Math.max(1, Number(state.page || 1) - 1)
    pagerFlow.save(ctx, { ...state, page: newPage })
    return renderPagerScreen(ctx, newPage, state.pageSize || 3)
  })

  component(pagerFlow.id("next"), async (ctx) => {
    const state = pagerFlow.load(ctx)
    const newPage = Number(state.page || 1) + 1
    pagerFlow.save(ctx, { ...state, page: newPage })
    return renderPagerScreen(ctx, newPage, state.pageSize || 3)
  })

  // ══════════════════════════════════════════════════════════════════════════
  //  /demo-cards — card gallery with select navigation
  // ══════════════════════════════════════════════════════════════════════════

  command("demo-cards", {
    description: "Browse products as a card gallery with select navigation",
  }, async (ctx) => {
    const selectedId = store.PRODUCTS[0] ? store.PRODUCTS[0].id : ""
    cardFlow.save(ctx, { selectedId })
    return renderCardScreen(ctx, selectedId)
  })

  // Alias: /browse does the same as /demo-cards
  command("browse", {
    description: "Alias for /demo-cards — browse products",
  }, async (ctx) => {
    const selectedId = store.PRODUCTS[0] ? store.PRODUCTS[0].id : ""
    cardFlow.save(ctx, { selectedId })
    return renderCardScreen(ctx, selectedId)
  })

  function renderCardScreen(ctx, selectedId) {
    const products = store.PRODUCTS
    const selected = products.find((p) => p.id === selectedId) || products[0]

    if (!selected) {
      return ui.ok("No products available.")
    }

    return ui.message()
      .ephemeral()
      .content("Product catalog — select a product to inspect it.")
      .embed(
        ui.card(selected.name)
          .description(selected.description)
          .color(selected.stock > 0 ? 0x57F287 : 0xED4245)
          .meta("Price", `$${selected.price.toFixed(2)}`, true)
          .meta("Stock", selected.stock > 0 ? `${selected.stock} in stock` : "Out of stock", true)
          .meta("Category", selected.category, true)
          .meta("ID", selected.id)
      )
      .row(
        ui.select(cardFlow.id("select"))
          .placeholder("Choose a product")
          .optionEntries(products.map((p) => ({
            id: p.id,
            label: p.name,
            description: `$${p.price.toFixed(2)} · ${p.stock > 0 ? "In stock" : "Out of stock"}`,
          })), selected.id)
      )
      .row(
        ui.button(cardFlow.id("buy"), "Buy", "primary"),
        ui.button(cardFlow.id("info"), "Info", "secondary"),
        ui.button(cardFlow.id("share"), "Share", "success"),
      )
      .build()
  }

  component(cardFlow.id("select"), async (ctx) => {
    const selectedId = firstValue(ctx.values)
    if (!selectedId) return ui.error("No product selected.")
    cardFlow.save(ctx, { selectedId })
    return renderCardScreen(ctx, selectedId)
  })

  component(cardFlow.id("buy"), async (ctx) => {
    const state = cardFlow.load(ctx)
    const product = store.getProduct(state.selectedId)
    if (!product) return ui.error("No product selected.")
    return ui.confirm("showcase:buy:confirm", "showcase:buy:cancel", {
      title: "Purchase confirmation",
      body: `Buy **${product.name}** for $${product.price.toFixed(2)}?`,
      confirmLabel: "Buy now",
      confirmStyle: "success",
    })
  })

  component("showcase:buy:confirm", async (ctx) => {
    const state = cardFlow.load(ctx)
    const product = store.getProduct(state.selectedId)
    return ui.message()
      .ephemeral()
      .content(`🎉 You bought **${product ? product.name : "the item"}**! (Just kidding, this is a demo.)`)
      .embed(ui.embed("Purchase confirmed").color(0x57F287).description("In a real bot, you would process the order here."))
      .build()
  })

  component("showcase:buy:cancel", async (ctx) => {
    return ui.ok("Purchase cancelled.")
  })

  component(cardFlow.id("info"), async (ctx) => {
    const state = cardFlow.load(ctx)
    const product = store.getProduct(state.selectedId)
    if (!product) return ui.error("No product selected.")
    return ui.message()
      .ephemeral()
      .content(`Product info: **${product.name}**`)
      .embed(
        ui.embed(product.name)
          .color(0x5865F2)
          .description(product.description)
          .field("ID", product.id, true)
          .field("Price", `$${product.price.toFixed(2)}`, true)
          .field("Stock", `${product.stock} units`, true)
          .field("Category", product.category, true)
      )
      .build()
  })

  component(cardFlow.id("share"), async (ctx) => {
    const state = cardFlow.load(ctx)
    const product = store.getProduct(state.selectedId)
    if (!product) return ui.error("No product selected.")
    await ctx.defer({ ephemeral: true })
    const channelId = ctx.channel && ctx.channel.id
    if (channelId) {
      await ctx.discord.channels.send(channelId, {
        content: `🛒 Check out **${product.name}** — $${product.price.toFixed(2)}\n${product.description}`,
      })
    }
    await ctx.edit({
      content: `Shared **${product.name}** to the channel.`,
    })
  })

  // ══════════════════════════════════════════════════════════════════════════
  //  /demo-selects — all select menu types
  // ══════════════════════════════════════════════════════════════════════════

  command("demo-selects", {
    description: "Showcase all select menu types (string, user, role, channel, mentionable)",
  }, async () => {
    return ui.message()
      .ephemeral()
      .content("All select menu types in one message:")
      .embed(
        ui.embed("Select Menu Showcase")
          .color(0x5865F2)
          .description("Discord supports five types of select menus. Each row below is a different type.")
          .field("String select", "Choose from predefined options", true)
          .field("User select", "Pick a server member", true)
          .field("Role select", "Pick a server role", true)
          .field("Channel select", "Pick a channel", true)
          .field("Mentionable", "Pick a user or role", true)
      )
      .row(
        ui.select("showcase:select:string")
          .placeholder("String select — pick a fruit")
          .option("🍎 Apple", "apple", "A red fruit")
          .option("🍌 Banana", "banana", "A yellow fruit")
          .option("🍇 Grape", "grape", "A purple fruit")
      )
      .row(ui.userSelect("showcase:select:user").placeholder("User select — pick a member").build())
      .row(ui.roleSelect("showcase:select:role").placeholder("Role select — pick a role").build())
      .row(ui.channelSelect("showcase:select:channel").placeholder("Channel select — pick a channel").build())
      .row(ui.mentionableSelect("showcase:select:mentionable").placeholder("Mentionable — pick user or role").build())
      .build()
  })

  component("showcase:select:string", async (ctx) => {
    const value = firstValue(ctx.values) || "none"
    return ui.ok(`You selected the fruit: **${value}**`)
  })

  component("showcase:select:user", async (ctx) => {
    const userId = firstValue(ctx.values) || "unknown"
    return ui.ok(`You selected user ID: **${userId}**`)
  })

  component("showcase:select:role", async (ctx) => {
    const roleId = firstValue(ctx.values) || "unknown"
    return ui.ok(`You selected role ID: **${roleId}**`)
  })

  component("showcase:select:channel", async (ctx) => {
    const channelId = firstValue(ctx.values) || "unknown"
    return ui.ok(`You selected channel ID: **${channelId}**`)
  })

  component("showcase:select:mentionable", async (ctx) => {
    const id = firstValue(ctx.values) || "unknown"
    return ui.ok(`You selected mentionable ID: **${id}**`)
  })

  // ══════════════════════════════════════════════════════════════════════════
  //  /demo-alias — alias registration demo
  // ══════════════════════════════════════════════════════════════════════════

  command("demo-alias", {
    description: "Showcase alias registration (same as /demo-alias-alt)",
  }, async () => {
    return ui.message()
      .ephemeral()
      .content("This command is an alias demo. Both `/demo-alias` and `/demo-alias-alt` share the same handler.")
      .embed(
        ui.embed("Alias Registration")
          .color(0x5865F2)
          .description("Use `ui.alias()` to register the same command handler under multiple names for discoverability.")
          .field("Primary", "/demo-alias", true)
          .field("Alias", "/demo-alias-alt", true)
      )
      .build()
  })

  command("demo-alias-alt", {
    description: "Alias of /demo-alias — demonstrates alias registration",
  }, async () => {
    return ui.message()
      .ephemeral()
      .content("This command is an alias demo. Both `/demo-alias` and `/demo-alias-alt` share the same handler.")
      .embed(
        ui.embed("Alias Registration")
          .color(0x5865F2)
          .description("Use `ui.alias()` to register the same command handler under multiple names for discoverability.")
          .field("Primary", "/demo-alias", true)
          .field("Alias", "/demo-alias-alt", true)
      )
      .build()
  })
})

// ── Utility functions ────────────────────────────────────────────────────────

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
