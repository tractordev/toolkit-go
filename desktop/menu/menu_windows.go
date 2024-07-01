package menu

import (
	"tractor.dev/toolkit-go/desktop/win32"
)

type menu struct {
	popupMenu win32.HMENU
	menu      win32.HMENU
}

func (m *menu) HMenu() win32.HMENU {
	return m.menu
}

func (m *menu) HPopupMenu() win32.HMENU {
	return m.popupMenu
}

func (m *menu) unload() {
	//
	// NOTE(nick): from the win32 docs:
	// DestroyMenu is recursive, that is, it will destroy the menu and all its submenus.
	//
	// @see https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-destroymenu
	//

	if m.popupMenu != 0 {
		win32.DestroyMenu(m.popupMenu)
		m.popupMenu = 0
	}

	if m.menu != 0 {
		win32.DestroyMenu(m.menu)
		m.menu = 0
	}
}

func (m *menu) reload(items []Item) {
	// @Cleanup: maybe just dynamically create the win32 menu each time it's needed?
	m.popupMenu = createMenu(true, items)
	m.menu = createMenu(false, items)
}

func (m *menu) popup() int {
	hwnd := win32.GetActiveWindow()

	if hwnd == 0 {
		return 0
	}

	win32.SetForegroundWindow(hwnd)

	mousePosition := win32.POINT{}
	win32.GetCursorPos(&mousePosition)

	var flags win32.UINT = win32.TPM_RIGHTBUTTON | win32.TPM_NONOTIFY | win32.TPM_RETURNCMD
	result := win32.TrackPopupMenu(m.popupMenu, flags, int32(mousePosition.X), int32(mousePosition.Y), 0, hwnd, nil)

	return int(result)
}

func createMenu(popup bool, items []Item) win32.HMENU {
	var menu win32.HMENU
	if popup {
		menu = win32.CreatePopupMenu()
	} else {
		menu = win32.CreateMenu()
	}

	if menu != win32.NULL {
		for _, it := range items {

			var info win32.MENUITEMINFO

			if it.Separator {
				info = win32.MakeMenuItemSeparator()
			} else {
				title := it.Title
				accel := it.Accelerator // @Incomplete: should this string be massaged at all?

				if len(it.Accelerator) > 0 {
					title += "\t" + accel
				}
				info = win32.MakeMenuItem(it.ID, title, it.Disabled, it.Selected, it.Selected == true)

				if !it.Disabled && len(it.SubMenu) > 0 {
					submenu := createMenu(popup, it.SubMenu)
					win32.AppendSubmenu(submenu, &info)
				}
			}

			win32.InsertMenuItemW(menu, win32.UINT(win32.GetMenuItemCount(menu)), 1, &info)
		}
	}

	return menu
}
