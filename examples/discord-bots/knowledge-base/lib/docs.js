const ARTICLES = {
  architecture: "The bot repository runner discovers named bot implementations and routes slash commands by owning bot.",
  onboarding: "Start with bots list, then bots help knowledge-base, then bots run knowledge-base.",
  runtime: "Each selected bot gets its own JavaScript runtime and store."
};

function search(query) {
  query = String(query || "").toLowerCase();
  return Object.entries(ARTICLES)
    .filter(([key, value]) => key.includes(query) || value.toLowerCase().includes(query))
    .map(([key, value]) => ({ key, excerpt: value }));
}

function article(name) {
  return ARTICLES[name] || "Article not found.";
}

module.exports = { search, article };
