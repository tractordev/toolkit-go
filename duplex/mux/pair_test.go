package mux

import (
	"errors"
	"io"
	"runtime"
	"sync"
	"testing"
	"time"
)

func awaitResult(t *testing.T, ch <-chan error) error {
	t.Helper()
	select {
	case err := <-ch:
		return err
	case <-time.After(2 * time.Second):
		t.Fatal("operation did not terminate")
		return nil
	}
}

func TestBufferedPipeWriterBlockedClose(t *testing.T) {
	r, p := io.Pipe()
	w := newBufferedPipeWriter(p, 4)
	defer r.Close()
	defer p.Close()
	for i := 0; i < 5; i++ {
		if _, err := w.Write([]byte("x")); err != nil {
			t.Fatal(err)
		}
	}
	result := make(chan error, 1)
	go func() { _, err := w.Write([]byte("blocked")); result <- err }()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := awaitResult(t, result); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("write: %v", err)
	}
	if _, err := r.Read(make([]byte, 1)); err != io.EOF {
		t.Fatalf("pump not cancelled: %v", err)
	}
	if _, err := w.Write(nil); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("late write: %v", err)
	}
}

func TestBufferedPipeWriterReaderError(t *testing.T) {
	r, p := io.Pipe()
	w := newBufferedPipeWriter(p, 4)
	defer r.Close()
	defer p.Close()
	sentinel := errors.New("reader failed")
	for i := 0; i < 5; i++ {
		w.Write([]byte("x"))
	}
	result := make(chan error, 1)
	go func() { _, err := w.Write([]byte("blocked")); result <- err }()
	r.CloseWithError(sentinel)
	if err := awaitResult(t, result); err != sentinel {
		t.Fatalf("blocked write: %v", err)
	}
	if _, err := w.Write(nil); err != sentinel {
		t.Fatalf("retained error: %v", err)
	}
}

func TestBufferedPipeWriterConcurrentWriteClose(t *testing.T) {
	r, p := io.Pipe()
	w := newBufferedPipeWriter(p, 4)
	defer r.Close()
	defer p.Close()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); <-start; w.Write([]byte("x")) }()
	}
	close(start)
	runtime.Gosched()
	w.Close()
	done := make(chan error, 1)
	go func() { wg.Wait(); done <- nil }()
	awaitResult(t, done)
}

func TestBufferedPipeWriterCopyFIFO(t *testing.T) {
	r, p := io.Pipe()
	w := newBufferedPipeWriter(p, 4)
	defer r.Close()
	defer w.Close()
	data := []byte("abc")
	w.Write(data)
	copy(data, "zzz")
	w.Write([]byte("def"))
	got := make([]byte, 6)
	if _, err := io.ReadFull(r, got); err != nil {
		t.Fatal(err)
	}
	if string(got) != "abcdef" {
		t.Fatalf("got %q", got)
	}
}

func TestBufferedPipeWriterConcurrentClose(t *testing.T) {
	r, p := io.Pipe()
	w := newBufferedPipeWriter(p, 4)
	defer r.Close()
	defer p.Close()
	start := make(chan struct{})
	results := make(chan error, 32)
	for i := 0; i < 32; i++ {
		go func() { <-start; results <- w.Close() }()
	}
	close(start)
	successes := 0
	for i := 0; i < 32; i++ {
		err := awaitResult(t, results)
		if err == nil {
			successes++
		} else if err != io.ErrClosedPipe {
			t.Fatal(err)
		}
	}
	if successes != 1 {
		t.Fatalf("successful closes: %d", successes)
	}
}

