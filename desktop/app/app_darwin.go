package app

import (
	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/macos/foundation"
	"tractor.dev/toolkit-go/desktop"
	"tractor.dev/toolkit-go/desktop/menu"
)

var (
	sharedApp appkit.Application
)

func init() {
	sharedApp = appkit.Application_SharedApplication()
}

type app struct {
	Options
}

func (a *app) run(didFinish func()) {
	// mainBundle := foundation.Bundle_MainBundle()
	// bundleClass := mainBundle.Class()
	// bundleClass.AddMethod("__bundleIdentifier", func(self objc.Object) objc.Object {
	// 	if self.Ptr() == mainBundle.Ptr() {
	// 		return foundation.String_StringWithString(options.Identifier).Object
	// 	}
	// 	// After the swizzle this will point to the original method, and return the
	// 	// original bundle identifier.
	// 	return objc.Call[objc.Object](self, objc.Sel("__bundleIdentifier"))
	// })
	// bundleClass.Swizzle("bundleIdentifier", "__bundleIdentifier")

	delegate := &appkit.ApplicationDelegate{}
	delegate.SetApplicationShouldTerminateAfterLastWindowClosed(func(sender appkit.Application) bool {
		return !a.Options.Agent
		// for demo
		// if os.Getenv("MSOCK") != "" {
		// 	return false
		// }
		// return true
	})
	delegate.SetApplicationDidFinishLaunching(func(notification foundation.Notification) {
		sharedApp.ActivateIgnoringOtherApps(true) // not if accessory?
		if didFinish != nil {
			didFinish()
		}
	})
	delegate.SetApplicationWillFinishLaunching(func(notification foundation.Notification) {
		mainMenu := menu.Main()
		if mainMenu == nil {
			menu.SetMain(menu.New([]menu.Item{}))
			mainMenu = menu.Main()
			// sharedApp.SetMainMenu(mainMenu.Menu)
		}
		if a.Options.Accessory {
			sharedApp.SetActivationPolicy(appkit.ApplicationActivationPolicyAccessory)
		} else {
			// sharedApp.SetMainMenu(mainMenu.Menu)
			sharedApp.SetActivationPolicy(appkit.ApplicationActivationPolicyRegular)
		}
	})
	sharedApp.SetDelegate(delegate)

	// DelegateClass.AddMethod("menuClick:", func(self, sender objc.Object) {
	// 	event.Emit(event.Event{
	// 		Type:     event.MenuItem,
	// 		MenuItem: int(sender.Get("tag").Int()),
	// 	})
	// })

	if a.Options.DisableAutoSave != true {
		// setupWindowRestoreListener(options.Identifier)
	}

	// sharedApp.Run()
	desktop.AppReady <- true
}
