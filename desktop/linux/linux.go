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

type Window struct {
	Handle  unsafe.Pointer
	MaxSize Size
	MinSize Size
}

type Webview struct {
	Handle unsafe.Pointer
}

type Menu struct {
	Handle unsafe.Pointer
}

type MenuItem struct {
	Handle unsafe.Pointer
}

type Indicator struct {
	Handle unsafe.Pointer
}

type Monitor struct {
	Handle unsafe.Pointer
}

type Size struct {
	Width  int32
	Height int32
}

type Position struct {
	X int32
	Y int32
}

type Rectangle struct {
	Position Position
	Size     Size
}

type GdkGeometry struct {
	min_width   int32
	min_height  int32
	max_width   int32
	max_height  int32
	base_width  int32
	base_height int32
	width_inc   int32
	height_inc  int32
	min_aspect  float64
	max_aspect  float64
	win_gravity uint32
	_           [4]byte
}

type GdkRectangle struct {
	x      int32
	y      int32
	width  int32
	height int32
}

type GdkRGBA struct {
	red   int32
	green int32
	blue  int32
	alpha int32
}

type GdkEventConfigure struct {
	_type      int32
	window     unsafe.Pointer
	send_event int8
	x          int32
	y          int32
	width      int32
	height     int32
	_          [4]byte
}

type GdkEventWindowState struct {
	_type            int32
	window           unsafe.Pointer
	send_event       int8
	changed_mask     uint32
	new_window_state uint32
	_                [4]byte
}

type EventType int32

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
	UserData int32

	Position Position
	Size     Size
	FocusIn  bool
}

type Menu_Callback func(menuId int32)

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
	GtkWindowToplevel = iota
	GtkWindowPopup
)

// GdkWindowHints
const (
	GdkHintPos        = 1 << 0
	GdkHintMinSize    = 1 << 1
	GdkHintMaxSize    = 1 << 2
	GdkHintBaseSize   = 1 << 3
	GdkHintAspect     = 1 << 4
	GdkHintResizeInc  = 1 << 5
	GdkHintWinGravity = 1 << 6
	GdkHintUserPos    = 1 << 7
	GdkHintUserSize   = 1 << 8
)

// GdkWindowState
const (
	GdkWindowStateWithdrawn       = 1 << 0
	GdkWindowStateIconified       = 1 << 1
	GdkWindowStateMaximized       = 1 << 2
	GdkWindowStateSticky          = 1 << 3
	GdkWindowStateFullscreen      = 1 << 4
	GdkWindowStateAbove           = 1 << 5
	GdkWindowStateBelow           = 1 << 6
	GdkWindowStateFocused         = 1 << 7
	GdkWindowStateTiled           = 1 << 8
	GdkWindowStateTopTiled        = 1 << 9
	GdkWindowStateTopResizable    = 1 << 10
	GdkWindowStateRightTiled      = 1 << 11
	GdkWindowStateRightResizable  = 1 << 12
	GdkWindowStateBottomTiled     = 1 << 13
	GdkWindowStateBottomResizable = 1 << 14
	GdkWindowStateLeftTiled       = 1 << 15
	GdkWindowStateLeftResizable   = 1 << 16
)

// WebKitUserContentInjectedFrames
const (
	WebkitUserContentInjectAllFrames = iota
	WebkitUserContentInjectTopFrame
)

// WebKitUserScriptInjectionTime
const (
	WebkitUserScriptInjectAtDocumentStart = iota
	WebkitUserScriptInjectAtDocumentEnd
)

// AppIndicatorCategory
const (
	AppIndicatorCategoryApplicationStatus = iota
	AppIndicatorCategoryCommunications
	AppIndicatorCategorySystemServices
	AppIndicatorCategoryHardware
	AppIndicatorCategoryOther
)

// AppIndicatorStatus
const (
	AppIndicatorStatusPassive = iota
	AppIndicatorStatusActive
	AppIndicatorStatusAttention
)

