module.exports = function registerChannelModerationCommands({ command }) {
  command("mod-fetch-channel", {
    description: "Fetch the current channel using the host moderation utilities"
  }, async (ctx) => {
    const channelId = ctx.channel && ctx.channel.id;
    if (!channelId) {
      return { content: "This command requires a channel context.", ephemeral: true };
    }
    const channel = await ctx.discord.channels.fetch(channelId);
    return {
      content: `Fetched ${channel.id}.`,
      ephemeral: true,
      embeds: [{
        title: "Channel info",
        description: [
          `Name: ${String(channel.name || "(unknown)")}`,
          `Topic: ${String(channel.topic || "(empty)")}`,
          `Slowmode: ${String(channel.rateLimitPerUser || 0)}s`,
        ].join("\n"),
        color: 0x5865F2,
      }],
    };
  });

  command("mod-set-topic", {
    description: "Set the current channel topic using the host moderation utilities",
    options: {
      topic: { type: "string", description: "New topic", required: true },
    }
  }, async (ctx) => {
    const channelId = ctx.channel && ctx.channel.id;
    if (!channelId) {
      return { content: "This command requires a channel context.", ephemeral: true };
    }
    await ctx.discord.channels.setTopic(channelId, ctx.args.topic);
    return { content: "Updated channel topic.", ephemeral: true };
  });

  command("mod-set-slowmode", {
    description: "Set the current channel slowmode using the host moderation utilities",
    options: {
      seconds: { type: "integer", description: "Slowmode in seconds", required: true },
    }
  }, async (ctx) => {
    const channelId = ctx.channel && ctx.channel.id;
    if (!channelId) {
      return { content: "This command requires a channel context.", ephemeral: true };
    }
    await ctx.discord.channels.setSlowmode(channelId, ctx.args.seconds);
    return { content: `Updated slowmode to ${ctx.args.seconds} second(s).`, ephemeral: true };
  });
};
