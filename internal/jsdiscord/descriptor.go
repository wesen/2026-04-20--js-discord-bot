package jsdiscord

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type BotDescriptor struct {
	Name          string
	Description   string
	ScriptPath    string
	Metadata      map[string]any
	Commands      []CommandDescriptor
	Subcommands   []SubcommandDescriptor
	Events        []EventDescriptor
	Components    []ComponentDescriptor
	Modals        []ModalDescriptor
	Autocompletes []AutocompleteDescriptor
	RunSchema     *RunSchemaDescriptor
}

// RunSchemaDescriptor describes bot startup/runtime configuration fields.
type RunSchemaDescriptor struct {
	Sections []RunSectionDescriptor
}

type RunSectionDescriptor struct {
	Slug   string
	Title  string
	Fields []RunFieldDescriptor
}

type RunFieldDescriptor struct {
	Name         string
	InternalName string
	Type         string
	Help         string
	Required     bool
	Default      any
}

type CommandDescriptor struct {
	Name        string
	Description string
	Type        string
	Spec        map[string]any
}

type SubcommandDescriptor struct {
	RootName    string
	Name        string
	Description string
	Spec        map[string]any
}

type EventDescriptor struct {
	Name string
}

type ComponentDescriptor struct {
	CustomID string
}

type ModalDescriptor struct {
	CustomID string
}

type AutocompleteDescriptor struct {
	CommandName string
	OptionName  string
}

func InspectScript(ctx context.Context, scriptPath string) (*BotDescriptor, error) {
	loaded, err := LoadBot(ctx, scriptPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = loaded.Close(context.Background()) }()
	return loaded.Descriptor, nil
}

type LoadedBot struct {
	Descriptor *BotDescriptor
	Host       *Host
}

func LoadBot(ctx context.Context, scriptPath string, opts ...HostOption) (*LoadedBot, error) {
	host, err := NewHost(ctx, scriptPath, opts...)
	if err != nil {
		return nil, err
	}
	desc, err := host.Describe(ctx)
	if err != nil {
		_ = host.Close(context.Background())
		return nil, err
	}
	descriptor, err := descriptorFromDescribe(host.scriptPath, desc)
	if err != nil {
		_ = host.Close(context.Background())
		return nil, err
	}
	return &LoadedBot{Descriptor: descriptor, Host: host}, nil
}

func (b *LoadedBot) Close(ctx context.Context) error {
	if b == nil || b.Host == nil {
		return nil
	}
	return b.Host.Close(ctx)
}

func descriptorFromDescribe(scriptPath string, desc map[string]any) (*BotDescriptor, error) {
	scriptPath = strings.TrimSpace(scriptPath)
	metadata, _ := desc["metadata"].(map[string]any)
	metadata = cloneMap(metadata)
	name := mapString(metadata, "name")
	if name == "" {
		name = fallbackBotName(scriptPath)
	}
	botDesc := &BotDescriptor{
		Name:          name,
		Description:   mapString(metadata, "description"),
		ScriptPath:    scriptPath,
		Metadata:      metadata,
		Commands:      parseCommandDescriptors(desc["commands"]),
		Subcommands:   parseSubcommandDescriptors(desc["subcommands"]),
		Events:        parseEventDescriptors(desc["events"]),
		Components:    parseComponentDescriptors(desc["components"]),
		Modals:        parseModalDescriptors(desc["modals"]),
		Autocompletes: parseAutocompleteDescriptors(desc["autocompletes"]),
		RunSchema:     parseRunSchemaDescriptor(metadata["run"]),
	}
	if botDesc.Description == "" {
		botDesc.Description = mapString(metadata, "summary")
	}
	return botDesc, nil
}

