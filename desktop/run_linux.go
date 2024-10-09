package desktop

import (
	"time"

	"tractor.dev/toolkit-go/desktop/linux"
)

func start() {
	for isRunning.Load() {
		linux.LoadLibraries()
		linux.SetAllCFuncs()
		linux.OS_Init()

		linux.PollEvents()

		select {
		case fn := <-dispatchQueue:
			fn()
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}
}

func stop() {
	dispatch(func() {
		linux.UnloadLibraries()
	}, false)
	isRunning.Store(false)
}
