package lzss

import (
	"errors"
)

const (
	minMatchLength = 3
	maxMatchLength = 18

	logWindowSize = 12
	windowSize    = 1 << logWindowSize
	windowMask    = windowSize - 1
	initPointer   = windowSize - maxMatchLength
)

type Window struct {
	buffer  []byte
	pointer int
}

func NewWindow() *Window {
	return &Window{
		buffer:  make([]byte, windowSize),
		pointer: initPointer,
	}
}

// Assuming: len(buffer) <= maxMatchLength
func (window *Window) FindMatch(buffer []byte) (int, int) {
	// XXX: bounds checking buffer optimizes away a call to runtime.panicindex
	// when accessing buffer[0]
	if len(buffer) <= 0 {
		return 0, 0
	}

	length := 0
	offset := 0

	curLength := 0

	pointer := window.pointer

	for {
		pointer = (pointer - 1) & windowMask
		if pointer == window.pointer {
			break
		}

		if buffer[0] != window.buffer[pointer] {
			continue
		}

		for curLength = 1; curLength < len(buffer); curLength++ {
			tempPointer := (pointer + curLength) & windowMask
			if tempPointer == window.pointer {
				break
			}

			if buffer[curLength] != window.buffer[tempPointer] {
				break
			}
		}

		if curLength > length {
			length = curLength
			offset = pointer

			if offset == (window.pointer-length)&windowMask {
				for ; curLength < len(buffer); curLength++ {
					tempPointer := (offset + curLength%length) & windowMask
					if buffer[curLength] != window.buffer[tempPointer] {
						break
					}
				}

				length = curLength
			}

			if length >= maxMatchLength {
				break
			}
		}
	}

	return length, offset
}

func (window *Window) ReadByte(offset *int) (byte, error) {
	if *offset < 0 || *offset >= windowSize {
		return 0, errors.New("Window.ReadByte: index out of range")
	}

	b := window.buffer[*offset]
	*offset = (*offset + 1) & windowMask

	return b, nil
}

func (window *Window) WriteByte(b byte) {
	window.buffer[window.pointer] = b
	window.pointer = (window.pointer + 1) & windowMask
}

func (window *Window) WriteBytes(bytes []byte) {
	for i := 0; i < len(bytes); i++ {
		window.WriteByte(bytes[i])
	}
}
