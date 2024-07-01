package shell

import "tractor.dev/toolkit-go/desktop/linux"

func showNotification(n Notification) {
}

func showMessage(msg MessageDialog) bool {
	return false
}

func showFilePicker(fd FileDialog) []string {
	return []string{}
}

func clipboard() string {
	return linux.OS_GetClipboardText()
}

func setClipboard(text string) bool {
	return linux.OS_SetClipboardText(text)
}
