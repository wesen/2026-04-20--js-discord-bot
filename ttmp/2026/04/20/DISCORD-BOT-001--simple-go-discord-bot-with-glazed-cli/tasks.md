# Tasks

## Done

- [x] Create the ticket workspace
- [x] Draft the Go Discord bot architecture and implementation guide
- [x] Document the required Discord credentials and how to obtain them
- [x] Record a diary entry for the work completed so far
- [x] Relate the ticket documents and validate the workspace

## Planned implementation tasks

### 1. Project bootstrap
- [x] Initialize the Go module and repository layout
- [x] Add a project-level `.gitignore` to keep `.envrc`, build artifacts, and local binaries out of git
- [x] Create a safe `.envrc.example` with placeholder Discord variables

### 2. Glazed CLI foundation
- [x] Create the Glazed root command with help/logging wiring
- [x] Add `run`, `sync-commands`, and `validate-config` subcommands
- [x] Define shared command settings for output/logging/config flags
- [x] Wire the Glazed env source middleware so `.envrc`-exported variables populate command settings
- [x] Confirm CLI help text explains which values come from flags vs environment

### 3. Config and validation
- [x] Load `DISCORD_BOT_TOKEN`, `DISCORD_APPLICATION_ID`, and `DISCORD_GUILD_ID` from environment/flags
- [x] Validate required values before opening a Discord session
- [x] Add explicit error messages for missing token, missing app ID, or invalid guild ID
- [x] Keep any secrets out of logs and structured output

### 4. Discord session and commands
- [x] Create a Discord session wrapper with clean connect/disconnect behavior
- [x] Register basic event handlers for ready, interaction, and shutdown paths
- [x] Add a `/ping` slash command that returns `pong`
- [x] Add a minimal `/echo` slash command for testing interaction payloads
- [x] Implement guild-scoped command syncing for development
- [x] Leave room for global command sync later without changing the command contract

### 5. Smoke tests and review
- [x] Add a local validation checklist for `validate-config`, `sync-commands`, and `run`
- [x] Verify env loading against the `.envrc` shell environment
- [x] Run formatting and tests before each commit
- [x] Update the diary after each meaningful milestone

## Next

- [x] Start Project bootstrap and Glazed CLI foundation
- [x] Implement the Discord session and slash-command flow
- [x] Add smoke tests and validation notes
