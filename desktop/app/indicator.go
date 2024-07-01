package app

import "tractor.dev/toolkit-go/desktop/menu"

func NewIndicator(icon []byte, items []menu.Item) error {
	newIndicator(icon, items)
	return nil
}
