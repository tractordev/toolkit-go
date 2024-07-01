package shell

import (
	"tractor.dev/toolkit-go/desktop/win32"
)

func showNotification(n Notification) {
}

func showMessage(msg MessageDialog) bool {
	var flags win32.UINT = 0

	switch msg.Level {
	case "error":
		flags |= win32.MB_ICONERROR
	case "warning":
		flags |= win32.MB_ICONWARNING
	default:
		flags |= win32.MB_ICONINFORMATION
	}

	switch msg.Buttons {
	case "okcancel":
		flags |= win32.MB_OKCANCEL
	case "yesno":
		flags |= win32.MB_YESNO
	default:
		flags |= win32.MB_OK
	}

	return win32.MessageBox(win32.NULL, msg.Body, msg.Title, flags) == win32.IDOK
}

func showFilePicker(fd FileDialog) []string {
	return []string{}
}

func clipboard() string {
	return win32.OS_GetClipboardText()
}

func setClipboard(text string) bool {
	return win32.OS_SetClipboardText(text)
}
