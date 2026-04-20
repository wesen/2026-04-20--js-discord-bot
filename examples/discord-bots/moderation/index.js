const { defineBot } = require("discord");
const registerOverviewCommands = require("./lib/register-overview-commands");
const registerMessageModerationCommands = require("./lib/register-message-moderation-commands");
const registerChannelModerationCommands = require("./lib/register-channel-moderation-commands");
const registerMemberModerationCommands = require("./lib/register-member-moderation-commands");
const registerModerationEvents = require("./lib/register-events");

module.exports = defineBot((api) => {
  api.configure({
    name: "moderation",
    description: "Moderation helpers with embeds and ephemeral replies",
    category: "moderation"
  });

  registerOverviewCommands(api);
  registerMessageModerationCommands(api);
  registerChannelModerationCommands(api);
  registerMemberModerationCommands(api);
  registerModerationEvents(api);
});
