package botcli

import (
	"strings"
)

// runtimeFieldInternalName converts a CLI flag name (kebab-case) to a
// JavaScript property name (snake_case). This is the same logic used by
// jsdiscord.runtimeFieldInternalName so that config values parsed from
// CLI flags match the keys bots expect in ctx.config.
//
// Examples:
//   "db-path"       -> "db_path"
//   "APIKey"        -> "api_key"
//   "batch-size"    -> "batch_size"
//   "simple"        -> "simple"
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
