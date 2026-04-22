package botcli

import (
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
)

type testSection struct {
	name        string
	slug        string
	description string
	prefix      string
	definitions *fields.Definitions
}

func (t *testSection) GetDefinitions() *fields.Definitions { return t.definitions }
func (t *testSection) GetName() string                     { return t.name }
func (t *testSection) GetDescription() string              { return t.description }
func (t *testSection) GetPrefix() string                   { return t.prefix }
func (t *testSection) GetSlug() string                     { return t.slug }

func createTestSection(slug, name string, defs ...*fields.Definition) values.Section {
	definitions := fields.NewDefinitions()
	for _, def := range defs {
		definitions.Set(def.Name, def)
	}
	return &testSection{
		name:        name,
		slug:        slug,
		definitions: definitions,
	}
}

func createTestSectionValues(section values.Section, vals map[string]interface{}) *values.SectionValues {
	sectionValues, err := values.NewSectionValues(section)
	if err != nil {
		panic(err)
	}
	for key, value := range vals {
		definition, ok := section.GetDefinitions().Get(key)
		if !ok {
			panic("definition " + key + " missing")
		}
		parsed := &fields.FieldValue{Definition: definition}
		if err := parsed.Update(value); err != nil {
			panic(err)
		}
		sectionValues.Fields.Set(key, parsed)
	}
	return sectionValues
}

func TestBuildRuntimeConfig(t *testing.T) {
	section := createTestSection("default", "Default",
		fields.New("db-path", fields.TypeString),
		fields.New("api-key", fields.TypeString),
		fields.New("batch-size", fields.TypeInteger),
	)
	sectionVals := createTestSectionValues(section, map[string]interface{}{
		"db-path":    "./data/test.db",
		"api-key":    "secret123",
		"batch-size": 42,
	})

	parsed := values.New()
	parsed.Set("default", sectionVals)

	config := buildRuntimeConfig(parsed)

	if got := config["db_path"]; got != "./data/test.db" {
		t.Errorf("db_path = %v, want ./data/test.db", got)
	}
	if got := config["api_key"]; got != "secret123" {
		t.Errorf("api_key = %v, want secret123", got)
	}
	if got := config["batch_size"]; got != 42 {
		t.Errorf("batch_size = %v, want 42", got)
	}
}

func TestBuildRuntimeConfigEmpty(t *testing.T) {
	config := buildRuntimeConfig(nil)
	if len(config) != 0 {
		t.Fatalf("config = %d, want 0", len(config))
	}

	config = buildRuntimeConfig(values.New())
	if len(config) != 0 {
		t.Fatalf("config = %d, want 0", len(config))
	}
}
