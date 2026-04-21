module.exports = function registerMessageModerationCommands({ command }) {
  command("mod-list-messages", {
    description: "List a bounded page of messages in the current channel using the host history utilities",
    options: {
      before_message_id: { type: "string", description: "List messages before this message ID", required: false },
      after_message_id: { type: "string", description: "List messages after this message ID", required: false },
      around_message_id: { type: "string", description: "List messages around this message ID", required: false },
      limit: { type: "integer", description: "Maximum messages to list", required: false },
    }
  }, async (ctx) => {
    const channelId = ctx.channel && ctx.channel.id;
    if (!channelId) {
      return { content: "This command requires a channel context.", ephemeral: true };
    }
    const messages = await ctx.discord.messages.list(channelId, {
      before: ctx.args.before_message_id || "",
      after: ctx.args.after_message_id || "",
      around: ctx.args.around_message_id || "",
      limit: ctx.args.limit || 10,
    });
    const description = messages.length > 0
      ? messages.map((message) => `• ${message.id} — ${String(message.content || "(empty)")}`).join("\n")
      : "No messages returned.";
    return {
      content: `Fetched ${messages.length} message(s).`,
      ephemeral: true,
      embeds: [{ title: "Message history", description, color: 0x5865F2 }],
    };
  });

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

  command("mod-bulk-delete", {
    description: "Bulk delete message IDs in the current channel using the host moderation utilities",
    options: {
      message_ids: { type: "string", description: "Comma-separated message IDs", required: true },
    }
  }, async (ctx) => {
    const channelId = ctx.channel && ctx.channel.id;
    if (!channelId) {
      return { content: "This command requires a channel context.", ephemeral: true };
    }
    const messageIds = String(ctx.args.message_ids || "")
      .split(",")
      .map((id) => id.trim())
      .filter(Boolean);
    await ctx.discord.messages.bulkDelete(channelId, messageIds);
    return { content: `Bulk deleted ${messageIds.length} message(s).`, ephemeral: true };
  });
};
