package shell

import (
	"reflect"
	"strings"
	"sync"

	"golang.design/x/hotkey"
	"tractor.dev/toolkit-go/desktop/event"
	"tractor.dev/toolkit-go/desktop/keycode"
)

var hotkeys sync.Map
var once sync.Once
var resetLoop chan bool

func shortcutLoop() {
	for {
		var accels []string
		var cases []reflect.SelectCase
		hotkeys.Range(func(key, value interface{}) bool {
			hk := value.(*hotkey.Hotkey)
			accels = append(accels, key.(string))
			cases = append(cases, reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(hk.Keydown()),
			})
			return true
		})
		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(resetLoop),
		})
		chosen, _, ok := reflect.Select(cases)
		if !ok {
			continue
		}
		if chosen > len(accels)-1 {
			// resetLoop was selected
			continue
		}
		event.Emit(event.Event{
			Type:     event.Shortcut,
			Shortcut: accels[chosen],
		})
	}
}

func registerShortcut(accelerator string) bool {
	if isShortcutRegistered(accelerator) {
		return false
	}

	var mods []hotkey.Modifier
	var key hotkey.Key
	for _, accel := range strings.Split(strings.ToUpper(accelerator), "+") {
		code := keycode.FromString(accel)
		if keycode.IsModifier(code) {
			mods = append(mods, keycode.HotkeyModifier(code))
			continue
		}
		key = hotkey.Key(keycode.Scancode(code))
		break
	}

	hk := hotkey.New(mods, key)
	if err := hk.Register(); err != nil {
		return false
	}
	hotkeys.Store(accelerator, hk)
	once.Do(func() {
		resetLoop = make(chan bool, 1)
		go shortcutLoop()
	})
	resetLoop <- true
	return true
}

func isShortcutRegistered(accelerator string) (exists bool) {
	_, exists = hotkeys.Load(strings.ToUpper(accelerator))
	return
}

func unregisterShortcut(accelerator string) bool {
	v, exists := hotkeys.Load(strings.ToUpper(accelerator))
	if !exists {
		return false
	}
	hk := v.(*hotkey.Hotkey)
	if err := hk.Unregister(); err != nil {
		return false
	}
	hotkeys.Delete(strings.ToUpper(accelerator))
	resetLoop <- true
	return true
}

func unregisterAllShortcuts() {
	hotkeys.Range(func(key, value interface{}) bool {
		hk := value.(*hotkey.Hotkey)
		hk.Unregister()
		hotkeys.Delete(key)
		return true
	})
	resetLoop <- true
}
