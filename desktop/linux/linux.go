//go:build linux

package linux

import (
	"fmt"
	"log"
	"os"
	"sync"
	"unsafe"
)

/*
#cgo linux pkg-config: gtk+-3.0 webkit2gtk-4.0 ayatana-appindicator3-0.1

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

//
// Exports
//

func OS_Init() {
	C.gtk_init_check(nil, nil)
}

func PollEvents() {
	C.gtk_main_iteration_do(0) // 0 = non-blocking
}

func Window_New() Window {
	result := Window{}
	result.Handle = Window_FromWidget(C.gtk_window_new(C.GTK_WINDOW_TOPLEVEL))
	return result
}

func Webview_New() Webview {
	result := Webview{}
	result.Handle = Webview_FromWidget(C.webkit_web_view_new())
	return result
}

func (window *Window) Pointer() uintptr {
	return (uintptr)(unsafe.Pointer(window.Handle))
}

func (window *Window) AddWebview(webview Webview) {
	C.gtk_container_add(Window_GTK_CONTAINER(window.Handle), Webview_GTK_WIDGET(webview.Handle))
	C.gtk_widget_grab_focus(Webview_GTK_WIDGET(webview.Handle))
}

func (window *Window) Show() {
	C.gtk_widget_show_all(Window_GTK_WIDGET(window.Handle))
}

func (window *Window) Hide() {
	C.gtk_widget_hide(Window_GTK_WIDGET(window.Handle))
}

func (window *Window) Destroy() {
	if window.Handle != nil {
		C.gtk_widget_destroy(Window_GTK_WIDGET(window.Handle))
		window.Handle = nil
	}
}

func (window *Window) SetTransparent(transparent bool) {
	C.gtk_window_set_transparent(window.Handle, toCBool(transparent))
}

func (window *Window) SetTitle(title string) {
	ctitle := C.CString(title)
	defer C.free(unsafe.Pointer(ctitle))

	C.gtk_window_set_title(window.Handle, ctitle)
}

func (window *Window) SetDecorated(decorated bool) {
	C.gtk_window_set_decorated(window.Handle, toCBool(decorated))
}

func (window *Window) GetSize() Size {
	result := Size{}

	width := C.int(0)
	height := C.int(0)
	C.gtk_window_get_size(window.Handle, &width, &height)

	result.Width = int(width)
	result.Height = int(height)

	return result
}

func (window *Window) GetPosition() Position {
	result := Position{}

	x := C.int(0)
	y := C.int(0)
	C.gtk_window_get_position(window.Handle, &x, &y)

	result.X = int(x)
	result.Y = int(y)

	return result
}

func (window *Window) SetResizable(resizable bool) {
	C.gtk_window_set_resizable(window.Handle, toCBool(resizable))
}

func (window *Window) SetSize(width int, height int) {
	C.gtk_window_resize(window.Handle, C.int(width), C.int(height))
}

func (window *Window) SetPosition(x int, y int) {

	C.gtk_window_move(window.Handle, C.int(x), C.int(y))
}

func (window *Window) SetMinSize(width int, height int) {
	g := C.GdkGeometry{}
	g.min_width = C.int(width)
	g.min_height = C.int(height)
	C.gtk_window_set_geometry_hints(window.Handle, nil, &g, C.GDK_HINT_MIN_SIZE)
}

func (window *Window) SetMaxSize(width int, height int) {
	g := C.GdkGeometry{}
	g.max_width = C.int(width)
	g.max_height = C.int(height)
	C.gtk_window_set_geometry_hints(window.Handle, nil, &g, C.GDK_HINT_MAX_SIZE)
}

func (window *Window) SetAlwaysOnTop(always bool) {
	C.gtk_window_set_keep_above(window.Handle, toCBool(always))
}

func (window *Window) Focus() {
	C.gtk_widget_grab_focus(Window_GTK_WIDGET(window.Handle))
}

func (window *Window) IsVisible() bool {
	return fromCBool(C.gtk_widget_is_visible(Window_GTK_WIDGET(window.Handle)))
}

func (window *Window) SetMinimized(minimized bool) {
	if minimized {
		C.gtk_window_iconify(window.Handle)
	} else {
		C.gtk_window_deiconify(window.Handle)
	}
}

func (window *Window) SetMaximized(maximized bool) {
	if maximized {
		C.gtk_window_maximize(window.Handle)
	} else {
		C.gtk_window_unmaximize(window.Handle)
	}
}

func (window *Window) SetFullscreen(fullscreen bool) {
	if fullscreen {
		C.gtk_window_fullscreen(window.Handle)
	} else {
		C.gtk_window_unfullscreen(window.Handle)
	}
}

func (window *Window) Center() {
	size := window.GetSize()
	root := C.gdk_screen_get_root_window(C.gdk_screen_get_default())

	screenWidth := C.int(0)
	screenHeight := C.int(0)
	C.gdk_window_get_geometry(root, nil, nil, &screenWidth, &screenHeight)

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
	defer C.free(unsafe.Pointer(cpath))

	buffer := C.gdk_pixbuf_new_from_file(cpath, nil)

	if buffer != nil {
		C.gtk_window_set_icon(window.Handle, buffer)
		return true
	} else {
		log.Println("[SetIconFromBytes] Failed to load PixBuf from file!")
	}

	return false
}

// https://docs.gtk.org/gdk3/union.Event.html
// https://api.gtkd.org/gdk.c.types.GdkEventType.html

//export go_event_callback
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
	defer C.free(unsafe.Pointer(cevent))

	C._g_signal_connect(Window_GTK_WIDGET(window.Handle), cevent, C.go_event_callback, C.int(userData))
}

func SetGlobalEventCallback(callback Event_Callback) {
	globalEventCallback = callback
}

func (webview *Webview) RegisterCallback(name string, callback func(result string)) int {
	manager := C.webkit_web_view_get_user_content_manager(webview.Handle)

	cevent := C.CString(fmt.Sprintf("script-message-received::%s", name))
	defer C.free(unsafe.Pointer(cevent))

	cexternal := C.CString(name)
	defer C.free(unsafe.Pointer(cexternal))

	index := wc_register(callback)
	C._g_signal_connect(WebKitUserContentManager_GTK_WIDGET(manager), cevent, C.go_webview_callback, C.int(index))
	C.webkit_user_content_manager_register_script_message_handler(manager, cexternal)

	return int(index)
}

func (webview *Webview) UnregisterCallback(callback int) {
	// @Incomplete: remove script handler

	wc_unregister(callback)
}

func (webview *Webview) Destroy() {
	if webview.Handle != nil {
		C.gtk_widget_destroy(Webview_GTK_WIDGET(webview.Handle))
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
	settings := C.webkit_web_view_get_settings(webview.Handle)

	C.webkit_settings_set_javascript_can_access_clipboard(settings, toCBool(config.CanAccessClipboard))
	C.webkit_settings_set_enable_write_console_messages_to_stdout(settings, toCBool(config.WriteConsoleToStdout))
	C.webkit_settings_set_enable_developer_extras(settings, toCBool(config.DeveloperTools))
}

func (webview *Webview) Eval(js string) {
	cjs := C.CString(js)
	defer C.free(unsafe.Pointer(cjs))

	C.webkit_web_view_run_javascript(webview.Handle, cjs, nil, nil, nil)
}

func (webview *Webview) SetHtml(html string, baseUri string) {
	chtml := C.CString(html)
	defer C.free(unsafe.Pointer(chtml))

	cbaseUri := C.CString(baseUri)
	defer C.free(unsafe.Pointer(cbaseUri))

	C.webkit_web_view_load_html(webview.Handle, chtml, cbaseUri)
}

func (webview *Webview) Navigate(url string) {
	curl := C.CString(url)
	defer C.free(unsafe.Pointer(curl))

	C.webkit_web_view_load_uri(webview.Handle, curl)
}

func (webview *Webview) AddScript(js string) {
	manager := C.webkit_web_view_get_user_content_manager(webview.Handle)

	cjs := C.CString(js)
	defer C.free(unsafe.Pointer(cjs))

	script := C.webkit_user_script_new(
		cjs,
		C.WEBKIT_USER_CONTENT_INJECT_TOP_FRAME,
		C.WEBKIT_USER_SCRIPT_INJECT_AT_DOCUMENT_START,
		nil,
		nil,
	)

	C.webkit_user_content_manager_add_script(manager, script)
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

	display := C.gdk_display_get_default()
	if display == nil {
		return make([]Monitor, 0)
	}

	n := int(C.gdk_display_get_n_monitors(display))

	result := make([]Monitor, n)

	for i := 0; i < n; i++ {
		monitor := C.gdk_display_get_monitor(display, C.int(i))

		result[i] = Monitor{
			Handle: monitor,
		}
	}

	return result
}

func (monitor *Monitor) Geometry() Rectangle {
	rect := C.GdkRectangle{}
	C.gdk_monitor_get_geometry(monitor.Handle, &rect)

	return Rectangle{
		Position: Position{X: int(rect.x), Y: int(rect.y)},
		Size:     Size{Width: int(rect.width), Height: int(rect.height)},
	}
}

func (monitor *Monitor) ScaleFactor() int {
	return int(C.gdk_monitor_get_scale_factor(monitor.Handle))
}

func (monitor *Monitor) Name() string {
	manufacturer := C.GoString(C.gdk_monitor_get_manufacturer(monitor.Handle))
	model := C.GoString(C.gdk_monitor_get_model(monitor.Handle))
	return manufacturer + " " + model
}

func (monitor *Monitor) RefreshRate() int {
	// NOTE(nick): in milli-Hertz (60Hz = 60000)
	return int(C.gdk_monitor_get_refresh_rate(monitor.Handle)) / 1000
}

func (monitor *Monitor) IsPrimary() bool {
	return fromCBool(C.gdk_monitor_is_primary(monitor.Handle))
}

//
// Indicator
//

func Indicator_New(id string, pngIconPath string, menu Menu) Indicator {
	cid := C.CString(id)
	defer C.free(unsafe.Pointer(cid))

	handle := C.app_indicator_new(cid, C.CString(""), C.APP_INDICATOR_CATEGORY_APPLICATION_STATUS)
	C.app_indicator_set_status(handle, C.APP_INDICATOR_STATUS_ACTIVE)

	//app_indicator_set_title(global_app_indicator, title);
	//app_indicator_set_label(global_app_indicator, title, "");

	if len(pngIconPath) > 0 {
		cIconPath := C.CString(pngIconPath)
		defer C.free(unsafe.Pointer(cIconPath))

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
	result.Handle = Menu_FromWidget(C.gtk_menu_new())
	return result
}

func (menu *Menu) Destroy() {
	if menu.Handle != nil {
		C.gtk_widget_destroy(Menu_GTK_WIDGET(menu.Handle))
		menu.Handle = nil
	}
}

func MenuItem_New(id int, title string, disabled bool, checked bool, separator bool) MenuItem {
	var widget *C.struct__GtkWidget = nil

	if separator {
		widget = C.gtk_separator_menu_item_new()
		C.gtk_widget_show(widget)
	} else {
		ctitle := C.CString(title)
		defer C.free(unsafe.Pointer(ctitle))

		if checked {
			widget = C.gtk_check_menu_item_new_with_label(ctitle)

			C.gtk_check_menu_item_set_active(CheckMenuItem_FromWidget(widget), toCBool(checked))
		} else {
			widget = C.gtk_menu_item_new_with_label(ctitle)
		}

		C.gtk_widget_set_sensitive(widget, toCBool(!disabled))

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
		defer C.free(unsafe.Pointer(cactivate))

		C._g_signal_connect(widget, cactivate, C.go_menu_callback, C.int(id))

		C.gtk_widget_show(widget)
	}

	result := MenuItem{}
	result.Handle = MenuItem_FromWidget(widget)
	return result
}

func (menu *Menu) AppendItem(item MenuItem) {
	C.gtk_menu_shell_append(Menu_GTK_MENU_SHELL(menu.Handle), MenuItem_GTK_WIDGET(item.Handle))
}

func (item *MenuItem) SetSubmenu(child Menu) {
	C.gtk_menu_item_set_submenu(item.Handle, Menu_GTK_WIDGET(child.Handle))
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
	clipboard := C.gtk_clipboard_get(C.GDK_SELECTION_CLIPBOARD)
	text := C.gtk_clipboard_wait_for_text(clipboard)

	return C.GoString(text)
}

func OS_SetClipboardText(text string) bool {
	ctext := C.CString(text)
	defer C.free(unsafe.Pointer(ctext))

	clipboard := C.gtk_clipboard_get(C.GDK_SELECTION_CLIPBOARD)

	C.gtk_clipboard_set_text(clipboard, ctext, -1)

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
