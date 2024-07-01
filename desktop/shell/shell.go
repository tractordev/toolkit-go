package shell

type Notification struct {
	Title    string
	Subtitle string // for MacOS only
	Body     string
	/*
		Silent   bool
	*/
}

type FileDialog struct {
	Title     string
	Directory string
	Filename  string
	Mode      string   // pickfile, pickfiles, pickfolder, savefile
	Filters   []string // each string is comma delimited (go,rs,toml) with optional label prefix (text:go,txt)
}

type MessageDialog struct {
	Title   string
	Body    string
	Level   string // info, warning, error
	Buttons string // ok, okcancel, yesno
}

func ShowNotification(n Notification) {
	showNotification(n)
}

func ShowMessage(msg MessageDialog) bool {
	return showMessage(msg)
}

func ShowFilePicker(fd FileDialog) []string {
	return showFilePicker(fd)
}

func Clipboard() string {
	return clipboard()
}

func SetClipboard(text string) bool {
	return setClipboard(text)
}

func RegisterShortcut(accelerator string) {
	// hotkey does its own dispatch so
	// this avoids a deadlock
	go registerShortcut(accelerator)
}

func IsShortcutRegistered(accelerator string) bool {
	return isShortcutRegistered(accelerator)
}

func UnregisterShortcut(accelerator string) bool {
	return unregisterShortcut(accelerator)
}

func UnregisterAllShortcuts() {
	unregisterAllShortcuts()
}
