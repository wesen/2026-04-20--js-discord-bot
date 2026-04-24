# Tasks

## Research deliverables

- [x] Create docmgr ticket `DISCORD-BOT-023`
- [x] Write the primary design / implementation guide for Discord helper verbs and jsverbs live debugging
- [x] Create and update the investigation diary
- [x] Relate key files, update changelog, and validate with `docmgr doctor`
- [x] Upload the ticket bundle to reMarkable and verify the remote listing

## Recommended future implementation phases

- [ ] Add a lazy top-level `discord-bot verbs` subtree
- [ ] Add shared repository discovery for both bots and verbs (`--repository` or equivalent unified flag)
- [ ] Add a helper-verb runtime factory with a custom invoker
- [ ] Add low-level read-only Discord inspection verbs through `require("discord-cli")`
- [ ] Add one bot-simulation probe module for `dispatchCommand` / `dispatchComponent` / `dispatchModal`
- [ ] Add writable helper verbs under `verbs-rw/` with dry-run / write-safety policy and operator docs
