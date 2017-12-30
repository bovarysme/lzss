package lzss

import (
	"io"
)

type writer struct {
	w      io.Writer
	window *Window

	mask byte

	// In the worst case scenario, 17 bytes have to be stored before being
	// flushed to the underlying io.Writer: one byte for the current flags, and
	// 16 bytes for 8 matches (2 bytes each).
	buffer [17]byte
	n      int

	err error
}

// NewWriter creates a new io.WriteCloser.
// Writes to the returned io.WriteCloser are compressed and written to w.
//
// It is the caller's responsibility to call Close on the io.WriteCloser when
// done, as writes may be buffered and not flushed until Close.
func NewWriter(w io.Writer) io.WriteCloser {
	return &writer{
		w:      w,
		window: NewWindow(),

		mask: 1,

		n: 1,
	}
}

// Write writes a compressed form of buffer to the underlying io.Writer. The
// compressed bytes are not necessarily flushed until the io.WriteCloser is closed.
func (w *writer) Write(buffer []byte) (int, error) {
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

			w.buffer[0] |= w.mask
		} else {
			w.buffer[w.n] = byte(offset & 0xff)
			w.n++

			w.buffer[w.n] = byte((offset&0xf00)>>4 | (length - minMatchLength))
			w.n++
		}

		w.mask <<= 1
		if w.mask == 0 {
			nn, err := w.flushBuffer()
			if err != nil {
				return n, err
			}

			n += nn
		}

		w.window.WriteBytes(buffer[pointer : pointer+length])

		pointer += length
	}

	return n, nil
}

// Close closes the io.WriteCloser, flushing any unwritten data to the underlying
// io.Writer. It does not close the underlying io.Writer.
func (w *writer) Close() error {
	if w.err != nil {
		return w.err
	}

	_, err := w.flushBuffer()
	if err != nil {
		return err
	}

	w.err = ErrClosed

	return nil
}

// flushBuffer flushes the io.WriteCloser to its underlying io.Writer.
func (w *writer) flushBuffer() (int, error) {
	n, err := w.w.Write(w.buffer[:w.n])
	if err != nil {
		return n, err
	}

	w.mask = 1

	w.buffer[0] = 0
	w.n = 1

	return n, nil
}
