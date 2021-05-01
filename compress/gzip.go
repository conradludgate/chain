package compress

import (
	"compress/gzip"
	"io"
)

func GZIPCompress(w io.Writer) (io.Writer, error) {
	return gzip.NewWriter(w), nil
}

func GZIPDecompress(r io.Reader) (io.Reader, error) {
	return gzip.NewReader(r)
}
