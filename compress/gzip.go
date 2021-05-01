package compress

import (
	"compress/gzip"
	"io"
)

type GZIPConfig struct {
	Header gzip.Header
	level  int
}

func (c GZIPConfig) WithLevel(level int) GZIPConfig {
	// Annoyingly, 'gzip.DefaultCompression` is the contant -1... not 0
	// Go ints default to 0 so adding 1 makes them inline
	c.level = level + 1
	return c
}

func (c GZIPConfig) Compress(w io.Writer) (io.Writer, error) {
	gw, err := gzip.NewWriterLevel(w, c.level-1)
	if err != nil {
		return nil, err
	}
	gw.Header = c.Header
	return gw, nil
}

func (c GZIPConfig) Decompress(r io.Reader) (io.Reader, error) {
	return gzip.NewReader(r)
}
