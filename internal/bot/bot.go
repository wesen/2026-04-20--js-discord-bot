package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"

	appconfig "github.com/manuel/wesen/2026-04-20--js-discord-bot/internal/config"
)

// Bot wraps the Discord session and bot behavior.
type Bot struct {
	cfg     appconfig.Settings
	session *discordgo.Session
}

// New creates a Discord session and wires bot handlers.
func New(cfg appconfig.Settings) (*Bot, error) {
	session, err := discordgo.New("Bot " + strings.TrimSpace(cfg.BotToken))
	if err != nil {
		return nil, fmt.Errorf("create discord session: %w", err)
	}

	session.Identify.Intents = discordgo.IntentsGuilds

	b := &Bot{
		cfg:     cfg,
		session: session,
	}

	session.AddHandler(b.handleReady)
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
	if b.session == nil {
		return nil
	}
	if err := b.session.Close(); err != nil {
		return fmt.Errorf("close discord session: %w", err)
	}
	return nil
}

// SyncCommands registers the slash commands for the configured scope.
func (b *Bot) SyncCommands() ([]*discordgo.ApplicationCommand, error) {
	commands := b.applicationCommands()
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

func (b *Bot) applicationCommands() []*discordgo.ApplicationCommand {
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
	}
}

func (b *Bot) handleReady(_ *discordgo.Session, ready *discordgo.Ready) {
	log.Info().
		Str("user", ready.User.Username).
		Str("user_id", ready.User.ID).
		Msg("discord bot connected")
}

func (b *Bot) handleInteractionCreate(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	if interaction.Type != discordgo.InteractionApplicationCommand {
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
