const { defineBot } = require("discord");
module.exports = defineBot(({ command, configure }) => {
  configure({ name: "duplicate-name", description: "Fixture A" });
  command("dupe-a", async () => ({ content: "a" }));
});
