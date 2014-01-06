// broadcast package
// code from dotcloud/docker
package main

import (
	"bytes"
	"io"
	"sync"
)

type StreamWriter struct {
	wc     io.WriteCloser
	stream string
}

type WriteBroadcaster struct {
	sync.Mutex
	buf     *bytes.Buffer
	writers map[StreamWriter]bool
}

func NewWriteBroadcaster() *WriteBroadcaster {
	return &WriteBroadcaster{
		writers: make(map[StreamWriter]bool),
		buf:     bytes.NewBuffer(nil),
	}
}

func (w *WriteBroadcaster) AddWriter(writer io.WriteCloser, stream string) {
	w.Lock()
	sw := StreamWriter{wc: writer, stream: stream}
	w.writers[sw] = true
	w.Unlock()
}

func (wb *WriteBroadcaster) NewReader(name string) ([]byte, *io.PipeReader) {
	r, w := io.Pipe()
	wb.AddWriter(w, name)
	return wb.buf.Bytes(), r
}
func (w *WriteBroadcaster) Write(p []byte) (n int, err error) {
	w.Lock()
	defer w.Unlock()
	w.buf.Write(p)
	for sw := range w.writers {
		if n, err := sw.wc.Write(p); err != nil || n != len(p) {
			// On error, evict the writer
			delete(w.writers, sw)
		}
	}
	return len(p), nil
}

func (w *WriteBroadcaster) CloseWriters() error {
	w.Lock()
	defer w.Unlock()
	for sw := range w.writers {
		sw.wc.Close()
	}
	w.writers = make(map[StreamWriter]bool)
	return nil
}

// nop writer
type NopWriter struct{}

func (*NopWriter) Write(buf []byte) (int, error) {
	return len(buf), nil
}

type nopWriteCloser struct {
	io.Writer
}

func (w *nopWriteCloser) Close() error { return nil }

func NopWriteCloser(w io.Writer) io.WriteCloser {
	return &nopWriteCloser{w}
}
