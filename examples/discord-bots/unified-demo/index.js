const { defineBot } = require("discord");

module.exports = defineBot(({ command, event, configure }) => {
  configure({
    name: "unified-demo",
    description: "Example bot that also exposes jsverbs metadata",
    category: "examples"
  });

  command("unified-ping", {
    description: "Show that ctx.config values are available inside bot handlers"
  }, async (ctx) => {
    const dbPath = String((ctx.config && ctx.config.db_path) || "(unset)");
    const apiKey = String((ctx.config && ctx.config.api_key) || "(unset)");
    return {
      content: `Unified demo is alive. db_path=${dbPath}, api_key=${apiKey}`,
      ephemeral: true,
    };
  });

  event("ready", async (ctx) => {
    ctx.log.info("unified-demo ready", {
      db_path: ctx.config && ctx.config.db_path,
      api_key: ctx.config && ctx.config.api_key,
    });
  });
});

function status() {
  return {
    active: true,
    mode: "unified-demo",
    note: "This verb runs through jsverbs with the Discord registrar available",
  };
}

__verb__("status", {
  short: "Return unified-demo metadata as structured output",
  output: "glaze",
});

function run() {
  return { status: "host-managed" };
}

__verb__("run", {
  short: "Run the unified-demo Discord bot",
  output: "text",
  fields: {
    "bot-token": {
      type: "string",
      required: true,
      help: "Discord bot token"
    },
    "application-id": {
      type: "string",
      required: true,
      help: "Discord application/client ID"
    },
    "guild-id": {
      type: "string",
      help: "Optional guild ID for development sync"
    },
    "db-path": {
      type: "string",
      default: "./examples/discord-bots/unified-demo/data/demo.sqlite",
      help: "SQLite path exposed to the bot as ctx.config.db_path"
    },
    "api-key": {
      type: "string",
      help: "Example external API key exposed as ctx.config.api_key"
    }
  }
});
