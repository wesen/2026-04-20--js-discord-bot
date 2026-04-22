const { defineBot } = require("discord");

module.exports = defineBot(({ command, configure }) => {
  configure({ name: "demo", description: "Demo bot" });
  command("ping", async () => ({ content: "pong" }));
});

function status() {
  return { active: true, commands: 1 };
}

__verb__("status", {
  short: "Check demo bot status",
  output: "glaze"
});

function run() {
  return { status: "placeholder" };
}

__verb__("run", {
  short: "Run the demo bot",
  output: "text",
  fields: {
    "bot-token": { type: "string", required: true, help: "Discord bot token" },
    "api-key": { type: "string", help: "External API key" }
  }
});
