package lzss

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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
	var testCases = []struct {
		filename       string
		compressedSize int
	}{
		{"gettysburg.txt", 1059},
		{"Mark.Twain-Tom.Sawyer.txt", 207910},
		{"purple_heart.raw", 6536},
	}

	for _, testCase := range testCases {
		basename := testCase.filename
		name := strings.TrimSuffix(basename, filepath.Ext(basename)) + ".output"
		path := filepath.Join("testdata", name)

		file, err := os.Create(path)
		if err != nil {
			t.Fatalf("os.Create: %v", err)
		}

		writer := NewWriter(file)

		want := readFile(t, testCase.filename)
		n, err := writer.Write(want)
		if err != nil {
			t.Fatalf("Writer.Write: %v", err)
		}

		err = writer.Close()
		if err != nil {
			t.Fatalf("Writer.Close: %v", err)
		}

		if n != testCase.compressedSize {
			t.Errorf("%q: wrote %d bytes, want %d", testCase.filename, n, testCase.compressedSize)
		}

		err = file.Close()
		if err != nil {
			t.Fatalf("file.Close: %v", err)
		}

		compressed := readFile(t, name)
		br := bytes.NewReader(compressed)
		reader := NewReader(br)

		got, err := ioutil.ReadAll(reader)
		if err != nil {
			t.Fatalf("Reader.Read: %v", err)
		}

		if !bytes.Equal(got, want) {
			t.Errorf("%q: compressing and then decompressing is not the identity function",
				testCase.filename)
		}
	}
}
