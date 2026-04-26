package botcli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/discord-bot/internal/jsdiscord"
)

const BotRepositoryFlag = "bot-repository"

type Bootstrap struct {
	Repositories []Repository
}

type Repository struct {
	Name      string
	Source    string
	SourceRef string
	RootDir   string
}

type DiscoveredBot struct {
	Repository Repository
	Descriptor *jsdiscord.BotDescriptor
}

func (b DiscoveredBot) Validate() error {
	if b.Descriptor == nil {
		return fmt.Errorf("bot descriptor is nil")
	}
	if strings.TrimSpace(b.Descriptor.Name) == "" {
		return fmt.Errorf("bot descriptor name is empty")
	}
	if strings.TrimSpace(b.Descriptor.ScriptPath) == "" {
		return fmt.Errorf("bot descriptor script path is empty")
	}
	return nil
}

func (b DiscoveredBot) Name() string {
	if b.Descriptor == nil {
		return ""
	}
	return b.Descriptor.Name
}

func (b DiscoveredBot) Description() string {
	if b.Descriptor == nil {
		return ""
	}
	return b.Descriptor.Description
}

func (b DiscoveredBot) ScriptPath() string {
	if b.Descriptor == nil {
		return ""
	}
	return b.Descriptor.ScriptPath
}

func (b DiscoveredBot) SourceLabel() string {
	if b.Descriptor == nil || strings.TrimSpace(b.Descriptor.ScriptPath) == "" {
		return b.Repository.Name
	}
	if rel, err := filepath.Rel(b.Repository.RootDir, b.Descriptor.ScriptPath); err == nil {
		return filepath.ToSlash(rel)
	}
	return b.Descriptor.ScriptPath
}

func (b DiscoveredBot) CommandNames() []string {
	if b.Descriptor == nil {
		return nil
	}
	ret := make([]string, 0, len(b.Descriptor.Commands))
	for _, command := range b.Descriptor.Commands {
		if strings.TrimSpace(command.Name) == "" {
			continue
		}
		ret = append(ret, command.Name)
	}
	return ret
}

func (b DiscoveredBot) EventNames() []string {
	if b.Descriptor == nil {
		return nil
	}
	ret := make([]string, 0, len(b.Descriptor.Events))
	for _, event := range b.Descriptor.Events {
		if strings.TrimSpace(event.Name) == "" {
			continue
		}
		ret = append(ret, event.Name)
	}
	return ret
}
