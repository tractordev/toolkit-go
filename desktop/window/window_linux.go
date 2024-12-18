package window

import (
	"log"
	"sync"

	"tractor.dev/toolkit-go/desktop/event"
	"tractor.dev/toolkit-go/desktop/linux"
)

type Window struct {
	win     linux.Window
	webview linux.Webview

	callbackId int

	prevPosition linux.Position
	prevSize     linux.Size

	window
}

func (w *Window) Load() {
	w.win = linux.Window_New()

	size := w.Options.Size

	// NOTE(nick): set default size
	if size.Width == 0 && size.Height == 0 {
		monitors := linux.Monitors()
		if len(monitors) > 0 {
			m := monitors[0]

			geom := m.Geometry()
			size.Width = float64(geom.Size.Width) * 0.8
			size.Height = float64(geom.Size.Height) * 0.8
		}
	}

	w.win.SetSize(int32(size.Width), int32(size.Height))

	if w.Options.MinSize.Width != 0 || w.Options.MinSize.Height != 0 {
		w.win.SetMinSize(int32(w.Options.MinSize.Width), int32(w.Options.MinSize.Height))
	}

	if w.Options.MaxSize.Width != 0 || w.Options.MaxSize.Height != 0 {
		w.win.SetMaxSize(int32(w.Options.MaxSize.Width), int32(w.Options.MaxSize.Height))
	}

	if w.Options.Center {
		w.win.Center()
	} else {
		//TODO
		w.win.SetPosition(int32(w.Options.Position.X), int32(w.Options.Position.Y))
	}

	if w.Options.Frameless {
		w.win.SetDecorated(false)
	}

	if w.Options.Fullscreen {
		w.win.SetFullscreen(true)
	}

	if w.Options.Maximized {
		w.win.SetMaximized(true)
	}

	w.win.SetResizable(w.Options.Resizable)

	if w.Options.Title != "" {
		w.win.SetTitle(w.Options.Title)
	}

	if w.Options.AlwaysOnTop {
		w.win.SetAlwaysOnTop(true)
	}

	if len(w.Options.Icon) > 0 {
		w.win.SetIconFromBytes(w.Options.Icon)
	}

	webview := linux.Webview_New()
	webview.SetSettings(linux.DefaultWebviewSettings())

	myCallback := func(result string) {
		log.Println("Callback from JavaScript!!", result)
	}
	callbackId := webview.RegisterCallback("apptron", myCallback)
	webview.Eval("webkit.messageHandlers.apptron.postMessage(JSON.stringify({ hello: 42 }));")

	w.win.AddWebview(webview)

	if w.Options.Transparent {
		w.win.SetTransparent(true)
		webview.SetTransparent(true)
	}

	if w.Options.URL != "" {
		webview.Navigate(w.Options.URL)
	}

	if w.Options.HTML != "" {
		webview.SetHtml(w.Options.HTML, "")
	}

	if w.Options.Script != "" {
		webview.AddScript(w.Options.Script)
	}

	if !w.Options.Hidden {
		w.win.Show()
	}

	w.win.BindEventCallback()

	w.webview = webview
	w.callbackId = callbackId

	event.Emit(event.Event{
		Type:     event.Created,
		Window:   w,
		Size:     w.GetInnerSize(),
		Position: w.GetOuterPosition(),
	})

}

func (w *Window) Unload() {
	if w.callbackId != 0 {
		w.webview.UnregisterCallback(w.callbackId)
		w.callbackId = 0
	}

	w.webview.Destroy()
	w.win.Destroy()
}

func (w *Window) Focus() {
	w.win.Focus()
}

func (w *Window) SetVisible(visible bool) {
	if visible {
		w.win.Show()
	} else {
		w.win.Hide()
	}
}

func (w *Window) IsVisible() bool {
	return w.win.IsVisible()
}

func (w *Window) SetMaximized(maximized bool) {
	w.win.SetMaximized(maximized)
}

func (w *Window) SetMinimized(minimized bool) {
	w.win.SetMinimized(minimized)
}

func (w *Window) SetFullscreen(fullscreen bool) {
	w.win.SetFullscreen(fullscreen)
}

func (w *Window) SetSize(size Size) {
	w.win.SetSize(int32(size.Width), int32(size.Height))
}

func (w *Window) SetMinSize(size Size) {
	w.win.SetMinSize(int32(size.Width), int32(size.Height))
}

func (w *Window) SetMaxSize(size Size) {
	w.win.SetMaxSize(int32(size.Width), int32(size.Height))
}

func (w *Window) SetResizable(resizable bool) {
	w.win.SetResizable(resizable)
}

func (w *Window) SetAlwaysOnTop(always bool) {
	w.win.SetAlwaysOnTop(always)
}

func (w *Window) SetPosition(position Position) {
	w.win.SetPosition(int32(position.X), int32(position.Y))
}

func (w *Window) SetTitle(title string) {
	w.win.SetTitle(title)
}

func (w *Window) GetOuterPosition() Position {
	pos := w.win.GetPosition()
	return Position{
		X: float64(pos.X),
		Y: float64(pos.Y),
	}
}

func (w *Window) GetOuterSize() Size {
	size := w.win.GetOuterSize()
	return Size{
		Width:  float64(size.Width),
		Height: float64(size.Height),
	}
}

func (w *Window) GetInnerSize() Size {
	size := w.win.GetSize()
	return Size{
		Width:  float64(size.Width),
		Height: float64(size.Height),
	}
}

var ptrLookup sync.Map

func findWindow(win linux.Window) *Window {
	v, ok := ptrLookup.Load(win.Pointer())
	if ok {
		return v.(*Window)
	}
	return nil
}

func init() {
	linux.SetGlobalEventCallback(func(it linux.Event) {

		if win := findWindow(it.Window); win != nil {
			if it.Type == linux.Delete {
				event.Emit(event.Event{
					Type: event.Destroyed,
					// Window: win.Handle,
				})
			}

			if it.Type == linux.Destroy {
				event.Emit(event.Event{
					Type: event.Close,
					// Window: win.Handle,
				})
			}

			if it.Type == linux.Configure {
				if it.Position.X != win.prevPosition.X || it.Position.Y != win.prevPosition.Y {
					event.Emit(event.Event{
						Type: event.Moved,
						// Window:   win.Handle,
						Position: win.GetOuterPosition(),
					})

					win.prevPosition = it.Position
				}

				if it.Size.Width != win.prevSize.Width || it.Size.Height != win.prevSize.Height {
					event.Emit(event.Event{
						Type: event.Resized,
						// Window: win.Handle,
						Size: win.GetOuterSize(),
					})

					win.prevSize = it.Size
				}
			}

			if it.Type == linux.FocusChange {
				if it.FocusIn {
					event.Emit(event.Event{
						Type: event.Focused,
						// Window: win.Handle,
					})
				} else {
					event.Emit(event.Event{
						Type: event.Blurred,
						// Window: win.Handle,
					})
				}
			}
		}

	})
}
