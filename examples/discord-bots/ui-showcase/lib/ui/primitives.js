// UI DSL primitives — generic builder helpers for Discord payloads
//
// These builders produce plain Discord API objects. They do not manage state
// or routing; they just make payload construction less verbose and more readable.

// ── Message builder ──────────────────────────────────────────────────────────

/**
 * Start building an interaction response or channel message payload.
 *
 *   ui.message()
 *     .content("Hello")
 *     .ephemeral()
 *     .embed(ui.embed("Title").description("desc"))
 *     .row(ui.button("id", "Click", "primary"))
 *     .build()
 */
function message() {
  const parts = {
    _content: "",
    _ephemeral: false,
    _embeds: [],
    _components: [],
    _files: [],
    _tts: false,
  }

  const builder = {
    content(text) {
      parts._content = String(text || "")
      return builder
    },

    ephemeral(flag) {
      parts._ephemeral = flag !== false
      return builder
    },

    tts(flag) {
      parts._tts = flag !== false
      return builder
    },

    embed(e) {
      if (e) {
        // Auto-build if a builder was passed instead of a plain object
        const built = typeof e.build === 'function' ? e.build() : e
        parts._embeds.push(built)
      }
      return builder
    },

    row(...components) {
      const flat = Array.isArray(components[0]) ? components[0] : components
      if (flat.length > 0) {
        parts._components.push({ type: "actionRow", components: flat })
      }
      return builder
    },

    rows(rowList) {
      for (const r of rowList) {
        if (r && r.components) {
          parts._components.push(r)
        } else if (Array.isArray(r)) {
          builder.row(r)
        }
      }
      return builder
    },

    file(name, content, contentType) {
      const f = { name: String(name), content: String(content) }
      if (contentType) f.contentType = contentType
      parts._files.push(f)
      return builder
    },

    build() {
      const payload = {}
      if (parts._content) payload.content = parts._content
      if (parts._embeds.length > 0) payload.embeds = parts._embeds
      if (parts._components.length > 0) payload.components = parts._components
      if (parts._files.length > 0) payload.files = parts._files
      if (parts._tts) payload.tts = true
      if (parts._ephemeral) payload.ephemeral = true
      return payload
    },
  }

  return builder
}

// ── Embed builder ────────────────────────────────────────────────────────────

/**
 * Build a Discord embed.
 *
 *   ui.embed("Search results")
 *     .color(0x5865F2)
 *     .description("Found 5 items")
 *     .field("Status", "OK", true)
 *     .footer("Page 1/3")
 */
function embed(title) {
  const parts = {
    _title: title || "",
    _description: "",
    _color: 0,
    _fields: [],
    _footer: null,
    _url: "",
    _thumbnail: "",
    _image: "",
    _author: null,
    _timestamp: false,
  }

  const builder = {
    title(text) {
      parts._title = String(text || "")
      return builder
    },

    description(text) {
      parts._description = String(text || "")
      return builder
    },

    color(value) {
      parts._color = Number(value || 0)
      return builder
    },

    url(href) {
      parts._url = String(href || "")
      return builder
    },

    thumbnail(url) {
      parts._thumbnail = String(url || "")
      return builder
    },

    image(url) {
      parts._image = String(url || "")
      return builder
    },

    field(name, value, inline) {
      parts._fields.push({
        name: String(name || ""),
        value: String(value || ""),
        inline: Boolean(inline),
      })
      return builder
    },

    fields(list) {
      for (const f of list || []) {
        builder.field(f.name, f.value, f.inline)
      }
      return builder
    },

    footer(text, iconUrl) {
      parts._footer = { text: String(text || "") }
      if (iconUrl) parts._footer.icon_url = String(iconUrl)
      return builder
    },

    author(name, iconUrl, url) {
      parts._author = { name: String(name || "") }
      if (iconUrl) parts._author.icon_url = String(iconUrl)
      if (url) parts._author.url = String(url)
      return builder
    },

    timestamp(flag) {
      parts._timestamp = flag !== false
      return builder
    },

    build() {
      const e = {}
      if (parts._title) e.title = parts._title
      if (parts._description) e.description = parts._description
      if (parts._color) e.color = parts._color
      if (parts._url) e.url = parts._url
      if (parts._thumbnail) e.thumbnail = { url: parts._thumbnail }
      if (parts._image) e.image = { url: parts._image }
      if (parts._fields.length > 0) e.fields = parts._fields
      if (parts._footer) e.footer = parts._footer
      if (parts._author) e.author = parts._author
      if (parts._timestamp) e.timestamp = new Date().toISOString()
      return e
    },
  }

  return builder
}

// ── Component builders ───────────────────────────────────────────────────────

/**
 * Build a button component.
 *
 *   ui.button("my:id", "Click me", "primary")
 *   ui.button("my:link", "Visit", "link").url("https://example.com")
 */