// GdkEventType
const (
	GdkNothing           = -1
	GdkDelete            = 0
	GdkDestroy           = 1
	GdkExpose            = 2
	GdkMotionNotify      = 3
	GdkButtonPress       = 4
	Gdk2ButtonPress      = 5
	GdkDoubleButtonPress = Gdk2ButtonPress
	Gdk3ButtonPress      = 6
	GdkTripleButtonPress = Gdk3ButtonPress
	GdkButtonRelease     = 7
	GdkKeyPress          = 8
	GdkKeyRelease        = 9
	GdkEnterNotify       = 10
	GdkLeaveNotify       = 11
	GdkFocusChange       = 12
	GdkConfigure         = 13
	GdkMap               = 14
	GdkUnmap             = 15
	GdkPropertyNotify    = 16
	GdkSelectionClear    = 17
	GdkSelectionRequest  = 18
	GdkSelectionNotify   = 19
	GdkProximityIn       = 20
	GdkProximityOut      = 21
	GdkDragEnter         = 22
	GdkDragLeave         = 23
	GdkDragMotion        = 24
	GdkDragStatus        = 25
	GdkDropStart         = 26
	GdkDropFinished      = 27
	GdkClientEvent       = 28
	GdkVisibilityNotify  = 29
	GdkScroll            = 31
	GdkWindowState       = 32
	GdkSetting           = 33
	GdkOwnerChange       = 34
	GdkGrabBroken        = 35
	GdkDamage            = 36
	GdkTouchBegin        = 37
	GdkTouchUpdate       = 38
	GdkTouchEnd          = 39
	GdkTouchCancel       = 40
	GdkTouchpadSwipe     = 41
	GdkTouchpadPinch     = 42
	GdkPadButtonPress    = 43
	GdkPadButtonRelease  = 44
	GdkPadRing           = 45
	GdkPadStrip          = 46
	GdkPadGroupMode      = 47
	GdkEventLast         = 48
)

/*
* PureGo Gtk Bindings
 */

var (
	LibCFree func(unsafe.Pointer)
)

var (
	GSignalConnectData func(
		instance unsafe.Pointer,
		detailed_signal string,
		c_handler uintptr,
		data unsafe.Pointer,
		//These two last arguments are ignored, so they only have generic types here
		destroy_data unsafe.Pointer,
		connect_flags uint32,
	)
)

var (
	//TODO add parameter types in docs
	GtkCheckMenuItemNewWithLabel func(label string) unsafe.Pointer
	GtkCheckMenuItemSetActive    func(checkMenuItem unsafe.Pointer, isActive bool)
	//GdkAtom is basically a pointer to 'struct _GdkAtom' (gdktypes.h)
	//unsafe.Pointer is used in its place and gdk_atom_intern will be used to get the atoms for strings
	//here since there is no way to use macros like GDK_SELECTION_CLIPBOARD
	GtkClipboardGet           func(selection unsafe.Pointer) unsafe.Pointer
	GtkClipboardSetText       func(clipboard unsafe.Pointer, text string, length int32)
	GtkClipboardWaitForText   func(clipboard unsafe.Pointer) string
	GtkContainerAdd           func(container unsafe.Pointer, widget unsafe.Pointer)
	GtkInitCheck              func(argc unsafe.Pointer, argv unsafe.Pointer)
	GtkMain                   func()
	GtkMainIterationDo        func(blocking bool)
	GtkMenuItemNewWithLabel   func(label string) unsafe.Pointer
	GtkMenuItemSetSubmenu     func(menuItem unsafe.Pointer, submenu unsafe.Pointer)
	GtkMenuNew                func() unsafe.Pointer
	GtkMenuShellAppend        func(menuShell unsafe.Pointer, child unsafe.Pointer)
	GtkSeparatorMenuItemNew   func() unsafe.Pointer
	GtkWidgetDestroy          func(widget unsafe.Pointer)
	GtkWidgetGrabFocus        func(widget unsafe.Pointer)
	GtkWidgetGetWindow        func(widget unsafe.Pointer) unsafe.Pointer
	GtkWidgetHide             func(widget unsafe.Pointer)
	GtkWidgetIsVisible        func(widget unsafe.Pointer) bool
	GtkWidgetSetAppPaintable  func(window unsafe.Pointer, appPaintable bool)
	GtkWidgetSetSensitive     func(widget unsafe.Pointer, sensitive bool)
	GtkWidgetSetVisual        func(window unsafe.Pointer, visual unsafe.Pointer)
	GtkWidgetShow             func(widget unsafe.Pointer)
	GtkWidgetShowAll          func(widget unsafe.Pointer)
	GtkWindowDeiconify        func(window unsafe.Pointer)
	GtkWindowFullscreen       func(window unsafe.Pointer)
	GtkWindowGetPosition      func(window unsafe.Pointer, x, y *int32)
	GtkWindowGetSize          func(window unsafe.Pointer, width, height *int32)
	GtkWindowIconify          func(window unsafe.Pointer)
	GtkWindowMaximize         func(window unsafe.Pointer)
	GtkWindowMove             func(window unsafe.Pointer, x, y int32)
	GtkWindowNew              func(windowType uint32) unsafe.Pointer
	GtkWindowResize           func(window unsafe.Pointer, width, height int32)
	GtkWindowSetDecorated     func(window unsafe.Pointer, setting bool)
	GtkWindowSetGeometryHints func(
		window unsafe.Pointer,
		geometryWidget unsafe.Pointer,
		geometry *GdkGeometry,
		geomMask uint32)
	GtkWindowSetIcon      func(window unsafe.Pointer, icon unsafe.Pointer)
	GtkWindowSetKeepAbove func(window unsafe.Pointer, setting bool)
	GtkWindowSetResizable func(window unsafe.Pointer, resizable bool)
	GtkWindowSetTitle     func(window unsafe.Pointer, title string)
	GtkWindowUnfullscreen func(window unsafe.Pointer)
	GtkWindowUnmaximize   func(window unsafe.Pointer)
)

