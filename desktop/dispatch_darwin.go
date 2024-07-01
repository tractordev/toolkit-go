package desktop

import gcd "github.com/progrium/darwinkit/dispatch"

func dispatch(fn func(), async bool) {
	if async {
		gcd.MainQueue().DispatchAsync(fn)
	} else {
		gcd.MainQueue().DispatchSync(fn)
	}
}
