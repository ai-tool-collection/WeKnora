//go:build bindings

package main

import (
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
)

// Generate Wails bindings without starting the API or database.
func main() {
	app := NewApp()
	_ = wails.Run(&options.App{
		Title: "Knowledge Hub",
		Bind:  []interface{}{app},
	})
}
