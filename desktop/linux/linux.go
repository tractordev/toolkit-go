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

#include <stdint.h>
#include <gtk/gtk.h>
#include <JavaScriptCore/JavaScript.h>
#include <webkit2/webkit2.h>
#include <libayatana-appindicator/app-indicator.h>
#include <string.h>
*/
import "C"

type Window struct {
	Handle *C.GtkWindow
}

type Webview struct {
	Handle *C.WebKitWebView
}

type Menu struct {
	Handle *C.GtkMenu
}

type MenuItem struct {
	Handle *C.GtkMenuItem
}

type Indicator struct {
	Handle *C.AppIndicator
}

type Monitor struct {
	Handle *C.GdkMonitor
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
	GSignalConnectData func (
		instance *C.GtkWidget,
		detailed_signal string,
		c_handler uintptr,
		data unsafe.Pointer,
		//These two last arguments are ignored, so they only have generic types here
		destroy_data unsafe.Pointer,
		connect_flags uint32,
	)
)

var (
	//TODO put these in lexical order
	//TODO indentation
	//TODO transfer types as well
	GtkMain 						  func ()
	GtkInitCheck 					  func (argc unsafe.Pointer, argv unsafe.Pointer)
	GtkMainIterationDo 				  func (blocking bool)
	GtkWindowNew 					  func (window_type uint32) *C.GtkWidget
	GtkContainerAdd 				  func (container *C.GtkContainer, widget *C.GtkWidget)
	GtkWidgetGrabFocus 				  func (widget *C.GtkWidget)
	GtkWidgetShowAll 				  func (widget *C.GtkWidget)
	GtkWidgetHide   				  func (widget *C.GtkWidget)
	GtkWidgetDestroy 				  func (widget *C.GtkWidget)
	GtkWindowSetDecorated 			  func (window *C.GtkWindow, setting bool)
	GtkWindowGetSize 				  func (window *C.GtkWindow, width *C.int, height *C.int)
	GtkCheckMenuItemNewWithLabel      func (label string) *C.GtkWidget
    GtkCheckMenuItemSetActive         func (checkMenuItem *C.GtkCheckMenuItem, is_active bool)
	//GdkAtom is basically a pointer to 'struct _GdkAtom' (gdktypes.h)
    GtkClipboardGet                   func (selection C.GdkAtom) *C.GtkClipboard
    GtkClipboardSetText               func (clipboard *C.GtkClipboard, text string, length int)
    GtkClipboardWaitForText           func (clipboard *C.GtkClipboard) string
    GtkMenuItemNewWithLabel           func (label string) *C.GtkWidget
    GtkMenuItemSetSubmenu             func (menu_item *C.GtkMenuItem, submenu *C.GtkWidget)
    GtkMenuNew                        func () *C.GtkWidget
    GtkMenuShellAppend                func (menu_shell *C.GtkMenuShell, child *C.GtkWidget)
    GtkSeparatorMenuItemNew           func () *C.GtkWidget
    GtkWidgetIsVisible                func (widget *C.GtkWidget) bool
    GtkWidgetSetSensitive             func (widget *C.GtkWidget, sensitive bool)
    GtkWidgetShow                     func (widget *C.GtkWidget)
    GtkWindowDeiconify                func (window *C.GtkWindow)
    GtkWindowIconify                  func (window *C.GtkWindow)
    GtkWindowFullscreen               func (window *C.GtkWindow)
    GtkWindowUnfullscreen             func (window *C.GtkWindow)
    GtkWindowGetPosition              func (window *C.GtkWindow, x, y *C.int)
    GtkWindowMaximize                 func (window *C.GtkWindow)
    GtkWindowUnmaximize               func (window *C.GtkWindow)
    GtkWindowMove                     func (window *C.GtkWindow, x, y int)
    GtkWindowResize                   func (window *C.GtkWindow, width, height int)
    GtkWindowSetGeometryHints         func (
		window *C.GtkWindow,
		geometry_widget *C.GtkWidget,
		geometry *C.GdkGeometry,
		geom_mask uint32)
    GtkWindowSetIcon                  func (window *C.GtkWindow, icon *C.GdkPixbuf)
    GtkWindowSetKeepAbove             func (window *C.GtkWindow, setting bool)
    GtkWindowSetResizable             func (window *C.GtkWindow, resizable bool)
    GtkWindowSetTitle                 func (window *C.GtkWindow, title string)
	GtkWidgetSetAppPaintable          func(window *C.GtkWidget, app_paintable bool)
	GtkWidgetSetVisual          	  func(window *C.GtkWindow, visual *C.GdkVisual)
)

