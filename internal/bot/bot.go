package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"

	appconfig "github.com/manuel/wesen/2026-04-20--js-discord-bot/internal/config"
	"github.com/manuel/wesen/2026-04-20--js-discord-bot/internal/jsdiscord"
)

// Bot wraps the Discord session and bot behavior.
type Bot struct {
	cfg     appconfig.Settings
	session *discordgo.Session
	jsHost  *jsdiscord.MultiHost
}

// New creates a Discord session and wires bot handlers using settings-derived scripts.
func New(cfg appconfig.Settings) (*Bot, error) {
	scripts := []string{}
	if strings.TrimSpace(cfg.BotScript) != "" {
		scripts = append(scripts, strings.TrimSpace(cfg.BotScript))
	}
	return NewWithScripts(cfg, scripts)
}

// NewWithScripts creates a Discord session and wires bot handlers for one or more explicit bot scripts.
func NewWithScripts(cfg appconfig.Settings, scripts []string) (*Bot, error) {
	session, err := discordgo.New("Bot " + strings.TrimSpace(cfg.BotToken))
	if err != nil {
		return nil, fmt.Errorf("create discord session: %w", err)
	}

	session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	var jsHost *jsdiscord.MultiHost
	if len(scripts) > 0 {
		jsHost, err = jsdiscord.NewMultiHost(context.Background(), scripts)
		if err != nil {
			return nil, fmt.Errorf("load javascript bot scripts: %w", err)
		}
	}

	b := &Bot{
		cfg:     cfg,
		session: session,
		jsHost:  jsHost,
	}

	session.AddHandler(b.handleReady)
	session.AddHandler(b.handleGuildCreate)
	session.AddHandler(b.handleMessageCreate)
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

	return created, nil
}

func (b *Bot) applicationCommands() ([]*discordgo.ApplicationCommand, error) {
	if b.jsHost != nil {
		return b.jsHost.ApplicationCommands(context.Background())
	}
	return []*discordgo.ApplicationCommand{
		{
			Name:        "ping",
			Description: "Reply with pong",
		},
		{
			Name:        "echo",
			Description: "Echo text back to you",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "text",
					Description: "Text to echo back",
					Required:    true,
				},
			},
		},
	}, nil
}

func (b *Bot) handleReady(session *discordgo.Session, ready *discordgo.Ready) {
	log.Info().
		Str("user", ready.User.Username).
		Str("user_id", ready.User.ID).
		Str("bot_script", strings.TrimSpace(b.cfg.BotScript)).
		Msg("discord bot connected")

	if b.jsHost != nil {
		if err := b.jsHost.DispatchReady(context.Background(), session, ready); err != nil {
			log.Error().Err(err).Msg("failed to dispatch ready event to javascript bots")
		}
	}
}

func (b *Bot) handleGuildCreate(session *discordgo.Session, guild *discordgo.GuildCreate) {
	if b.jsHost != nil {
		if err := b.jsHost.DispatchGuildCreate(context.Background(), session, guild); err != nil {
			log.Error().Err(err).Msg("failed to dispatch guildCreate event to javascript bots")
		}
	}
}

func (b *Bot) handleMessageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message == nil || message.Message == nil || message.Author == nil || message.Author.Bot {
		return
	}
	if b.jsHost != nil {
		if err := b.jsHost.DispatchMessageCreate(context.Background(), session, message); err != nil {
			log.Error().Err(err).Msg("failed to dispatch messageCreate event to javascript bots")
		}
	}
}

func (b *Bot) handleInteractionCreate(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if interaction.Type != discordgo.InteractionApplicationCommand {
		return
	}

	if b.jsHost != nil {
		if err := b.jsHost.DispatchInteraction(context.Background(), session, interaction); err != nil {
			log.Error().Err(err).Msg("failed to dispatch interaction to javascript bots")
		}
		return
	}

	data := interaction.ApplicationCommandData()
	switch data.Name {
	case "ping":
		if err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "pong"},
		}); err != nil {
			log.Error().Err(err).Msg("failed to respond to ping command")
		}
	case "echo":
		text := ""
		if len(data.Options) > 0 {
			text = fmt.Sprint(data.Options[0].Value)
		}
		if strings.TrimSpace(text) == "" {
			text = "(empty)"
		}
		if err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: text},
		}); err != nil {
			log.Error().Err(err).Msg("failed to respond to echo command")
		}
	default:
		if err := session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "unknown command"},
		}); err != nil {
			log.Error().Err(err).Msg("failed to respond to unknown command")
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
