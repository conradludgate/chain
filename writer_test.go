package chain_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/conradludgate/chain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const inputUpper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func TestWriterChainA(t *testing.T) {
	output := bytes.NewBuffer(nil)

	chainW, err := chain.NewWriteBuilder(RemoveABC).
		Then(ToLower).
		WritingTo(chain.NopWriteCloser{Writer: output})
	require.Nil(t, err)

	n, err := io.WriteString(chainW, inputUpper)
	require.Nil(t, err)
	assert.Equal(t, len(inputUpper), n)

	assert.Equal(t, "...defghijklmnopqrstuvwxyz", output.String())
}

func TestWriterChainB(t *testing.T) {
	output := bytes.NewBuffer(nil)

	chainW, err := chain.NewWriteBuilder(ToLower).
		Then(RemoveABC).
		WritingTo(chain.NopWriteCloser{Writer: output})
	require.Nil(t, err)

	n, err := io.WriteString(chainW, inputUpper)
	require.Nil(t, err)
	assert.Equal(t, len(inputUpper), n)

	assert.Equal(t, "abcdefghijklmnopqrstuvwxyz", output.String())
}

func RemoveABC(w io.WriteCloser) (io.WriteCloser, error) {
	return removeABC{w}, nil
}

type removeABC struct{ io.WriteCloser }

func (r removeABC) Write(p []byte) (n int, err error) {
	q := make([]byte, len(p))
	copy(q, p)
	for i, v := range q {
		if v >= 'A' && v <= 'C' {
			q[i] = '.'
		}
	}
	return r.WriteCloser.Write(q)
}

func ToLower(w io.WriteCloser) (io.WriteCloser, error) {
	return toLower{w}, nil
}

type toLower struct{ io.WriteCloser }

func (r toLower) Write(p []byte) (n int, err error) {
	q := make([]byte, len(p))
	copy(q, p)
	for i, v := range q {
		if v >= 'A' && v <= 'Z' {
			q[i] = v - 'A' + 'a'
		}
	}
	return r.WriteCloser.Write(q)
}
