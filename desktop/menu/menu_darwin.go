package menu

import (
	"io"
	"time"

	"github.com/progrium/darwinkit/helper/action"
	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/objc"
	"tractor.dev/toolkit-go/desktop/event"
)

type menu struct {
	*appkit.Menu
}

func (m *menu) unload() {
	if m.Menu != nil {
		m.Menu.Release()
		m.Menu = nil
	}
}

func (m *menu) reload(items []Item) {
	menu := appkit.NewMenu()
	menu.SetAutoenablesItems(true)

	for _, i := range items {
		menu.AddItem(newMenuItem(&i))
	}

	m.Menu = &menu
	m.Menu.Retain()
}

func (m *menu) popup() int {
	ch := make(chan int, 1)
	event.Listen(time.Now(), func(e event.Event) error {
		if e.Type == event.MenuItem {
			ch <- e.MenuItem
		}
		return io.EOF
	})
	if m.Menu.PopUpMenuPositioningItemAtLocationInView(nil, appkit.Event_MouseLocation(), nil) {
		return <-ch
	}
	return 0
}

func newMenuItem(i *Item) appkit.MenuItem {
	if i.Separator {
		return appkit.MenuItem_SeparatorItem()
	}

	item := appkit.NewMenuItem()
	title := i.Title
	// if title == "" && i.obj != nil {
	// 	title = entity.Name(i.obj)
	// }
	item.SetTitle(title)
	item.SetTag(i.ID)
	item.SetEnabled(!i.Disabled)
	// item.SetToolTip(i.Tooltip)

	// Checked
	if i.Selected {
		item.SetState(appkit.ControlStateValueOn)
	}

	// Icon
	// if i.Icon != "" {
	// 	b, err := base64.StdEncoding.DecodeString(i.Icon)
	// 	if err == nil {
	// 		data := core.NSData_WithBytes(b, uint64(len(b)))
	// 		img := cocoa.NSImage_InitWithData(data)
	// 		img.SetSize(core.Size(16, 16))
	// 		item.SetImage(img)
	// 	}
	// }

	if !i.Disabled && len(i.SubMenu) == 0 {
		// special item titles
		if title == "Quit" {
			item.SetTarget(appkit.Application_SharedApplication())
			item.SetAction(objc.Sel("terminate:"))

		} else if i.OnClick != nil {
			action.Set(item, action.Handler(func(sender objc.Object) {
				i.OnClick()
			}))

		}
		// else if i.obj != nil {
		// 	action.Set(item, action.Handler(func(sender objc.Object) {
		// 		if err := node.Activate(context.Background(), i.obj); err != nil {
		// 			log.Println(err)
		// 		}
		// 	}))
		// }
	}

	items := subItems(*i)
	if len(items) > 0 {
		sub := appkit.NewMenu()
		sub.SetTitle(title)
		sub.SetAutoenablesItems(true)
		for _, i := range items {
			sub.AddItem(newMenuItem(&i))
		}
		item.SetSubmenu(sub)
	}

	return item
}
