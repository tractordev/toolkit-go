package window

import (
	"sync"

	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/macos/foundation"
	"github.com/progrium/darwinkit/macos/webkit"
	"github.com/progrium/darwinkit/objc"
)

type Window struct {
	moveOffset foundation.Point
	*appkit.Window

	window
}

func (w *Window) Unload() {
	if w.Window != nil {
		w.Window.Close()
		w.Window = nil
	}
}

func (w *Window) Load() {
	screen := appkit.Screen_MainScreen().Frame()
	size := w.Options.Size
	pos := w.Options.Position
	pos.Y = screen.Size.Height - pos.Y
	if size.Width == 0 {
		size.Width = 800
	}
	if size.Height == 0 {
		size.Height = 600
	}
	frame := foundation.Rect{
		Origin: foundation.Point{X: pos.X, Y: pos.Y - size.Height},
		Size:   foundation.Size{Width: size.Width, Height: size.Height},
	}

	win := appkit.NewWindowWithContentRectStyleMaskBackingDefer(
		frame,
		appkit.TitledWindowMask,
		appkit.BackingStoreBuffered,
		false,
	)

	if w.Options.Center {
		pos.X = (screen.Size.Width / 2) - (size.Width / 2)
		pos.Y = (screen.Size.Height / 2) - (size.Height / 2)
		frame = foundation.Rect{
			Origin: foundation.Point{X: pos.X, Y: pos.Y},
			Size:   foundation.Size{Width: size.Width, Height: size.Height},
		}
	}

	if w.Options.Hidden {
		frame = foundation.Rect{
			Origin: foundation.Point{X: pos.X, Y: pos.Y},
			Size:   foundation.Size{Width: 0, Height: 0},
		}
	}

	wkconf := webkit.NewWebViewConfiguration()
	objc.Call[objc.Void](wkconf.Preferences(), objc.Sel("setValue:forKey:"), foundation.Number_NumberWithBool(true), foundation.NewStringWithString("developerExtrasEnabled"))

	wv := webkit.NewWebViewWithFrameConfiguration(foundation.Rect{}, wkconf)
	wv.SetAutoresizingMask(appkit.ViewHeightSizable | appkit.ViewWidthSizable)
	if w.Options.URL != "" {
		req := foundation.NewURLRequestWithURL(foundation.NewURLWithString(w.Options.URL))
		wv.LoadRequest(req)
	} else if w.Options.HTML != "" {
		url := foundation.NewURLWithString("http://localhost")
		wv.LoadHTMLStringBaseURL(w.Options.HTML, url)
	}

	mask := appkit.TitledWindowMask | appkit.ClosableWindowMask | appkit.MiniaturizableWindowMask
	if w.Options.Frameless {
		mask = appkit.BorderlessWindowMask
	}
	if w.Options.Resizable {
		mask = mask | appkit.ResizableWindowMask
	}
	win.SetStyleMask(mask)

	if w.Options.Title != "" {
		win.SetTitle(w.Options.Title)
	} else {
		win.SetMovableByWindowBackground(true)
		win.SetTitlebarAppearsTransparent(true)
	}

	// todo: transparent
	// if options.Transparent {
	// 	nswin.SetBackgroundColor(cocoa.NSColor_Clear())
	// 	nswin.SetOpaque(false)
	// 	wv.SetOpaque(false)
	// 	wv.SetBackgroundColor(cocoa.NSColor_Clear())
	// 	wv.SetValueForKey(mac.False, mac.String("drawsBackground"))
	// }

	// todo: windowview/keydown hack
	// view := objc.Get("WindowView").Alloc().Init()

	// view := appkit.NewView()
	// view.AddSubview(wv)
	win.SetContentView(wv)
	win.MakeFirstResponder(wv)

	if w.Options.AlwaysOnTop {
		win.SetLevel(appkit.MainMenuWindowLevel)
	}

	// todo: delegate
	// delegate := objc.Get("WindowDelegate").Alloc().Init()
	// nswin.SetDelegate_(delegate)

	win.SetFrameDisplay(frame, true)
	//appkit.RunningApplication_CurrentApplication().ActivateWithOptions(appkit.ApplicationActivateAllWindows | appkit.ApplicationActivateIgnoringOtherApps)
	win.MakeKeyAndOrderFront(nil)
	//win.MakeKeyWindow()

	w.Window = &win
	w.Window.Retain()

	// event.Emit(event.Event{
	// 	Type:     event.Created,
	// 	Window:   w,
	// 	Size:     w.GetInnerSize(),
	// 	Position: w.GetOuterPosition(),
	// })
}

