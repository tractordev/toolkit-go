package app

import (
	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/macos/foundation"
	"tractor.dev/toolkit-go/desktop/menu"
)

type Indicator struct {
	Icon  []byte
	Items []menu.Item

	*appkit.StatusItem
}

func (i *Indicator) Unload() {
	if i.StatusItem != nil {
		i.StatusItem.Release()
		i.StatusItem = nil
	}
}

func (i *Indicator) SetMenu(m menu.Menu) {
	i.Items = nil
	i.StatusItem.SetMenu(*m.Menu)
}

func (i *Indicator) SetItems(items []menu.Item) {
	i.Items = items
	i.StatusItem.SetMenu(menu.New(items).Menu)
}

func (i *Indicator) Load() {
	image := appkit.NewImageWithData(i.Icon)
	image.SetSize(foundation.Size{Width: 16, Height: 16})
	image.SetTemplate(true)

	item := appkit.StatusBar_SystemStatusBar().StatusItemWithLength(appkit.VariableStatusItemLength)
	item.Button().SetImage(image)
	item.Button().SetImagePosition(appkit.ImageOnly)

	i.StatusItem = &item
	i.StatusItem.Retain()

	// use Items if set
	if i.Items != nil {
		i.SetItems(i.Items)
		return
	}

	// otherwise look for sibling menu
	// if i.com == nil {
	// 	return
	// }
	// if m := node.Get[*menu.Menu](i.com, node.Include{Siblings: true}); m != nil {
	// 	m.Reload()
	// 	i.StatusItem.SetMenu(m)
	// }
}
