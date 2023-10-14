package watchfs

import (
	"fmt"
	"io/fs"
	"log"
	"strings"
	"sync"
	"time"

	"tractor.dev/toolkit-go/engine/fs/watchfs/watcher"
)

var Interval time.Duration = 500 * time.Millisecond

type EventType uint

const (
	EventError EventType = 1 << iota
	EventCreate
	EventWrite
	EventRemove
	EventRename
	EventChmod
	EventMove
)

type Event struct {
	Type    EventType
	Path    string
	OldPath string
	Err     error
	fs.FileInfo
}

func (e Event) String() string {
	if e.Type == EventError {
		return fmt.Sprintf("{Error %s}", e.Err.Error())
	}
	return fmt.Sprintf("{%s %s %s}", map[EventType]string{
		EventCreate: "create",
		EventWrite:  "write",
		EventRemove: "remove",
		EventRename: "rename",
		EventChmod:  "chmod",
		EventMove:   "move",
	}[e.Type], e.Path, e.OldPath)
}

type Config struct {
	Recursive bool
	EventMask uint
	Ignores   []string
	Handler   func(Event)
}

func Join(watches ...*Watch) <-chan Event {
	joined := make(chan Event)
	var wg sync.WaitGroup
	wg.Add(len(watches))
	for _, w := range watches {
		go func(ch <-chan Event) {
			for v := range ch {
				joined <- v
			}
			wg.Done()
		}(w.Iter())
	}
	go func() {
		wg.Wait()
		close(joined)
	}()
	return joined
}

type Watch struct {
	path    string
	cfg     Config
	inbox   chan Event
	unwatch func(*Watch)
}

func (w *Watch) Iter() <-chan Event {
	return w.inbox
}

func (w *Watch) Close() {
	w.unwatch(w)
}

type FS struct {
	fs.FS
	watcher *watcher.Watcher
	watches map[*Watch]struct{}
	mu      sync.Mutex
}

func New(fsys fs.StatFS) *FS {
	return &FS{
		FS:      fsys,
		watcher: watcher.New(fsys),
		watches: make(map[*Watch]struct{}),
	}
}

func (f *FS) matchWatches(path string) (watches []*Watch) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for w := range f.watches {
		if strings.HasPrefix(path, w.path) {
			watches = append(watches, w)
		}
	}
	return
}

func (f *FS) startWatcher() {
	go f.watcher.Start(Interval)
	for {
		select {
		case event := <-f.watcher.Event:
			// log.Println(event)
			for _, w := range f.matchWatches(event.Path) {
				// TODO: eventmask + ignores
				e := Event{
					FileInfo: event.FileInfo,
					Path:     event.Path,
					OldPath:  event.OldPath,
					Type: map[watcher.Op]EventType{
						watcher.Create: EventCreate,
						watcher.Write:  EventWrite,
						watcher.Move:   EventMove,
						watcher.Remove: EventRemove,
						watcher.Rename: EventRename,
						watcher.Chmod:  EventChmod,
					}[event.Op],
				}
				if w.cfg.Handler != nil {
					w.cfg.Handler(e)
				} else {
					w.inbox <- e
				}
			}
		case err := <-f.watcher.Error:
			if err == watcher.ErrWatchedFileDeleted {
				log.Println(err)
			} else {
				panic(err)
			}
		case <-f.watcher.Closed:
			return
		}
	}
}

func (f *FS) unwatch(w *Watch) {

	if w.cfg.Recursive {
		f.watcher.RemoveRecursive(w.path)
	} else {
		f.watcher.Remove(w.path)
	}

	f.mu.Lock()
	delete(f.watches, w)
	f.mu.Unlock()
}

func (f *FS) Watch(name string, cfg *Config) (*Watch, error) {
	if wfs, ok := f.FS.(interface {
		Watch(name string, cfg *Config) (*Watch, error)
	}); ok {
		return wfs.Watch(name, cfg)
	}

	if !f.watcher.IsRunning() {
		go f.startWatcher()
	}

	if cfg == nil {
		cfg = &Config{}
	}

	if cfg.Recursive {
		if err := f.watcher.AddRecursive(name); err != nil {
			return nil, err
		}
	} else {
		if err := f.watcher.Add(name); err != nil {
			return nil, err
		}
	}

	w := &Watch{
		path:    name,
		cfg:     *cfg,
		inbox:   make(chan Event),
		unwatch: f.unwatch,
	}

	f.mu.Lock()
	f.watches[w] = struct{}{}
	f.mu.Unlock()

	return w, nil
}
