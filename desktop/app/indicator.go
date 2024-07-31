package app

import "tractor.dev/toolkit-go/desktop/menu"

func NewIndicator(icon []byte, items []menu.Item) *Indicator {
	i := &Indicator{
		Icon:  icon,
		Items: items,
	}
	i.Reload()
	return nil
}

func (i *Indicator) Reload() {
	i.Unload()
	i.Load()
}
