package desktop

func Dispatch(fn func()) {
	dispatch(fn, false)
}

func DispatchAsync(fn func()) {
	dispatch(fn, true)
}