func (w *Window) Focus() {
	if !w.IsMiniaturized() {
		w.MakeKeyAndOrderFront(nil)
		appkit.Application_SharedApplication().ActivateIgnoringOtherApps(true)
	}
}

func (w *Window) SetVisible(visible bool) {
	if visible {
		w.MakeKeyAndOrderFront(nil)
	} else {
		w.OrderOut(nil)
	}
}

func (w *Window) IsVisible() bool {
	return w.Window.IsVisible()
}

func (w *Window) SetMaximized(maximized bool) {
	// TODO: if true and is zoomed, return
	// TODO: https://github.com/tauri-apps/tao/blob/dev/src/platform_impl/macos/util/async.rs#L150
}

func (w *Window) SetMinimized(minimized bool) {
	if w.IsMiniaturized() == minimized {
		return
	}
	if minimized {
		w.Miniaturize(nil)
	} else {
		w.Deminiaturize(nil)
	}
}

func (w *Window) SetFullscreen(fullscreen bool) {
	// TODO: https://github.com/tauri-apps/tao/blob/dev/src/platform_impl/macos/window.rs#L784
}

func (w *Window) SetSize(size Size) {
	w.SetContentSize(foundation.Size{Width: size.Width, Height: size.Height})
}

func (w *Window) SetMinSize(size Size) {
	w.SetContentMinSize(foundation.Size{Width: size.Width, Height: size.Height})
}

func (w *Window) SetMaxSize(size Size) {
	w.SetContentMaxSize(foundation.Size{Width: size.Width, Height: size.Height})
}

func (w *Window) SetResizable(resizable bool) {
	// TODO: If fullscreen?
	mask := w.StyleMask()
	if resizable {
		mask = mask | appkit.ResizableWindowMask
	} else {
		mask = mask & appkit.ResizableWindowMask
	}
	w.SetStyleMask(mask)
}

func (w *Window) SetAlwaysOnTop(always bool) {
	if always {
		w.SetLevel(appkit.FloatingWindowLevel)
	} else {
		w.SetLevel(appkit.NormalWindowLevel)
	}
}

func (w *Window) SetPosition(position Position) {
	screenRect := w.Screen().Frame()
	// NOTE(nick): Y is inverted on MacOS
	position.Y = screenRect.Size.Height - position.Y

	// @Robustness: this implicitly relies on the frame size now that Y is inverted
	w.SetFrameTopLeftPoint(foundation.Point{X: position.X, Y: position.Y})
}

func (w *Window) SetTitle(title string) {
	w.Window.SetTitle(title)
}

func (w *Window) GetOuterPosition() Position {
	frame := w.Frame()
	screenRect := w.Screen().Frame()
	return Position{
		X: frame.Origin.X,
		Y: screenRect.Size.Height - (frame.Origin.Y + frame.Size.Height),
	}
}

func (w *Window) GetOuterSize() Size {
	frame := w.Frame()
	return Size{
		Width:  frame.Size.Width,
		Height: frame.Size.Height,
	}
}

func (w *Window) GetInnerSize() Size {
	// TODO(nick): adjust window rect
	return w.GetOuterSize()
}

var ptrLookup sync.Map

func findWindow(win objc.Object) *Window {
	v, ok := ptrLookup.Load(win.Ptr())
	if ok {
		return v.(*Window)
	}
	return nil
}

type WindowView struct {
	objc.Object `objc:"WindowView : NSView"`
}

