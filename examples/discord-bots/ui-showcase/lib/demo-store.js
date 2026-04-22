// In-memory demo data store for the UI showcase bot.
//
// Provides articles, products, and tasks that the various demo screens
// can search, paginate, review, and display.

const ARTICLES = [
  { id: "art-1", title: "Getting Started with Discord Bots", summary: "Learn how to build your first Discord bot using the JavaScript DSL.", category: "tutorial", status: "verified", tags: ["discord", "bot", "getting-started"], confidence: 0.95, author: "alice" },
  { id: "art-2", title: "UI DSL Design Patterns", summary: "Common patterns for building elegant Discord bot UIs with builders and flows.", category: "design", status: "verified", tags: ["ui", "dsl", "patterns"], confidence: 0.92, author: "bob" },
  { id: "art-3", title: "Stateful Screen Flows", summary: "How to manage per-user state across button clicks and select menus.", category: "tutorial", status: "draft", tags: ["state", "flow", "screen"], confidence: 0.80, author: "alice" },
  { id: "art-4", title: "Modal Forms Best Practices", summary: "Tips for building user-friendly modal forms in Discord bots.", category: "guide", status: "review", tags: ["modal", "form", "ux"], confidence: 0.75, author: "carol" },
  { id: "art-5", title: "Embed Color Theory", summary: "Choose colors that match your bot's personality and message intent.", category: "design", status: "verified", tags: ["embed", "color", "design"], confidence: 0.88, author: "bob" },
  { id: "art-6", title: "Pagination in Discord", summary: "Techniques for paginating search results and long lists.", category: "tutorial", status: "draft", tags: ["pagination", "search", "ux"], confidence: 0.70, author: "dave" },
  { id: "art-7", title: "Confirmation Dialogs", summary: "When and how to use inline confirmation before destructive actions.", category: "guide", status: "verified", tags: ["confirm", "dialog", "safety"], confidence: 0.85, author: "alice" },
  { id: "art-8", title: "Select Menu Patterns", summary: "Using string selects, user selects, role selects, and channel selects effectively.", category: "tutorial", status: "review", tags: ["select", "menu", "components"], confidence: 0.78, author: "carol" },
  { id: "art-9", title: "Alias Registration", summary: "Register the same command under multiple names for discoverability.", category: "reference", status: "verified", tags: ["alias", "command", "registration"], confidence: 0.90, author: "bob" },
  { id: "art-10", title: "Autocomplete Helpers", summary: "Provide smart suggestions while users type slash-command options.", category: "tutorial", status: "draft", tags: ["autocomplete", "search", "suggestions"], confidence: 0.65, author: "dave" },
  { id: "art-11", title: "Card Galleries", summary: "Display collections of items as browsable cards with select navigation.", category: "design", status: "review", tags: ["card", "gallery", "navigation"], confidence: 0.72, author: "alice" },
  { id: "art-12", title: "Outbound Operations", summary: "Send messages, manage channels, and interact with Discord resources.", category: "reference", status: "verified", tags: ["outbound", "api", "operations"], confidence: 0.93, author: "carol" },
]

const PRODUCTS = [
  { id: "prod-1", name: "Widget Pro", price: 29.99, description: "A professional-grade widget for all your widgeting needs.", stock: 42, category: "widgets" },
  { id: "prod-2", name: "Gizmo Lite", price: 14.99, description: "The lightweight version of our popular gizmo.", stock: 128, category: "gizmos" },
  { id: "prod-3", name: "Doodad Max", price: 49.99, description: "Premium doodad with all features unlocked.", stock: 7, category: "doodads" },
  { id: "prod-4", name: "Thingamajig", price: 9.99, description: "Simple, reliable, and always useful.", stock: 256, category: "basics" },
  { id: "prod-5", name: "Whatchamacallit", price: 34.99, description: "You'll know what to call it once you see it.", stock: 0, category: "specialty" },
  { id: "prod-6", name: "Doohickey Plus", price: 22.49, description: "Enhanced doohickey with extra features.", stock: 63, category: "doohickeys" },
]

