package encoding

import (
	"encoding/base64"
	"io"
)

type Base64Config struct {
	Encoding *base64.Encoding
}

func (cfg Base64Config) Encode(w io.Writer) (io.Writer, error) {
	return base64.NewEncoder(cfg.Encoding, w), nil
}

func (cfg Base64Config) Decode(r io.Reader) (io.Reader, error) {
	return base64.NewDecoder(cfg.Encoding, r), nil
}
