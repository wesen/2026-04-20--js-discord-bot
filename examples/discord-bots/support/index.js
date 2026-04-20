const { defineBot } = require("discord");

module.exports = defineBot(({ command, event, configure }) => {
  configure({
    name: "support",
    description: "Support workflows with deferred replies and follow-ups",
    category: "operations"
  });

  command("support-ticket", {
    description: "Create a support ticket preview",
    options: {
      topic: {
        type: "string",
        description: "Ticket topic",
        required: true,
      }
    }
  }, async (ctx) => {
    await ctx.defer({ ephemeral: true });
    await ctx.edit({
      content: `Drafted a support ticket for ${ctx.args.topic}.`,
      embeds: [{
        title: "Support Ticket Draft",
        description: "Deferred + edited interaction response from the support bot.",
        color: 0x57F287,
      }],
      components: [{
        type: "actionRow",
        components: [{
          type: "button",
          style: "link",
          label: "Support Handbook",
          url: "https://example.com/support-handbook"
        }]
      }]
    });
    await ctx.followUp({ content: "A follow-up reminder was also sent.", ephemeral: true });
  });

  command("support-status", {
    description: "Show the support bot status"
  }, async () => {
    return {
      content: "Support systems nominal.",
      embeds: [{ title: "Support Status", description: "Everything looks good.", color: 0x3498DB }]
    };
  });

  event("guildCreate", async (ctx) => {
    ctx.log.info("support bot saw guildCreate", { guild: ctx.guild && ctx.guild.name });
  });

  event("messageCreate", async (ctx) => {
    const content = (ctx.message && ctx.message.content || "").trim();
    if (content === "!support") {
      await ctx.reply({ content: "Support bot received your message trigger." });
    }
  });
});
