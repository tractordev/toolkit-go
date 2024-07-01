package main

import (
	"tractor.dev/toolkit-go/desktop"
	"tractor.dev/toolkit-go/desktop/app"
	"tractor.dev/toolkit-go/desktop/window"
)

func main() {
	desktop.Start(func() {
		app.Run(app.Options{}, func() {
			window.New(window.Options{
				URL:   "https://google.com",
				Title: "Hello",
				Size: desktop.Size{
					Width:  800,
					Height: 600,
				},
				Resizable: true,
				Center:    true,
			})
		})
	})
}
