package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"

	appconfig "github.com/go-go-golems/discord-bot/internal/config"
	"github.com/go-go-golems/discord-bot/internal/jsdiscord"
)

// Bot wraps the Discord session and bot behavior.
type Bot struct {
	cfg     appconfig.Settings
	session *discordgo.Session
	jsHost  *jsdiscord.Host
}

// New creates a Discord session and wires bot handlers using settings-derived script selection.
func New(cfg appconfig.Settings) (*Bot, error) {
	return NewWithScript(cfg, strings.TrimSpace(cfg.BotScript), nil)
}

// NewWithScript creates a Discord session and wires bot handlers for one explicit bot script.
func NewWithScript(cfg appconfig.Settings, script string, runtimeConfig map[string]any, hostOpts ...jsdiscord.HostOption) (*Bot, error) {
	session, err := discordgo.New("Bot " + strings.TrimSpace(cfg.BotToken))
	if err != nil {
		return nil, fmt.Errorf("create discord session: %w", err)
	}

	session.Identify.Intents = discordgo.IntentsGuilds

	script = strings.TrimSpace(script)
	if script == "" {
		return nil, fmt.Errorf("javascript bot script is required; use discord-bot bots <bot> run or pass --bot-script")
	}

	loaded, err := jsdiscord.LoadBot(context.Background(), script, hostOpts...)
	if err != nil {
		return nil, fmt.Errorf("load javascript bot script: %w", err)
	}
	jsHost := loaded.Host
	jsHost.SetRuntimeConfig(runtimeConfig)
	if loaded.Descriptor != nil {
		session.Identify.Intents = intentsForDescriptor(loaded.Descriptor)

		log.Info().
			Str("bot", loaded.Descriptor.Name).
			Str("script", loaded.Descriptor.ScriptPath).
			Strs("commands", commandNames(loaded.Descriptor.Commands)).
			Strs("events", eventNames(loaded.Descriptor.Events)).
			Msg("loaded javascript bot implementation")
	}

	b := &Bot{
		cfg:     cfg,
		session: session,
		jsHost:  jsHost,
	}

	session.AddHandler(b.handleReady)
	session.AddHandler(b.handleGuildCreate)
	session.AddHandler(b.handleGuildMemberAdd)
	session.AddHandler(b.handleGuildMemberUpdate)
	session.AddHandler(b.handleGuildMemberRemove)
	session.AddHandler(b.handleMessageCreate)
	session.AddHandler(b.handleMessageUpdate)
	session.AddHandler(b.handleMessageDelete)
	session.AddHandler(b.handleReactionAdd)
	session.AddHandler(b.handleReactionRemove)
	session.AddHandler(b.handleInteractionCreate)

	return b, nil
}

// Open connects the bot to Discord.
func (b *Bot) Open() error {
	if err := b.session.Open(); err != nil {
		return fmt.Errorf("open discord session: %w", err)
	}
	return nil
}

// Close disconnects the bot from Discord.
func (b *Bot) Close() error {
	var retErr error
	if b.session != nil {
		if err := b.session.Close(); err != nil {
			retErr = fmt.Errorf("close discord session: %w", err)
		}
	}
	if b.jsHost != nil {
		if err := b.jsHost.Close(context.Background()); err != nil && retErr == nil {
			retErr = fmt.Errorf("close javascript bot host: %w", err)
		}
	}
	return retErr
}

// SyncCommands registers the slash commands for the configured scope.
func (b *Bot) SyncCommands() ([]*discordgo.ApplicationCommand, error) {
	commands, err := b.applicationCommands()
	if err != nil {
		return nil, err
	}
	guildID := strings.TrimSpace(b.cfg.GuildID)

	created, err := b.session.ApplicationCommandBulkOverwrite(b.cfg.ApplicationID, guildID, commands)
	if err != nil {
		if guildID == "" {
			return nil, fmt.Errorf("sync global commands: %w", err)
		}
		return nil, fmt.Errorf("sync guild commands for %s: %w", guildID, err)
	}

	commandNames := make([]string, 0, len(created))
	for _, command := range created {
		if command == nil {
			continue
		}
		commandNames = append(commandNames, command.Name)
	}
	log.Info().
		Str("scope", syncScopeLabel(guildID)).
		Strs("commands", commandNames).
		Msg("synced discord application commands")

	return created, nil
}

func (b *Bot) applicationCommands() ([]*discordgo.ApplicationCommand, error) {
	if b.jsHost == nil {
		return nil, fmt.Errorf("javascript bot host is not configured")
	}
	return b.jsHost.ApplicationCommands(context.Background())
}

func commandNames(commands []jsdiscord.CommandDescriptor) []string {
	ret := make([]string, 0, len(commands))
	for _, command := range commands {
		if strings.TrimSpace(command.Name) == "" {
			continue
		}
		ret = append(ret, command.Name)
	}
	return ret
}

func eventNames(events []jsdiscord.EventDescriptor) []string {
	ret := make([]string, 0, len(events))
	for _, event := range events {
		if strings.TrimSpace(event.Name) == "" {
			continue
		}
		ret = append(ret, event.Name)
	}
	return ret
}

