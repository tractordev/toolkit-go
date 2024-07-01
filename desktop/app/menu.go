package app

import "tractor.dev/toolkit-go/desktop/menu"

func Menu() *menu.Menu {
	return menu.Main()
}

func SetMenu(m *menu.Menu) {
	setMenu(m)
}
