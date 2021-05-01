package encoding

import (
	"encoding/hex"
	"io"
)

type h struct{}

var Hex h

func (h) Encode(w io.Writer) (io.Writer, error) {
	return hex.NewEncoder(w), nil
}

func (h) Decode(r io.Reader) (io.Reader, error) {
	return hex.NewDecoder(r), nil
}
