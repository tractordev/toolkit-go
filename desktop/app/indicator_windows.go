package app

import (
	"tractor.dev/toolkit-go/desktop/event"
	"tractor.dev/toolkit-go/desktop/menu"
	"tractor.dev/toolkit-go/desktop/win32"
)

func newIndicator(icon []byte, items []menu.Item) {
	menu := menu.New(items)
	onClick := func(id int32) {
		event.Emit(event.Event{
			Type:     event.MenuItem,
			MenuItem: int(id),
		})
	}
	win32.NewTrayMenu(menu.HPopupMenu(), icon, onClick)
}
