package main

import (
	"embed"
	"fmt"
	"log"
	"runtime/debug"

	app "db-desktop/backend/handler"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

// safeRun 包装函数，添加 panic 恢复机制
func safeRun() {
	defer func() {
		if r := recover(); r != nil {
			// 打印 panic 信息和堆栈
			fmt.Printf("🚨 PANIC RECOVERED: %v\n", r)
			fmt.Printf("📚 Stack trace:\n%s\n", debug.Stack())
			log.Fatalf("Application crashed due to panic: %v", r)
		}
	}()

	runApp()
}

// runApp 实际的应用程序逻辑
func runApp() {
	// Create an instance of the app structure
	appInstance := app.NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "db-desktop",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        appInstance.Startup,
		Bind: []interface{}{
			appInstance,
		},
	})

	if err != nil {
		log.Fatal("Error:", err.Error())
	}
}

func main() {
	safeRun()
}
