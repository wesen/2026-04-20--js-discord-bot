package jsdiscord

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type MultiHost struct {
	bots            []*LoadedBot
	botByName       map[string]*LoadedBot
	botByCommand    map[string]*LoadedBot
	selectedScripts []string
}

func NewMultiHost(ctx context.Context, scriptPaths []string) (*MultiHost, error) {
	cleanPaths := normalizeScriptPaths(scriptPaths)
	if len(cleanPaths) == 0 {
		return nil, fmt.Errorf("no bot scripts selected")
	}
	ret := &MultiHost{
		bots:            []*LoadedBot{},
		botByName:       map[string]*LoadedBot{},
		botByCommand:    map[string]*LoadedBot{},
		selectedScripts: cleanPaths,
	}
	for _, scriptPath := range cleanPaths {
		loaded, err := LoadBot(ctx, scriptPath)
		if err != nil {
			_ = ret.Close(context.Background())
			return nil, err
		}
		name := strings.TrimSpace(loaded.Descriptor.Name)
		if name == "" {
			_ = loaded.Close(context.Background())
			_ = ret.Close(context.Background())
			return nil, fmt.Errorf("bot script %s has empty descriptor name", scriptPath)
		}
		if previous, ok := ret.botByName[name]; ok {
			_ = loaded.Close(context.Background())
			_ = ret.Close(context.Background())
			return nil, fmt.Errorf("duplicate bot name %q from %s and %s", name, previous.Descriptor.ScriptPath, loaded.Descriptor.ScriptPath)
		}
		for _, command := range loaded.Descriptor.Commands {
			commandName := strings.TrimSpace(command.Name)
			if commandName == "" {
				continue
			}
			if previous, ok := ret.botByCommand[commandName]; ok {
				_ = loaded.Close(context.Background())
				_ = ret.Close(context.Background())
				return nil, fmt.Errorf("duplicate slash command %q from bots %s and %s", commandName, previous.Descriptor.Name, loaded.Descriptor.Name)
			}
			ret.botByCommand[commandName] = loaded
		}
		ret.botByName[name] = loaded
		ret.bots = append(ret.bots, loaded)
	}
	sort.Slice(ret.bots, func(i, j int) bool { return ret.bots[i].Descriptor.Name < ret.bots[j].Descriptor.Name })
	return ret, nil
}

func normalizeScriptPaths(paths []string) []string {
	seen := map[string]struct{}{}
	ret := make([]string, 0, len(paths))
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		ret = append(ret, path)
	}
	sort.Strings(ret)
	return ret
}

func (m *MultiHost) SelectedScripts() []string {
	if m == nil {
		return nil
	}
	return append([]string(nil), m.selectedScripts...)
}

func (m *MultiHost) Descriptors() []*BotDescriptor {
	if m == nil {
		return nil
	}
	ret := make([]*BotDescriptor, 0, len(m.bots))
	for _, bot := range m.bots {
		if bot != nil && bot.Descriptor != nil {
			ret = append(ret, bot.Descriptor)
		}
	}
	return ret
}

func (m *MultiHost) Close(ctx context.Context) error {
	if m == nil {
		return nil
	}
	var retErr error
	for _, bot := range m.bots {
		if bot == nil {
			continue
		}
		if err := bot.Close(ctx); err != nil {
			retErr = errors.Join(retErr, err)
		}
	}
	return retErr
}

func (m *MultiHost) ApplicationCommands(ctx context.Context) ([]*discordgo.ApplicationCommand, error) {
	if m == nil {
		return nil, nil
	}
	ret := []*discordgo.ApplicationCommand{}
	for _, bot := range m.bots {
		commands, err := bot.Host.ApplicationCommands(ctx)
		if err != nil {
			return nil, err
		}
		ret = append(ret, commands...)
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].Name < ret[j].Name })
	return ret, nil
}

func (m *MultiHost) DispatchReady(ctx context.Context, session *discordgo.Session, ready *discordgo.Ready) error {
	return m.dispatchEvent(ctx, "ready", func(bot *LoadedBot) error {
		return bot.Host.DispatchReady(ctx, session, ready)
	})
}

func (m *MultiHost) DispatchGuildCreate(ctx context.Context, session *discordgo.Session, guild *discordgo.GuildCreate) error {
	return m.dispatchEvent(ctx, "guildCreate", func(bot *LoadedBot) error {
		return bot.Host.DispatchGuildCreate(ctx, session, guild)
	})
}

func (m *MultiHost) DispatchMessageCreate(ctx context.Context, session *discordgo.Session, message *discordgo.MessageCreate) error {
	return m.dispatchEvent(ctx, "messageCreate", func(bot *LoadedBot) error {
		return bot.Host.DispatchMessageCreate(ctx, session, message)
	})
}

func (m *MultiHost) DispatchInteraction(ctx context.Context, session *discordgo.Session, interaction *discordgo.InteractionCreate) error {
	if m == nil || interaction == nil {
		return nil
	}
	data := interaction.ApplicationCommandData()
	bot := m.botByCommand[data.Name]
	if bot == nil {
		return fmt.Errorf("no loaded bot owns slash command %q", data.Name)
	}
	return bot.Host.DispatchInteraction(ctx, session, interaction)
}

func (m *MultiHost) dispatchEvent(ctx context.Context, eventName string, fn func(*LoadedBot) error) error {
	if m == nil {
		return nil
	}
	var retErr error
	for _, bot := range m.bots {
		if bot == nil || !botHandlesEvent(bot, eventName) {
			continue
		}
		if err := fn(bot); err != nil {
			retErr = errors.Join(retErr, err)
		}
	}
	return retErr
}

func botHandlesEvent(bot *LoadedBot, eventName string) bool {
	if bot == nil || bot.Descriptor == nil {
		return false
	}
	eventName = strings.TrimSpace(eventName)
	for _, event := range bot.Descriptor.Events {
		if strings.TrimSpace(event.Name) == eventName {
			return true
		}
	}
	return false
}
