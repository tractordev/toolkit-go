package app

import (
	"sync/atomic"

	"tractor.dev/toolkit-go/desktop"
	"tractor.dev/toolkit-go/desktop/event"
	"tractor.dev/toolkit-go/desktop/win32"
)

func init() {
	// @see https://github.com/glfw/glfw/blob/master/src/win32_init.c#L692

	// NOTE(nick): the exact snippet from GLFW is:
	/*
			if (_glfwIsWindows10Version1703OrGreaterWin32())
		        SetProcessDpiAwarenessContext(DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2);
		    else if (IsWindows8Point1OrGreater())
		        SetProcessDpiAwareness(PROCESS_PER_MONITOR_DPI_AWARE);
		    else if (IsWindowsVistaOrGreater())
		        SetProcessDPIAware();

	*/
	// BUT, I think it's sufficient to just check if these proceedures are loaded?
	// @Robustness: test this assumption

	success := win32.SetProcessDpiAwarenessContext(win32.DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2)
	if !success {
		success = win32.SetProcessDpiAwareness(win32.PROCESS_PER_MONITOR_DPI_AWARE)
	}
	if !success {
		success = win32.SetProcessDPIAware()
	}
}

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
