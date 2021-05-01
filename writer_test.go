package chain_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/conradludgate/chain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriterChainA(t *testing.T) {
	output := bytes.NewBuffer(nil)
	input := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	chainW, err := chain.NewWriteBuilder(RemoveABC).Then(ToLower).WritingTo(output)
	require.Nil(t, err)

	n, err := io.WriteString(chainW, input)
	require.Nil(t, err)
	assert.Equal(t, len(input), n)

	assert.Equal(t, "...defghijklmnopqrstuvwxyz", output.String())
}

func TestWriterChainB(t *testing.T) {
	output := bytes.NewBuffer(nil)
	input := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	chainW, err := chain.NewWriteBuilder(ToLower).Then(RemoveABC).WritingTo(output)
	require.Nil(t, err)

	n, err := io.WriteString(chainW, input)
	require.Nil(t, err)
	assert.Equal(t, len(input), n)

	assert.Equal(t, "abcdefghijklmnopqrstuvwxyz", output.String())
}

func RemoveABC(w io.Writer) (io.Writer, error) {
	return removeABC{w}, nil
}

type removeABC struct{ io.Writer }

func (r removeABC) Write(p []byte) (n int, err error) {
	q := make([]byte, len(p))
	copy(q, p)
	for i, v := range q {
		if v >= 'A' && v <= 'C' {
			q[i] = '.'
		}
	}
	return r.Writer.Write(q)
}

func ToLower(w io.Writer) (io.Writer, error) {
	return toLower{w}, nil
}

type toLower struct{ io.Writer }

func (r toLower) Write(p []byte) (n int, err error) {
	q := make([]byte, len(p))
	copy(q, p)
	for i, v := range q {
		if v >= 'A' && v <= 'Z' {
			q[i] = v - 'A' + 'a'
		}
	}
	return r.Writer.Write(q)
}
