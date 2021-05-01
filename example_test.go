package chain_test

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/conradludgate/chain"
	"github.com/conradludgate/chain/archive"
	"github.com/conradludgate/chain/cipher"
	"github.com/conradludgate/chain/compress"
	"github.com/conradludgate/chain/encoding"
)

func ExampleNewWriteBuilder() {
	key, _ := hex.DecodeString("6368616e676520746869732070617373")
	aes := cipher.AESConfig{Key: key}
	gzip := compress.GZIPConfig{}

	output := bytes.NewBuffer(nil)

	w, _ := chain.NewWriteBuilder(aes.Encrypt).
		Then(gzip.Compress).
		WritingTo(output)

	io.WriteString(w, "hello world")
	w.Close()

	r, _ := chain.ReadingFrom(output).
		Then(gzip.Decompress).
		Finally(aes.Decrypt)

	b, _ := io.ReadAll(r)
	defer r.Close()

	fmt.Println(string(b))
	// Output: hello world

}

func ExampleWriterBuilder_IntoFS() {
	key, _ := hex.DecodeString("6368616e676520746869732070617373")
	aes := cipher.AESConfig{Key: key}
	zip := archive.ZipConfig{}

	output := bytes.NewBuffer(nil)

	wfs, _ := chain.NewWriteBuilder(aes.Encrypt).
		IntoFS(zip.FSWriter).
		WritingTo(output)

	w, _ := wfs.Create("hello.txt")

	io.WriteString(w, "hello world")
	w.Close()
	wfs.Close()

	rfs, _ := chain.ReadingFrom(output).
		AsFS(zip.FSReader).
		Finally(aes.Decrypt)

	r, _ := rfs.Open("hello.txt")

	b, _ := io.ReadAll(r)

	r.Close()
	rfs.Close()

	fmt.Println(string(b))
	// Output: hello world

}

func ExampleReaderBuilder_AsFS() {
	zip := archive.ZipConfig{}
	b64 := encoding.Base64Config{Encoding: base64.RawStdEncoding}

	fs, _ := chain.ReadingFromFS(chain.OS{RootDir: "./example"}).
		Open("archive.zip").
		AsFS(zip.FSReader).
		Finally(b64.Decode)
	defer fs.Close()

	r, _ := fs.Open("hello.txt")
	defer r.Close()
	io.Copy(os.Stdout, r)
}

func ExampleReadingFromFS() {
	b64 := encoding.Base64Config{Encoding: base64.RawStdEncoding}

	fs, _ := chain.ReadingFromFS(chain.OS{RootDir: "./example"}).Finally(b64.Decode)
	defer fs.Close()

	r, _ := fs.Open("hello.txt")
	defer r.Close()
	io.Copy(os.Stdout, r)
}
