package chain_test

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"io"
	"testing"

	"github.com/conradludgate/chain"
	"github.com/conradludgate/chain/archive"
	"github.com/conradludgate/chain/cipher"
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

func TestWritingToFS(t *testing.T) {
	key, err := hex.DecodeString("6368616e676520746869732070617373")
	require.Nil(t, err)
	aes := cipher.AESConfig{Key: key}

	wfs := chain.NewWriteBuilder(aes.Encrypt).
		WritingToFS(chain.OS{RootDir: "."})

	w, err := wfs.Create("hello.txt")
	require.Nil(t, err)

	_, err = io.WriteString(w, "hello world")
	require.Nil(t, err)
	err = w.Close()
	require.Nil(t, err)
	err = wfs.Close()
	require.Nil(t, err)
}
