package lzss

import (
	"bufio"
	"errors"
	"io"
)

var ErrClosed = errors.New("lzss: reader/writer is closed")

type ByteReadReader interface {
	io.ByteReader
	io.Reader
}

type Reader struct {
	r      ByteReadReader
	window *Window

	flags int
	match [2]byte

	buffer [maxMatchLength]byte
	n      int
	toRead []byte

	err error
}

func NewReader(r io.Reader) *Reader {
	reader := new(Reader)

	br, ok := r.(ByteReadReader)
	if ok {
		reader.r = br
	} else {
		reader.r = bufio.NewReader(r)
	}

	reader.window = NewWindow()
	reader.flags = 1

	return reader
}

func (r *Reader) Read(buffer []byte) (int, error) {
	if r.err != nil {
		return 0, r.err
	}

	if len(r.toRead) > 0 {
		n := copy(buffer, r.toRead)
		r.toRead = r.toRead[n:]

		return n, nil
	}

	n, err := r.read(buffer)

	return n, err
}

func (r *Reader) read(buffer []byte) (int, error) {
	n := 0

	for n < len(buffer) {
		if r.flags == 1 {
			b, err := r.r.ReadByte()
			if err != nil {
				return n, err
			}

			r.flags = 0x100 | int(b)
		}

		if r.flags&1 == 1 {
			b, err := r.r.ReadByte()
			if err != nil {
				return n, err
			}

			buffer[n] = b
			n++

			r.window.WriteByte(b)
		} else {
			_, err := r.r.Read(r.match[:])
			if err != nil {
				return n, err
			}

			offset := int(r.match[1])&0xf0<<4 | int(r.match[0])
			length := int(r.match[1])&0x0f + minMatchLength

			for i := 0; i < length; i++ {
				b, err := r.window.ReadByte(&offset)
				if err != nil {
					return n, err
				}

				if n < len(buffer) {
					buffer[n] = b
					n++
				} else {
					r.buffer[r.n] = b
					r.n++
				}

				r.window.WriteByte(b)
			}
		}

		r.flags >>= 1
	}

	r.toRead = r.buffer[:r.n]
	r.n = 0

	return n, nil
}

func (r *Reader) Close() error {
	if r.err != nil {
		return r.err
	}

	r.err = ErrClosed

	return nil
}
