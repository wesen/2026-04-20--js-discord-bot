// Package framework provides a simple embedding API for running a single Discord bot
// with a JavaScript authoring runtime inside any Go application.
//
// The main entrypoint is New(), which accepts functional options:
//
//	bot, err := framework.New(
//	    framework.WithCredentialsFromEnv(),
//	    framework.WithScript("./my-bot/index.js"),
//	    framework.WithSyncOnStart(true),
//	)
//	bot.Run(ctx)
//
// For custom native modules, use WithRuntimeModuleRegistrars.
package framework

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-go-golems/go-go-goja/engine"
	appbot "github.com/go-go-golems/discord-bot/internal/bot"
	appconfig "github.com/go-go-golems/discord-bot/internal/config"
	"github.com/go-go-golems/discord-bot/internal/jsdiscord"
)

// Credentials holds the explicit Discord settings needed to run one bot.
type Credentials struct {
	BotToken      string
	ApplicationID string
	GuildID       string
	PublicKey     string
	ClientID      string
	ClientSecret  string
}

// Option configures the public single-bot framework constructor.
type Option func(*Config) error

// Config describes the simple single-bot embedding path.
type Config struct {
	Credentials             Credentials
	ScriptPath              string
	RuntimeConfig           map[string]any
	SyncOnStart             bool
	RuntimeModuleRegistrars []engine.RuntimeModuleRegistrar
}

// Bot is the public single-bot wrapper around the internal runtime.
type Bot struct {
	cfg   Config
	inner *appbot.Bot
}

// New creates one explicit bot instance without any repository scanning.
func New(opts ...Option) (*Bot, error) {
	cfg := Config{
		RuntimeConfig: map[string]any{},
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}
	cfg.ScriptPath = strings.TrimSpace(cfg.ScriptPath)
	if cfg.ScriptPath == "" {
		return nil, fmt.Errorf("framework script path is required; use framework.WithScript(...) or pass --bot-script to the standalone CLI")
	}

	settings := appconfig.Settings{
		BotToken:      strings.TrimSpace(cfg.Credentials.BotToken),
		ApplicationID: strings.TrimSpace(cfg.Credentials.ApplicationID),
		GuildID:       strings.TrimSpace(cfg.Credentials.GuildID),
		PublicKey:     strings.TrimSpace(cfg.Credentials.PublicKey),
		ClientID:      strings.TrimSpace(cfg.Credentials.ClientID),
		ClientSecret:  strings.TrimSpace(cfg.Credentials.ClientSecret),
		BotScript:     cfg.ScriptPath,
	}
	if err := settings.Validate(); err != nil {
		return nil, err
	}

	hostOpts := []jsdiscord.HostOption{}
	if len(cfg.RuntimeModuleRegistrars) > 0 {
		hostOpts = append(hostOpts, jsdiscord.WithRuntimeModuleRegistrars(cfg.RuntimeModuleRegistrars...))
	}

	inner, err := appbot.NewWithScript(settings, cfg.ScriptPath, cloneMap(cfg.RuntimeConfig), hostOpts...)
	if err != nil {
		return nil, err
	}
	return &Bot{cfg: cfg, inner: inner}, nil
}

// WithScript configures the explicit JavaScript bot script path.
func WithScript(path string) Option {
	return func(cfg *Config) error {
		cfg.ScriptPath = strings.TrimSpace(path)
		return nil
	}
}

// WithCredentials configures explicit Discord credentials.
func WithCredentials(credentials Credentials) Option {
	return func(cfg *Config) error {
		cfg.Credentials = credentials
		return nil
	}
}

// WithCredentialsFromEnv loads Discord credentials from the same env vars as the CLI.
func WithCredentialsFromEnv() Option {
	return func(cfg *Config) error {
		cfg.Credentials = Credentials{
			BotToken:      os.Getenv("DISCORD_BOT_TOKEN"),
			ApplicationID: os.Getenv("DISCORD_APPLICATION_ID"),
			GuildID:       os.Getenv("DISCORD_GUILD_ID"),
			PublicKey:     os.Getenv("DISCORD_PUBLIC_KEY"),
			ClientID:      os.Getenv("DISCORD_CLIENT_ID"),
			ClientSecret:  os.Getenv("DISCORD_CLIENT_SECRET"),
		}
		return nil
	}
}

// WithRuntimeConfig injects arbitrary runtime config values into ctx.config.
func WithRuntimeConfig(runtimeConfig map[string]any) Option {
	return func(cfg *Config) error {
		cfg.RuntimeConfig = cloneMap(runtimeConfig)
		return nil
	}
}

// WithSyncOnStart enables a command sync before opening the gateway session.
func WithSyncOnStart(enabled bool) Option {
	return func(cfg *Config) error {
		cfg.SyncOnStart = enabled
		return nil
	}
}

// WithRuntimeModuleRegistrars appends custom per-runtime native module registrars.
func WithRuntimeModuleRegistrars(registrars ...engine.RuntimeModuleRegistrar) Option {
	return func(cfg *Config) error {
		for i, registrar := range registrars {
			if registrar == nil {
				return fmt.Errorf("runtime module registrar at index %d is nil", i)
			}
		}
		cfg.RuntimeModuleRegistrars = append(cfg.RuntimeModuleRegistrars, registrars...)
		return nil
	}
}

// Open optionally syncs commands, then opens the Discord gateway session.
func (b *Bot) Open() error {
	if b == nil || b.inner == nil {
		return fmt.Errorf("framework bot is not initialized")
	}
	if b.cfg.SyncOnStart {
		if _, err := b.inner.SyncCommands(); err != nil {
			return err
		}
	}
	return b.inner.Open()
}

// SyncCommands manually syncs the bot's application commands.
func (b *Bot) SyncCommands() error {
	if b == nil || b.inner == nil {
		return fmt.Errorf("framework bot is not initialized")
	}
	_, err := b.inner.SyncCommands()
	return err
}

// Run opens the session and blocks until the context is canceled.
func (b *Bot) Run(ctx context.Context) error {
	if err := b.Open(); err != nil {
		return err
	}
	defer func() { _ = b.Close() }()
	<-ctx.Done()
	return nil
}

// Close shuts down the bot runtime and Discord session.
func (b *Bot) Close() error {
	if b == nil || b.inner == nil {
		return nil
	}
	return b.inner.Close()
}

func cloneMap(input map[string]any) map[string]any {
	if len(input) == 0 {
		return map[string]any{}
	}
	ret := make(map[string]any, len(input))
	for k, v := range input {
		ret[k] = v
	}
	return ret
}