func intentsForDescriptor(desc *jsdiscord.BotDescriptor) discordgo.Intent {
	intents := discordgo.IntentsGuilds
	if desc == nil {
		return intents
	}
	for _, event := range desc.Events {
		switch strings.TrimSpace(event.Name) {
		case "guildMemberAdd", "guildMemberUpdate", "guildMemberRemove":
			intents |= discordgo.IntentsGuildMembers
		case "messageCreate", "messageUpdate":
			intents |= discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent
		case "messageDelete":
			intents |= discordgo.IntentsGuildMessages
		case "reactionAdd", "reactionRemove":
			intents |= discordgo.IntentsGuildMessages | discordgo.IntentsGuildMessageReactions
		}
	}
	return intents
}

func syncScopeLabel(guildID string) string {
	guildID = strings.TrimSpace(guildID)
	if guildID == "" {
		return "global"
	}
	return "guild:" + guildID
}

func (b *Bot) handleReady(session *discordgo.Session, ready *discordgo.Ready) {
	log.Info().
		Str("user", ready.User.Username).
		Str("user_id", ready.User.ID).
		Str("bot_script", strings.TrimSpace(b.cfg.BotScript)).
		Msg("discord bot connected")

	if b.jsHost != nil {
		if err := b.jsHost.DispatchReady(context.Background(), session, ready); err != nil {
			log.Error().Err(err).Msg("failed to dispatch ready event to javascript bot")
		}
	}
}

func (b *Bot) handleGuildCreate(session *discordgo.Session, guild *discordgo.GuildCreate) {
	if b.jsHost != nil {
		if err := b.jsHost.DispatchGuildCreate(context.Background(), session, guild); err != nil {
			log.Error().Err(err).Msg("failed to dispatch guildCreate event to javascript bot")
		}
	}
}

func (b *Bot) handleGuildMemberAdd(session *discordgo.Session, member *discordgo.GuildMemberAdd) {
	if member == nil {
		return
	}
	if member.User != nil && member.User.Bot {
		return
	}
	if b.jsHost != nil {
		if err := b.jsHost.DispatchGuildMemberAdd(context.Background(), session, member); err != nil {
			log.Error().Err(err).Msg("failed to dispatch guildMemberAdd event to javascript bot")
		}
	}
}

func (b *Bot) handleGuildMemberUpdate(session *discordgo.Session, member *discordgo.GuildMemberUpdate) {
	if member == nil {
		return
	}
	if member.User != nil && member.User.Bot {
		return
	}
	if b.jsHost != nil {
		if err := b.jsHost.DispatchGuildMemberUpdate(context.Background(), session, member); err != nil {
			log.Error().Err(err).Msg("failed to dispatch guildMemberUpdate event to javascript bot")
		}
	}
}

func (b *Bot) handleGuildMemberRemove(session *discordgo.Session, member *discordgo.GuildMemberRemove) {
	if member == nil {
		return
	}
	if member.User != nil && member.User.Bot {
		return
	}
	if b.jsHost != nil {
		if err := b.jsHost.DispatchGuildMemberRemove(context.Background(), session, member); err != nil {
			log.Error().Err(err).Msg("failed to dispatch guildMemberRemove event to javascript bot")
		}
	}
}

func (b *Bot) handleMessageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message == nil || message.Message == nil || message.Author == nil || message.Author.Bot {
		return
	}
	if b.jsHost != nil {
		if err := b.jsHost.DispatchMessageCreate(context.Background(), session, message); err != nil {
			log.Error().Err(err).Msg("failed to dispatch messageCreate event to javascript bot")
		}
	}
}

func (b *Bot) handleMessageUpdate(session *discordgo.Session, message *discordgo.MessageUpdate) {
	if message == nil {
		return
	}
	if message.Author != nil && message.Author.Bot {
		return
	}
	if b.jsHost != nil {
		if err := b.jsHost.DispatchMessageUpdate(context.Background(), session, message); err != nil {
			log.Error().Err(err).Msg("failed to dispatch messageUpdate event to javascript bot")
		}
	}
}

func (b *Bot) handleMessageDelete(session *discordgo.Session, message *discordgo.MessageDelete) {
	if message == nil {
		return
	}
	if message.Author != nil && message.Author.Bot {
		return
	}
	if b.jsHost != nil {
		if err := b.jsHost.DispatchMessageDelete(context.Background(), session, message); err != nil {
			log.Error().Err(err).Msg("failed to dispatch messageDelete event to javascript bot")
		}
	}
}

func (b *Bot) handleReactionAdd(session *discordgo.Session, reaction *discordgo.MessageReactionAdd) {
	if reaction == nil {
		return
	}
	if b.jsHost != nil {
		if err := b.jsHost.DispatchReactionAdd(context.Background(), session, reaction); err != nil {
			log.Error().Err(err).Msg("failed to dispatch reactionAdd event to javascript bot")
		}
	}
}

func (b *Bot) handleReactionRemove(session *discordgo.Session, reaction *discordgo.MessageReactionRemove) {
	if reaction == nil {
		return
	}
	if b.jsHost != nil {
		if err := b.jsHost.DispatchReactionRemove(context.Background(), session, reaction); err != nil {
			log.Error().Err(err).Msg("failed to dispatch reactionRemove event to javascript bot")
		}
	}
}

func (b *Bot) handleInteractionCreate(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if b.jsHost != nil {
		if err := b.jsHost.DispatchInteraction(context.Background(), session, interaction); err != nil {
			log.Error().Err(err).Msg("failed to dispatch interaction to javascript bot")
		}
	}
}

// Run blocks until the context is canceled.
func (b *Bot) Run(ctx context.Context) error {
	if err := b.Open(); err != nil {
		return err
	}
	defer func() {
		_ = b.Close()
	}()

	<-ctx.Done()
	return nil
}
