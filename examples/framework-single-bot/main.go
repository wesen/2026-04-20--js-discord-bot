package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/go-go-golems/discord-bot/pkg/framework"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	bot, err := framework.New(
		framework.WithCredentialsFromEnv(),
		framework.WithScript(filepath.Join("examples", "discord-bots", "unified-demo", "index.js")),
		framework.WithRuntimeConfig(map[string]any{
			"db_path": "./examples/discord-bots/unified-demo/data/demo.sqlite",
			"api_key": "local-dev-key",
		}),
		framework.WithSyncOnStart(true),
	)
	if err != nil {
		log.Fatalf("create framework bot: %v", err)
	}

	log.Printf("starting bot from %s", filepath.Join("examples", "discord-bots", "unified-demo", "index.js"))
	if err := bot.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("run framework bot: %v", err)
	}
}
