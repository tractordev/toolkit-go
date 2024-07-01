package shell

import (
	"strings"

	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/macos/foundation"
	"github.com/progrium/darwinkit/objc"
	"github.com/progrium/macdriver/cocoa"
	"github.com/progrium/macdriver/core"
)

func showNotification(n Notification) {
	notification := objc.Call[objc.Object](objc.GetClass("NSUserNotification"), objc.Sel("new"))
	objc.Call[objc.Void](notification, objc.Sel("setTitle:"), foundation.String_StringWithString(n.Title))
	objc.Call[objc.Void](notification, objc.Sel("setInformativeText:"), foundation.String_StringWithString(n.Body))

	center := objc.Call[objc.Object](foundation.UserNotificationCenterClass, objc.Sel("defaultUserNotificationCenter"))
	objc.Call[objc.Void](center, objc.Sel("deliverNotification:"), notification)
	notification.Release()
}

func showMessage(msg MessageDialog) bool {
	alert := appkit.NewAlert()

	switch msg.Level {
	case "error":
		alert.SetAlertStyle(appkit.AlertStyleCritical)
	case "warning":
		alert.SetAlertStyle(appkit.AlertStyleWarning)
	default:
		alert.SetAlertStyle(appkit.AlertStyleInformational)
	}

	switch msg.Buttons {
	case "ok":
		alert.AddButtonWithTitle("OK")
	case "okcancel":
		alert.AddButtonWithTitle("OK")
		alert.AddButtonWithTitle("Cancel")
	case "yesno":
		alert.AddButtonWithTitle("Yes")
		alert.AddButtonWithTitle("No")
	}

	alert.SetMessageText(msg.Title)
	alert.SetInformativeText(msg.Body)

	return alert.RunModal() == 1000
}

func showFilePicker(fd FileDialog) []string {
	if fd.Mode == "savefile" {
		return showSavePicker(fd)
	}
	return showOpenPicker(fd)
}

func showSavePicker(fd FileDialog) []string {
	picker := appkit.SavePanel_SavePanel()
	if fd.Filename != "" {
		picker.SetNameFieldStringValue(fd.Filename)
	}
	if fd.Directory != "" {
		url := foundation.URL_FileURLWithPathIsDirectory(fd.Directory, true)
		picker.SetDirectoryURL(url)
	}
	if fd.Filters != nil {
		var filters []objc.IObject
		for _, entry := range fd.Filters {
			kvp := strings.Split(entry, ":")
			var idx int
			if len(kvp) > 1 {
				idx = 1
			}
			for _, ext := range strings.Split(kvp[idx], ",") {
				filters = append(filters, core.String(ext))
			}
		}
		objc.Call[objc.Void](picker, objc.Sel("setAllowedFileTypes:"), foundation.Array_ArrayWithArray(filters))
	}
	picker.SetTitle(fd.Title)
	if picker.RunModal() == 1 {
		return []string{picker.URL().Path()}
	}
	return []string{}
}

func showOpenPicker(fd FileDialog) []string {
	picker := appkit.OpenPanel_OpenPanel()
	switch fd.Mode {
	case "pickfiles":
		picker.SetAllowsMultipleSelection(true)
	case "pickfolder":
		picker.SetCanChooseDirectories(true)
	default: // pickfile
	}
	if fd.Filename != "" {
		picker.SetNameFieldStringValue(fd.Filename)
	}
	if fd.Directory != "" {
		url := foundation.URL_FileURLWithPathIsDirectory(fd.Directory, true)
		picker.SetDirectoryURL(url)
	}
	if fd.Filters != nil {
		var filters []objc.IObject
		for _, entry := range fd.Filters {
			kvp := strings.Split(entry, ":")
			var idx int
			if len(kvp) > 1 {
				idx = 1
			}
			for _, ext := range strings.Split(kvp[idx], ",") {
				filters = append(filters, core.String(ext))
			}
		}
		objc.Call[objc.Void](picker, objc.Sel("setAllowedFileTypes:"), foundation.Array_ArrayWithArray(filters))
	}
	picker.SetTitle(fd.Title)
	if picker.RunModal() == 1 {
		var paths []string
		for _, url := range picker.URLs() {
			paths = append(paths, url.Path())
		}
		return paths
	}
	return []string{}
}

func clipboard() string {
	pb := appkit.Pasteboard_GeneralPasteboard()
	return pb.StringForType(appkit.PasteboardTypeString)
}

func setClipboard(text string) bool {
	pb := appkit.Pasteboard_GeneralPasteboard()
	pb.ClearContents()
	pb.SetStringForType(text, cocoa.NSPasteboardTypeString)
	return true
}
