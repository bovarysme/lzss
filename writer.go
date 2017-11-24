package lzss

import (
	"errors"
	"io"
)

var ErrWriterClosed = errors.New("lzss: writer is closed")

type Writer struct {
	w      io.WriteSeeker
	window *Window

	flags  int
	offset int64
	match  [2]byte

	init bool
	err  error
}

func NewWriter(w io.WriteSeeker) *Writer {
	return &Writer{
		w:      w,
		window: NewWindow(),

		flags: 1,
	}
}

func (w *Writer) Write(buffer []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}

	n := 0

	if !w.init {
		w.init = true

		err := w.initFlags()
		if err != nil {
			return n, err
		}

		n++
	}

	pointer := 0

	for pointer < len(buffer) {
		length, offset := w.window.FindMatch(buffer[pointer : pointer+maxMatchLength])

		if length < minMatchLength {
			length = 1

			_, err := w.w.Write(buffer[pointer : pointer+1])
			if err != nil {
				return n, err
			}

			n++
			w.flags = w.flags<<1 | 1
		} else {
			w.match[0] = byte(offset & 0xff)
			w.match[1] = byte((offset&0xf00)>>4 | (length - minMatchLength))
			_, err := w.w.Write(w.match[:])
			if err != nil {
				return n, err
			}

			n += 2
			w.flags = w.flags<<1 | 0
		}

		if w.flags>>8 == 1 {
			err := w.writeFlags()
			if err != nil {
				return n, err
			}

			err = w.writeByte(0)
			if err != nil {
				return n, err
			}

			n++
		}

		w.window.WriteBytes(buffer[pointer : pointer+length])

		pointer += length
	}

	return n, nil
}

func (w *Writer) Close() error {
	if w.err != nil {
		return w.err
	}

	for w.flags>>8 != 1 {
		w.flags = w.flags<<1 | 1
	}

	err := w.writeFlags()
	if err != nil {
		return err
	}

	w.err = ErrWriterClosed

	return nil
}

func (w *Writer) writeByte(b byte) error {
	var buffer [1]byte
	buffer[0] = b

	_, err := w.w.Write(buffer[:])

	return err
}

func (w *Writer) initFlags() error {
	offset, err := w.w.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	w.offset = offset

	err = w.writeByte(0)

	return err
}

func (w *Writer) writeFlags() error {
	var flags byte

	offset, err := w.w.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	_, err = w.w.Seek(w.offset, io.SeekStart)
	if err != nil {
		return err
	}

	flags = reverse(byte(w.flags & 0xff))

	err = w.writeByte(flags)
	if err != nil {
		return err
	}

	_, err = w.w.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}
	w.offset = offset

	w.flags = 1

	return nil
}

func reverse(b byte) byte {
	b = b&0x55<<1 | b&0xaa>>1
	b = b&0x33<<2 | b&0xcc>>2
	b = b&0x0f<<4 | b&0xf0>>4

	return b
}
