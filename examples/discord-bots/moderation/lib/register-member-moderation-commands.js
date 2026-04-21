module.exports = function registerMemberModerationCommands({ command }) {
  command("mod-fetch-member", {
    description: "Fetch one guild member using the host lookup utilities",
    options: {
      user_id: { type: "string", description: "Target user ID", required: true },
    }
  }, async (ctx) => {
    const guildId = ctx.guild && ctx.guild.id;
    if (!guildId) {
      return { content: "This command must be used in a guild.", ephemeral: true };
    }
    const member = await ctx.discord.members.fetch(guildId, ctx.args.user_id);
    return {
      content: `Fetched member ${member.id}.`,
      ephemeral: true,
      embeds: [{
        title: "Member info",
        description: [
          `ID: ${String(member.id || ctx.args.user_id)}`,
          `Nick: ${String(member.nick || "(none)")}`,
          `Roles: ${Array.isArray(member.roles) ? member.roles.length : 0}`,
          `Pending: ${String(Boolean(member.pending))}`,
        ].join("\n"),
        color: 0x5865F2,
      }],
    };
  });

  command("mod-list-members", {
    description: "List a small page of guild members using the host lookup utilities",
    options: {
      limit: { type: "integer", description: "Maximum members to list", required: false },
      after_user_id: { type: "string", description: "Pagination cursor user ID", required: false },
    }
  }, async (ctx) => {
    const guildId = ctx.guild && ctx.guild.id;
    if (!guildId) {
      return { content: "This command must be used in a guild.", ephemeral: true };
    }
    const members = await ctx.discord.members.list(guildId, {
      after: ctx.args.after_user_id || "",
      limit: ctx.args.limit || 10,
    });
    const preview = members.slice(0, 10).map((member) => {
      const user = member.user || {};
      return `${String(member.id || "(unknown)")} — ${String(user.username || member.nick || "(unknown)")}`;
    }).join("\n");
    return {
      content: `Fetched ${members.length} member(s).`,
      ephemeral: true,
      embeds: [{
        title: "Member preview",
        description: preview || "No members returned.",
        color: 0x57F287,
      }],
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
};