function button(customId, label, style) {
  const b = {
    type: "button",
    customId: String(customId || ""),
    label: String(label || ""),
    style: resolveButtonStyle(style || "secondary"),
  }

  function emoji(name) {
    b.emoji = typeof name === "string" ? { name } : name
    return b
  }

  function disabled(flag) {
    b.disabled = flag !== false
    return b
  }

  function url(href) {
    b.url = String(href || "")
    return b
  }

  function build() {
    return b
  }

  b.emoji = emoji
  b.disabled = disabled
  b.url = url
  b.build = build
  return b
}

function resolveButtonStyle(style) {
  const map = {
    primary: "primary",
    secondary: "secondary",
    success: "success",
    danger: "danger",
    link: "link",
  }
  return map[String(style || "secondary").toLowerCase()] || "secondary"
}

/**
 * Build a select menu component.
 *
 *   ui.select("my:select")
 *     .placeholder("Choose one")
 *     .option("A", "a")
 *     .option("B", "b")
 */
function selectMenu(customId) {
  const s = {
    type: "select",
    customId: String(customId || ""),
    placeholder: "",
    options: [],
    minValues: undefined,
    maxValues: undefined,
    disabled: false,
  }

  const chain = {
    placeholder(text) {
      s.placeholder = String(text || "")
      return chain
    },

    option(label, value, description, isDefault) {
      const opt = {
        label: String(label || ""),
        value: String(value || ""),
      }
      if (description) opt.description = String(description)
      if (isDefault) opt.default = true
      s.options.push(opt)
      return chain
    },

    options(list) {
      for (const o of list || []) {
        if (typeof o === "object" && o.label) {
          chain.option(o.label, o.value, o.description, o.default)
        }
      }
      return chain
    },

    optionEntries(entries, selectedId) {
      for (const entry of entries || []) {
        chain.option(
          truncate(entry.label || entry.title || entry.name || entry.id || "Item", 100),
          String(entry.value || entry.id),
          truncate(entry.description || "", 100),
          entry.id === selectedId || entry.value === selectedId
        )
      }
      return chain
    },

    minValues(n) {
      s.minValues = Number(n)
      return chain
    },

    maxValues(n) {
      s.maxValues = Number(n)
      return chain
    },

    disabled(flag) {
      s.disabled = flag !== false
      return chain
    },

    build() {
      return s
    },
  }

  return chain
}

/**
 * Build a user select menu.
 */
function userSelect(customId) {
  const s = {
    type: "userSelect",
    customId: String(customId || ""),
    placeholder: "",
    defaultUsers: [],
  }
  const chain = {
    placeholder(text) { s.placeholder = String(text || ""); return chain },
    build() { return s },
  }
  return chain
}

/**
 * Build a role select menu.
 */
function roleSelect(customId) {
  const s = {
    type: "roleSelect",
    customId: String(customId || ""),
    placeholder: "",
    defaultRoles: [],
  }
  const chain = {
    placeholder(text) { s.placeholder = String(text || ""); return chain },
    build() { return s },
  }
  return chain
}

/**
 * Build a channel select menu.
 */
function channelSelect(customId) {
  const s = {
    type: "channelSelect",
    customId: String(customId || ""),
    placeholder: "",
    channelTypes: [],
  }
  const chain = {
    placeholder(text) { s.placeholder = String(text || ""); return chain },
    channelTypes(types) { s.channelTypes = types; return chain },
    build() { return s },
  }
  return chain
}

/**
 * Build a mentionable select menu (users + roles).
 */
function mentionableSelect(customId) {
  const s = {
    type: "mentionableSelect",
    customId: String(customId || ""),
    placeholder: "",
  }
  const chain = {
    placeholder(text) { s.placeholder = String(text || ""); return chain },
    build() { return s },
  }
  return chain
}

// ── Row helpers ──────────────────────────────────────────────────────────────

/**
 * Wrap components into an action row.
 *
 *   ui.row(ui.button("a", "A"), ui.button("b", "B"))
 */
function row(...components) {
  const flat = Array.isArray(components[0]) ? components[0] : components
  return { type: "actionRow", components: flat }
}

/**
 * Build multiple action rows from arrays of components.
 */
function rows(...componentLists) {
  return componentLists.map((list) => {
    const flat = Array.isArray(list) ? list : [list]
    return { type: "actionRow", components: flat }
  })
}

// ── Pager builder ────────────────────────────────────────────────────────────

/**
 * Build a previous/next pager row.
 *
 *   ui.pager("ns:prev", "ns:next", { hasPrevious: true, hasNext: true })
 */
function pager(previousId, nextId, controls) {
  const c = controls || {}
  return row(
    button(previousId, "◀ Previous", "secondary").disabled(!c.hasPrevious),
    button(nextId, "Next ▶", "secondary").disabled(!c.hasNext)
  )
}

// ── Action bar builder ───────────────────────────────────────────────────────

