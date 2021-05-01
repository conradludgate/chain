package encoding

import (
	"encoding/base64"
	"io"

	"github.com/conradludgate/chain"
)

type Base64Config struct {
	Encoding *base64.Encoding
}

func (cfg Base64Config) Encode(w io.WriteCloser) (io.WriteCloser, error) {
	return chain.WriteCloser2{
		WriteCloser: base64.NewEncoder(cfg.Encoding, w),
		Closer:      w,
	}, nil
}

func (cfg Base64Config) Decode(r io.ReadCloser) (io.ReadCloser, error) {
	return chain.ReadCloser{
		Reader: base64.NewDecoder(cfg.Encoding, r),
		Closer: r,
	}, nil
}
