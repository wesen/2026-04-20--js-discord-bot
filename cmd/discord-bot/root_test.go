package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRootHelpLoadsEmbeddedDocs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains string
	}{
		{
			name:     "api reference",
			args:     []string{"help", "discord-js-bot-api-reference"},
			contains: "Discord JavaScript Bot API Reference",
		},
		{
			name:     "tutorial",
			args:     []string{"help", "build-and-run-discord-js-bots"},
			contains: "Build and Run Discord JavaScript Bots",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := newRootCommand()
			require.NoError(t, err)

			var stdout bytes.Buffer
			root.SetOut(&stdout)
			root.SetErr(&stdout)
			root.SetArgs(tt.args)

			err = root.Execute()
			require.NoError(t, err)
			require.Contains(t, stdout.String(), tt.contains)
		})
	}
}
