const { defineBot } = require("discord");

module.exports = defineBot(({ command, event, configure }) => {
  configure({
    name: "announcements",
    description: "Announcement previews from a root-level bot script",
    category: "communications"
  });

  command("announce-preview", {
    description: "Preview an announcement message",
    options: {
      title: {
        type: "string",
        description: "Announcement title",
        required: true,
      }
    }
  }, async (ctx) => {
    return {
      content: `Preview ready for ${ctx.args.title}.`,
      embeds: [{ title: ctx.args.title, description: "Announcement preview generated from a root-level bot script.", color: 0x95A5A6 }]
    };
  });

  event("ready", async (ctx) => {
    ctx.log.info("announcements bot ready", { user: ctx.me && ctx.me.username });
  });
});
