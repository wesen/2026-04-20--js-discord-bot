package jsdiscord

import (
	"testing"
)

func TestJsverbsPolyfillsAllowVerbInBotScript(t *testing.T) {
	scriptPath := writeBotScript(t, `
		const { defineBot } = require("discord");

		__verb__("status", {
			short: "Check bot status",
			output: "glaze"
		});

		module.exports = defineBot(({ command, configure }) => {
			configure({ name: "polyfill-test", description: "Tests __verb__ polyfill" });
			command("ping", async () => ({ content: "pong" }));
		});
	`)

	handle := loadTestBot(t, scriptPath)

	desc, err := handle.Describe(nil)
	if err != nil {
		t.Fatalf("describe: %v", err)
	}
	if got := desc["metadata"].(map[string]any)["name"]; got != "polyfill-test" {
		t.Fatalf("bot name = %v", got)
	}
}

func TestJsverbsPolyfillsAllowSectionInBotScript(t *testing.T) {
	scriptPath := writeBotScript(t, `
		const { defineBot } = require("discord");

		__section__("filters", {
			title: "Filters",
			fields: {
				state: { type: "choice", choices: ["open", "closed"] }
			}
		});

		module.exports = defineBot(({ configure }) => {
			configure({ name: "section-polyfill-test", description: "Tests __section__ polyfill" });
		});
	`)

	handle := loadTestBot(t, scriptPath)

	desc, err := handle.Describe(nil)
	if err != nil {
		t.Fatalf("describe: %v", err)
	}
	if got := desc["metadata"].(map[string]any)["name"]; got != "section-polyfill-test" {
		t.Fatalf("bot name = %v", got)
	}
}

func TestJsverbsPolyfillsAllowPackageInBotScript(t *testing.T) {
	scriptPath := writeBotScript(t, `
		const { defineBot } = require("discord");

		__package__({
			name: "test-pkg",
			parents: ["bots"]
		});

		module.exports = defineBot(({ configure }) => {
			configure({ name: "package-polyfill-test", description: "Tests __package__ polyfill" });
		});
	`)

	handle := loadTestBot(t, scriptPath)

	desc, err := handle.Describe(nil)
	if err != nil {
		t.Fatalf("describe: %v", err)
	}
	if got := desc["metadata"].(map[string]any)["name"]; got != "package-polyfill-test" {
		t.Fatalf("bot name = %v", got)
	}
}
