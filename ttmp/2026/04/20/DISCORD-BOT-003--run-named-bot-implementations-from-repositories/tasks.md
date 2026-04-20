# Tasks

## Planned implementation tasks

### 1. Ticket and design package
- [x] Create the ticket workspace
- [x] Write the implementation/design guide for the bot repository runner model
- [x] Write a reference doc for CLI usage and example repositories
- [x] Start and maintain a detailed diary

### 2. Discovery model
- [x] Define what counts as a discoverable bot implementation in a repository
- [x] Implement repository scanning for bot scripts (`index.js` packages and root-level `.js` files)
- [x] Load bot metadata from `configure(...)` / `describe()` rather than jsverbs function scanning
- [x] Derive stable bot names and descriptions
- [x] Reject duplicate bot names across repositories

### 3. Runtime model
- [x] Add a multi-bot composition layer that can hold one or more JS bot hosts
- [x] Route slash commands to the owning bot by command name
- [x] Fan out non-command events (`ready`, `guildCreate`, `messageCreate`) to all loaded bots
- [x] Reject duplicate slash-command names across selected bots
- [x] Add a host constructor path for loading multiple bot scripts at once

### 4. CLI surface
- [x] Replace the current verb-oriented `internal/botcli` discovery model with bot-oriented discovery
- [x] Implement `discord-bot bots list`
- [x] Implement `discord-bot bots help <bot>`
- [x] Implement `discord-bot bots run <bot...>` for one or more named bots
- [x] Add config flags/env loading for `bots run` so it can start the real Discord host
- [x] Add an optional `--sync-on-start` flag for selected named bots

### 5. Example bot repositories
- [x] Add a multi-bot example repository under `examples/discord-bots`
- [x] Add at least one bot that uses relative `require()` helpers
- [x] Add at least one bot that exercises deferred/edit/follow-up interaction flows
- [x] Add at least one bot that exercises embeds/components
- [x] Add at least one bot that exercises `messageCreate` and/or `guildCreate`
- [x] Add README/playbook notes explaining how to list/help/run the example bots

### 6. Tests and validation
- [x] Add tests for repository discovery and duplicate bot-name rejection
- [x] Add tests for duplicate slash-command rejection across multiple selected bots
- [x] Add tests for `bots list` output
- [x] Add tests for `bots help <bot>` output
- [x] Add tests for `bots run <bot...>` resolution and runner selection without opening Discord
- [x] Re-run focused and full repository validation

## Done

- [x] Create the ticket workspace
- [x] Write the implementation/design guide for the bot repository runner model
- [x] Write a reference doc for CLI usage and example repositories
- [x] Start and maintain a detailed diary