func fallbackBotName(scriptPath string) string {
	if strings.TrimSpace(scriptPath) == "" {
		return "bot"
	}
	base := filepath.Base(scriptPath)
	if strings.EqualFold(base, "index.js") {
		return filepath.Base(filepath.Dir(scriptPath))
	}
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func parseDescriptors[T any](raw any, parseFn func(map[string]any) (T, bool), lessFn func(T, T) bool) []T {
	snapshots := commandSnapshots(raw)
	ret := make([]T, 0, len(snapshots))
	for _, item := range snapshots {
		mapping, _ := item.(map[string]any)
		if len(mapping) == 0 {
			continue
		}
		desc, ok := parseFn(mapping)
		if !ok {
			continue
		}
		ret = append(ret, desc)
	}
	sort.Slice(ret, func(i, j int) bool { return lessFn(ret[i], ret[j]) })
	return ret
}

func parseCommandDescriptors(raw any) []CommandDescriptor {
	return parseDescriptors(raw, func(mapping map[string]any) (CommandDescriptor, bool) {
		spec, _ := mapping["spec"].(map[string]any)
		cmdType := ""
		if spec != nil {
			cmdType = strings.TrimSpace(fmt.Sprint(spec["type"]))
		}
		return CommandDescriptor{
			Name:        mapString(mapping, "name"),
			Description: mapString(spec, "description"),
			Type:        cmdType,
			Spec:        cloneMap(spec),
		}, true
	}, func(a, b CommandDescriptor) bool { return a.Name < b.Name })
}

func parseSubcommandDescriptors(raw any) []SubcommandDescriptor {
	return parseDescriptors(raw, func(mapping map[string]any) (SubcommandDescriptor, bool) {
		spec, _ := mapping["spec"].(map[string]any)
		return SubcommandDescriptor{
			RootName:    mapString(mapping, "rootName"),
			Name:        mapString(mapping, "name"),
			Description: mapString(spec, "description"),
			Spec:        cloneMap(spec),
		}, true
	}, func(a, b SubcommandDescriptor) bool {
		if a.RootName != b.RootName {
			return a.RootName < b.RootName
		}
		return a.Name < b.Name
	})
}

func parseEventDescriptors(raw any) []EventDescriptor {
	return parseDescriptors(raw, func(mapping map[string]any) (EventDescriptor, bool) {
		return EventDescriptor{Name: mapString(mapping, "name")}, true
	}, func(a, b EventDescriptor) bool { return a.Name < b.Name })
}

func parseComponentDescriptors(raw any) []ComponentDescriptor {
	return parseDescriptors(raw, func(mapping map[string]any) (ComponentDescriptor, bool) {
		customID := mapString(mapping, "customId")
		if customID == "" {
			return ComponentDescriptor{}, false
		}
		return ComponentDescriptor{CustomID: customID}, true
	}, func(a, b ComponentDescriptor) bool { return a.CustomID < b.CustomID })
}

func parseModalDescriptors(raw any) []ModalDescriptor {
	return parseDescriptors(raw, func(mapping map[string]any) (ModalDescriptor, bool) {
		customID := mapString(mapping, "customId")
		if customID == "" {
			return ModalDescriptor{}, false
		}
		return ModalDescriptor{CustomID: customID}, true
	}, func(a, b ModalDescriptor) bool { return a.CustomID < b.CustomID })
}

func parseAutocompleteDescriptors(raw any) []AutocompleteDescriptor {
	return parseDescriptors(raw, func(mapping map[string]any) (AutocompleteDescriptor, bool) {
		commandName := mapString(mapping, "commandName")
		optionName := mapString(mapping, "optionName")
		if commandName == "" || optionName == "" {
			return AutocompleteDescriptor{}, false
		}
		return AutocompleteDescriptor{CommandName: commandName, OptionName: optionName}, true
	}, func(a, b AutocompleteDescriptor) bool {
		if a.CommandName != b.CommandName {
			return a.CommandName < b.CommandName
		}
		return a.OptionName < b.OptionName
	})
}

func parseRunSchemaDescriptor(raw any) *RunSchemaDescriptor {
	mapping, _ := raw.(map[string]any)
	if len(mapping) == 0 {
		return nil
	}
	sections := make([]RunSectionDescriptor, 0)
	if fieldsSection := parseRunFieldSection("run", "Run settings", mapping["fields"]); fieldsSection != nil {
		sections = append(sections, *fieldsSection)
	}
	if rawSections, ok := mapping["sections"].(map[string]any); ok {
		keys := make([]string, 0, len(rawSections))
		for key := range rawSections {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			sectionMapping, _ := rawSections[key].(map[string]any)
			title := mapString(sectionMapping, "title")
			if title == "" {
				title = key
			}
			section := parseRunFieldSection(key, title, sectionMapping["fields"])
			if section != nil {
				sections = append(sections, *section)
			}
		}
	}
	if len(sections) == 0 {
		return nil
	}
	return &RunSchemaDescriptor{Sections: sections}
}

func parseRunFieldSection(slug, title string, raw any) *RunSectionDescriptor {
	mapping, _ := raw.(map[string]any)
	if len(mapping) == 0 {
		return nil
	}
	keys := make([]string, 0, len(mapping))
	for key := range mapping {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	fields := make([]RunFieldDescriptor, 0, len(keys))
	for _, key := range keys {
		fieldMapping, _ := mapping[key].(map[string]any)
		if len(fieldMapping) == 0 {
			continue
		}
		fieldType := mapString(fieldMapping, "type")
		if fieldType == "" {
			fieldType = "string"
		}
		fields = append(fields, RunFieldDescriptor{
			Name:         key,
			InternalName: runtimeFieldInternalName(key),
			Type:         fieldType,
			Help:         mapString(fieldMapping, "help"),
			Required:     mapBool(fieldMapping, "required"),
			Default:      mapAny(fieldMapping, "default"),
		})
	}
	if len(fields) == 0 {
		return nil
	}
	return &RunSectionDescriptor{Slug: slug, Title: title, Fields: fields}
}

func runtimeFieldInternalName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	var out []rune
	for i, r := range name {
		switch {
		case r == '-':
			out = append(out, '_')
		case r >= 'A' && r <= 'Z':
			if i > 0 && len(out) > 0 && out[len(out)-1] != '_' {
				out = append(out, '_')
			}
			out = append(out, r+'a'-'A')
		default:
			out = append(out, r)
		}
	}
	return strings.Trim(strings.TrimSpace(string(out)), "_")
}

func mapString(mapping map[string]any, key string) string {
	if len(mapping) == 0 {
		return ""
	}
	value, ok := mapping[key]
	if !ok || value == nil {
		return ""
	}
	text := strings.TrimSpace(fmt.Sprint(value))
	if text == "<nil>" {
		return ""
	}
	return text
}

func mapBool(mapping map[string]any, key string) bool {
	if len(mapping) == 0 {
		return false
	}
	value, ok := mapping[key].(bool)
	if !ok {
		return false
	}
	return value
}

func mapAny(mapping map[string]any, key string) any {
	if len(mapping) == 0 {
		return nil
	}
	return mapping[key]
}
