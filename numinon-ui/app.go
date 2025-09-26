package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// GetSystemInfo returns basic system information
type SystemInfo struct {
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	HostName string `json:"hostname"`
	Time     string `json:"current_time"`
}

func (a *App) GetSystemInfo() SystemInfo {
	hostname, _ := os.Hostname()

	return SystemInfo{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		HostName: hostname,
		Time:     time.Now().Format("15:04:05"),
	}
}
