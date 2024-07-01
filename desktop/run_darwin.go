package desktop

import (
	"github.com/progrium/darwinkit/macos/appkit"
)

var AppReady = make(chan bool)

func start() {
	// loop := foundation.RunLoop_MainRunLoop()
	// loop.Run()
	<-AppReady
	appkit.Application_SharedApplication().Run()
}

func stop() {
	app := appkit.Application_SharedApplication()
	app.Terminate(nil)
}
