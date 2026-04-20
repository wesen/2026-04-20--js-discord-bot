const { defineBot } = require("discord");
const docs = require("./lib/docs");

module.exports = defineBot(({ command, event, configure }) => {
  configure({
    name: "knowledge-base",
    description: "Search and summarize internal docs from JavaScript",
    category: "knowledge",
    run: {
      fields: {
        index_path: {
          type: "string",
          help: "Optional path label for the active docs index",
          default: "builtin-docs"
        },
        read_only: {
          type: "bool",
          help: "Disable write operations for future knowledge-base mutations",
          default: true
        }
      }
    }
  });

  command("kb-search", {
    description: "Search the knowledge base",
    options: {
      query: {
        type: "string",
        description: "Search query",
        required: true,
      }
    }
  }, async (ctx) => {
    const matches = docs.search(ctx.args.query);
    const indexPath = ctx.config && ctx.config.index_path || "builtin-docs";
    if (matches.length === 0) {
      return { content: `No docs found for ${ctx.args.query} in ${indexPath}.`, ephemeral: true };
    }
    return {
      content: `Found ${matches.length} document(s) for ${ctx.args.query} in ${indexPath}.`,
      embeds: [{
        title: "Knowledge Base Search",
        description: matches.map((m) => `**${m.key}** — ${m.excerpt}`).join("\n"),
        color: 0x5865F2,
      }]
    };
  });

  command("kb-article", {
    description: "Fetch one knowledge base article",
    options: {
      name: {
        type: "string",
        description: "Article name",
        required: true,
      }
    }
  }, async (ctx) => {
    return {
      content: docs.article(ctx.args.name),
      embeds: [{ title: `Article: ${ctx.args.name}`, color: 0x57F287 }]
    };
  });

  event("ready", async (ctx) => {
    ctx.log.info("knowledge-base bot ready", { user: ctx.me && ctx.me.username });
  });

  event("messageCreate", async (ctx) => {
    const content = (ctx.message && ctx.message.content || "").trim();
    if (content === "!kb") {
      await ctx.reply({
        content: "Knowledge base bot is online.",
        embeds: [{ title: "KB", description: "Try /kb-search or /kb-article.", color: 0xFEE75C }]
      });
    }
  });
});
