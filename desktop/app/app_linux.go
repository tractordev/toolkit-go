package app

import (
	"sync/atomic"

	"tractor.dev/toolkit-go/desktop"
	"tractor.dev/toolkit-go/desktop/event"
)

type app struct {
	Options
}

func (a *app) run(didFinish func()) {
	if a.Options.DisableAutoSave != true {
		// setupWindowRestoreListener(options.Identifier)
	}

	// NOTE(nick): MacOS-style window behavior
	if a.Options.Agent == false {
		var windowCount int64

		event.Listen("__APPTRON_Platform_listener__", func(e event.Event) error {
			if e.Type == event.Created {
				atomic.AddInt64(&windowCount, 1)
			}

			if e.Type == event.Destroyed {
				if atomic.AddInt64(&windowCount, -1) == 0 {
					desktop.Stop()
				}
			}

			return nil
		})
	}

	if didFinish != nil {
		desktop.DispatchAsync(didFinish)
	}
}
