package event

import (
	"tractor.dev/toolkit-go/desktop"
)

var EventHandler func(event Event)

type Event struct {
	Type     Type
	Window   any
	Position desktop.Position
	Size     desktop.Size
	MenuItem int
	Shortcut string
}