/**
 * Build a row of action buttons from a definition array.
 *
 *   ui.actions([
 *     { id: "verify", label: "Verify", style: "success" },
 *     { id: "reject", label: "Reject", style: "danger" },
 *   ])
 */
function actions(definitions) {
  const buttons = (definitions || []).map((def) => {
    return button(def.id || def.customId, def.label, def.style || "secondary")
  })
  return row(...buttons)
}

// ── Form/modal builder ───────────────────────────────────────────────────────

/**
 * Build a modal (form) payload.
 *
 *   ui.form("my:modal", "My Form")
 *     .text("name", "Name").required().min(3).max(100)
 *     .textarea("desc", "Description")
 *     .build()
 */
function form(customId, title) {
  const fields = []
  const formTitle = String(title || "Form")
  const formId = String(customId || "form")

  let currentField = null

  function pushCurrent() {
    if (currentField) {
      fields.push(currentField)
      currentField = null
    }
  }

  const builder = {
    text(id, label) {
      pushCurrent()
      currentField = {
        type: "textInput",
        customId: String(id),
        label: String(label || id),
        style: "short",
        required: false,
        value: "",
      }
      return builder
    },

    textarea(id, label) {
      pushCurrent()
      currentField = {
        type: "textInput",
        customId: String(id),
        label: String(label || id),
        style: "paragraph",
        required: false,
        value: "",
      }
      return builder
    },

    required(flag) {
      if (currentField) currentField.required = flag !== false
      return builder
    },

    value(text) {
      if (currentField) currentField.value = String(text || "")
      return builder
    },

    placeholder(text) {
      if (currentField) currentField.placeholder = String(text || "")
      return builder
    },

    min(n) {
      if (currentField) currentField.minLength = Number(n)
      return builder
    },

    max(n) {
      if (currentField) currentField.maxLength = Number(n)
      return builder
    },

    build() {
      pushCurrent()
      const components = fields.map((f) => ({
        type: "actionRow",
        components: [f],
      }))
      return {
        customId: formId,
        title: formTitle,
        components,
      }
    },
  }

  return builder
}

// ── Confirmation dialog ──────────────────────────────────────────────────────

/**
 * Build an inline confirmation dialog as an ephemeral message with
 * confirm/cancel buttons.
 *
 *   ui.confirm("ns:confirm", "ns:cancel", {
 *     title: "Delete item?",
 *     body: "This cannot be undone.",
 *     confirmLabel: "Delete",
 *     cancelLabel: "Cancel",
 *     confirmStyle: "danger",
 *   })
 */
function confirm(confirmId, cancelId, options) {
  const opts = options || {}
  return message()
    .ephemeral()
    .content(opts.body || "Are you sure?")
    .embed(
      embed(opts.title || "Confirm")
        .description(opts.body || "Please confirm or cancel this action.")
        .color(0xFEE75C)
        .build()
    )
    .row(
      button(confirmId, opts.confirmLabel || "Confirm", opts.confirmStyle || "danger"),
      button(cancelId, opts.cancelLabel || "Cancel", "secondary")
    )
    .build()
}

// ── Utility helpers ──────────────────────────────────────────────────────────

function truncate(text, maxLen) {
  const str = String(text || "")
  if (str.length <= (maxLen || 100)) return str
  return str.slice(0, (maxLen || 100) - 3).trim() + "..."
}

function ok(content) {
  return { content: String(content || "Done."), ephemeral: true }
}

function error(content) {
  return { content: "⚠️ " + String(content || "Something went wrong."), ephemeral: true }
}

function emptyResults(query) {
  return message()
    .ephemeral()
    .content(query ? `No results found for **${query}**.` : "No results found.")
    .build()
}

// ── Card helper ──────────────────────────────────────────────────────────────

/**
 * Build a card-style embed for displaying an item with standard fields.
 *
 *   ui.card("Item Title")
 *     .description("A brief description")
 *     .color(0x57F287)
 *     .meta("Status", "Active", true)
 *     .meta("ID", "abc-123")
 *     .footer("From the demo store")
 */
function card(title) {
  const e = embed(title)
  const originalBuild = e.build

  const chain = {
    description(text) { e.description(text); return chain },
    color(value) { e.color(value); return chain },
    field(name, value, inline) { e.field(name, value, inline); return chain },
    footer(text, iconUrl) { e.footer(text, iconUrl); return chain },
    author(name, iconUrl, url) { e.author(name, iconUrl, url); return chain },
    meta(name, value, inline) { e.field(name, value, inline); return chain },
    build() { return originalBuild.call(e) },
  }

  return chain
}

// ── Exports ──────────────────────────────────────────────────────────────────

module.exports = {
  message,
  embed,
  button,
  select: selectMenu,
  userSelect,
  roleSelect,
  channelSelect,
  mentionableSelect,
  row,
  rows,
  pager,
  actions,
  form,
  confirm,
  card,
  truncate,
  ok,
  error,
  emptyResults,
}