func TestBufferedPipeWriterDuplexErrorClosesReader(t *testing.T) {
	downstream, p := io.Pipe()
	w := newBufferedPipeWriter(p, 4)
	reader, peer := io.Pipe()
	defer downstream.Close()
	defer p.Close()
	defer reader.Close()
	defer peer.Close()
	sentinel := errors.New("downstream failure")
	for i := 0; i < 5; i++ {
		w.Write([]byte("x"))
	}
	failed := make(chan error, 1)
	go func() { _, err := w.Write([]byte("queued")); failed <- err }()
	downstream.CloseWithError(sentinel)
	if err := awaitResult(t, failed); err != sentinel {
		t.Fatalf("writer terminal: %v", err)
	}
	d := &ioduplex{w, reader}
	blocked := make(chan error, 1)
	go func() { _, err := d.Read(make([]byte, 1)); blocked <- err }()
	if err := d.Close(); err != io.ErrClosedPipe {
		t.Fatalf("close error: %v", err)
	}
	if err := awaitResult(t, blocked); err != io.ErrClosedPipe {
		t.Fatalf("read error: %v", err)
	}
}

// closeErrorEnd lets both halves report independent Close failures.
type closeErrorEnd struct {
	name  string
	err   error
	calls *[]string
}

func (e closeErrorEnd) Read([]byte) (int, error)    { return 0, io.EOF }
func (e closeErrorEnd) Write(p []byte) (int, error) { return len(p), nil }
func (e closeErrorEnd) Close() error {
	*e.calls = append(*e.calls, e.name)
	return e.err
}

func TestIODuplexCloseErrorPrecedence(t *testing.T) {
	writerErr := errors.New("writer close failed")
	readerErr := errors.New("reader close failed")
	for _, tc := range []struct {
		name     string
		writeErr error
		want     error
	}{
		{"both fail", writerErr, writerErr},
		{"reader fails", nil, readerErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls []string
			d := &ioduplex{
				closeErrorEnd{"writer", tc.writeErr, &calls},
				closeErrorEnd{"reader", readerErr, &calls},
			}
			if err := d.Close(); err != tc.want {
				t.Fatalf("Close = %v, want %v", err, tc.want)
			}
			if len(calls) != 2 || calls[0] != "writer" || calls[1] != "reader" {
				t.Fatalf("close calls = %v, want [writer reader]", calls)
			}
		})
	}
}

type readStartedPipe struct {
	*io.PipeReader
	started chan struct{}
	once    sync.Once
}

func (r *readStartedPipe) Read(p []byte) (int, error) {
	r.once.Do(func() { close(r.started) })
	return r.PipeReader.Read(p)
}

func TestBufferedPipeWriterFailureSessionWait(t *testing.T) {
	downstream, p := io.Pipe()
	w := newBufferedPipeWriter(p, 1)
	reader, peer := io.Pipe()
	defer downstream.Close()
	defer w.Close()
	defer reader.Close()
	defer peer.Close()
	input := &readStartedPipe{PipeReader: reader, started: make(chan struct{})}
	s := New(&ioduplex{w, input})
	defer s.Close()
	select {
	case <-input.started:
	case <-time.After(2 * time.Second):
		t.Fatal("session never started reading")
	}

	// No downstream reader consumes this payload. Closing the downstream
	// reader must make the pump retain its error before the next write returns.
	if _, err := w.Write([]byte("payload")); err != nil {
		t.Fatal(err)
	}
	sentinel := errors.New("downstream failed")
	downstream.CloseWithError(sentinel)
	failed := make(chan error, 1)
	go func() {
		// At most one payload can remain in the queue; writes may succeed
		// until the pump observes the downstream failure.
		for i := 0; i < 3; i++ {
			if _, err := w.Write([]byte("next")); err != nil {
				failed <- err
				return
			}
		}
		failed <- nil
	}()
	if err := awaitResult(t, failed); err != sentinel {
		t.Fatalf("terminal writer error = %v, want %v", err, sentinel)
	}

	// A terminal writer makes Close fail; the read half must still close
	// so the actual session decoder and Wait can terminate.
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	waited := make(chan error, 1)
	go func() { waited <- s.Wait() }()
	if err := awaitResult(t, waited); err != io.ErrClosedPipe {
		t.Fatalf("Wait = %v, want %v", err, io.ErrClosedPipe)
	}
}
