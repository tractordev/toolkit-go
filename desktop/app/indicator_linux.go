package app

import (
	"fmt"
	"log"
	"os"

	"tractor.dev/toolkit-go/desktop/event"
	"tractor.dev/toolkit-go/desktop/linux"
	"tractor.dev/toolkit-go/desktop/menu"
)

var globalTrayId = 0

func newIndicator(icon []byte, items []menu.Item) {
	//
	// NOTE(nick): it seems like libappindicator warns about the "tmp" directory:
	//
	// libappindicator-WARNING **: 15:49:46.793: Using '/tmp' paths in SNAP environment will lead to unreadable resources
	//
	f, err := os.CreateTemp("", "apptron__icon-*.png")
	if err != nil {
		log.Println("[NewIndicator] Failed to create temporary icon file!")
		return
	}

	_, err = f.Write(icon)
	if err != nil {
		log.Println("[NewIndicator] Failed to create write icon bytes!")
		return
	}

	// @Incomplete @Leak: should remove tmp png file when deleting indicator
	//defer os.Remove(f.Name())

	globalTrayId += 1
	trayId := fmt.Sprintf("tray_%d", globalTrayId)

	trayIconPath := f.Name()

	menu := menu.New(items)
	linux.Indicator_New(trayId, trayIconPath, menu.Menu)

	linux.SetGlobalMenuCallback(func(menuId int) {
		event.Emit(event.Event{
			Type:     event.MenuItem,
			MenuItem: menuId,
		})
	})
}
