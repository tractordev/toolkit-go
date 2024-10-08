package main

import (
	"fmt"
	"time"

	"tractor.dev/toolkit-go/desktop"
	"tractor.dev/toolkit-go/desktop/app"
	"tractor.dev/toolkit-go/desktop/window"
)

func Test(w *window.Window) {
	opWithWait := func(op func(), wait time.Duration, message string) {
		op()
		fmt.Println(message)
		time.Sleep(wait)
	}
	fmt.Println("Starting Test...")
	// Wait for the window to load
	time.Sleep(5 * time.Second)

	duration := 2 * time.Second
	opWithWait(func() { w.SetVisible(false) }, duration, "SetVisible(false)...")
	opWithWait(func() { fmt.Print(w.IsVisible(), " ") }, duration, "IsVisible()...")
	opWithWait(func() { w.SetVisible(true) }, duration, "SetVisible(true)...")
	opWithWait(func() { fmt.Print(w.IsVisible(), " ") }, duration, "IsVisible()...")
	opWithWait(func() { w.SetMaximized(true) }, duration, "SetMaximized(true)...")
	opWithWait(func() { w.SetMaximized(false) }, duration, "SetMaximized(false)...")
	opWithWait(func() { w.SetFullscreen(true) }, duration, "SetFullscreen(true)...")
	opWithWait(func() { w.SetFullscreen(false) }, duration, "SetFullscreen(false)...")

	// `false`` condition is not functional on GNOME (intended behavior)
	// comment out the following line if you are using GNOME to continue the test
	// or manually bring up the window after minimizing it
	opWithWait(func() { w.SetMinimized(true) }, duration, "SetMinimized(true)...")
	opWithWait(func() { w.SetMinimized(false) }, duration, "SetMinimized(false)...")

	// size
	opWithWait(func() { fmt.Print(w.GetInnerSize(), " ") }, duration, "GetInnerSize()...")
	opWithWait(func() { fmt.Print(w.GetOuterSize(), " ") }, duration, "GetOuterSize()...")
	opWithWait(func() { w.SetSize(desktop.Size{Width: 600, Height: 600}) }, duration, "SetSize(600, 600)...")
	opWithWait(func() { fmt.Print(w.GetInnerSize(), " ") }, duration, "GetInnerSize()...")
	opWithWait(func() { fmt.Print(w.GetOuterSize(), " ") }, duration, "GetOuterSize()...")

	// position
	opWithWait(func() { fmt.Print(w.GetOuterPosition(), " ") }, duration, "GetOuterPosition()...")
	opWithWait(func() { w.SetPosition(desktop.Position{X: 100, Y: 100}) }, duration, "SetPosition(100, 100)...")
	opWithWait(func() { fmt.Print(w.GetOuterPosition(), " ") }, duration, "GetOuterPosition()...")

	opWithWait(func() { w.SetTitle("HelloTest") }, duration, "SetTitle('HelloTest')...")

	// always on top
	opWithWait(func() { w.SetAlwaysOnTop(true) }, duration * 3, "SetAlwaysOnTop(true) [Try focusing on another window!]...")
	opWithWait(func() { w.SetAlwaysOnTop(false) }, duration * 3, "SetAlwaysOnTop(false) [Try focusing on another window!]...")

	// Has issues, working on it #8
	// opWithWait(func() { w.SetResizable(false) }, duration * 3, "SetResizable(false) [Try resizing the window!]...")
	// opWithWait(func() { w.SetResizable(true) }, duration * 3, "SetResizable(true) [Try resizing the window!]...")

	opWithWait(func() { w.SetMinSize(desktop.Size{Width: 400, Height: 400}) }, duration * 3, "SetMinSize(400, 400) [Try resizing the window!]...")
	opWithWait(func() { w.SetMinSize(desktop.Size{Width: 0, Height: 0}) }, duration * 3, "SetMinSize(undone) [Try resizing the window!]...")

	opWithWait(func() { w.SetMaxSize(desktop.Size{Width: 1000, Height: 1000}) }, duration * 3, "SetMaxSize(1000, 1000) [Try resizing the window!]...")
	opWithWait(func() { w.SetMaxSize(desktop.Size{Width: 0, Height: 0}) }, duration * 3, "SetMaxSize(undone) [Try resizing the window!]...")

	// focus (not sure what the test case would be)
	opWithWait(func() { w.Focus() }, duration, "Focus()...")

	// unload and load
	opWithWait(func() { desktop.Dispatch(func() { w.Reload() }) }, duration, "Reload()...")

	fmt.Println("Test Done!")
}

func main() {
	desktop.Start(func() {
		app.Run(app.Options{}, func() {
			w := window.New(window.Options{
				URL:   "https://google.com",
				Title: "Hello",
				Size: desktop.Size{
					Width:  800,
					Height: 800,
				},
				Visible:     true,
				Center:      false,
				Transparent: true,
				Resizable:   true,
			})
			// Unload and Load the window here
			w.Reload()

			go Test(w)
		})
	})

}