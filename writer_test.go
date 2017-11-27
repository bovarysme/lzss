package lzss

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func readFile(tb testing.TB, name string) []byte {
	path := filepath.Join("testdata", name)
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		tb.Fatal(err)
	}

	return bytes
}

func TestReverse(t *testing.T) {
	for i := byte(0); i < 255; i++ {
		id := reverse(reverse(i))
		if i != id {
			t.Fatalf("identity is %d, want %d", id, i)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	var filenames = []string{
		"gettysburg.txt",
		"Mark.Twain-Tom.Sawyer.txt",
		"purple_heart.raw",
	}

	for _, filename := range filenames {
		buffer := new(bytes.Buffer)

		writer := NewWriter(buffer)

		want := readFile(t, filename)
		_, err := writer.Write(want)
		if err != nil {
			t.Fatalf("Writer.Write: %v", err)
		}

		err = writer.Close()
		if err != nil {
			t.Fatalf("Writer.Close: %v", err)
		}

		reader := NewReader(buffer)

		got, err := ioutil.ReadAll(reader)
		if err != nil {
			t.Fatalf("Reader.Read: %v", err)
		}

		err = reader.Close()
		if err != nil {
			t.Fatalf("Reader.Close: %v", err)
		}

		if !bytes.Equal(got, want) {
			t.Errorf("%q: compressing and then decompressing is not the identity function",
				filename)
		}
	}
}

func TestWriterClosed(t *testing.T) {
	buffer := new(bytes.Buffer)

	writer := NewWriter(buffer)

	_, err := writer.Write([]byte("hello, hello, world\n"))
	if err != nil {
		t.Fatalf("Writer.Write: %v", err)
	}

	err = writer.Close()
	if err != nil {
		t.Fatalf("Writer.Close: %v", err)
	}

	_, err = writer.Write([]byte("hello again, world\n"))
	if err != ErrClosed {
		t.Fatalf("Writer.Write after Close: %v", err)
	}

	err = writer.Close()
	if err != ErrClosed {
		t.Fatalf("Writer.Close after Close: %v", err)
	}
}
