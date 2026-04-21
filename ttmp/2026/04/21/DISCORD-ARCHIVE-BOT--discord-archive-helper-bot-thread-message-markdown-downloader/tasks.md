# Tasks

## TODO

- [ ] Add tasks here

- [ ] Read and understand Discord API concepts (bots, intents, permissions)
- [ ] Set up bot application in Discord Developer Portal
- [ ] Implement message fetching with pagination (channels + threads)
- [ ] Implement markdown rendering for messages, embeds, attachments, reactions
- [ ] Build CLI interface for channel/thread selection and output path
- [ ] Add file download support (attachments + CDN)
- [ ] Write tests and validation for markdown output
- [ ] Document usage and archive conventions
- [x] Implement bot scaffold: create examples/discord-bots/archive-helper/ with index.js and lib/ directory
- [x] Implement fetchAllMessages() helper with before_message_id support and pagination loop
- [x] Implement markdown rendering: renderMessage(), renderArchive(), discordToMarkdown(), sanitize()
- [x] Implement /archive-channel slash command with limit and before_message_id options
- [x] Implement "Archive Thread" messageCommand (right-click any message in thread)
- [x] Add configure() with default_limit runtime config field
- [x] Test bot locally: run with --sync-on-start, verify commands appear in Discord
- [x] Write archive-helper README with usage examples and screenshots