const TASKS = [
  { id: "task-1", title: "Review form DSL builder", priority: "high", status: "todo", assignee: "alice" },
  { id: "task-2", title: "Fix pager edge case", priority: "medium", status: "in-progress", assignee: "bob" },
  { id: "task-3", title: "Add select menu examples", priority: "low", status: "done", assignee: "carol" },
  { id: "task-4", title: "Test confirmation dialogs", priority: "high", status: "todo", assignee: "dave" },
  { id: "task-5", title: "Document screen helper", priority: "medium", status: "in-progress", assignee: "alice" },
  { id: "task-6", title: "Optimize embed builder", priority: "low", status: "todo", assignee: "bob" },
  { id: "task-7", title: "Add card gallery demo", priority: "medium", status: "done", assignee: "carol" },
  { id: "task-8", title: "Refine autocomplete UX", priority: "high", status: "in-progress", assignee: "dave" },
]

function searchArticles(query, limit) {
  const q = String(query || "").toLowerCase().trim()
  const l = Math.max(1, Math.min(Number(limit || 25), 50))
  if (!q) return ARTICLES.slice(0, l)
  return ARTICLES.filter((a) =>
    a.title.toLowerCase().includes(q) ||
    a.summary.toLowerCase().includes(q) ||
    a.tags.some((t) => t.toLowerCase().includes(q)) ||
    a.category.toLowerCase().includes(q) ||
    a.status.toLowerCase().includes(q)
  ).slice(0, l)
}

function getArticle(id) {
  return ARTICLES.find((a) => a.id === id || a.title.toLowerCase() === String(id || "").toLowerCase()) || null
}

function searchProducts(query) {
  const q = String(query || "").toLowerCase().trim()
  if (!q) return PRODUCTS
  return PRODUCTS.filter((p) =>
    p.name.toLowerCase().includes(q) ||
    p.description.toLowerCase().includes(q) ||
    p.category.toLowerCase().includes(q)
  )
}

function getProduct(id) {
  return PRODUCTS.find((p) => p.id === id) || null
}

function listTasks(status) {
  const s = String(status || "").toLowerCase().trim()
  if (!s) return TASKS
  return TASKS.filter((t) => t.status === s)
}

function getTask(id) {
  return TASKS.find((t) => t.id === id) || null
}

function setTaskStatus(id, status) {
  const task = getTask(id)
  if (task) task.status = status
  return task
}

function articleSuggestions(focused) {
  const q = String(focused || "").toLowerCase().trim()
  return ARTICLES
    .filter((a) => !q || a.title.toLowerCase().includes(q) || a.summary.toLowerCase().includes(q))
    .map((a) => ({ name: `${a.title} (${a.status})`, value: a.id }))
    .slice(0, 25)
}

function productSuggestions(focused) {
  const q = String(focused || "").toLowerCase().trim()
  return PRODUCTS
    .filter((p) => !q || p.name.toLowerCase().includes(q))
    .map((p) => ({ name: `${p.name} — $${p.price}`, value: p.id }))
    .slice(0, 25)
}

function statusColor(status) {
  switch (String(status || "").toLowerCase()) {
    case "verified":
    case "done":
      return 0x57F287
    case "review":
    case "in-progress":
      return 0x5865F2
    case "stale":
    case "todo":
      return 0xFEE75C
    case "rejected":
      return 0xED4245
    default:
      return 0x95A5A6
  }
}

function priorityColor(priority) {
  switch (String(priority || "").toLowerCase()) {
    case "high": return 0xED4245
    case "medium": return 0xFEE75C
    case "low": return 0x57F287
    default: return 0x95A5A6
  }
}

module.exports = {
  ARTICLES,
  PRODUCTS,
  TASKS,
  searchArticles,
  getArticle,
  searchProducts,
  getProduct,
  listTasks,
  getTask,
  setTaskStatus,
  articleSuggestions,
  productSuggestions,
  statusColor,
  priorityColor,
}
