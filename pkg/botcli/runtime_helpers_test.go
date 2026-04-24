package botcli

import (
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/stretchr/testify/require"
)

func TestBuildRuntimeConfigExcludesHostManagedAndGlobalFields(t *testing.T) {
	defaultSection, err := schema.NewSection(schema.DefaultSlug, "Arguments")
	require.NoError(t, err)
	defaultSection.AddFields(
		fields.New("bot-token", fields.TypeString),
		fields.New("application-id", fields.TypeString),
		fields.New("guild-id", fields.TypeString),
		fields.New("sync-on-start", fields.TypeBool),
		fields.New("db-path", fields.TypeString),
		fields.New("api-key", fields.TypeString),
	)
	defaultValues, err := values.NewSectionValues(defaultSection,
		values.WithFieldValue("bot-token", "super-secret-token"),
		values.WithFieldValue("application-id", "app-123"),
		values.WithFieldValue("guild-id", "guild-456"),
		values.WithFieldValue("sync-on-start", true),
		values.WithFieldValue("db-path", "./knowledge.sqlite"),
		values.WithFieldValue("api-key", "service-key"),
	)
	require.NoError(t, err)

	globalSection, err := schema.NewSection(schema.GlobalDefaultSlug, "Global")
	require.NoError(t, err)
	globalSection.AddFields(
		fields.New("config-file", fields.TypeString),
		fields.New("print-schema", fields.TypeBool),
	)
	globalValues, err := values.NewSectionValues(globalSection,
		values.WithFieldValue("config-file", "/tmp/config.yaml"),
		values.WithFieldValue("print-schema", true),
	)
	require.NoError(t, err)

	parsedValues := values.New(
		values.WithSectionValues(schema.DefaultSlug, defaultValues),
		values.WithSectionValues(schema.GlobalDefaultSlug, globalValues),
	)

	runtimeConfig := buildRuntimeConfig(parsedValues)

	require.Equal(t, map[string]any{
		"db_path": "./knowledge.sqlite",
		"api_key": "service-key",
	}, runtimeConfig)
	require.NotContains(t, runtimeConfig, "bot_token")
	require.NotContains(t, runtimeConfig, "application_id")
	require.NotContains(t, runtimeConfig, "guild_id")
	require.NotContains(t, runtimeConfig, "sync_on_start")
	require.NotContains(t, runtimeConfig, "config_file")
	require.NotContains(t, runtimeConfig, "print_schema")
}
