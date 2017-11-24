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

func (window *Window) FindMatch(buffer []byte) (int, int) {
	length := 0
	offset := 0

	pointer := (window.pointer - 1) & windowMask

	for pointer != window.pointer {
		currentLength := 0
		for i := 0; i < len(buffer); i++ {
			tempPointer := (pointer + i) & windowMask
			if tempPointer == window.pointer {
				break
			}

			if buffer[i] != window.buffer[tempPointer] {
				break
			}

			currentLength++
		}

		if currentLength > length {
			length = currentLength
			offset = pointer

			if offset == (window.pointer-length)&windowMask {
				currentLength := 0
				repetitions := 1
				for i := length; i < len(buffer); i++ {
					tempPointer := (offset + currentLength) & windowMask
					if buffer[i] != window.buffer[tempPointer] {
						break
					}

					currentLength++
					if currentLength >= length {
						currentLength = 0
						repetitions++
					}

					if length*repetitions+currentLength >= maxMatchLength {
						break
					}
				}

				length = length*repetitions + currentLength
			}
		}

		if length >= maxMatchLength {
			break
		}

		pointer = (pointer - 1) & windowMask
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
