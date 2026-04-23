const { defineBot } = require("discord");
const app = require("app");

module.exports = defineBot(({ command, event, configure }) => {
  configure({
    name: "framework-custom-module",
    description: app.description(),
    category: "examples",
  });

  command("app-info", {
    description: "Show data coming from the Go-side app module",
  }, async () => {
    return {
      content: `${app.greeting()} (${app.name()})`,
      ephemeral: true,
    };
  });

  event("ready", async (ctx) => {
    ctx.log.info("framework-custom-module ready", {
      app_name: app.name(),
      app_description: app.description(),
    });
  });
});
