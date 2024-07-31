package app

import (
	"tractor.dev/toolkit-go/desktop/event"
	"tractor.dev/toolkit-go/desktop/menu"
	"tractor.dev/toolkit-go/desktop/win32"
)

type Indicator struct {
	Icon  []byte
	Items []menu.Item
}

func (i *Indicator) Unload() {
	// todo
}

func (i *Indicator) SetMenu(m menu.Menu) {
	i.Items = nil
	// i.StatusItem.SetMenu(m)
	// TODO
}

func (i *Indicator) SetItems(items []menu.Item) {
	i.Items = items
	// i.StatusItem.SetMenu(menu.New(items))
	// TODO
}

func (i *Indicator) Load() {
	menu := menu.New(i.Items)
	onClick := func(id int32) {
		event.Emit(event.Event{
			Type:     event.MenuItem,
			MenuItem: int(id),
		})
	}
	win32.NewTrayMenu(menu.HPopupMenu(), i.Icon, onClick)
}
