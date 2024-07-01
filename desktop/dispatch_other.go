//go:build !darwin

package desktop

var dispatchQueue = make(chan func())

func dispatch(fn func(), async bool) {
	if async {
		dispatchQueue <- fn
	} else {
		done := make(chan bool, 1)
		dispatchQueue <- func() {
			fn()
			done <- true
		}
		<-done
	}
}
