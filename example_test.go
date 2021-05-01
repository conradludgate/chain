package chain_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/conradludgate/chain"
	"github.com/conradludgate/chain/cipher"
	"github.com/conradludgate/chain/compress"
)

func Example() {
	key, _ := hex.DecodeString("6368616e676520746869732070617373")
	aes := cipher.AESConfig{Key: key}

	output := bytes.NewBuffer(nil)

	w, _ := chain.NewWriteBuilder(aes.Encrypt).
		Then(compress.GZIPCompress).
		WritingTo(output)

	io.WriteString(w, "hello world")
	w.Close()

	r, _ := chain.ReadingFrom(output).
		Then(compress.GZIPDecompress).
		Finally(aes.Decrypt)

	b, _ := io.ReadAll(r)
	r.Close()

	fmt.Println(string(b))
	// Output: hello world
}