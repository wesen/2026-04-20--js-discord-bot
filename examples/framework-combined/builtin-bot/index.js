const { defineBot } = require("discord");

module.exports = defineBot(({ command, event, configure }) => {
  configure({
    name: "framework-combined-builtin",
    description: "Built-in bot used by the combined framework + botcli example",
    category: "examples",
  });

  command("combined-builtin-ping", {
    description: "Show that the explicit built-in bot is alive",
  }, async (ctx) => {
    const mode = String((ctx.config && ctx.config.mode) || "(unset)");
    const source = String((ctx.config && ctx.config.source) || "(unset)");
    return {
      content: `framework-combined built-in bot is alive. mode=${mode}, source=${source}`,
      ephemeral: true,
    };
  });

  event("ready", async (ctx) => {
    ctx.log.info("framework-combined builtin ready", {
      mode: ctx.config && ctx.config.mode,
      source: ctx.config && ctx.config.source,
    });
  });
});
