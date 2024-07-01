package desktop

import (
	"time"

	"tractor.dev/toolkit-go/desktop/win32"
)

func init() {
	win32.OS_Init()
}

func start() {
	for isRunning.Load() {
		win32.PollEvents()

		select {
		case fn := <-dispatchQueue:
			fn()
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}
	win32.RemoveAllTrayMenus()
	win32.ExitProcess(0)
}

func stop() {
	isRunning.Store(false)
}
