---
Title: Glazed Documentation Feedback and Env-Loading Gotcha
Ticket: DISCORD-BOT-001
Status: active
Topics:
    - backend
    - chat
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../corporate-headquarters/glazed/pkg/cli/cobra-parser.go
      Note: Parser config comments and override behavior that should be clearer
    - Path: ../../../../../../../corporate-headquarters/glazed/pkg/doc/topics/21-cmds-middlewares.md
      Note: Best home for the sharp-edge warning about custom middleware chains
    - Path: ../../../../../../../corporate-headquarters/glazed/pkg/doc/topics/24-config-files.md
      Note: Clarify distinction between env loading and config-file loading
    - Path: ../../../../../../../corporate-headquarters/glazed/pkg/doc/tutorials/05-build-first-command.md
      Note: Quick-start tutorial that currently mentions MiddlewaresFunc and should show the safe pattern
    - Path: ../../../../../../../corporate-headquarters/glazed/pkg/doc/tutorials/config-files-quickstart.md
      Note: Could include a debug/precedence example for env and config source inspection
    - Path: ../../../../../../../corporate-headquarters/glazed/pkg/doc/tutorials/migrating-from-viper-to-config-files.md
      Note: Useful migration context for env/config precedence and where docs can link back
ExternalSources: []
Summary: A maintainer-facing writeup of the Glazed env-loading confusion encountered while building the bot CLI.
LastUpdated: 2026-04-20T10:30:00-04:00
WhatFor: Give Glazed maintainers a concrete example of where the docs and API shape were easy to misread.
WhenToUse: Use when proposing documentation clarifications for Glazed CLI env parsing and middleware behavior.
---


# Glazed Documentation Feedback and Env-Loading Gotcha

## Goal

Document the specific point of confusion I hit while building the bot CLI so it can be turned into a documentation improvement for Glazed. This is not Discord-specific; the core issue is about how the Glazed Cobra parser handles environment loading when a command uses the default parser path versus a custom middleware function.

## Executive Summary

I built a CLI that relied on shell-exported environment variables from a local `.envrc` file. I expected the configuration to be loaded automatically once I set the app name/prefix, but a custom `MiddlewaresFunc` accidentally replaced the default parsing chain. The CLI compiled and the flags looked right, yet the config values were missing at runtime.

The missing documentation point is simple: the docs should make it unmistakably clear that `MiddlewaresFunc` is an override, not an additive hook, and that the built-in env loading path only stays in effect when the default parser path is used or when env loading is explicitly re-added.

## What I struggled with

I initially treated `CobraParserConfig.MiddlewaresFunc` like a place to add app-specific behavior on top of the default Glazed chain. Instead, it replaced the default middleware construction.

That meant:

- the command still built
- the help text still looked correct
- the env variables were present in the shell
- but `validate-config` still reported missing required values

The symptom looked like a missing secrets problem, which was misleading. The real issue was parser configuration shape, not shell setup.

## What would have helped

### 1) A hard warning in the Cobra parser docs

The docs should explicitly say:

> If you provide `MiddlewaresFunc`, you are responsible for including all desired sources yourself. The default chain is not merged in automatically.

A sentence like that would have stopped the mistake immediately.

### 2) A complete “env-backed CLI” example

There should be one minimal example that shows the recommended path for a CLI that reads from environment variables:

- define fields and settings
- set `AppName` for env prefixing
- do **not** override `MiddlewaresFunc` unless necessary
- show `--print-parsed-fields` or similar debug output for verification

That example should emphasize the happy path, not only the advanced customization path.

### 3) A small precedence table

A table like this would clarify how the parser behaves:

| Scenario | Effect |
| --- | --- |
| Default Cobra parser path | includes the built-in env handling when `AppName` is set |
| Custom `MiddlewaresFunc` provided | default chain is replaced |
| Need env + custom behavior | add env middleware explicitly in the custom chain |

This would make the override behavior visible at a glance.

### 4) A note in the config/docs tutorial that env discovery is separate from config discovery