var (
    GdkScreenGetRootWindow            func (window *C.GdkScreen) *C.GdkWindow
	GdkScreenGetDefault				  func () *C.GdkScreen
	GdkDisplayGetDefault          	  func() *C.GdkDisplay
    GdkDisplayGetMonitor          	  func(display *C.GdkDisplay, monitorNum C.int) *C.GdkMonitor
    GdkDisplayGetNMonitors        	  func(display *C.GdkDisplay) C.int
	//GdkRectangle is not always a struct, and in this case it's a typedef to a definition in cairo
	//so C. prefix is not used here
	//TODO once cgo is completely out, these defitions will follow. So this won't even matter, probably.
    GdkMonitorGetGeometry         	  func(monitor *C.GdkMonitor, rect *C.GdkRectangle)
    GdkMonitorGetManufacturer      	  func(monitor *C.GdkMonitor) string
    GdkMonitorGetModel                func(monitor *C.GdkMonitor) string
    GdkMonitorGetRefreshRate          func(monitor *C.GdkMonitor) int
    GdkMonitorGetScaleFactor          func(monitor *C.GdkMonitor) int
    GdkMonitorIsPrimary               func(monitor *C.GdkMonitor) bool
    GdkPixbufNewFromFile          	  func(filename string, err **C.GError) *C.GdkPixbuf
    GdkWindowGetGeometry          	  func(window *C.GdkWindow, x, y, width, height *C.int)
	GdkScreenGetRgbaVisual            func(window *C.GdkScreen) *C.GdkVisual
	GdkScreenIsComposited             func(screen *C.GdkScreen) bool
)

var (
    WebkitSettingsSetEnableDeveloperExtras                  func(settings *C.WebKitSettings, enable bool)
    WebkitSettingsSetEnableWriteConsoleMessagesToStdout   	func(settings *C.WebKitSettings, enable bool)
    WebkitSettingsSetJavascriptCanAccessClipboard           func(settings *C.WebKitSettings, enable bool)
    WebkitUserContentManagerAddScript                       func(manager *C.WebKitUserContentManager, script *C.WebKitUserScript)
    WebkitUserContentManagerRegisterScriptMessageHandler    func(manager *C.WebKitUserContentManager, name string) bool
    WebkitUserScriptNew                                     func(
		source string,
	 	injected_frames uint32,
	 	injected_time uint32,
		allow_list []string,
		block_list []string,
	) *C.WebKitUserScript
    WebkitWebViewEvaluateJavascript                         func(
		web_view *C.WebKitWebView,
		script string,
		length int,
		//ignoring these atm, so they don't have string type
		world_name unsafe.Pointer,
		source_uri unsafe.Pointer,
		cancellable *C.GCancellable,
		callback *C.GAsyncReadyCallback,
		user_data unsafe.Pointer,
	)
    WebkitWebViewGetSettings                                func(web_view *C.WebKitWebView) *C.WebKitSettings
    WebkitWebViewGetUserContentManager                      func(web_view *C.WebKitWebView) *C.WebKitUserContentManager
    WebkitWebViewLoadHtml                                   func(web_view *C.WebKitWebView, content string, base_uri string)
    WebkitWebViewLoadUri                                    func(web_view *C.WebKitWebView, uri string)
    WebkitWebViewNew                                        func() *C.GtkWidget
	WebkitWebViewSetBackgroundColor							func(web_view *C.WebKitWebView, rgba *C.GdkRGBA)
	WebkitJavascriptResultGetJsValue						func(js_result *C.WebKitJavascriptResult) *C.JSCValue
)

