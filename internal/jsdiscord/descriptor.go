package jsdiscord

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type BotDescriptor struct {
	Name        string
	Description string
	ScriptPath  string
	Metadata    map[string]any
	Commands    []CommandDescriptor
	Events      []EventDescriptor
}

type CommandDescriptor struct {
	Name        string
	Description string
	Spec        map[string]any
}

type EventDescriptor struct {
	Name string
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
		Name:        name,
		Description: mapString(metadata, "description"),
		ScriptPath:  scriptPath,
		Metadata:    metadata,
		Commands:    parseCommandDescriptors(desc["commands"]),
		Events:      parseEventDescriptors(desc["events"]),
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
