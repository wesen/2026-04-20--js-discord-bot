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

  command("mod-add-role", {
    description: "Add a role to a guild member using the host moderation API",
    options: {
      user_id: { type: "string", description: "Target user ID", required: true },
      role_id: { type: "string", description: "Role ID to add", required: true },
    }
  }, async (ctx) => {
    const guildId = ctx.guild && ctx.guild.id;
    if (!guildId) {
      return { content: "This command must be used in a guild.", ephemeral: true };
    }
    await ctx.discord.members.addRole(guildId, ctx.args.user_id, ctx.args.role_id);
    return { content: `Added role ${ctx.args.role_id} to ${ctx.args.user_id}.`, ephemeral: true };
  });

  command("mod-timeout", {
    description: "Timeout a guild member for a number of seconds using the host moderation API",
    options: {
      user_id: { type: "string", description: "Target user ID", required: true },
      duration_seconds: { type: "integer", description: "Timeout duration in seconds", required: true },
    }
  }, async (ctx) => {
    const guildId = ctx.guild && ctx.guild.id;
    if (!guildId) {
      return { content: "This command must be used in a guild.", ephemeral: true };
    }
    await ctx.discord.members.timeout(guildId, ctx.args.user_id, { durationSeconds: ctx.args.duration_seconds });
    return { content: `Timed out ${ctx.args.user_id} for ${ctx.args.duration_seconds} seconds.`, ephemeral: true };
  });

  command("mod-kick", {
    description: "Kick a guild member using the host moderation API",
    options: {
      user_id: { type: "string", description: "Target user ID", required: true },
      reason: { type: "string", description: "Kick reason", required: false },
    }
  }, async (ctx) => {
    const guildId = ctx.guild && ctx.guild.id;
    if (!guildId) {
      return { content: "This command must be used in a guild.", ephemeral: true };
    }
    await ctx.discord.members.kick(guildId, ctx.args.user_id, { reason: ctx.args.reason || "" });
    return { content: `Kicked ${ctx.args.user_id}.`, ephemeral: true };
  });

  command("mod-ban", {
    description: "Ban a guild member using the host moderation API",
    options: {
      user_id: { type: "string", description: "Target user ID", required: true },
      reason: { type: "string", description: "Ban reason", required: false },
      delete_message_days: { type: "integer", description: "How many days of messages to delete", required: false },
    }
  }, async (ctx) => {
    const guildId = ctx.guild && ctx.guild.id;
    if (!guildId) {
      return { content: "This command must be used in a guild.", ephemeral: true };
    }
    await ctx.discord.members.ban(guildId, ctx.args.user_id, {
      reason: ctx.args.reason || "",
      deleteMessageDays: ctx.args.delete_message_days || 0,
    });
    return { content: `Banned ${ctx.args.user_id}.`, ephemeral: true };
  });

  command("mod-unban", {
    description: "Unban a guild member using the host moderation API",
    options: {
      user_id: { type: "string", description: "Target user ID", required: true },
    }
  }, async (ctx) => {
    const guildId = ctx.guild && ctx.guild.id;
    if (!guildId) {
      return { content: "This command must be used in a guild.", ephemeral: true };
    }
    await ctx.discord.members.unban(guildId, ctx.args.user_id);
    return { content: `Unbanned ${ctx.args.user_id}.`, ephemeral: true };
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
