package botcli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	appbot "github.com/manuel/wesen/2026-04-20--js-discord-bot/internal/bot"
	appconfig "github.com/manuel/wesen/2026-04-20--js-discord-bot/internal/config"
)

type RunRequest struct {
	Config      appconfig.Settings
	Bots        []DiscoveredBot
	SyncOnStart bool
	Out         io.Writer
}

var runSelectedBotsFn = runSelectedBots

func runSelectedBots(ctx context.Context, request RunRequest) error {
	scripts := make([]string, 0, len(request.Bots))
	for _, bot := range request.Bots {
		scripts = append(scripts, bot.ScriptPath())
	}
	instance, err := appbot.NewWithScripts(request.Config, scripts)
	if err != nil {
		return err
	}
	if request.SyncOnStart {
		commands, err := instance.SyncCommands()
		if err != nil {
			_ = instance.Close()
			return err
		}
		if request.Out != nil {
			_, _ = fmt.Fprintf(request.Out, "synced %d commands\n", len(commands))
		}
	}
	return instance.Run(ctx)
}

func runContext(parent context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	return ctx, func() {
		stop()
		cancel()
	}
}

func settingsFromValues(botToken, applicationID, guildID, publicKey, clientID, clientSecret string) appconfig.Settings {
	return appconfig.Settings{
		BotToken:      strings.TrimSpace(botToken),
		ApplicationID: strings.TrimSpace(applicationID),
		GuildID:       strings.TrimSpace(guildID),
		PublicKey:     strings.TrimSpace(publicKey),
		ClientID:      strings.TrimSpace(clientID),
		ClientSecret:  strings.TrimSpace(clientSecret),
		BotScript:     "",
	}
}
