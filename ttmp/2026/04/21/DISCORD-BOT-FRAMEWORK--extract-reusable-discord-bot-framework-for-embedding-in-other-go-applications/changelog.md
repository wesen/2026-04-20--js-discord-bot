# Changelog

## 2026-04-21

- Initial workspace created


## 2026-04-21

Created framework extraction design document with three options (Interface-based, Functional Options, Hybrid). Recommended Option C (Hybrid). Uploaded to reMarkable at /ai/2026/04/22/DISCORD-BOT-FRAMEWORK.

### Related Files

- /home/manuel/code/wesen/2026-04-20--js-discord-bot/ttmp/2026/04/21/DISCORD-BOT-FRAMEWORK--extract-reusable-discord-bot-framework-for-embedding-in-other-go-applications/design-doc/01-framework-extraction-design-and-implementation-guide.md — Design document with three options


## 2026-04-22

Added jsverbs integration and RuntimeFactory override to design document. Analyzed loupedeck codebase for reference implementation pattern. Added Repository, RepositoryDiscovery, VerbRegistry, and RuntimeFactory concepts. Re-uploaded to reMarkable as v2.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/go-go-goja/pkg/jsverbs — Reference for the scan/build/invoke pipeline
- /home/manuel/code/wesen/corporate-headquarters/loupedeck/cmd/loupedeck/cmds/verbs/bootstrap.go — Reference for repository discovery pattern
- /home/manuel/code/wesen/corporate-headquarters/loupedeck/runtime/js/runtime.go — Reference for custom runtime factory pattern

## 2026-04-22

Started implementation Track A for the framework split by improving the standalone single-bot path. Added `--sync-on-start` to `discord-bot run`, wired it to sync commands before opening the gateway session, added root help coverage, and refined the framework ticket task list to treat the single-bot path as a first-class track alongside public `botcli` extraction.

### Related Files

- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/cmd/discord-bot/commands.go — Standalone `run` command now supports `--sync-on-start`
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/cmd/discord-bot/root_test.go — Root help regression test for `--sync-on-start`
- /home/manuel/workspaces/2026-04-22/discord-bot-framework/2026-04-20--js-discord-bot/ttmp/2026/04/21/DISCORD-BOT-FRAMEWORK--extract-reusable-discord-bot-framework-for-embedding-in-other-go-applications/tasks.md — Track A/Track B split with first single-bot subtask checked off