var (
	JscValueToString		func (*C.JSCValue) string
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

func GetJSCLibPath() string {
	return "/usr/lib/libjavascriptcoregtk-4.1.so"
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

	purego.RegisterLibFunc(&GSignalConnectData, libgtk, "g_signal_connect_data")

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
	purego.RegisterLibFunc(&GtkWidgetSetAppPaintable, libgtk, "gtk_widget_set_app_paintable")
	purego.RegisterLibFunc(&GtkWidgetSetVisual, libgtk, "gtk_widget_set_visual")

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
	purego.RegisterLibFunc(&GdkScreenGetRgbaVisual, libgtk, "gdk_screen_get_rgba_visual")
	purego.RegisterLibFunc(&GdkScreenIsComposited, libgtk, "gdk_screen_is_composited")


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
	purego.RegisterLibFunc(&WebkitWebViewSetBackgroundColor, libwebgtk, "webkit_web_view_set_background_color")
	purego.RegisterLibFunc(&WebkitJavascriptResultGetJsValue, libwebgtk, "webkit_javascript_result_get_js_value")

	libjsc, err := purego.Dlopen(GetJSCLibPath(), purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		panic(err)
	}

	//LibJavascriptCore functions
	purego.RegisterLibFunc(&JscValueToString, libwebgtk, "jsc_value_to_string")

	//TODO where and when to close files
	purego.Dlclose(libc)
	purego.Dlclose(libgtk)
	purego.Dlclose(libwebgtk)
	purego.Dlclose(libjsc)
}

/*
* User defined Gtk functions
*/
func GtkWindowSetTransparent(window *C.GtkWindow, transparent bool) {
	fmt.Println("setting transparent")
	if transparent {
		GtkWidgetSetAppPaintable(Window_GTK_WIDGET(window), true)
		screen := GdkScreenGetDefault()
		visual := GdkScreenGetRgbaVisual(screen)

		if visual != nil && GdkScreenIsComposited(screen) {
			GtkWidgetSetVisual(window, visual)
		}
	} else {
		GtkWidgetSetAppPaintable(Window_GTK_WIDGET(window), false)
		GtkWidgetSetVisual(window, nil)
	}
}

func GtkWebViewSetTransparent(webview *C.WebKitWebView, transparent bool) {
	color := C.GdkRGBA{}
	color.red = 1.0
	color.green = 1.0
	color.blue = 1.0
	color.alpha = 1.0

	if transparent {
		color.alpha = 0
	}
}

func StringFromJsResult(result *C.WebKitJavascriptResult) string {
	value := WebkitJavascriptResultGetJsValue(result)
	return JscValueToString(value)
}

// A simple go implementation of `g_signal_connect`
func g_signal_connect(
	instance *C.GtkWidget,
	detailed_signal string,
	c_handler uintptr,
	data unsafe.Pointer,
) {
	GSignalConnectData(instance, detailed_signal, c_handler, data, nil, 0);
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

func (window *Window) SetTransparent(transparent bool) {
	GtkWindowSetTransparent(window.Handle, transparent)
}

func (window *Window) SetTitle(title string) {
	GtkWindowSetTitle(window.Handle, title)
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
	buffer := GdkPixbufNewFromFile(iconPath, nil)

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

func go_event_callback(window *C.GtkWindow, event *C.GdkEvent, arg C.int) {
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
			configure := (*C.GdkEventConfigure)(unsafe.Pointer(event))

			result.Type = Configure
			result.Position = Position{X: int(configure.x), Y: int(configure.y)}
			result.Size = Size{Width: int(configure.width), Height: int(configure.height)}
		}

		/*
			if eventType == C.GDK_FOCUS_CHANGE {
				focusChange := (*C.GdkEventFocus)(unsafe.Pointer(event))

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
			windowState := (*C.GdkEventWindowState)(unsafe.Pointer(event))

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
	g_signal_connect(Window_GTK_WIDGET(window.Handle), "event", purego.NewCallback(go_event_callback), unsafe.Pointer(&userData))
}

func SetGlobalEventCallback(callback Event_Callback) {
	globalEventCallback = callback
}

func (webview *Webview) RegisterCallback(name string, callback func(result string)) int {
	manager := WebkitWebViewGetUserContentManager(webview.Handle)

	event := fmt.Sprintf("script-message-received::%s", name)

	index := wc_register(callback)
	g_signal_connect(WebKitUserContentManager_GTK_WIDGET(manager), event, purego.NewCallback(go_webview_callback), unsafe.Pointer(&index))
	WebkitUserContentManagerRegisterScriptMessageHandler(manager, name)

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
	WebkitWebViewEvaluateJavascript(webview.Handle, js, len(js), nil, nil, nil, nil, nil)
}

func (webview *Webview) SetHtml(html string, baseUri string) {
	WebkitWebViewLoadHtml(webview.Handle, html, baseUri)
}

func (webview *Webview) Navigate(url string) {
	WebkitWebViewLoadUri(webview.Handle, url)
}

func (webview *Webview) AddScript(js string) {
	manager := WebkitWebViewGetUserContentManager(webview.Handle)

	script := WebkitUserScriptNew(
		js,
		WEBKIT_USER_CONTENT_INJECT_TOP_FRAME,
		WEBKIT_USER_SCRIPT_INJECT_AT_DOCUMENT_START,
		nil,
		nil,
	)

	WebkitUserContentManagerAddScript(manager, script)
}

func (webview *Webview) SetTransparent(transparent bool) {
	GtkWebViewSetTransparent(webview.Handle, transparent)
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
	manufacturer := GdkMonitorGetManufacturer(monitor.Handle)
	model := GdkMonitorGetModel(monitor.Handle)
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

//TODO
//func Indicator_New(id string, pngIconPath string, menu Menu) Indicator {
//	cid := C.CString(id)
//	defer LibCFree(unsafe.Pointer(cid))
//
//	handle := C.app_indicator_new(cid, C.CString(""), C.APP_INDICATOR_CATEGORY_APPLICATION_STATUS)
//	C.app_indicator_set_status(handle, C.APP_INDICATOR_STATUS_ACTIVE)
//
//	//app_indicator_set_title(global_app_indicator, title);
//	//app_indicator_set_label(global_app_indicator, title, "");
//
//	if len(pngIconPath) > 0 {
//		cIconPath := C.CString(pngIconPath)
//		defer LibCFree(unsafe.Pointer(cIconPath))
//
//		C.app_indicator_set_icon_full(handle, cIconPath, C.CString(""))
//	}
//
//	if menu.Handle != nil {
//		C.app_indicator_set_menu(handle, menu.Handle)
//	}
//
//	result := Indicator{}
//	result.Handle = handle
//	return result
//}

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
	var widget *C.GtkWidget = nil

	if separator {
		widget = GtkSeparatorMenuItemNew()
		GtkWidgetShow(widget)
	} else {
		if checked {
			widget = GtkCheckMenuItemNewWithLabel(title)

			GtkCheckMenuItemSetActive(CheckMenuItem_FromWidget(widget), checked)
		} else {
			widget = GtkCheckMenuItemNewWithLabel(title)
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

		g_signal_connect(widget, "activate", purego.NewCallback(go_menu_callback), unsafe.Pointer(&id))

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

func go_menu_callback(item *C.GtkMenuItem, menuId C.int) {
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

func go_webview_callback(manager *C.WebKitUserContentManager, result *C.WebKitJavascriptResult, arg C.int) {
	fn := wc_lookup(int(arg))
	str := StringFromJsResult(result)
	if fn != nil {
		fn(str)
	}
}

func OS_GetClipboardText() string {
	//TODO Resolve this macro
	clipboard := GtkClipboardGet(C.GDK_SELECTION_CLIPBOARD)
	return GtkClipboardWaitForText(clipboard)
}

func OS_SetClipboardText(text string) bool {
	clipboard := GtkClipboardGet(C.GDK_SELECTION_CLIPBOARD)

	GtkClipboardSetText(clipboard, text, -1)

	// @Incomplete: is there a way to check if set_text succeeded?
	return true
}

//
// Helpers
//

func Menu_GTK_WIDGET(it *C.GtkMenu) *C.GtkWidget {
	return (*C.GtkWidget)(unsafe.Pointer(it))
}

func Menu_FromWidget(it *C.GtkWidget) *C.GtkMenu {
	return (*C.GtkMenu)(unsafe.Pointer(it))
}

func Menu_GTK_MENU_SHELL(it *C.GtkMenu) *C.GtkMenuShell {
	return (*C.GtkMenuShell)(unsafe.Pointer(it))
}

func MenuItem_GTK_WIDGET(it *C.GtkMenuItem) *C.GtkWidget {
	return (*C.GtkWidget)(unsafe.Pointer(it))
}

func MenuItem_FromWidget(it *C.GtkWidget) *C.GtkMenuItem {
	return (*C.GtkMenuItem)(unsafe.Pointer(it))
}

func CheckMenuItem_FromWidget(it *C.GtkWidget) *C.GtkCheckMenuItem {
	return (*C.GtkCheckMenuItem)(unsafe.Pointer(it))
}

func Window_FromWidget(it *C.GtkWidget) *C.GtkWindow {
	return (*C.GtkWindow)(unsafe.Pointer(it))
}

func Webview_FromWidget(it *C.GtkWidget) *C.WebKitWebView {
	return (*C.WebKitWebView)(unsafe.Pointer(it))
}

func Window_GTK_WIDGET(it *C.GtkWindow) *C.GtkWidget {
	return (*C.GtkWidget)(unsafe.Pointer(it))
}

func Window_GTK_CONTAINER(it *C.GtkWindow) *C.GtkContainer {
	return (*C.GtkContainer)(unsafe.Pointer(it))
}

func Webview_GTK_WIDGET(it *C.WebKitWebView) *C.GtkWidget {
	return (*C.GtkWidget)(unsafe.Pointer(it))
}

func WebKitUserContentManager_GTK_WIDGET(it *C.WebKitUserContentManager) *C.GtkWidget {
	return (*C.GtkWidget)(unsafe.Pointer(it))
}
