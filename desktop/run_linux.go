package desktop

import (
	"time"

	"tractor.dev/toolkit-go/desktop/linux"
)

func init() {
	linux.SetAllCFuncs()
	linux.OS_Init()
}

func start() {
	for isRunning.Load() {
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
	isRunning.Store(false)
}
