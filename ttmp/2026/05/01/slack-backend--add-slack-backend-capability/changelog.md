# Changelog

## 2026-05-01

- Initial workspace created


## 2026-05-01

Created Slack backend ticket and added API/Block Kit reference focused on preserving the existing JavaScript layer while adapting transport behavior in Go.

### Related Files

- /Users/kball/git/go-go-golems/discord-bot/ttmp/2026/05/01/slack-backend--add-slack-backend-capability/reference/01-slack-api-and-block-kit-reference.md — New Slack API and Block Kit research reference


## 2026-05-01

Recorded resolved Slack backend design decisions: preserve Discord JS naming, generate manifest, use HTTP endpoints, persist adapter state in SQLite, support one command option, and inline file exports.

### Related Files

- /Users/kball/git/go-go-golems/discord-bot/ttmp/2026/05/01/slack-backend--add-slack-backend-capability/design-doc/01-slack-backend-design-decisions.md — New decision record for Slack backend implementation
- /Users/kball/git/go-go-golems/discord-bot/ttmp/2026/05/01/slack-backend--add-slack-backend-capability/reference/01-slack-api-and-block-kit-reference.md — Updated open questions into resolved design decisions


## 2026-05-01

Step 1: Added Slack HTTP backend prototype, manifest generation, SQLite adapter state, Block Kit rendering, and CLI commands (commit f55387a).

### Related Files

- /Users/kball/git/go-go-golems/discord-bot/cmd/discord-bot/slack_commands.go — Slack CLI entry points
- /Users/kball/git/go-go-golems/discord-bot/internal/jsdiscord/slack_backend.go — Slack backend implementation
- /Users/kball/git/go-go-golems/discord-bot/ttmp/2026/05/01/slack-backend--add-slack-backend-capability/reference/02-implementation-diary.md — Implementation diary entry


## 2026-05-01

Step 2: Adjusted Slack responder so public command replies create tracked messages and later ctx.edit calls use chat.update (commit 0bb8931).

### Related Files

- /Users/kball/git/go-go-golems/discord-bot/internal/jsdiscord/slack_backend.go — Updated Slack responder routing and message identity tracking
- /Users/kball/git/go-go-golems/discord-bot/internal/jsdiscord/slack_backend_test.go — Added fake Slack API response routing test
- /Users/kball/git/go-go-golems/discord-bot/ttmp/2026/05/01/slack-backend--add-slack-backend-capability/reference/02-implementation-diary.md — Recorded Step 2


## 2026-05-01

Added Slack backend setup and smoke-test playbook covering manifest generation, app install, slack-serve, command/button/modal checks, and SQLite inspection.

### Related Files

- /Users/kball/git/go-go-golems/discord-bot/ttmp/2026/05/01/slack-backend--add-slack-backend-capability/playbook/01-slack-backend-setup-and-smoke-test.md — New repeatable Slack setup and validation playbook


## 2026-05-01

Step 3: Omitted empty usage_hint fields from generated Slack manifests so Slack accepts commands without options (commit 8c426cd).

### Related Files

- /Users/kball/git/go-go-golems/discord-bot/internal/jsdiscord/slack_backend.go — Manifest generation fix
- /Users/kball/git/go-go-golems/discord-bot/ttmp/2026/05/01/slack-backend--add-slack-backend-capability/reference/02-implementation-diary.md — Recorded Step 3


## 2026-05-01

Step 4: Added app_mentions:read to generated Slack manifests because app_mention events require that scope (commit 4d4a2a).

### Related Files

- /Users/kball/git/go-go-golems/discord-bot/internal/jsdiscord/slack_backend.go — Manifest scope fix
- /Users/kball/git/go-go-golems/discord-bot/internal/jsdiscord/slack_backend_test.go — Regression assertion for app_mentions:read
- /Users/kball/git/go-go-golems/discord-bot/ttmp/2026/05/01/slack-backend--add-slack-backend-capability/reference/02-implementation-diary.md — Recorded Step 4


## 2026-05-01

Step 5: Slack backend now creates the SQLite state parent directory before opening the database (commit 5eba553).

### Related Files

- /Users/kball/git/go-go-golems/discord-bot/internal/jsdiscord/slack_backend.go — OpenSlackStore now creates missing parent directories
- /Users/kball/git/go-go-golems/discord-bot/ttmp/2026/05/01/slack-backend--add-slack-backend-capability/reference/02-implementation-diary.md — Recorded Step 5

