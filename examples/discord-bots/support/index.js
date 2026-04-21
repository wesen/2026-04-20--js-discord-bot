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

  command("support-fetch-thread", {
    description: "Fetch one thread using the host thread utilities",
    options: {
      thread_id: {
        type: "string",
        description: "Target thread ID",
        required: true,
      }
    }
  }, async (ctx) => {
    const thread = await ctx.discord.threads.fetch(ctx.args.thread_id);
    return {
      content: `Fetched thread ${thread.name || thread.id}.`,
      ephemeral: true,
      embeds: [{
        title: "Thread info",
        description: [
          `ID: ${String(thread.id)}`,
          `Parent: ${String(thread.parentID || "(unknown)")}`,
          `Archived: ${String(Boolean(thread.archived))}`,
          `Locked: ${String(Boolean(thread.locked))}`,
        ].join("\n"),
        color: 0x5865F2,
      }]
    };
  });

  command("support-join-thread", {
    description: "Join a thread using the host thread utilities",
    options: {
      thread_id: {
        type: "string",
        description: "Target thread ID",
        required: true,
      }
    }
  }, async (ctx) => {
    await ctx.discord.threads.join(ctx.args.thread_id);
    return { content: `Joined thread ${ctx.args.thread_id}.`, ephemeral: true };
  });

  command("support-leave-thread", {
    description: "Leave a thread using the host thread utilities",
    options: {
      thread_id: {
        type: "string",
        description: "Target thread ID",
        required: true,
      }
    }
  }, async (ctx) => {
    await ctx.discord.threads.leave(ctx.args.thread_id);
    return { content: `Left thread ${ctx.args.thread_id}.`, ephemeral: true };
  });

  command("support-start-thread", {
    description: "Start a support thread from the current channel using the host thread utilities",
    options: {
      name: {
        type: "string",
        description: "Thread name",
        required: true,
      },
      source_message_id: {
        type: "string",
        description: "Optional source message ID for message-thread creation",
        required: false,
      },
      type: {
        type: "string",
        description: "Thread type: public, private, or news",
        required: false,
      },
      auto_archive_duration: {
        type: "integer",
        description: "Auto archive duration in minutes",
        required: false,
      }
    }
  }, async (ctx) => {
    const channelId = ctx.channel && ctx.channel.id;
    if (!channelId) {
      return { content: "This command requires a channel context.", ephemeral: true };
    }
    const thread = await ctx.discord.threads.start(channelId, {
      name: ctx.args.name,
      messageId: ctx.args.source_message_id || "",
      type: ctx.args.type || "public",
      autoArchiveDuration: ctx.args.auto_archive_duration || 1440,
    });
    return {
      content: `Started thread ${thread.name || thread.id}.`,
      ephemeral: true,
      embeds: [{
        title: "Thread created",
        description: [
          `ID: ${String(thread.id)}`,
          `Parent: ${String(thread.parentID || channelId)}`,
          `Archived: ${String(Boolean(thread.archived))}`,
        ].join("\n"),
        color: 0x57F287,
      }]
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
