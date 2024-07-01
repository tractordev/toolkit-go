// TODO: update/move comment:
// osloop is intended to be a cross-platform abstraction for the GUI
// event loop that needs to run on the main thread. The idea is for the
// loop to co-exist with a normal Go program, so you'd call RunWith in
// main with a function representing what you'd normally have in main,
// which will be run in a new goroutine while the platform specific
// event loop runs in the main thread+goroutine. It also provides an
// abstraction to run code in the main thread using any platform specific
// mechanism, such as Apple's Grand Central Dispatch.
package desktop

import (
	"runtime"
	"sync/atomic"
)

var (
	isRunning atomic.Bool
)

func Start(fn func()) {
	if isRunning.Load() {
		panic("start already called")
	}

	runtime.LockOSThread()

	if fn != nil {
		go fn()
	}

	isRunning.Store(true)
	start()
}

func Stop() {
	stop()
}
