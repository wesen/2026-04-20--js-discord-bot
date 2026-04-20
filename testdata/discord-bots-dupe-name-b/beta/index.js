const { defineBot } = require("discord");
module.exports = defineBot(({ command, configure }) => {
  configure({ name: "duplicate-name", description: "Fixture B" });
  command("dupe-b", async () => ({ content: "b" }));
});
