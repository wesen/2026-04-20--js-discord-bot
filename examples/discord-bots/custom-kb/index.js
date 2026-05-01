const { defineBot } = require("discord")
const { createLinkStore } = require("./lib/store")
const ui = require("ui")

const store = createLinkStore()
const SEARCH_SELECT = "custom-kb:search:select"
const SEARCH_REFRESH = "custom-kb:search:refresh"
const ADD_MODAL = "custom-kb:add:submit"

module.exports = defineBot(({ command, event, component, modal, autocomplete, configure }) => {
  configure({
    name: "custom-kb",
    description: "Store useful links in SQLite and search them from Discord with the UI DSL",
    category: "knowledge",
    run: {
      fields: {
        "db-path": { type: "string", default: "./examples/discord-bots/custom-kb/data/custom-kb.sqlite", description: "SQLite database path for stored links" },
        "search-limit": { type: "int", default: 10, description: "Maximum links to show in search results" },
      },
    },
  })

  event("ready", async (ctx) => {
    store.ensure(ctx.config)
    ctx.log.info("custom-kb bot ready", { user: ctx.me && ctx.me.username, dbPath: store.configPath(), links: store.count(ctx.config) })
  })

  command("kb-add", { description: "Add a link to the custom knowledge base" }, async (ctx) => {
    await ctx.showModal(
      ui.form(ADD_MODAL, "Add KB Link")
        .text("url", "URL").required().min(3).max(500)
        .text("title", "Title").required().min(2).max(120)
        .textarea("summary", "Summary / why this matters").max(1000)
        .text("tags", "Tags (comma-separated)").max(200)
        .build()
    )
  })

  modal(ADD_MODAL, async (ctx) => {
    try {
      const link = store.addLink(ctx.config, linkPayloadFromValues(ctx))
      return renderLinkMessage("Saved link", link, true)
    } catch (err) {
      return ui.message().ephemeral().content(`Could not save link: ${err.message || err}`).build()
    }
  })

  command("kb-link", {
    description: "Add a link directly without opening the modal",
    options: {
      url: { type: "string", description: "URL to store", required: true },
      title: { type: "string", description: "Short title", required: true },
      summary: { type: "string", description: "Optional summary", required: false },
      tags: { type: "string", description: "Comma-separated tags", required: false },
    },
  }, async (ctx) => {
    const args = ctx.args || {}
    const link = store.addLink(ctx.config, linkPayloadFromArgs(ctx, args))
    return renderLinkMessage("Saved link", link, true)
  })

  command("kb-search", {
    description: "Search stored KB links",
    options: { query: { type: "string", description: "Search title, URL, summary, or tags", required: true, autocomplete: true } },
  }, async (ctx) => renderSearch(ctx, String((ctx.args || {}).query || "")))

  command("kb-list", { description: "List recent KB links" }, async (ctx) => renderSearch(ctx, ""))

  component(SEARCH_SELECT, async (ctx) => {
    const id = firstValue(ctx.values)
    const link = store.getLink(ctx.config, id)
    if (!link) return ui.message().ephemeral().content("That link is no longer available.").build()
    return renderLinkMessage("Selected link", link, true)
  })

  component(SEARCH_REFRESH, async (ctx) => renderSearch(ctx, String((ctx.store && ctx.store.get && ctx.store.get("custom-kb:last-query")) || "")))

  autocomplete("kb-search", "query", async (ctx) => {
    const query = String((ctx.args || {}).query || "")
    return store.search(ctx.config, query, 10).map((link) => ({ name: link.title.slice(0, 100), value: link.title.slice(0, 100) }))
  })
})

function renderSearch(ctx, query) {
  store.ensure(ctx.config)
  if (ctx.store && ctx.store.set) ctx.store.set("custom-kb:last-query", query)
  const limit = Number((ctx.config || {}).searchLimit || (ctx.config || {}).search_limit || 10)
  const links = store.search(ctx.config, query, limit)
  const title = query ? `KB search: ${query}` : "Recent KB links"
  let msg = ui.message()
    .ephemeral()
    .content(links.length ? `Found ${links.length} link(s).` : "No links found yet. Use /kb-add or /kb-link to store one.")
    .embed(
      ui.embed(title)
        .color(0x5865F2)
        .description(links.length ? links.map((link, i) => `${i + 1}. [${link.title}](${link.url})${link.tags.length ? ` — ${link.tags.map((t) => `#${t}`).join(" ")}` : ""}`).join("\n") : "The custom KB is empty or there were no matches.")
        .footer("custom-kb · SQLite-backed link search")
    )
  if (links.length) {
    msg = msg.row(
      ui.select(SEARCH_SELECT)
        .placeholder("Open a stored link")
        .optionEntries(links.map((link) => ({ id: link.id, label: link.title.slice(0, 100), description: link.url.slice(0, 100) })))
    )
  }
  return msg.row(ui.button(SEARCH_REFRESH, "Refresh", "secondary")).build()
}

function renderLinkMessage(heading, link, ephemeral) {
  let msg = ui.message().content(`[Open link](${link.url})`)
  if (ephemeral) msg = msg.ephemeral()
  return msg.embed(
    ui.embed(heading)
      .color(0x57F287)
      .description(link.summary || link.url)
      .field("Title", link.title, false)
      .field("URL", link.url, false)
      .field("Tags", link.tags.length ? link.tags.map((t) => `#${t}`).join(" ") : "(none)", true)
      .field("ID", link.id, true)
      .footer("custom-kb · stored in SQLite")
  ).build()
}

function linkPayloadFromValues(ctx) { return linkPayload(ctx, ctx.values || {}) }
function linkPayloadFromArgs(ctx, args) { return linkPayload(ctx, args || {}) }
function linkPayload(ctx, values) {
  return {
    url: values.url,
    title: values.title,
    summary: values.summary,
    tags: values.tags,
    addedBy: ctx.user && ctx.user.id,
    guildId: ctx.guild && ctx.guild.id,
    channelId: ctx.channel && ctx.channel.id,
  }
}
function firstValue(values) { return values && values.length ? String(values[0]) : "" }
