/*
Package lzss implements the Lempel-Ziv-Storer-Szymanski compressed data format,
described in J. A. Storer, ``Data Compression via Textual Substitution'',
Journal of the ACM, 29(4) (October 1982), pp 928-951.

This example illustrates how to write compressed data to a buffer:

	buffer := new(bytes.Buffer)
	writer := lzss.NewWriter(buffer)
	writer.Write([]byte("hello, hello, world\n"))
	writer.Close()

and how to read that data back:

	reader := lzss.NewReader(buffer)
	io.Copy(os.Stdout, reader)
	reader.Close()

*/
package lzss
