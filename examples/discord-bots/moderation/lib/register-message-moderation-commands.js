module.exports = function registerMessageModerationCommands({ command }) {
  command("mod-fetch-message", {
    description: "Fetch one message by ID using the host moderation utilities",
    options: {
      message_id: { type: "string", description: "Target message ID", required: true },
    }
  }, async (ctx) => {
    const channelId = ctx.channel && ctx.channel.id;
    if (!channelId) {
      return { content: "This command requires a channel context.", ephemeral: true };
    }
    const message = await ctx.discord.messages.fetch(channelId, ctx.args.message_id);
    return {
      content: `Fetched ${message.id}.`,
      ephemeral: true,
      embeds: [{
        title: "Fetched message",
        description: String(message.content || "(empty)"),
        color: 0x5865F2,
      }],
    };
  });

  command("mod-pin", {
    description: "Pin a message in the current channel using the host moderation utilities",
    options: {
      message_id: { type: "string", description: "Target message ID", required: true },
    }
  }, async (ctx) => {
    const channelId = ctx.channel && ctx.channel.id;
    if (!channelId) {
      return { content: "This command requires a channel context.", ephemeral: true };
    }
    await ctx.discord.messages.pin(channelId, ctx.args.message_id);
    return { content: `Pinned ${ctx.args.message_id}.`, ephemeral: true };
  });

  command("mod-unpin", {
    description: "Unpin a message in the current channel using the host moderation utilities",
    options: {
      message_id: { type: "string", description: "Target message ID", required: true },
    }
  }, async (ctx) => {
    const channelId = ctx.channel && ctx.channel.id;
    if (!channelId) {
      return { content: "This command requires a channel context.", ephemeral: true };
    }
    await ctx.discord.messages.unpin(channelId, ctx.args.message_id);
    return { content: `Unpinned ${ctx.args.message_id}.`, ephemeral: true };
  });

  command("mod-list-pins", {
    description: "List pinned messages in the current channel using the host moderation utilities"
  }, async (ctx) => {
    const channelId = ctx.channel && ctx.channel.id;
    if (!channelId) {
      return { content: "This command requires a channel context.", ephemeral: true };
    }
    const messages = await ctx.discord.messages.listPinned(channelId);
    const description = messages.length > 0
      ? messages.map((message) => `• ${message.id} — ${String(message.content || "(empty)")}`).join("\n")
      : "No pinned messages.";
    return {
      content: `Found ${messages.length} pinned message(s).`,
      ephemeral: true,
      embeds: [{ title: "Pinned messages", description, color: 0x57F287 }],
    };
  });
};
