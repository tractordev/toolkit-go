package mux

import (
	"io"
	"sync"
)

func Pair() (a, b Session) {
	ar, bw := io.Pipe()
	br, aw := io.Pipe()
	abuf := newBufferedPipeWriter(aw, 4)
	bbuf := newBufferedPipeWriter(bw, 4)
	a, _ = DialIO(abuf, ar)
	b, _ = DialIO(bbuf, br)
	return
}

type bufferedPipeWriter struct {
	dataCh  chan []byte
	closeCh chan struct{}
	spaceCh chan struct{}
	mu      sync.Mutex
	err     error
	pw      *io.PipeWriter
}

func newBufferedPipeWriter(pw *io.PipeWriter, bufferSize int) *bufferedPipeWriter {
	w := &bufferedPipeWriter{dataCh: make(chan []byte, bufferSize), closeCh: make(chan struct{}), spaceCh: make(chan struct{}), pw: pw}
	go func() {
		defer pw.Close()
		for {
			select {
			case <-w.closeCh:
				return
			case data := <-w.dataCh:
				w.mu.Lock()
				close(w.spaceCh)
				w.spaceCh = make(chan struct{})
				err := w.err
				w.mu.Unlock()
				if err != nil {
					return
				}
				if _, err := pw.Write(data); err != nil {
					w.stop(err)
					return
				}
			}
		}
	}()
	return w
}

// Write transfers ownership of a copy to the bounded queue, not delivery.
func (w *bufferedPipeWriter) Write(p []byte) (int, error) {
	for {
		w.mu.Lock()
		if w.err != nil {
			err := w.err
			w.mu.Unlock()
			return 0, err
		}
		select {
		case w.dataCh <- append([]byte(nil), p...):
			w.mu.Unlock()
			return len(p), nil
		default:
			space := w.spaceCh
			w.mu.Unlock()
			select {
			case <-w.closeCh:
			case <-space:
			}
		}
	}
}

func (w *bufferedPipeWriter) stop(err error) bool {
	w.mu.Lock()
	if w.err != nil {
		w.mu.Unlock()
		return false
	}
	w.err = err
	close(w.closeCh)
	// Close is abortive: release queued payloads without waiting for delivery.
	for {
		select {
		case <-w.dataCh:
		default:
			w.mu.Unlock()
			w.pw.Close()
			return true
		}
	}
}

func (w *bufferedPipeWriter) Close() error {
	if !w.stop(io.ErrClosedPipe) {
		return io.ErrClosedPipe
	}
	return nil
}
