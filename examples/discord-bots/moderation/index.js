const { defineBot } = require("discord");

module.exports = defineBot(({ command, event, configure }) => {
  configure({
    name: "moderation",
    description: "Moderation helpers with embeds and ephemeral replies",
    category: "moderation"
  });

  command("mod-summary", {
    description: "Show moderation summary for a channel"
  }, async () => {
    return {
      content: "Moderation summary prepared.",
      ephemeral: true,
      embeds: [{
        title: "Moderation Summary",
        description: "No new incidents detected.",
        color: 0xED4245,
      }],
      components: [{
        type: "actionRow",
        components: [{
          type: "button",
          style: "secondary",
          label: "Acknowledge",
          customId: "ack-mod-summary"
        }]
      }]
    };
  });

  command("mod-guidelines", {
    description: "Show moderation guidelines"
  }, async () => {
    return {
      content: "See the moderation guidelines below.",
      embeds: [{ title: "Guidelines", description: "Be calm, clear, and consistent.", color: 0xF1C40F }]
    };
  });

  event("messageCreate", async (ctx) => {
    const content = (ctx.message && ctx.message.content || "").trim();
    if (content === "!modping") {
      await ctx.reply({ content: "Moderation bot message trigger received." });
    }
  });

  event("messageUpdate", async (ctx) => {
    const before = String(ctx.before && ctx.before.content || "").trim();
    const after = String(ctx.message && ctx.message.content || "").trim();
    if (before === after || after === "") {
      return null;
    }
    ctx.log.info("moderation observed message edit", {
      messageId: ctx.message && ctx.message.id,
      before,
      after,
      channelId: ctx.channel && ctx.channel.id,
    });
    return null;
  });

  event("messageDelete", async (ctx) => {
    ctx.log.info("moderation observed message delete", {
      messageId: ctx.message && ctx.message.id,
      cachedContent: ctx.before && ctx.before.content,
      channelId: ctx.channel && ctx.channel.id,
    });
    return null;
  });

  event("reactionAdd", async (ctx) => {
    ctx.log.info("moderation observed reaction add", {
      messageId: ctx.reaction && ctx.reaction.messageId,
      emoji: ctx.reaction && ctx.reaction.emoji && ctx.reaction.emoji.name,
      userId: ctx.user && ctx.user.id,
    });
    return null;
  });

  event("reactionRemove", async (ctx) => {
    ctx.log.info("moderation observed reaction remove", {
      messageId: ctx.reaction && ctx.reaction.messageId,
      emoji: ctx.reaction && ctx.reaction.emoji && ctx.reaction.emoji.name,
      userId: ctx.user && ctx.user.id,
    });
    return null;
  });

  event("guildMemberAdd", async (ctx) => {
    ctx.log.info("moderation observed member join", {
      userId: ctx.user && ctx.user.id,
      guildId: ctx.guild && ctx.guild.id,
      roles: ctx.member && ctx.member.roles,
    });
    return null;
  });

  event("guildMemberUpdate", async (ctx) => {
    ctx.log.info("moderation observed member update", {
      userId: ctx.user && ctx.user.id,
      guildId: ctx.guild && ctx.guild.id,
      beforeRoles: ctx.before && ctx.before.roles,
      afterRoles: ctx.member && ctx.member.roles,
    });
    return null;
  });

  event("guildMemberRemove", async (ctx) => {
    ctx.log.info("moderation observed member leave", {
      userId: ctx.user && ctx.user.id,
      guildId: ctx.guild && ctx.guild.id,
    });
    return null;
  });
});
