package app

import (
	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/macos/foundation"
	"tractor.dev/toolkit-go/desktop/menu"
)

func newIndicator(icon []byte, items []menu.Item) {
	obj := appkit.StatusBar_SystemStatusBar().StatusItemWithLength(appkit.VariableStatusItemLength)
	obj.Retain()
	//obj.Button().SetTitle(i.Text)
	image := appkit.NewImageWithData(icon)
	image.SetSize(foundation.Size{Width: 16.0, Height: 16.0})
	image.SetTemplate(true)
	obj.Button().SetImage(image)
	obj.Button().SetImagePosition(appkit.ImageOnly)

	menu := menu.New(items)
	obj.SetMenu(menu.Menu)
}
