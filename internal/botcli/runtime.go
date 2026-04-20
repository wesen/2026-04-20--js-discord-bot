package botcli

import (
	"context"
	"encoding/json"
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
	Config            appconfig.Settings
	Bot               DiscoveredBot
	RuntimeConfig     map[string]any
	SyncOnStart       bool
	PrintParsedValues bool
	Out               io.Writer
}

var runSelectedBotsFn = runSelectedBots

func runSelectedBots(ctx context.Context, request RunRequest) error {
	if request.PrintParsedValues {
		return printRunRequest(request.Out, request)
	}
	instance, err := appbot.NewWithScript(request.Config, request.Bot.ScriptPath(), request.RuntimeConfig)
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

func printRunRequest(w io.Writer, request RunRequest) error {
	if w == nil {
		w = os.Stdout
	}
	payload := map[string]any{
		"config": map[string]any{
			"botToken":      request.Config.RedactedToken(),
			"applicationID": request.Config.ApplicationID,
			"guildID":       request.Config.GuildID,
			"publicKey":     nonEmptyMask(request.Config.PublicKey),
			"clientID":      request.Config.ClientID,
			"clientSecret":  nonEmptyMask(request.Config.ClientSecret),
		},
		"syncOnStart":   request.SyncOnStart,
		"bot":           botDebugSummary(request.Bot),
		"runtimeConfig": request.RuntimeConfig,
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(payload)
}

func botDebugSummary(bot DiscoveredBot) map[string]any {
	return map[string]any{
		"name":        bot.Name(),
		"description": bot.Description(),
		"scriptPath":  bot.ScriptPath(),
		"sourceLabel": bot.SourceLabel(),
		"commands":    bot.CommandNames(),
		"events":      bot.EventNames(),
	}
}

func nonEmptyMask(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return "***"
}
