# Tasks

## Done

- [x] Create the ticket workspace
- [x] Draft the architecture/design/reference docs for moving the sandbox here
- [x] Port the sandbox-style runtime-local store, builder, dispatch, and async settlement into `internal/jsdiscord`
- [x] Implement a real local `require("discord")` JavaScript bot API in this repository
- [x] Add a first live-host integration path so the Discord bot can load a JavaScript bot script
- [x] Sync slash commands from JavaScript bot metadata
- [x] Dispatch live Discord slash-command interactions into JavaScript handlers
- [x] Dispatch the Discord `ready` event into JavaScript handlers
- [x] Add tests for the local Discord JS runtime package

- [x] Add richer Discord response payload support beyond `content` and `ephemeral`
- [x] Add broader live Discord event coverage beyond `ready`

## Completed implementation details

- [x] Support embeds in slash-command and event responses
- [x] Support button/action-row components in slash-command and event responses
- [x] Support deferred interaction responses with optional ephemeral flags
- [x] Support editing deferred interaction responses from JavaScript
- [x] Support interaction follow-up messages from JavaScript
- [x] Dispatch `guildCreate` events into JavaScript handlers
- [x] Dispatch `messageCreate` events into JavaScript handlers
- [x] Update the example JS bot to demonstrate richer payloads and event handling
- [x] Add tests covering richer response payload normalization and event context wiring

## Next

- [ ] Add option validation/normalization for more Discord command option shapes
- [ ] Add a dedicated CLI or playbook for inspecting a JS bot script without opening the gateway
- [ ] Move or revert the corresponding sandbox code from `go-go-goja` once this repo is the clear source of truth