var (
	GdkAtomIntern             func(atomName string, onlyIfExists bool) unsafe.Pointer
	GdkDisplayGetDefault      func() unsafe.Pointer
	GdkDisplayGetMonitor      func(display unsafe.Pointer, monitorNum int32) unsafe.Pointer
	GdkDisplayGetNMonitors    func(display unsafe.Pointer) int32
	GdkMonitorGetGeometry     func(monitor unsafe.Pointer, rect *GdkRectangle)
	GdkMonitorGetManufacturer func(monitor unsafe.Pointer) string
	GdkMonitorGetModel        func(monitor unsafe.Pointer) string
	GdkMonitorGetRefreshRate  func(monitor unsafe.Pointer) int32
	GdkMonitorGetScaleFactor  func(monitor unsafe.Pointer) int32
	GdkMonitorIsPrimary       func(monitor unsafe.Pointer) bool
	GdkPixbufNewFromFile      func(filename string, err *unsafe.Pointer) unsafe.Pointer
	GdkScreenGetDefault       func() unsafe.Pointer
	GdkScreenGetRootWindow    func(window unsafe.Pointer) unsafe.Pointer
	GdkScreenGetRgbaVisual    func(window unsafe.Pointer) unsafe.Pointer
	GdkScreenIsComposited     func(screen unsafe.Pointer) bool
	GdkWindowGetFrameExtends  func(window unsafe.Pointer, rect *GdkRectangle)
	GdkWindowGetGeometry      func(window unsafe.Pointer, x, y, width, height *int32)
)

var (
	WebkitJavascriptResultGetJsValue                     func(jsResult unsafe.Pointer) unsafe.Pointer
	WebkitSettingsSetEnableDeveloperExtras               func(settings unsafe.Pointer, enable bool)
	WebkitSettingsSetEnableWriteConsoleMessagesToStdout  func(settings unsafe.Pointer, enable bool)
	WebkitSettingsSetJavascriptCanAccessClipboard        func(settings unsafe.Pointer, enable bool)
	WebkitUserContentManagerAddScript                    func(manager unsafe.Pointer, script unsafe.Pointer)
	WebkitUserContentManagerRegisterScriptMessageHandler func(manager unsafe.Pointer, name string) bool
	WebkitUserScriptNew                                  func(
		source string,
		injectedFrames uint32,
		injectedTime uint32,
		allowList []string,
		blockList []string,
	) unsafe.Pointer
	WebkitWebViewEvaluateJavascript func(
		webView unsafe.Pointer,
		script string,
		length int32,
		//ignoring these atm, so they don't have string type
		worldName unsafe.Pointer,
		sourceUri unsafe.Pointer,
		cancellable unsafe.Pointer,
		callback unsafe.Pointer,
		userData unsafe.Pointer,
	)
	WebkitWebViewGetSettings           func(webView unsafe.Pointer) unsafe.Pointer
	WebkitWebViewGetUserContentManager func(webView unsafe.Pointer) unsafe.Pointer
	WebkitWebViewLoadHtml              func(webView unsafe.Pointer, content string, baseUri string)
	WebkitWebViewLoadUri               func(webView unsafe.Pointer, uri string)
	WebkitWebViewNew                   func() unsafe.Pointer
	WebkitWebViewSetBackgroundColor    func(webView unsafe.Pointer, rgba *GdkRGBA)
)

var (
	JscValueToString func(unsafe.Pointer) string
)

