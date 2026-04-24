package botcli

import (
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
)

func runtimeFieldInternalName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	var out []rune
	var prevLower bool
	for i, r := range name {
		switch {
		case r == '-':
			if len(out) > 0 && out[len(out)-1] != '_' {
				out = append(out, '_')
			}
			prevLower = false
		case r >= 'A' && r <= 'Z':
			if i > 0 && prevLower && len(out) > 0 && out[len(out)-1] != '_' {
				out = append(out, '_')
			}
			out = append(out, r+'a'-'A')
			prevLower = false
		default:
			out = append(out, r)
			prevLower = true
		}
	}
	return strings.Trim(strings.TrimSpace(string(out)), "_")
}

func boolField(parsedValues *values.Values, sectionSlug, fieldName string) bool {
	if parsedValues == nil {
		return false
	}
	fv, ok := parsedValues.GetField(sectionSlug, fieldName)
	if !ok || fv == nil || fv.Value == nil {
		return false
	}
	value, ok := fv.Value.(bool)
	if !ok {
		return false
	}
	return value
}

func buildRuntimeConfig(parsedValues *values.Values) map[string]any {
	ret := map[string]any{}
	if parsedValues == nil {
		return ret
	}
	parsedValues.ForEach(func(slug string, sectionVals *values.SectionValues) {
		if slug != schema.DefaultSlug || sectionVals == nil || sectionVals.Fields == nil {
			return
		}
		sectionVals.Fields.ForEach(func(fieldName string, fv *fields.FieldValue) {
			if fv == nil || fv.Definition == nil || isHostManagedRuntimeField(fieldName) {
				return
			}
			configKey := runtimeFieldInternalName(fieldName)
			ret[configKey] = fv.Value
		})
	})
	return ret
}

func isHostManagedRuntimeField(name string) bool {
	switch strings.TrimSpace(name) {
	case "bot-token", "application-id", "guild-id", "sync-on-start":
		return true
	default:
		return false
	}
}
