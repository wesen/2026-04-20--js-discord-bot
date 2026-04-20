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

## Next

- [ ] Add richer Discord response payload support beyond `content` and `ephemeral`
- [ ] Add option validation/normalization for more Discord command option shapes
- [ ] Add a dedicated CLI or playbook for inspecting a JS bot script without opening the gateway
- [ ] Move or revert the corresponding sandbox code from `go-go-goja` once this repo is the clear source of truth
