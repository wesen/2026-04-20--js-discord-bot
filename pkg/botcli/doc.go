// Package botcli exposes the optional repo-driven Discord bot command layer.
//
// Use this package when a downstream Cobra application should discover named
// bot scripts from one or more repositories and mount a `bots` subtree that
// supports inventory, inspection, ordinary jsverbs, and host-managed bot runs.
//
// The public entrypoints are:
//   - BuildBootstrap(...) to resolve repositories from raw argv / env / defaults
//   - NewBotsCommand(...) to mount the repo-driven `bots` command tree
//
// Runtime customization has a deliberate "smallest hook first" shape:
//   - Use WithAppName(...) when only the dynamic env prefix should change.
//   - Use WithRuntimeModuleRegistrars(...) when bot scripts or ordinary jsverbs
//     just need extra Go-native require() modules such as `require("app")`.
//   - Use WithRuntimeFactory(...) only when ordinary jsverb runtime creation
//     itself must change, for example custom module roots, require behavior,
//     builder configuration, or a custom engine/runtime lifecycle.
//
// If the custom runtime behavior must also affect discovery and host-managed bot
// runs, implement HostOptionsProvider on the runtime factory so the same choice
// contributes jsdiscord host options as well.
package botcli
