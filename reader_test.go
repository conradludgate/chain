package chain_test

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/conradludgate/chain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReaderChainA(t *testing.T) {
	input := "abcdefghijklmnopqrstuvwxyz"
	chainR, err := chain.ReadingFrom(strings.NewReader(input)).Then(RemoveXYZ).Finally(ToUpper)
	require.Nil(t, err)
	b, err := ioutil.ReadAll(chainR)
	require.Nil(t, err)

	assert.Equal(t, "ABCDEFGHIJKLMNOPQRSTUVW...", string(b))
}

func TestReaderChainB(t *testing.T) {
	input := "abcdefghijklmnopqrstuvwxyz"
	chainR, err := chain.ReadingFrom(strings.NewReader(input)).Then(ToUpper).Finally(RemoveXYZ)
	require.Nil(t, err)
	b, err := ioutil.ReadAll(chainR)
	require.Nil(t, err)

	assert.Equal(t, "ABCDEFGHIJKLMNOPQRSTUVWXYZ", string(b))
}

func RemoveXYZ(w io.Reader) (io.Reader, error) {
	return removeXYZ{w}, nil
}

type removeXYZ struct{ io.Reader }

func (r removeXYZ) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	for i := 0; i < n; i++ {
		if p[i] >= 'x' && p[i] <= 'z' {
			p[i] = '.'
		}
	}
	return
}

func ToUpper(w io.Reader) (io.Reader, error) {
	return toUpper{w}, nil
}

type toUpper struct{ io.Reader }

func (r toUpper) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	for i := 0; i < n; i++ {
		if p[i] >= 'a' && p[i] <= 'z' {
			p[i] -= 'a' - 'A'
		}
	}
	return
}