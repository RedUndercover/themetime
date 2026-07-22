package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
)

//go:embed all:frontend/dist
var assets embed.FS

const applicationID = "io.github.themetime.ThemeTime"

func main() {
	app := &App{}
	tray := newTrayController(app)
	startTray, stopTray := tray.externalLoop()
	startTray()
	defer stopTray()
	err := wails.Run(&options.App{
		Title:     "ThemeTime",
		Width:     1440,
		Height:    900,
		MinWidth:  980,
		MinHeight: 680,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour:  options.NewRGBA(5, 14, 25, 255),
		HideWindowOnClose: true,
		OnStartup:         app.startup,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: applicationID,
			OnSecondInstanceLaunch: func(options.SecondInstanceData) {
				app.showWindow()
			},
		},
		Linux: &linux.Options{
			Icon:        themeTimeIconPNG(),
			ProgramName: applicationID,
		},
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
