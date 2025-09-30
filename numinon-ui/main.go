package main

import (
	"embed"
	"os"
	"os/signal"
	"syscall"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Handle OS signals in a goroutine
	go func() {
		sig := <-sigChan
		println("Received signal:", sig.String(), "- Shutting down gracefully...")

		// Trigger shutdown through the lifecycle manager
		if app.lifecycleManager != nil {
			app.lifecycleManager.ForceShutdown()
		} else {
			// Fallback if lifecycle manager isn't initialized
			os.Exit(0)
		}
	}()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "Numinon C2 Client",
		Width:  1200,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		OnBeforeClose:    app.beforeClose,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
		os.Exit(1)
	}
}
