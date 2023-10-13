package mux

import (
	"io"
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
	closed  bool
}

func newBufferedPipeWriter(pw *io.PipeWriter, bufferSize int) *bufferedPipeWriter {
	dataCh := make(chan []byte, bufferSize)
	closeCh := make(chan struct{})

	go func() {
		defer pw.Close()
		for {
			select {
			case data := <-dataCh:
				pw.Write(data)
			case <-closeCh:
				return
			}
		}
	}()

	return &bufferedPipeWriter{
		dataCh:  dataCh,
		closeCh: closeCh,
		closed:  false,
	}
}

func (w *bufferedPipeWriter) Write(p []byte) (n int, err error) {
	if w.closed {
		return 0, io.ErrClosedPipe
	}

	data := make([]byte, len(p))
	copy(data, p)
	w.dataCh <- data
	return len(p), nil
}

func (w *bufferedPipeWriter) Close() error {
	if w.closed {
		return io.ErrClosedPipe
	}
	w.closed = true
	close(w.closeCh)
	return nil
}
