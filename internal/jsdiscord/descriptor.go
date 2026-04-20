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

func LoadBot(ctx context.Context, scriptPath string) (*LoadedBot, error) {
	host, err := NewHost(ctx, scriptPath)
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

func parseCommandDescriptors(raw any) []CommandDescriptor {
	snapshots := commandSnapshots(raw)
	ret := make([]CommandDescriptor, 0, len(snapshots))
	for _, item := range snapshots {
		mapping, _ := item.(map[string]any)
		if len(mapping) == 0 {
			continue
		}
		spec, _ := mapping["spec"].(map[string]any)
		ret = append(ret, CommandDescriptor{
			Name:        mapString(mapping, "name"),
			Description: mapString(spec, "description"),
			Spec:        cloneMap(spec),
		})
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].Name < ret[j].Name })
	return ret
}

func parseEventDescriptors(raw any) []EventDescriptor {
	snapshots := commandSnapshots(raw)
	ret := make([]EventDescriptor, 0, len(snapshots))
	for _, item := range snapshots {
		mapping, _ := item.(map[string]any)
		if len(mapping) == 0 {
			continue
		}
		ret = append(ret, EventDescriptor{Name: mapString(mapping, "name")})
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].Name < ret[j].Name })
	return ret
}

func parseComponentDescriptors(raw any) []ComponentDescriptor {
	snapshots := commandSnapshots(raw)
	ret := make([]ComponentDescriptor, 0, len(snapshots))
	for _, item := range snapshots {
		mapping, _ := item.(map[string]any)
		if len(mapping) == 0 {
			continue
		}
		customID := mapString(mapping, "customId")
		if customID == "" {
			continue
		}
		ret = append(ret, ComponentDescriptor{CustomID: customID})
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].CustomID < ret[j].CustomID })
	return ret
}

func parseModalDescriptors(raw any) []ModalDescriptor {
	snapshots := commandSnapshots(raw)
	ret := make([]ModalDescriptor, 0, len(snapshots))
	for _, item := range snapshots {
		mapping, _ := item.(map[string]any)
		if len(mapping) == 0 {
			continue
		}
		customID := mapString(mapping, "customId")
		if customID == "" {
			continue
		}
		ret = append(ret, ModalDescriptor{CustomID: customID})
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].CustomID < ret[j].CustomID })
	return ret
}

func parseAutocompleteDescriptors(raw any) []AutocompleteDescriptor {
	snapshots := commandSnapshots(raw)
	ret := make([]AutocompleteDescriptor, 0, len(snapshots))
	for _, item := range snapshots {
		mapping, _ := item.(map[string]any)
		if len(mapping) == 0 {
			continue
		}
		commandName := mapString(mapping, "commandName")
		optionName := mapString(mapping, "optionName")
		if commandName == "" || optionName == "" {
			continue
		}
		ret = append(ret, AutocompleteDescriptor{CommandName: commandName, OptionName: optionName})
	}
	sort.Slice(ret, func(i, j int) bool {
		if ret[i].CommandName != ret[j].CommandName {
			return ret[i].CommandName < ret[j].CommandName
		}
		return ret[i].OptionName < ret[j].OptionName
	})
	return ret
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