The docs should separate two concepts that are easy to conflate:

- environment variable loading
- config file discovery/loading

They are related, but not the same mechanism.

## What I changed in my app after learning this

I removed the custom middleware override and let Glazed build the default env-aware parsing path through `AppName`. That fixed the config-loading issue right away.

Then I added a dedicated validation command so I could confirm env wiring before even touching the Discord gateway/runtime side of the app.

## Docs in the Glazed repo that seem most important to update

These are the places I would update first, in order:

1. `pkg/cli/cobra-parser.go`
   - The code comments around `CobraParserConfig.MiddlewaresFunc` and `AppName` should say more clearly that a custom middleware function replaces the default chain.
   - The current comments hint at extensibility, but they do not clearly spell out the override behavior.

2. `pkg/doc/topics/21-cmds-middlewares.md`
   - This is the best place for the sharp-edge warning.
   - Add a dedicated subsection like “Using a custom `MiddlewaresFunc` safely” with a complete example.
   - Explain how to preserve env loading when customizing the chain.

3. `pkg/doc/tutorials/05-build-first-command.md`
   - The tutorial already shows `CobraParserConfig` and `MiddlewaresFunc` in a quick-start context.
   - It would benefit from one canonical env-backed example that uses `AppName` and either avoids a custom middleware override or explicitly re-adds env loading.

4. `pkg/doc/topics/24-config-files.md`
   - This doc is strong on config-file precedence, but it would benefit from a clearer distinction between config-file loading and env loading in the Cobra parser path.
   - A short note near the Cobra integration section could prevent readers from assuming all configuration sources are automatically merged.

5. `pkg/doc/tutorials/config-files-quickstart.md`
   - If there is room for one more improvement, this tutorial could include a “debugging config source precedence” section with `--print-parsed-fields` and env-source visibility.

## Docs I read while debugging this

These are the Glazed docs and source files that shaped my understanding:

- `pkg/doc/topics/21-cmds-middlewares.md`
- `pkg/doc/topics/24-config-files.md`
- `pkg/doc/tutorials/05-build-first-command.md`
- `pkg/doc/tutorials/config-files-quickstart.md`
- `pkg/doc/tutorials/migrating-from-viper-to-config-files.md`
- `pkg/doc/topics/13-sections-and-values.md`
- `pkg/cli/cobra-parser.go`
- `pkg/cli/helpers.go`
- `pkg/cmds/sources/update.go`

## Suggested maintainer-facing wording

Here is the main sentence I wish had been present in the docs:

> `MiddlewaresFunc` replaces the default Cobra parsing chain. If you customize it, you must explicitly include any sources you still want, such as environment loading.

A companion example could be:

```go
cli.WithParserConfig(cli.CobraParserConfig{
    AppName: "myapp",
    // Leave MiddlewaresFunc nil if you want the default env-aware path.
})
```

And if customization is required:

```go
cli.WithParserConfig(cli.CobraParserConfig{
    AppName: "myapp",
    MiddlewaresFunc: func(parsed *values.Values, cmd *cobra.Command, args []string) ([]cmd_sources.Middleware, error) {
        return []cmd_sources.Middleware{
            cmd_sources.FromCobra(cmd, fields.WithSource("cobra")),
            cmd_sources.FromArgs(args, fields.WithSource("arguments")),
            cmd_sources.FromEnv("MYAPP", fields.WithSource("env")),
            cmd_sources.FromDefaults(fields.WithSource(fields.SourceDefaults)),
        }, nil
    },
})
```

## Why this matters

This is an easy trap for application authors:

- the API looks flexible
- the app compiles
- the help output appears correct
- but one config source silently disappears

That makes the bug feel like a missing environment secret when it is actually a parser configuration issue.

A stronger doc warning would save time for anyone building a CLI with Glazed and environment-backed settings.

## Related

- `design-doc/01-implementation-and-architecture-guide.md`
- `reference/02-discord-credentials-and-setup.md`
- `reference/01-diary.md`
- `playbook/01-local-validation-and-smoke-test-checklist.md`
