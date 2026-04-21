package botcli

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/manuel/wesen/2026-04-20--js-discord-bot/internal/jsdiscord"
)

func printRunSchema(w io.Writer, bot DiscoveredBot) error {
	if w == nil || bot.Descriptor == nil || bot.Descriptor.RunSchema == nil || len(bot.Descriptor.RunSchema.Sections) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(w, "Run config:"); err != nil {
		return err
	}
	for _, section := range bot.Descriptor.RunSchema.Sections {
		heading := section.Title
		if strings.TrimSpace(heading) == "" {
			heading = section.Slug
		}
		if _, err := fmt.Fprintf(w, "  [%s]\n", heading); err != nil {
			return err
		}
		fields_ := append([]jsdiscord.RunFieldDescriptor(nil), section.Fields...)
		sort.Slice(fields_, func(i, j int) bool { return fields_[i].Name < fields_[j].Name })
		for _, field := range fields_ {
			line := fmt.Sprintf("    - %s (--%s) <%s>", field.Name, strings.ReplaceAll(field.InternalName, "_", "-"), field.Type)
			if field.Required {
				line += " required"
			}
			if field.Help != "" {
				line += " — " + field.Help
			}
			if field.Default != nil {
				line += fmt.Sprintf(" (default: %v)", field.Default)
			}
			if _, err := fmt.Fprintln(w, line); err != nil {
				return err
			}
		}
	}
	return nil
}
