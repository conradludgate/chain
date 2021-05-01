package chain_test

import (
	"bytes"
	"encoding/base64"
	"io"
	"testing"

	"github.com/conradludgate/chain"
	"github.com/conradludgate/chain/archive"
	"github.com/conradludgate/chain/encoding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadingFromFS(t *testing.T) {
	b64 := encoding.Base64Config{Encoding: base64.RawStdEncoding}

	fs, err := chain.ReadingFromFS(chain.OS{RootDir: "./example"}).Finally(b64.Decode)
	require.Nil(t, err)
	defer func() {
		require.Nil(t, fs.Close())
	}()

	r, _ := fs.Open("hello.txt")
	defer func() {
		require.Nil(t, fs.Close())
	}()

	output := bytes.NewBuffer(nil)
	_, err = io.Copy(output, r)
	require.Nil(t, err)

	assert.Equal(t, "hello world\n", output.String())
}

func TestReadingFromFS_Open(t *testing.T) {
	zip := archive.ZipConfig{}
	b64 := encoding.Base64Config{Encoding: base64.RawStdEncoding}

	r, err := chain.ReadingFromFS(chain.OS{RootDir: "./example"}).
		Open("archive.zip").
		AsFS(zip.FSReader).
		Open("hello.txt").
		Finally(b64.Decode)
	require.Nil(t, err)

	output := bytes.NewBuffer(nil)
	_, err = io.Copy(output, r)
	require.Nil(t, err)

	require.Nil(t, r.Close())

	assert.Equal(t, "hello world\n", output.String())
}
