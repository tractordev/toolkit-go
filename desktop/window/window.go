package window

import (
	"tractor.dev/toolkit-go/desktop"
)

type window struct {
	ID string
	Options

	// com entity.Node
}

type Size = desktop.Size
type Position = desktop.Position

type Options struct {
	AlwaysOnTop bool
	Frameless   bool
	Fullscreen  bool
	Size        Size
	MinSize     Size
	MaxSize     Size
	Maximized   bool
	Position    Position
	Resizable   bool
	Title       string
	Transparent bool
	Visible     bool
	Hidden      bool
	Center      bool
	Icon        []byte
	URL         string
	HTML        string
	Script      string
	ID          string
}

func New(opts Options) *Window {
	w := &Window{window: window{Options: opts}}
	//w.Reload()
	return w
}

// func (w *window) ComponentAttached(com entity.Node) {
// 	w.com = com
// }

func (w *Window) Reload() {
	w.Unload()
	w.Load()
}
