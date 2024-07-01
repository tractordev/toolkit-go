package app

import (
	"context"
	"log"

	"tractor.dev/toolkit-go/desktop"
)

var DefaultApp *App
var DefaultIdentifier = "dev.tractor.App"

func Run(opts Options, didFinish func()) *App {
	if opts.Identifier == "" {
		opts.Identifier = DefaultIdentifier
	}
	DefaultApp = &App{app: app{
		Options: opts,
	}}
	DefaultApp.Run(didFinish)
	return DefaultApp
}

type Options struct {
	Identifier      string
	Agent           bool // app should not terminate when last window closes
	Accessory       bool // app should not be task switchable
	DisableAutoSave bool // disable window position saving and restoring
}

type App struct {
	app

	// obj      manifold.Node
	launched bool
}

func (a *App) Activate(ctx context.Context) error {
	// app.obj = manifold.FromContext(ctx).Parent()
	desktop.DispatchAsync(func() {
		a.Run(func() {
			a.Reload()
		})
	})
	return nil
}

// func (app *App) Signaled(s manifold.Signal) {
// 	if app.launched &&
// 		s.Name == "SetAttr" &&
// 		s.Args[0] == "enabled" &&
// 		strings.HasPrefix(manifold.Receiver(s).ComponentType(), "tractor.dev/hack/pkg/desktop/") {

// 		desktop.DispatchAsync(func() {
// 			app.Reload()
// 		})

// 	}
// }

// must be called from main thread
func (a *App) Reload() {
	// if a.obj == nil {
	// 	return
	// }
	// for _, i := range node.GetAll[*Indicator](a.obj, node.Include{Children: true}) {
	// 	i.Reload()
	// }
	// for _, w := range node.GetAll[*window.Window](a.obj, node.Include{Children: true, Disabled: true}) {
	// 	w.Reload()
	// }
}

func (a *App) Run(didFinish func()) {
	if a.launched {
		log.Println("application already launched")
		return
	}
	a.run(didFinish)
}
