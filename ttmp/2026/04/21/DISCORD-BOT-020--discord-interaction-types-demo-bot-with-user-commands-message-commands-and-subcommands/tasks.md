# Tasks

## TODO

- [ ] Add tasks here

- [x] Download Discord application commands docs via defuddle and store in sources
- [x] Add userCommand, messageCommand, and subcommand to JS bot API runtime
- [x] Update Go descriptor and command sync to handle user/message/subcommand types
- [x] Update Go interaction dispatch for user commands, message commands, and subcommands
- [ ] Create demo bot examples/discord-bots/interaction-types with all command variants
- [x] Write detailed design and implementation guide for intern onboarding
- [x] Upload guide and source bundle to reMarkable
- [x] FIX bot.go: Add userCommands, messageCommands, subcommands fields to botDraft struct
- [x] FIX bot.go: Ensure newBotDraft initializes all command slices
- [x] FIX bot.go: Wire dispatchSubcommand into finalize() bot object
- [x] FIX host_commands.go: Map command type 'user'/'message' to discordgo ApplicationCommandType
- [x] FIX host_commands.go: Add sub_command and sub_command_group to optionTypeFromSpec
- [x] FIX host_dispatch.go: Route user context menu commands to userCommand handlers with resolved target
- [x] FIX host_dispatch.go: Route message context menu commands to messageCommand handlers with resolved target
- [x] FIX host_dispatch.go: Route subcommands to dispatchSubcommand with flattened args
- [x] VERIFY go build ./... compiles cleanly after all framework changes
- [x] CREATE demo bot: examples/discord-bots/interaction-types/index.js with hello, echo, admin kick/ban subcommands, Show Avatar userCommand, Quote Message messageCommand
- [x] TEST demo bot locally with bots run interaction-types --sync-on-start
- [x] UPDATE design doc with final implementation details and any deviations from original plan
- [x] UPDATE pkg/doc/topics/discord-js-bot-api-reference.md to document userCommand, messageCommand, subcommand APIs
- [x] UPDATE examples/discord-bots/README.md to list interaction-types bot
- [x] UPLOAD final docs to reMarkable
