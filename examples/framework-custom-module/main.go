package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/dop251/goja"
	noderequire "github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/go-go-goja/engine"
	"github.com/go-go-golems/discord-bot/pkg/framework"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	bot, err := framework.New(
		framework.WithCredentialsFromEnv(),
		framework.WithScript(filepath.Join("examples", "framework-custom-module", "bot", "index.js")),
		framework.WithRuntimeModuleRegistrars(appRegistrar{}),
		framework.WithSyncOnStart(true),
	)
	if err != nil {
		log.Fatalf("create framework bot with custom module: %v", err)
	}

	log.Printf("starting custom-module bot from %s", filepath.Join("examples", "framework-custom-module", "bot", "index.js"))
	if err := bot.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("run framework bot with custom module: %v", err)
	}
}

type appRegistrar struct{}

func (appRegistrar) ID() string {
	return "framework-custom-module-app"
}

func (appRegistrar) RegisterRuntimeModules(_ *engine.RuntimeModuleContext, reg *noderequire.Registry) error {
	reg.RegisterNativeModule("app", func(vm *goja.Runtime, moduleObj *goja.Object) {
		exports := moduleObj.Get("exports").(*goja.Object)
		_ = exports.Set("name", func() string { return "framework-custom-module" })
		_ = exports.Set("description", func() string { return "Custom Go module injected into a single embedded bot" })
		_ = exports.Set("greeting", func() string { return "hello from the Go-side app module" })
	})
	return nil
}
