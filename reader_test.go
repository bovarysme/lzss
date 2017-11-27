package lzss

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestRead(t *testing.T) {
	compressed := []byte("\x7fhello, \xee\xf4\xffworld\n")
	want := []byte("hello, hello, world\n")

	br := bytes.NewReader(compressed)
	reader := NewReader(br)

	got, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Fatalf("Reader.Read: %v", err)
	}

	if !bytes.Equal(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestReaderClosed(t *testing.T) {
	br := new(bytes.Reader)
	reader := NewReader(br)

	err := reader.Close()
	if err != nil {
		t.Fatalf("Reader.Close: %v", err)
	}

	_, err = reader.Read([]byte{})
	if err != ErrClosed {
		t.Fatalf("Reader.Read after Close: %v", err)
	}

	err = reader.Close()
	if err != ErrClosed {
		t.Fatalf("Reader.Close after Close: %v", err)
	}
}