func (v *WindowView) keyDown(event objc.Object) {
	// no-op, otherwise not having a keyDown in
	// responder chain makes system beep on keyDown
	// when not handled in javascript.
}

func init() {
	// viewClass := objc.NewClassFromStruct(WindowView{})
	// viewClass.AddMethod("keyDown:", (*WindowView).keyDown)
	// objc.RegisterClass(viewClass)

	// DelegateClass := objc.NewClass("WindowDelegate", "NSObject")
	// DelegateClass.AddMethod("windowDidMove:", func(self, notif objc.Object) {
	// 	if win := findWindow(notif.Get("object")); win != nil {
	// 		event.Emit(event.Event{
	// 			Type:     event.Moved,
	// 			Window:   win,
	// 			Position: win.GetOuterPosition(),
	// 		})
	// 	}
	// })
	// DelegateClass.AddMethod("windowDidResize:", func(self, notif objc.Object) {
	// 	if win := findWindow(notif.Get("object")); win != nil {
	// 		event.Emit(event.Event{
	// 			Type:   event.Resized,
	// 			Window: win,
	// 			Size:   win.GetOuterSize(),
	// 		})
	// 	}
	// })
	// DelegateClass.AddMethod("windowDidBecomeKey:", func(self, notif objc.Object) {
	// 	if win := findWindow(notif.Get("object")); win != nil {
	// 		event.Emit(event.Event{
	// 			Type:   event.Focused,
	// 			Window: win,
	// 		})
	// 	}
	// })
	// DelegateClass.AddMethod("windowDidResignKey:", func(self, notif objc.Object) {
	// 	if win := findWindow(notif.Get("object")); win != nil {
	// 		event.Emit(event.Event{
	// 			Type:   event.Blurred,
	// 			Window: win,
	// 		})
	// 	}
	// })
	// DelegateClass.AddMethod("windowShouldClose:", func(sender objc.Object) bool {
	// 	// not sure this is right
	// 	if win := findWindow(sender); win != nil {
	// 		event.Emit(event.Event{
	// 			Type:   event.Close,
	// 			Window: win,
	// 		})
	// 	}
	// 	return true
	// })
	// DelegateClass.AddMethod("windowWillClose:", func(self, notif objc.Object) {
	// 	// maybe this isn't what should trigger this event
	// 	if win := findWindow(notif.Get("object")); win != nil {
	// 		event.Emit(event.Event{
	// 			Type:   event.Destroyed,
	// 			Window: win,
	// 		})
	// 	}
	// })
	// DelegateClass.AddMethod("userContentController:didReceiveScriptMessage:", func(self, cc, msg objc.Object) {
	// 	msgDict := mac.NSDictionary_fromRef(msg.Get("body"))
	// 	win := findWindow(msg.Get("webView").Get("window"))
	// 	if win == nil {
	// 		return
	// 	}
	// 	action := msgDict.ObjectForKey(mac.String("action")).String()
	// 	switch action {
	// 	case "minimize":
	// 		// TODO: known issue this doesnt work for frameless windows...
	// 		// 			 sort of defeats the point, but i'm sure theres a way
	// 		win.SetMinimized(true)
	// 	case "maximize":
	// 		// TODO: not tested since setmaximized is not implemented
	// 		win.SetMaximized(true)
	// 	case "close":
	// 		win.Unload()
	// 	case "move":
	// 		pos := win.GetOuterPosition()
	// 		mouseLoc := appkit.Event_MouseLocation()
	// 		win.moveOffset = mac.NSPoint{
	// 			X: mouseLoc.X - pos.X,
	// 			Y: mouseLoc.Y - pos.Y,
	// 		}
	// 	case "moving":
	// 		mouseLoc := appkit.Event_MouseLocation()
	// 		win.SetPosition(Position{
	// 			X: mouseLoc.X - win.moveOffset.X,
	// 			Y: mouseLoc.Y - win.moveOffset.Y,
	// 		})
	// 	default:
	// 	}
	// })
	// objc.RegisterClass(DelegateClass)
}
