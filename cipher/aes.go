package cipher

import (
	"crypto/aes"
	"crypto/cipher"
	"io"

	"github.com/conradludgate/chain"
)

type AESConfig struct {
	Key []byte
	IV  [aes.BlockSize]byte
}

func (cfg AESConfig) Stream() (cipher.Stream, error) {
	block, err := aes.NewCipher(cfg.Key)
	if err != nil {
		return nil, err
	}
	var iv [aes.BlockSize]byte
	copy(iv[:], cfg.IV[:])
	return cipher.NewOFB(block, iv[:]), nil
}

func (cfg AESConfig) Encrypt(w io.WriteCloser) (io.WriteCloser, error) {
	s, err := cfg.Stream()
	return cipher.StreamWriter{S: s, W: w}, err
}

func (cfg AESConfig) Decrypt(r io.ReadCloser) (io.ReadCloser, error) {
	s, err := cfg.Stream()
	return chain.ReadCloser{
		Reader: cipher.StreamReader{S: s, R: r},
		Closer: r,
	}, err
}
