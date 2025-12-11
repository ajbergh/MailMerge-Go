/*
MailMerge Go - Outlook Email Merge Application

This is the main entry point for the MailMerge Go application.
The application is built using Wails v2, which provides a bridge between
Go backend services and a React-based frontend.

Features:
  - Import contacts from CSV and Excel files
  - Compose personalized emails with merge fields
  - Send bulk emails through Microsoft Outlook
  - Track sending progress in real-time
  - Export send logs to CSV

Architecture:
  - Backend: Go with COM automation for Outlook integration
  - Frontend: React with TypeScript and Vite
  - Framework: Wails v2 for native desktop packaging

Copyright (c) 2025. All rights reserved.
*/
package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

// main initializes and runs the Wails application.
// It configures the application window, binds the backend App struct
// to the frontend, and starts the event loop.
func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	// The App struct is bound to the frontend, making all its exported
	// methods available as JavaScript functions
	err := wails.Run(&options.App{
		Title:     "MailMerge Go - Outlook Email Merge",
		Width:     1280,
		Height:    800,
		MinWidth:  1024,
		MinHeight: 700,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
