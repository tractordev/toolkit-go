//go:build linux

package linux

import (
	"fmt"
	"log"
	"os"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

/*
#cgo linux pkg-config: gtk+-3.0 webkit2gtk-4.1 ayatana-appindicator3-0.1

#include "linux.h"
*/
import "C"

type Window struct {
	Handle *C.struct__GtkWindow
}

type Webview struct {
	Handle *C.struct__WebKitWebView
}

type Menu struct {
	Handle *C.struct__GtkMenu
}

type MenuItem struct {
	Handle *C.struct__GtkMenuItem
}

type Indicator struct {
	Handle *C.struct__AppIndicator
}

type Monitor struct {
	Handle *C.struct__GdkMonitor
}

type Size struct {
	Width  int
	Height int
}

type Position struct {
	X int
	Y int
}

type Rectangle struct {
	Position Position
	Size     Size
}

type EventType int

const (
	None EventType = iota
	Delete
	Destroy
	Configure
	FocusChange
)

type Event struct {
	Type     EventType
	Window   Window
	UserData int

	Position Position
	Size     Size
	FocusIn  bool
}

type Menu_Callback func(menuId int)

var globalMenuCallback Menu_Callback

type Event_Callback func(event Event)

var globalEventCallback Event_Callback

// NOTE(nick): there are quiet a lot of these!
//
// @see https://webkitgtk.org/reference/webkit2gtk/stable/WebKitSettings.html
type WebviewSetings struct {
	CanAccessClipboard   bool
	WriteConsoleToStdout bool
	DeveloperTools       bool
}

// GtkWindowType
const (
	GTK_WINDOW_TOPLEVEL = 0
	GTK_WINDOW_POPUP    = 1
)

// GdkWindowHints
const (
	GDK_HINT_POS	     = 1 << 0
	GDK_HINT_MIN_SIZE    = 1 << 1
	GDK_HINT_MAX_SIZE    = 1 << 2
	GDK_HINT_BASE_SIZE   = 1 << 3
	GDK_HINT_ASPECT      = 1 << 4
	GDK_HINT_RESIZE_INC  = 1 << 5
	GDK_HINT_WIN_GRAVITY = 1 << 6
	GDK_HINT_USER_POS    = 1 << 7
	GDK_HINT_USER_SIZE   = 1 << 8
)

// WebKitUserContentInjectedFrames
const (
	WEBKIT_USER_CONTENT_INJECT_ALL_FRAMES = 0
	WEBKIT_USER_CONTENT_INJECT_TOP_FRAME  = 1
)

// WebKitUserScriptInjectionTime
const (
	WEBKIT_USER_SCRIPT_INJECT_AT_DOCUMENT_START = 0
	WEBKIT_USER_SCRIPT_INJECT_AT_DOCUMENT_END   = 1
)

/*
* PureGo Gtk Bindings
*/

var (
	LibCFree func (unsafe.Pointer)
)

var (
	//TODO put these in lexical order
	//TODO indentation
	//TODO convert all *C.char to string
	//TODO transfer types as well
	//TODO using 'struct__' is not needed because most are typedefs to structs
	GtkMain 						  func ()
	GtkInitCheck 					  func (argc unsafe.Pointer, argv unsafe.Pointer)
	GtkMainIterationDo 				  func (blocking bool)
	GtkWindowNew 					  func (window_type uint32) *C.struct__GtkWidget
	GtkContainerAdd 				  func (container *C.struct__GtkContainer, widget *C.struct__GtkWidget)
	GtkWidgetGrabFocus 				  func (widget *C.struct__GtkWidget)
	GtkWidgetShowAll 				  func (widget *C.struct__GtkWidget)
	GtkWidgetHide   				  func (widget *C.struct__GtkWidget)
	GtkWidgetDestroy 				  func (widget *C.struct__GtkWidget)
	GtkWindowSetDecorated 			  func (window *C.struct__GtkWindow, setting bool)
	GtkWindowGetSize 				  func (window *C.struct__GtkWindow, width *C.int, height *C.int)
	GtkCheckMenuItemNewWithLabel      func (label *C.char) *C.struct__GtkWidget
    GtkCheckMenuItemSetActive         func (checkMenuItem *C.struct__GtkCheckMenuItem, is_active bool)
	//GdkAtom is basically a pointer to 'struct _GdkAtom' (gdktypes.h)
    GtkClipboardGet                   func (selection C.GdkAtom) *C.struct__GtkClipboard
    GtkClipboardSetText               func (clipboard *C.struct__GtkClipboard, text *C.char, length int)
    GtkClipboardWaitForText           func (clipboard *C.struct__GtkClipboard) *C.char
    GtkMenuItemNewWithLabel           func (label *C.char) *C.struct__GtkWidget
    GtkMenuItemSetSubmenu             func (menu_item *C.struct__GtkMenuItem, submenu *C.struct__GtkWidget)
    GtkMenuNew                        func () *C.struct__GtkWidget
    GtkMenuShellAppend                func (menu_shell *C.struct__GtkMenuShell, child *C.struct__GtkWidget)
    GtkSeparatorMenuItemNew           func () *C.struct__GtkWidget
    GtkWidgetIsVisible                func (widget *C.struct__GtkWidget) bool
    GtkWidgetSetSensitive             func (widget *C.struct__GtkWidget, sensitive bool)
    GtkWidgetShow                     func (widget *C.struct__GtkWidget)
    GtkWindowDeiconify                func (window *C.struct__GtkWindow)
    GtkWindowIconify                  func (window *C.struct__GtkWindow)
    GtkWindowFullscreen               func (window *C.struct__GtkWindow)
    GtkWindowUnfullscreen             func (window *C.struct__GtkWindow)
    GtkWindowGetPosition              func (window *C.struct__GtkWindow, x, y *C.int)
    GtkWindowMaximize                 func (window *C.struct__GtkWindow)
    GtkWindowUnmaximize               func (window *C.struct__GtkWindow)
    GtkWindowMove                     func (window *C.struct__GtkWindow, x, y int)
    GtkWindowResize                   func (window *C.struct__GtkWindow, width, height int)
    GtkWindowSetGeometryHints         func (
		window *C.struct__GtkWindow,
		geometry_widget *C.struct__GtkWidget,
		geometry *C.struct__GdkGeometry,
		geom_mask uint32)
    GtkWindowSetIcon                  func (window *C.struct__GtkWindow, icon *C.struct__GdkPixbuf)
    GtkWindowSetKeepAbove             func (window *C.struct__GtkWindow, setting bool)
    GtkWindowSetResizable             func (window *C.struct__GtkWindow, resizable bool)
    GtkWindowSetTitle                 func (window *C.struct__GtkWindow, title *C.char)
)

var (
    GdkScreenGetRootWindow            func (window *C.struct__GdkScreen) *C.struct__GdkWindow
	GdkScreenGetDefault				  func () *C.struct__GdkScreen
	GdkDisplayGetDefault          	  func() *C.struct__GdkDisplay
    GdkDisplayGetMonitor          	  func(display *C.struct__GdkDisplay, monitorNum C.int) *C.struct__GdkMonitor
    GdkDisplayGetNMonitors        	  func(display *C.struct__GdkDisplay) C.int
	//GdkRectangle is not always a struct, and in this case it's a typedef to a definition in cairo
	//so C.struct__ prefix is not used here
	//TODO once cgo is completely out, these defitions will follow. So this won't even matter, probably.
    GdkMonitorGetGeometry         	  func(monitor *C.struct__GdkMonitor, rect *C.GdkRectangle)
    GdkMonitorGetManufacturer      	  func(monitor *C.struct__GdkMonitor) *C.char
    GdkMonitorGetModel                func(monitor *C.struct__GdkMonitor) *C.char
    GdkMonitorGetRefreshRate          func(monitor *C.struct__GdkMonitor) int
    GdkMonitorGetScaleFactor          func(monitor *C.struct__GdkMonitor) int
    GdkMonitorIsPrimary               func(monitor *C.struct__GdkMonitor) bool
    GdkPixbufNewFromFile          	  func(filename *C.char, err **C.struct__GError) *C.struct__GdkPixbuf
    GdkWindowGetGeometry          	  func(window *C.struct__GdkWindow, x, y, width, height *C.int)
)

var (
    WebkitSettingsSetEnableDeveloperExtras                  func(settings *C.struct__WebKitSettings, enable bool)
    WebkitSettingsSetEnableWriteConsoleMessagesToStdout   	func(settings *C.struct__WebKitSettings, enable bool)
    WebkitSettingsSetJavascriptCanAccessClipboard           func(settings *C.struct__WebKitSettings, enable bool)
    WebkitUserContentManagerAddScript                       func(manager *C.struct__WebKitUserContentManager, script *C.struct__WebKitUserScript)
    WebkitUserContentManagerRegisterScriptMessageHandler    func(manager *C.struct__WebKitUserContentManager, name *C.char) bool
    WebkitUserScriptNew                                     func(
		source *C.char,
	 	injected_frames uint32,
	 	injected_time uint32,
		allow_list *C.char,
		block_list *C.char,
	) *C.struct__WebKitUserScript
    WebkitWebViewEvaluateJavascript                         func(
		web_view *C.struct__WebKitWebView,
		script *C.char,
		length int,
		world_name *C.char,
		source_uri *C.char,
		cancellable *C.struct__GCancellable,
		callback *C.struct__GAsyncReadyCallback,
		user_data unsafe.Pointer,
	)
    WebkitWebViewGetSettings                                func(web_view *C.struct__WebKitWebView) *C.struct__WebKitSettings
    WebkitWebViewGetUserContentManager                      func(web_view *C.struct__WebKitWebView) *C.struct__WebKitUserContentManager
    WebkitWebViewLoadHtml                                   func(web_view *C.struct__WebKitWebView, content *C.char, base_uri *C.char)
    WebkitWebViewLoadUri                                    func(web_view *C.struct__WebKitWebView, uri *C.char)
    WebkitWebViewNew                                        func() *C.struct__GtkWidget
)


//TODO make these dynamically find libraries
//Find out if it's statically linkable
func GetGTKPath() string {
	return "/usr/lib/libgtk-3.so"
}

func GetCLibPath() string {
	return "libc.so.6"
}

func GetWebkitGtkLibbPath() string {
	return "/usr/lib/libwebkit2gtk-4.1.so"
}

func SetAllCFuncs() {
	libc, err := purego.Dlopen(GetCLibPath(), purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		panic(err)
	}

	libgtk, err := purego.Dlopen(GetGTKPath(), purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		panic(err)
	}

	//LibC functions
	purego.RegisterLibFunc(&LibCFree, libc, "free")

	//Gtk functions
	purego.RegisterLibFunc(&GtkMain, libgtk, "gtk_main")
	purego.RegisterLibFunc(&GtkInitCheck, libgtk, "gtk_init_check")
	purego.RegisterLibFunc(&GtkMainIterationDo, libgtk, "gtk_main_iteration_do")
	purego.RegisterLibFunc(&GtkWindowNew, libgtk, "gtk_window_new")
	purego.RegisterLibFunc(&GtkWindowNew, libgtk, "gtk_window_new")
	purego.RegisterLibFunc(&GtkContainerAdd, libgtk, "gtk_container_add")
	purego.RegisterLibFunc(&GtkWidgetGrabFocus, libgtk, "gtk_widget_grab_focus")
	purego.RegisterLibFunc(&GtkWidgetShowAll, libgtk, "gtk_widget_show_all")
	purego.RegisterLibFunc(&GtkWidgetHide, libgtk, "gtk_widget_hide")
	purego.RegisterLibFunc(&GtkWidgetDestroy, libgtk, "gtk_widget_destroy")
	purego.RegisterLibFunc(&GtkWindowSetDecorated, libgtk, "gtk_window_set_decorated")
	purego.RegisterLibFunc(&GtkWindowGetSize, libgtk, "gtk_window_get_size")
	purego.RegisterLibFunc(&GtkCheckMenuItemNewWithLabel, libgtk, "gtk_check_menu_item_new_with_label")
	purego.RegisterLibFunc(&GtkCheckMenuItemSetActive, libgtk, "gtk_check_menu_item_set_active")
	purego.RegisterLibFunc(&GtkClipboardGet, libgtk, "gtk_clipboard_get")
	purego.RegisterLibFunc(&GtkClipboardSetText, libgtk, "gtk_clipboard_set_text")
	purego.RegisterLibFunc(&GtkClipboardWaitForText, libgtk, "gtk_clipboard_wait_for_text")
	purego.RegisterLibFunc(&GtkMenuItemNewWithLabel, libgtk, "gtk_menu_item_new_with_label")
	purego.RegisterLibFunc(&GtkMenuItemSetSubmenu, libgtk, "gtk_menu_item_set_submenu")
	purego.RegisterLibFunc(&GtkMenuNew, libgtk, "gtk_menu_new")
	purego.RegisterLibFunc(&GtkMenuShellAppend, libgtk, "gtk_menu_shell_append")
	purego.RegisterLibFunc(&GtkSeparatorMenuItemNew, libgtk, "gtk_separator_menu_item_new")
	purego.RegisterLibFunc(&GtkWidgetIsVisible, libgtk, "gtk_widget_is_visible")
	purego.RegisterLibFunc(&GtkWidgetSetSensitive, libgtk, "gtk_widget_set_sensitive")
	purego.RegisterLibFunc(&GtkWidgetShow, libgtk, "gtk_widget_show")
	purego.RegisterLibFunc(&GtkWindowDeiconify, libgtk, "gtk_window_deiconify")
	purego.RegisterLibFunc(&GtkWindowFullscreen, libgtk, "gtk_window_fullscreen")
	purego.RegisterLibFunc(&GtkWindowGetPosition, libgtk, "gtk_window_get_position")
	purego.RegisterLibFunc(&GtkWindowIconify, libgtk, "gtk_window_iconify")
	purego.RegisterLibFunc(&GtkWindowMaximize, libgtk, "gtk_window_maximize")
	purego.RegisterLibFunc(&GtkWindowMove, libgtk, "gtk_window_move")
	purego.RegisterLibFunc(&GtkWindowResize, libgtk, "gtk_window_resize")
	purego.RegisterLibFunc(&GtkWindowSetGeometryHints, libgtk, "gtk_window_set_geometry_hints")
	purego.RegisterLibFunc(&GtkWindowSetIcon, libgtk, "gtk_window_set_icon")
	purego.RegisterLibFunc(&GtkWindowSetKeepAbove, libgtk, "gtk_window_set_keep_above")
	purego.RegisterLibFunc(&GtkWindowSetResizable, libgtk, "gtk_window_set_resizable")
	purego.RegisterLibFunc(&GtkWindowSetTitle, libgtk, "gtk_window_set_title")
	purego.RegisterLibFunc(&GtkWindowUnfullscreen, libgtk, "gtk_window_unfullscreen")
	purego.RegisterLibFunc(&GtkWindowUnmaximize, libgtk, "gtk_window_unmaximize")

	//Gdk functions
	purego.RegisterLibFunc(&GdkScreenGetRootWindow, libgtk, "gdk_screen_get_root_window")
	purego.RegisterLibFunc(&GdkScreenGetDefault, libgtk, "gdk_screen_get_default")
	purego.RegisterLibFunc(&GdkDisplayGetDefault, libgtk, "gdk_display_get_default")
	purego.RegisterLibFunc(&GdkDisplayGetMonitor, libgtk, "gdk_display_get_monitor")
	purego.RegisterLibFunc(&GdkDisplayGetNMonitors, libgtk, "gdk_display_get_n_monitors")
	purego.RegisterLibFunc(&GdkMonitorGetGeometry, libgtk, "gdk_monitor_get_geometry")
	purego.RegisterLibFunc(&GdkMonitorGetManufacturer, libgtk, "gdk_monitor_get_manufacturer")
	purego.RegisterLibFunc(&GdkMonitorGetModel, libgtk, "gdk_monitor_get_model")
	purego.RegisterLibFunc(&GdkMonitorGetRefreshRate, libgtk, "gdk_monitor_get_refresh_rate")
	purego.RegisterLibFunc(&GdkMonitorGetScaleFactor, libgtk, "gdk_monitor_get_scale_factor")
	purego.RegisterLibFunc(&GdkMonitorIsPrimary, libgtk, "gdk_monitor_is_primary")
	purego.RegisterLibFunc(&GdkPixbufNewFromFile, libgtk, "gdk_pixbuf_new_from_file")
	purego.RegisterLibFunc(&GdkWindowGetGeometry, libgtk, "gdk_window_get_geometry")


	libwebgtk, err := purego.Dlopen(GetWebkitGtkLibbPath(), purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		panic(err)
	}
	//Webkit functions
	purego.RegisterLibFunc(&WebkitSettingsSetEnableDeveloperExtras, libwebgtk, "webkit_settings_set_enable_developer_extras")
	purego.RegisterLibFunc(&WebkitSettingsSetEnableWriteConsoleMessagesToStdout, libwebgtk, "webkit_settings_set_enable_write_console_messages_to_stdout")
	purego.RegisterLibFunc(&WebkitSettingsSetJavascriptCanAccessClipboard, libwebgtk, "webkit_settings_set_javascript_can_access_clipboard")
	purego.RegisterLibFunc(&WebkitUserContentManagerAddScript, libwebgtk, "webkit_user_content_manager_add_script")
	purego.RegisterLibFunc(&WebkitUserContentManagerRegisterScriptMessageHandler, libwebgtk, "webkit_user_content_manager_register_script_message_handler")
	purego.RegisterLibFunc(&WebkitUserScriptNew, libwebgtk, "webkit_user_script_new")
	purego.RegisterLibFunc(&WebkitWebViewEvaluateJavascript, libwebgtk, "webkit_web_view_evaluate_javascript")
	purego.RegisterLibFunc(&WebkitWebViewGetSettings, libwebgtk, "webkit_web_view_get_settings")
	purego.RegisterLibFunc(&WebkitWebViewGetUserContentManager, libwebgtk, "webkit_web_view_get_user_content_manager")
	purego.RegisterLibFunc(&WebkitWebViewLoadHtml, libwebgtk, "webkit_web_view_load_html")
	purego.RegisterLibFunc(&WebkitWebViewLoadUri, libwebgtk, "webkit_web_view_load_uri")
	purego.RegisterLibFunc(&WebkitWebViewNew, libwebgtk, "webkit_web_view_new")



	//TODO where and when to close files
	purego.Dlclose(libc)
	purego.Dlclose(libgtk)
}

//
// Exports
//



func OS_Init() {
	GtkInitCheck(nil, nil)
}

func PollEvents() {
	GtkMainIterationDo(false) // false = non-blocking
}

func Window_New() Window {
	result := Window{}
	result.Handle = Window_FromWidget(GtkWindowNew(GTK_WINDOW_TOPLEVEL))
	return result
}

func Webview_New() Webview {
	result := Webview{}
	//TODO cast in go
	result.Handle = Webview_FromWidget(WebkitWebViewNew())
	return result
}

func (window *Window) Pointer() uintptr {
	return (uintptr)(unsafe.Pointer(window.Handle))
}

func (window *Window) AddWebview(webview Webview) {
	GtkContainerAdd(Window_GTK_CONTAINER(window.Handle), Webview_GTK_WIDGET(webview.Handle))
	GtkWidgetGrabFocus(Webview_GTK_WIDGET(webview.Handle))
}

func (window *Window) Show() {
	GtkWidgetShowAll(Window_GTK_WIDGET(window.Handle))
}

func (window *Window) Hide() {
	GtkWidgetHide(Window_GTK_WIDGET(window.Handle))
}

func (window *Window) Destroy() {
	if window.Handle != nil {
		GtkWidgetDestroy(Window_GTK_WIDGET(window.Handle))
		window.Handle = nil
	}
}

//TODO transfer linux.h to purego
func (window *Window) SetTransparent(transparent bool) {
	C.gtk_window_set_transparent(window.Handle, toCBool(transparent))
}

func (window *Window) SetTitle(title string) {
	//TODO
	ctitle := C.CString(title)
	defer LibCFree(unsafe.Pointer(ctitle))

	GtkWindowSetTitle(window.Handle, ctitle)
}

func (window *Window) SetDecorated(decorated bool) {
	GtkWindowSetDecorated(window.Handle, decorated)
}

func (window *Window) GetSize() Size {
	result := Size{}

	//TODO
	width := C.int(0)
	height := C.int(0)
	GtkWindowGetSize(window.Handle, &width, &height)

	result.Width = int(width)
	result.Height = int(height)

	return result
}

func (window *Window) GetPosition() Position {
	result := Position{}

	x := C.int(0)
	y := C.int(0)
	GtkWindowGetPosition(window.Handle, &x, &y)

	result.X = int(x)
	result.Y = int(y)

	return result
}

func (window *Window) SetResizable(resizable bool) {
	GtkWindowSetResizable(window.Handle, resizable)
}

func (window *Window) SetSize(width int, height int) {
	GtkWindowResize(window.Handle, width, height)
}

func (window *Window) SetPosition(x int, y int) {
	GtkWindowMove(window.Handle, x, y)
}

func (window *Window) SetMinSize(width int, height int) {
	//TODO
	g := C.GdkGeometry{}
	g.min_width = C.int(width)
	g.min_height = C.int(height)
	GtkWindowSetGeometryHints(window.Handle, nil, &g, GDK_HINT_MIN_SIZE)
}

func (window *Window) SetMaxSize(width int, height int) {
	g := C.GdkGeometry{}
	g.max_width = C.int(width)
	g.max_height = C.int(height)
	GtkWindowSetGeometryHints(window.Handle, nil, &g, GDK_HINT_MAX_SIZE)
}

func (window *Window) SetAlwaysOnTop(always bool) {
	GtkWindowSetKeepAbove(window.Handle, always)
}

func (window *Window) Focus() {
	//TODO
	GtkWidgetGrabFocus(Window_GTK_WIDGET(window.Handle))
}

func (window *Window) IsVisible() bool {
	return GtkWidgetIsVisible(Window_GTK_WIDGET(window.Handle))
}

func (window *Window) SetMinimized(minimized bool) {
	if minimized {
		GtkWindowIconify(window.Handle)
	} else {
		GtkWindowDeiconify(window.Handle)
	}
}

func (window *Window) SetMaximized(maximized bool) {
	if maximized {
		GtkWindowMaximize(window.Handle)
	} else {
		GtkWindowUnmaximize(window.Handle)
	}
}

func (window *Window) SetFullscreen(fullscreen bool) {
	if fullscreen {
		GtkWindowFullscreen(window.Handle)
	} else {
		GtkWindowUnfullscreen(window.Handle)
	}
}

func (window *Window) Center() {
	size := window.GetSize()
	root := GdkScreenGetRootWindow(GdkScreenGetDefault())

	screenWidth := C.int(0)
	screenHeight := C.int(0)
	GdkWindowGetGeometry(root, nil, nil, &screenWidth, &screenHeight)

	nextPos := Position{
		X: (int(screenWidth) - size.Width) / 2,
		Y: (int(screenHeight) - size.Height) / 2,
	}

	window.SetPosition(nextPos.X, nextPos.Y)
}

func (window *Window) SetIconFromBytes(icon []byte) bool {
	//
	// @Cleanup: GTK doesn't seem to provide a way to load an icon from raw image bytes,
	// only from _parsed_ image pixels which is not exactly nice API composition
	//
	// https://docs.gtk.org/gdk-pixbuf/class.Pixbuf.html
	//
	f, err := os.CreateTemp("", "apptron__window_icon-*.png")
	if err != nil {
		log.Println("[SetIconFromBytes] Failed to create temporary icon file!")
		return false
	}

	_, err = f.Write(icon)
	if err != nil {
		log.Println("[SetIconFromBytes] Failed to create write icon bytes!")
		return false
	}

	iconPath := f.Name()
	cpath := C.CString(iconPath)
	defer LibCFree(unsafe.Pointer(cpath))

	buffer := GdkPixbufNewFromFile(cpath, nil)

	if buffer != nil {
		GtkWindowSetIcon(window.Handle, buffer)
		return true
	} else {
		log.Println("[SetIconFromBytes] Failed to load PixBuf from file!")
	}

	return false
}

// https://docs.gtk.org/gdk3/union.Event.html
// https://api.gtkd.org/gdk.c.types.GdkEventType.html

//export go_event_callback
//TODO callbacks
func go_event_callback(window *C.struct__GtkWindow, event *C.union__GdkEvent, arg C.int) {
	if globalEventCallback != nil {
		eventType := *(*C.int)(unsafe.Pointer(event))

		result := Event{}
		result.Window.Handle = window
		result.UserData = int(arg)

		if eventType == C.GDK_DELETE {
			result.Type = Delete
		}

		if eventType == C.GDK_DESTROY {
			result.Type = Destroy
		}

		if eventType == C.GDK_CONFIGURE {
			// NOTE(nick): Resize and move event
			configure := (*C.struct__GdkEventConfigure)(unsafe.Pointer(event))

			result.Type = Configure
			result.Position = Position{X: int(configure.x), Y: int(configure.y)}
			result.Size = Size{Width: int(configure.width), Height: int(configure.height)}
		}

		/*
			if eventType == C.GDK_FOCUS_CHANGE {
				focusChange := (*C.struct__GdkEventFocus)(unsafe.Pointer(event))

				result.Type = FocusChange
				result.FocusIn = fromCBool(C.int(focusChange.in))
			}
		*/

		//
		// NOTE(nick): window state change is similar to focus change,
		// but happens less frequently. for example, focus change is triggered
		// when dragging the window and when pressing super+tab (even if you navigate back)
		// to the same window
		//
		if eventType == C.GDK_WINDOW_STATE {
			windowState := (*C.struct__GdkEventWindowState)(unsafe.Pointer(event))

			// https://docs.gtk.org/gdk3/flags.WindowState.html
			if windowState.changed_mask&C.GDK_WINDOW_STATE_FOCUSED > 0 {
				focused := windowState.new_window_state&C.GDK_WINDOW_STATE_FOCUSED > 0

				result.Type = FocusChange
				result.FocusIn = focused
			}
		}

		if result.Type != None {
			globalEventCallback(result)
		}
	}
}

func (window *Window) BindEventCallback(userData int) {
	cevent := C.CString("event")
	defer LibCFree(unsafe.Pointer(cevent))

	//TODO
	C._g_signal_connect(Window_GTK_WIDGET(window.Handle), cevent, C.go_event_callback, C.int(userData))
}

func SetGlobalEventCallback(callback Event_Callback) {
	globalEventCallback = callback
}

func (webview *Webview) RegisterCallback(name string, callback func(result string)) int {
	manager := WebkitWebViewGetUserContentManager(webview.Handle)

	cevent := C.CString(fmt.Sprintf("script-message-received::%s", name))
	defer LibCFree(unsafe.Pointer(cevent))

	cexternal := C.CString(name)
	defer LibCFree(unsafe.Pointer(cexternal))

	index := wc_register(callback)
	C._g_signal_connect(WebKitUserContentManager_GTK_WIDGET(manager), cevent, C.go_webview_callback, C.int(index))
	WebkitUserContentManagerRegisterScriptMessageHandler(manager, cexternal)

	return int(index)
}

func (webview *Webview) UnregisterCallback(callback int) {
	// @Incomplete: remove script handler

	wc_unregister(callback)
}

func (webview *Webview) Destroy() {
	if webview.Handle != nil {
		GtkWidgetDestroy(Webview_GTK_WIDGET(webview.Handle))
		webview.Handle = nil
	}
}

func DefaultWebviewSettings() WebviewSetings {
	result := WebviewSetings{}
	result.CanAccessClipboard = true
	result.WriteConsoleToStdout = true
	result.DeveloperTools = true
	return result
}

func (webview *Webview) SetSettings(config WebviewSetings) {
	settings := WebkitWebViewGetSettings(webview.Handle)

	WebkitSettingsSetJavascriptCanAccessClipboard(settings, config.CanAccessClipboard)
	WebkitSettingsSetEnableWriteConsoleMessagesToStdout(settings, config.WriteConsoleToStdout)
	WebkitSettingsSetEnableDeveloperExtras(settings, config.DeveloperTools)
}

func (webview *Webview) Eval(js string) {
	cjs := C.CString(js)
	defer LibCFree(unsafe.Pointer(cjs))

	WebkitWebViewEvaluateJavascript(webview.Handle, cjs, len(js), nil, nil, nil, nil, nil)
}

func (webview *Webview) SetHtml(html string, baseUri string) {
	chtml := C.CString(html)
	defer LibCFree(unsafe.Pointer(chtml))

	cbaseUri := C.CString(baseUri)
	defer LibCFree(unsafe.Pointer(cbaseUri))

	WebkitWebViewLoadHtml(webview.Handle, chtml, cbaseUri)
}

func (webview *Webview) Navigate(url string) {
	curl := C.CString(url)
	defer LibCFree(unsafe.Pointer(curl))

	WebkitWebViewLoadUri(webview.Handle, curl)
}

func (webview *Webview) AddScript(js string) {
	manager := WebkitWebViewGetUserContentManager(webview.Handle)

	cjs := C.CString(js)
	defer LibCFree(unsafe.Pointer(cjs))

	script := WebkitUserScriptNew(
		cjs,
		WEBKIT_USER_CONTENT_INJECT_TOP_FRAME,
		WEBKIT_USER_SCRIPT_INJECT_AT_DOCUMENT_START,
		nil,
		nil,
	)

	WebkitUserContentManagerAddScript(manager, script)
}

func (webview *Webview) SetTransparent(transparent bool) {
	C.gtk_webview_set_transparent(webview.Handle, toCBool(transparent))
}

// https://docs.gtk.org/gdk3/class.Monitor.html

func Monitors() []Monitor {
	// @Incomplete: Should this be gdk_display_manager_list_displays instead?
	/*
		displays := C.gdk_display_manager_list_displays(C.gdk_display_manager_get())
		C.g_slist_free(displays)
	*/

	display := GdkDisplayGetDefault()
	if display == nil {
		return make([]Monitor, 0)
	}

	n := int(GdkDisplayGetNMonitors(display))

	result := make([]Monitor, n)

	for i := 0; i < n; i++ {
		monitor := GdkDisplayGetMonitor(display, C.int(i))

		result[i] = Monitor{
			Handle: monitor,
		}
	}

	return result
}

func (monitor *Monitor) Geometry() Rectangle {
	rect := C.GdkRectangle{}
	GdkMonitorGetGeometry(monitor.Handle, &rect)

	return Rectangle{
		Position: Position{X: int(rect.x), Y: int(rect.y)},
		Size:     Size{Width: int(rect.width), Height: int(rect.height)},
	}
}

func (monitor *Monitor) ScaleFactor() int {
	return int(GdkMonitorGetScaleFactor(monitor.Handle))
}

func (monitor *Monitor) Name() string {
	manufacturer := C.GoString(GdkMonitorGetManufacturer(monitor.Handle))
	model := C.GoString(GdkMonitorGetModel(monitor.Handle))
	return manufacturer + " " + model
}

func (monitor *Monitor) RefreshRate() int {
	// NOTE(nick): in milli-Hertz (60Hz = 60000)
	return int(GdkMonitorGetRefreshRate(monitor.Handle)) / 1000
}

func (monitor *Monitor) IsPrimary() bool {
	return GdkMonitorIsPrimary(monitor.Handle)
}

//
// Indicator
//

func Indicator_New(id string, pngIconPath string, menu Menu) Indicator {
	cid := C.CString(id)
	defer LibCFree(unsafe.Pointer(cid))

	handle := C.app_indicator_new(cid, C.CString(""), C.APP_INDICATOR_CATEGORY_APPLICATION_STATUS)
	C.app_indicator_set_status(handle, C.APP_INDICATOR_STATUS_ACTIVE)

	//app_indicator_set_title(global_app_indicator, title);
	//app_indicator_set_label(global_app_indicator, title, "");

	if len(pngIconPath) > 0 {
		cIconPath := C.CString(pngIconPath)
		defer LibCFree(unsafe.Pointer(cIconPath))

		C.app_indicator_set_icon_full(handle, cIconPath, C.CString(""))
	}

	if menu.Handle != nil {
		C.app_indicator_set_menu(handle, menu.Handle)
	}

	result := Indicator{}
	result.Handle = handle
	return result
}

func Menu_New() Menu {
	result := Menu{}
	result.Handle = Menu_FromWidget(GtkMenuNew())
	return result
}

func (menu *Menu) Destroy() {
	if menu.Handle != nil {
		GtkWidgetDestroy(Menu_GTK_WIDGET(menu.Handle))
		menu.Handle = nil
	}
}

func MenuItem_New(id int, title string, disabled bool, checked bool, separator bool) MenuItem {
	var widget *C.struct__GtkWidget = nil

	if separator {
		widget = GtkSeparatorMenuItemNew()
		GtkWidgetShow(widget)
	} else {
		ctitle := C.CString(title)
		defer LibCFree(unsafe.Pointer(ctitle))

		if checked {
			widget = GtkCheckMenuItemNewWithLabel(ctitle)

			GtkCheckMenuItemSetActive(CheckMenuItem_FromWidget(widget), checked)
		} else {
			widget = GtkCheckMenuItemNewWithLabel(ctitle)
		}

		GtkWidgetSetSensitive(widget, !disabled)

		//
		// NOTE(nick): accelerators seem to require a window and an accel_group
		// Are they even supported in the AppIndicator?
		// As far as I can tell they don't ever show up in the AppIndicator menu...
		//
		// @see https://github.com/bstpierre/gtk-examples/blob/master/c/accel.c
		//
		/*
		   GtkWindow *window = gtk_window_new(GTK_WINDOW_TOPLEVEL);
		   GtkAccelGroup *accel_group = gtk_accel_group_new();
		   gtk_window_add_accel_group(GTK_WINDOW(window), accel_group);

		   gtk_widget_add_accelerator(item, "activate", accel_group, GDK_KEY_F7, 0, GTK_ACCEL_VISIBLE);
		*/

		cactivate := C.CString("activate")
		defer LibCFree(unsafe.Pointer(cactivate))

		C._g_signal_connect(widget, cactivate, C.go_menu_callback, C.int(id))

		GtkWidgetShow(widget)
	}

	result := MenuItem{}
	result.Handle = MenuItem_FromWidget(widget)
	return result
}

func (menu *Menu) AppendItem(item MenuItem) {
	GtkMenuShellAppend(Menu_GTK_MENU_SHELL(menu.Handle), MenuItem_GTK_WIDGET(item.Handle))
}

func (item *MenuItem) SetSubmenu(child Menu) {
	GtkMenuItemSetSubmenu(item.Handle, Menu_GTK_WIDGET(child.Handle))
}

//export go_menu_callback
func go_menu_callback(item *C.struct__GtkMenuItem, menuId C.int) {
	if globalMenuCallback != nil {
		globalMenuCallback(int(menuId))
	}
}

func SetGlobalMenuCallback(callback Menu_Callback) {
	globalMenuCallback = callback
}

//
// Callbacks
//

type Webview_Callback func(str string)

var wc_mu sync.Mutex
var wc_index int
var wc_fns = make(map[int]Webview_Callback)

func wc_register(fn Webview_Callback) int {
	wc_mu.Lock()
	defer wc_mu.Unlock()
	wc_index++
	for wc_fns[wc_index] != nil {
		wc_index++
	}
	wc_fns[wc_index] = fn
	return wc_index
}

func wc_lookup(i int) Webview_Callback {
	wc_mu.Lock()
	defer wc_mu.Unlock()
	return wc_fns[i]
}

func wc_unregister(i int) {
	wc_mu.Lock()
	defer wc_mu.Unlock()
	delete(wc_fns, i)
}

//export go_webview_callback
func go_webview_callback(manager *C.struct__WebKitUserContentManager, result *C.struct__WebKitJavascriptResult, arg C.int) {
	fn := wc_lookup(int(arg))
	cstr := C.string_from_js_result(result)
	if fn != nil {
		fn(C.GoString(cstr))
	}
	C.g_free((C.gpointer)(unsafe.Pointer(cstr)))
}

func OS_GetClipboardText() string {
	//TODO Resolve this macro
	clipboard := GtkClipboardGet(C.GDK_SELECTION_CLIPBOARD)
	text := GtkClipboardWaitForText(clipboard)

	return C.GoString(text)
}

func OS_SetClipboardText(text string) bool {
	ctext := C.CString(text)
	defer LibCFree(unsafe.Pointer(ctext))

	clipboard := GtkClipboardGet(C.GDK_SELECTION_CLIPBOARD)

	GtkClipboardSetText(clipboard, ctext, -1)

	// @Incomplete: is there a way to check if set_text succeeded?
	return true
}

//
// Helpers
//

func toCBool(value bool) C.int {
	if value {
		return C.int(1)
	}
	return C.int(0)
}

func fromCBool(value C.int) bool {
	if int(value) == 0 {
		return false
	}

	return true
}

func Menu_GTK_WIDGET(it *C.struct__GtkMenu) *C.struct__GtkWidget {
	return (*C.struct__GtkWidget)(unsafe.Pointer(it))
}

func Menu_FromWidget(it *C.struct__GtkWidget) *C.struct__GtkMenu {
	return (*C.struct__GtkMenu)(unsafe.Pointer(it))
}

func Menu_GTK_MENU_SHELL(it *C.struct__GtkMenu) *C.struct__GtkMenuShell {
	return (*C.struct__GtkMenuShell)(unsafe.Pointer(it))
}

func MenuItem_GTK_WIDGET(it *C.struct__GtkMenuItem) *C.struct__GtkWidget {
	return (*C.struct__GtkWidget)(unsafe.Pointer(it))
}

func MenuItem_FromWidget(it *C.struct__GtkWidget) *C.struct__GtkMenuItem {
	return (*C.struct__GtkMenuItem)(unsafe.Pointer(it))
}

func CheckMenuItem_FromWidget(it *C.struct__GtkWidget) *C.struct__GtkCheckMenuItem {
	return (*C.struct__GtkCheckMenuItem)(unsafe.Pointer(it))
}

func Window_FromWidget(it *C.struct__GtkWidget) *C.struct__GtkWindow {
	return (*C.struct__GtkWindow)(unsafe.Pointer(it))
}

func Webview_FromWidget(it *C.struct__GtkWidget) *C.struct__WebKitWebView {
	return (*C.struct__WebKitWebView)(unsafe.Pointer(it))
}

func Window_GTK_WIDGET(it *C.struct__GtkWindow) *C.struct__GtkWidget {
	return (*C.struct__GtkWidget)(unsafe.Pointer(it))
}

func Window_GTK_CONTAINER(it *C.struct__GtkWindow) *C.struct__GtkContainer {
	return (*C.struct__GtkContainer)(unsafe.Pointer(it))
}

func Webview_GTK_WIDGET(it *C.struct__WebKitWebView) *C.struct__GtkWidget {
	return (*C.struct__GtkWidget)(unsafe.Pointer(it))
}

func WebKitUserContentManager_GTK_WIDGET(it *C.struct__WebKitUserContentManager) *C.struct__GtkWidget {
	return (*C.struct__GtkWidget)(unsafe.Pointer(it))
}
