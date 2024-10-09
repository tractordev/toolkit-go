package menu

import (
	"tractor.dev/toolkit-go/desktop/linux"
)

type menu struct {
	linux.Menu
	items []Item
}

func (m *menu) unload() {
	if m.Menu.Handle != nil {
		m.Menu.Destroy()
		m.Menu.Handle = nil
	}
}

func (m *menu) load() {
	m.Menu = createMenu(m.items)
}

// Should always be called before load
func (m *menu) AddItem(it Item) {
	m.items = append(m.items, it)
}

//TODO
func (m *menu) popup() int {
	return 0
}

func createMenu(items []Item) linux.Menu {
	menu := linux.Menu_New()

	if menu.Handle != nil {
		for _, it := range items {
			// @Incomplete: accelerators
			item := linux.MenuItem_New(it.ID, it.Title, it.Disabled, it.Selected, it.Separator)

			if !it.Disabled && len(it.SubMenu) > 0 {
				submenu := createMenu(it.SubMenu)
				item.SetSubmenu(submenu)
			}

			menu.AppendItem(item)
		}
	}
	return menu
}