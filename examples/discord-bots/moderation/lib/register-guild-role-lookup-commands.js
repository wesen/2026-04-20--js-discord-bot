module.exports = function registerGuildRoleLookupCommands({ command }) {
  command("mod-fetch-guild", {
    description: "Fetch the current guild using the host lookup utilities"
  }, async (ctx) => {
    const guildId = ctx.guild && ctx.guild.id;
    if (!guildId) {
      return { content: "This command must be used in a guild.", ephemeral: true };
    }
    const guild = await ctx.discord.guilds.fetch(guildId);
    return {
      content: `Fetched guild ${guild.name || guild.id}.`,
      ephemeral: true,
      embeds: [{
        title: "Guild info",
        description: [
          `ID: ${String(guild.id)}`,
          `Owner: ${String(guild.ownerID || "(unknown)")}`,
          `Members: ${String(guild.memberCount || 0)}`,
          `Verification: ${String(guild.verificationLevel || "(unknown)")}`,
        ].join("\n"),
        color: 0x5865F2,
      }],
    };
  });

  command("mod-list-roles", {
    description: "List guild roles using the host lookup utilities"
  }, async (ctx) => {
    const guildId = ctx.guild && ctx.guild.id;
    if (!guildId) {
      return { content: "This command must be used in a guild.", ephemeral: true };
    }
    const roles = await ctx.discord.roles.list(guildId);
    const preview = roles
      .slice()
      .sort((a, b) => Number(b.position || 0) - Number(a.position || 0))
      .slice(0, 10)
      .map((role) => `${String(role.name)} (${String(role.id)})`)
      .join("\n");
    return {
      content: `Found ${roles.length} role(s).`,
      ephemeral: true,
      embeds: [{
        title: "Role preview",
        description: preview || "No roles found.",
        color: 0x57F287,
      }],
    };
  });

  command("mod-fetch-role", {
    description: "Fetch one guild role using the host lookup utilities",
    options: {
      role_id: { type: "string", description: "Target role ID", required: true },
    }
  }, async (ctx) => {
    const guildId = ctx.guild && ctx.guild.id;
    if (!guildId) {
      return { content: "This command must be used in a guild.", ephemeral: true };
    }
    const role = await ctx.discord.roles.fetch(guildId, ctx.args.role_id);
    return {
      content: `Fetched role ${role.name || role.id}.`,
      ephemeral: true,
      embeds: [{
        title: "Role info",
        description: [
          `ID: ${String(role.id)}`,
          `Name: ${String(role.name || "(unknown)")}`,
          `Position: ${String(role.position || 0)}`,
          `Managed: ${String(Boolean(role.managed))}`,
          `Mentionable: ${String(Boolean(role.mentionable))}`,
        ].join("\n"),
        color: Number(role.color || 0x5865F2),
      }],
    };
  });
};
