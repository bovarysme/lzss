package lzss

import (
	"errors"
	"io"
)

var ErrWriterClosed = errors.New("lzss: writer is closed")

// A Writer takes data written to it and writes the compressed form of that
// data to an underlying io.Writer (see NewWriter).
type Writer struct {
	w      io.Writer
	window *Window

	flags int

	// In the worst case scenario, 17 bytes have to be stored before being
	// flushed to the underlying io.Writer: one byte for the current flags, and
	// 16 bytes for 8 matches (2 bytes each).
	buffer [17]byte
	n      int

	err error
}

// NewWriter creates a new Writer.
// Writes to the returned Writer are compressed and written to w.
//
// It is the caller's responsibility to call Close on the WriteCloser when
// done, as writes may be buffered and not flushed until Close.
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w:      w,
		window: NewWindow(),

		flags: 1,

		n: 1,
	}
}

// Write writes a compressed form of buffer to the underlying io.Writer. The
// compressed bytes are not necessarily flushed until the Writer is closed.
func (w *Writer) Write(buffer []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}

	n := 0
	pointer := 0

	for pointer < len(buffer) {
		end := pointer + maxMatchLength
		if end > len(buffer) {
			end = len(buffer)
		}

		length, offset := w.window.FindMatch(buffer[pointer:end])

		if length < minMatchLength {
			length = 1

			w.buffer[w.n] = buffer[pointer]
			w.n++

			w.flags = w.flags<<1 | 1
		} else {
			w.buffer[w.n] = byte(offset & 0xff)
			w.n++

			w.buffer[w.n] = byte((offset&0xf00)>>4 | (length - minMatchLength))
			w.n++

			w.flags = w.flags<<1 | 0
		}

		if w.flags>>8 == 1 {
			m, err := w.flushBuffer()
			if err != nil {
				return n, err
			}

			n += m
		}

		w.window.WriteBytes(buffer[pointer : pointer+length])

		pointer += length
	}

	return n, nil
}

// Close closes the Writer, flushing any unwritten data to the underlying
// io.Writer. It does not close the underlying io.Writer.
func (w *Writer) Close() error {
	if w.err != nil {
		return w.err
	}

	_, err := w.flushBuffer()
	if err != nil {
		return err
	}

	w.err = ErrWriterClosed

	return nil
}

// flushBuffer flushes the Writer to its underlying io.Writer.
func (w *Writer) flushBuffer() (int, error) {
	for w.flags>>8 != 1 {
		w.flags = w.flags<<1 | 1
	}

	w.buffer[0] = reverse(byte(w.flags & 0xff))

	n, err := w.w.Write(w.buffer[:w.n])
	if err != nil {
		return n, err
	}

	w.flags = 1
	w.n = 1

	return n, nil
}

// reverse reverses bits in a byte.
func reverse(b byte) byte {
	b = b&0x55<<1 | b&0xaa>>1
	b = b&0x33<<2 | b&0xcc>>2
	b = b&0x0f<<4 | b&0xf0>>4

	return b
}
