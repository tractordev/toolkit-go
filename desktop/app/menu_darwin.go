package app

import "tractor.dev/toolkit-go/desktop/menu"

func setMenu(m *menu.Menu) error {
	sharedApp.SetMainMenu(m.Menu)
	return menu.SetMain(m)
}
