//go:build !darwin

package app

import "tractor.dev/toolkit-go/desktop/menu"

func setMenu(men *menu.Menu) {
	menu.SetMain(men)
}