var (
	AppIndicatorNew         func(id string, iconName string, category uint32) unsafe.Pointer
	AppIndicatorSetIconFull func(self unsafe.Pointer, iconName string, iconDesc string)
	AppIndicatorSetLabel    func(self unsafe.Pointer, label string, guide string)
	AppIndicatorSetMenu     func(self unsafe.Pointer, menu unsafe.Pointer)
	AppIndicatorSetStatus   func(self unsafe.Pointer, status uint32)
	AppIndicatorSetTitle    func(self unsafe.Pointer, title string)
)

var (
	libc      uintptr
	libgtk    uintptr
	libwebgtk uintptr
	libjsc    uintptr
	libind    uintptr
)

func GetGTKPath() string {
	return "libgtk-3.so"
}

func GetCLibPath() string {
	return "libc.so.6"
}

func GetWebkitGtkLibbPath() string {
	return "libwebkit2gtk-4.1.so"
}

func GetJSCLibPath() string {
	return "libjavascriptcoregtk-4.1.so"
}

func GetAppIndicatorLibPath() string {
	return "libayatana-appindicator3.so.1"
}

func LoadLibraries() {
	load := func(libPathFunc func() string) uintptr {
		lib, err := purego.Dlopen(libPathFunc(), purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err != nil {
			panic(err)
		}
		return lib
	}

	libc = load(GetCLibPath)
	libgtk = load(GetGTKPath)
	libwebgtk = load(GetWebkitGtkLibbPath)
	libjsc = load(GetJSCLibPath)
	libind = load(GetAppIndicatorLibPath)
}

func UnloadLibraries() {
	purego.Dlclose(libc)
	purego.Dlclose(libgtk)
	purego.Dlclose(libwebgtk)
	purego.Dlclose(libjsc)
	purego.Dlclose(libind)
}

func SetAllCFuncs() {

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
	purego.RegisterLibFunc(&GtkWidgetGetWindow, libgtk, "gtk_widget_get_window")

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
	purego.RegisterLibFunc(&GdkWindowGetFrameExtends, libgtk, "gdk_window_get_frame_extents")
	purego.RegisterLibFunc(&GdkAtomIntern, libgtk, "gdk_atom_intern")

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

	//LibJavascriptCore functions
	purego.RegisterLibFunc(&JscValueToString, libjsc, "jsc_value_to_string")

	//LibAppIndicator functions
	purego.RegisterLibFunc(&AppIndicatorNew, libind, "app_indicator_new")
	purego.RegisterLibFunc(&AppIndicatorSetStatus, libind, "app_indicator_set_status")
	purego.RegisterLibFunc(&AppIndicatorSetTitle, libind, "app_indicator_set_title")
	purego.RegisterLibFunc(&AppIndicatorSetLabel, libind, "app_indicator_set_label")
	purego.RegisterLibFunc(&AppIndicatorSetMenu, libind, "app_indicator_set_menu")
	purego.RegisterLibFunc(&AppIndicatorSetIconFull, libind, "app_indicator_set_icon_full")
}

/*
* User defined Gtk functions
 */
func GtkWindowSetTransparent(window unsafe.Pointer, transparent bool) {
	fmt.Println("setting transparent")
	if transparent {
		GtkWidgetSetAppPaintable(window, true)
		screen := GdkScreenGetDefault()
		visual := GdkScreenGetRgbaVisual(screen)

		if visual != nil && GdkScreenIsComposited(screen) {
			GtkWidgetSetVisual(window, visual)
		}
	} else {
		GtkWidgetSetAppPaintable(window, false)
		GtkWidgetSetVisual(window, nil)
	}
}

func GtkWebViewSetTransparent(webview unsafe.Pointer, transparent bool) {
	color := GdkRGBA{}
	color.red = 1.0
	color.green = 1.0
	color.blue = 1.0
	color.alpha = 1.0

	if transparent {
		color.alpha = 0
	}

	WebkitWebViewSetBackgroundColor(webview, &color)
}

func StringFromJsResult(result unsafe.Pointer) string {
	value := WebkitJavascriptResultGetJsValue(result)
	return JscValueToString(value)
}

// A simple go implementation of `g_signal_connect`
func g_signal_connect(
	instance unsafe.Pointer,
	detailed_signal string,
	c_handler uintptr,
	data unsafe.Pointer,
) {
	GSignalConnectData(instance, detailed_signal, c_handler, data, nil, 0)
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
	result.Handle = GtkWindowNew(GtkWindowToplevel)
	return result
}

func Webview_New() Webview {
	result := Webview{}
	result.Handle = WebkitWebViewNew()
	return result
}

func (window *Window) Pointer() uintptr {
	return (uintptr)(unsafe.Pointer(window.Handle))
}

func (window *Window) AddWebview(webview Webview) {
	GtkContainerAdd(window.Handle, webview.Handle)
	GtkWidgetGrabFocus(webview.Handle)
}

func (window *Window) Show() {
	GtkWidgetShowAll(window.Handle)
}

func (window *Window) Hide() {
	GtkWidgetHide(window.Handle)
}

func (window *Window) Destroy() {
	if window.Handle != nil {
		GtkWidgetDestroy(window.Handle)
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

	GtkWindowGetSize(window.Handle, &result.Width, &result.Height)
	return result
}

func (window *Window) GetOuterSize() Size {
	gdk_window := GtkWidgetGetWindow(window.Handle)
	frame_extends := GdkRectangle{}
	GdkWindowGetFrameExtends(gdk_window, &frame_extends)
	return Size{
		Width:  int32(frame_extends.width),
		Height: int32(frame_extends.height),
	}
}

func (window *Window) GetPosition() Position {
	result := Position{}

	GtkWindowGetPosition(window.Handle, &result.X, &result.Y)
	return result
}

// TODO this works as intended but user shall be aware of gtk_window_set_resizable's behavior
// https://stackoverflow.com/a/3582628
func (window *Window) SetResizable(resizable bool) {
	GtkWindowSetResizable(window.Handle, resizable)
}

func (window *Window) SetSize(width int32, height int32) {
	GtkWindowResize(window.Handle, width, height)
}

func (window *Window) SetPosition(x int32, y int32) {
	GtkWindowMove(window.Handle, x, y)
}

func (window *Window) SetMinSize(width int32, height int32) {
	window.MinSize.Width = width
	window.MinSize.Height = height
	window.setGeometry()
}

func (window *Window) SetMaxSize(width int32, height int32) {
	window.MaxSize.Width = width
	window.MaxSize.Height = height
	window.setGeometry()
}

func (window *Window) setGeometry() {
	g := GdkGeometry{}
	var flags uint32 = 0
	if window.MaxSize.Width != 0 && window.MaxSize.Height != 0 {
		g.max_width = window.MaxSize.Width
		g.max_height = window.MaxSize.Height
		flags = flags | GdkHintMaxSize
	}
	if window.MinSize.Width != 0 && window.MinSize.Height != 0 {
		g.min_width = window.MinSize.Width
		g.min_height = window.MinSize.Width
		flags = flags | GdkHintMinSize
	}
	GtkWindowSetGeometryHints(window.Handle, nil, &g, flags)

}

func (window *Window) SetAlwaysOnTop(always bool) {
	GtkWindowSetKeepAbove(window.Handle, always)
}

func (window *Window) Focus() {
	//TODO
	GtkWidgetGrabFocus(window.Handle)
}

func (window *Window) IsVisible() bool {
	return GtkWidgetIsVisible(window.Handle)
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

	screenWidth := int32(0)
	screenHeight := int32(0)
	GdkWindowGetGeometry(root, nil, nil, &screenWidth, &screenHeight)

	nextPos := Position{
		X: (screenWidth - size.Width) / 2,
		Y: (screenHeight - size.Height) / 2,
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
// Caution: The parameter @event is a union whose first 4 bytes denoting the type regardles of what
// it is, this integer is alo present in "all" of the possible fields of the union. This is
// not a idiomatic go call, the @event is handled more like C raw poitner here.
// The function will first read that integer and deduce the union type, it will then cast
// this pointer to one of two possible union fields declared in go.
func go_event_callback(window unsafe.Pointer, event *int32, arg int32) {
	if globalEventCallback != nil {
		eventType := *event

		result := Event{}
		result.Window.Handle = window
		result.UserData = arg

		if eventType == GdkDelete {
			result.Type = Delete
		}

		if eventType == GdkDestroy {
			result.Type = Destroy
		}

		if eventType == GdkConfigure {
			// NOTE(nick): Resize and move event

			configure := (*GdkEventConfigure)(unsafe.Pointer(event))

			result.Type = Configure
			result.Position = Position{X: int32(configure.x), Y: int32(configure.y)}
			result.Size = Size{Width: int32(configure.width), Height: int32(configure.height)}
		}

		/*
			if eventType == C.GDK_FOCUS_CHANGE {
				focusChange := (unsafe.Pointer)(unsafe.Pointer(event))

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
		if eventType == GdkWindowState {
			windowState := (*GdkEventWindowState)(unsafe.Pointer(event))

			// https://docs.gtk.org/gdk3/flags.WindowState.html
			if windowState.changed_mask&GdkWindowStateFocused > 0 {
				focused := windowState.new_window_state&GdkWindowStateFocused > 0

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
	g_signal_connect(window.Handle, "event", purego.NewCallback(go_event_callback), unsafe.Pointer(&userData))
}

func SetGlobalEventCallback(callback Event_Callback) {
	globalEventCallback = callback
}

func (webview *Webview) RegisterCallback(name string, callback func(result string)) int {
	manager := WebkitWebViewGetUserContentManager(webview.Handle)

	event := fmt.Sprintf("script-message-received::%s", name)

	index := wc_register(callback)
	g_signal_connect(manager, event, purego.NewCallback(go_webview_callback), unsafe.Pointer(&index))
	WebkitUserContentManagerRegisterScriptMessageHandler(manager, name)

	return int(index)
}

func (webview *Webview) UnregisterCallback(callback int) {
	// @Incomplete: remove script handler

	wc_unregister(callback)
}

func (webview *Webview) Destroy() {
	if webview.Handle != nil {
		GtkWidgetDestroy(webview.Handle)
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
	//TODO int to int32 is lossy cast careful!
	WebkitWebViewEvaluateJavascript(webview.Handle, js, int32(len(js)), nil, nil, nil, nil, nil)
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
		WebkitUserContentInjectTopFrame,
		WebkitUserScriptInjectAtDocumentStart,
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
		monitor := GdkDisplayGetMonitor(display, int32(i))

		result[i] = Monitor{
			Handle: monitor,
		}
	}

	return result
}

func (monitor *Monitor) Geometry() Rectangle {
	rect := GdkRectangle{}
	GdkMonitorGetGeometry(monitor.Handle, &rect)

	return Rectangle{
		Position: Position{X: int32(rect.x), Y: int32(rect.y)},
		Size:     Size{Width: int32(rect.width), Height: int32(rect.height)},
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

func Indicator_New(id string, pngIconPath string, menu Menu) Indicator {
	handle := AppIndicatorNew(id, "", AppIndicatorCategoryApplicationStatus)
	AppIndicatorSetStatus(handle, AppIndicatorStatusActive)

	if len(pngIconPath) > 0 {
		AppIndicatorSetIconFull(handle, pngIconPath, "")
	}

	if menu.Handle != nil {
		AppIndicatorSetMenu(handle, menu.Handle)
	}

	result := Indicator{}
	result.Handle = handle
	return result
}

func Menu_New() Menu {
	result := Menu{}
	result.Handle = GtkMenuNew()
	return result
}

func (menu *Menu) Destroy() {
	if menu.Handle != nil {
		GtkWidgetDestroy(menu.Handle)
		menu.Handle = nil
	}
}

func MenuItem_New(id int, title string, disabled bool, checked bool, separator bool) MenuItem {
	var widget unsafe.Pointer = nil

	if separator {
		widget = GtkSeparatorMenuItemNew()
		GtkWidgetShow(widget)
	} else {
		if checked {
			widget = GtkCheckMenuItemNewWithLabel(title)

			GtkCheckMenuItemSetActive(widget, checked)
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
	result.Handle = widget
	return result
}

func (menu *Menu) AppendItem(item MenuItem) {
	GtkMenuShellAppend((unsafe.Pointer)(unsafe.Pointer(menu.Handle)), item.Handle)
}

func (item *MenuItem) SetSubmenu(child Menu) {
	GtkMenuItemSetSubmenu(item.Handle, child.Handle)
}

func go_menu_callback(item unsafe.Pointer, menuId int32) {
	if globalMenuCallback != nil {
		globalMenuCallback(menuId)
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

func go_webview_callback(manager unsafe.Pointer, result unsafe.Pointer, arg int32) {
	fn := wc_lookup(int(arg))
	str := StringFromJsResult(result)
	if fn != nil {
		fn(str)
	}
}

func OS_GetClipboardText() string {
	clipboard := GtkClipboardGet(GdkAtomIntern("CLIPBOARD", true))
	return GtkClipboardWaitForText(clipboard)
}

func OS_SetClipboardText(text string) bool {
	clipboard := GtkClipboardGet(GdkAtomIntern("CLIPBOARD", true))

	GtkClipboardSetText(clipboard, text, -1)

	// @Incomplete: is there a way to check if set_text succeeded?
	return true
}
