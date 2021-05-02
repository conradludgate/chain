# chain

Chain together io.Reader/Writer with ease

[![Docs](https://img.shields.io/github/v/tag/conradludgate/chain?label=docs&style=flat-square)](https://pkg.go.dev/github.com/conradludgate/chain)
[![Coverage](https://img.shields.io/codecov/c/gh/conradludgate/chain?style=flat-square)](https://app.codecov.io/gh/conradludgate/chain)

## Examples

### Simple chain

Let's say you want to compress some data using GZip, then encrypt that using AES,
finally writing the contents to a file, we can do the following

```go
// Setup the encoding configs
key, _ := hex.DecodeString("6368616e676520746869732070617373")
aes := cipher.AESConfig{Key: key}
gzip := compress.GZIPConfig{}

f, _ := os.Open("data.gzip.encrypted")

w, _ := chain.NewWriterBuilder(gzip.Compress).
    Then(aes.Encrypt).
    WritingTo(f)

// Closing `w` will automatically close `f`
defer w.Close()

io.WriteString(w, "hello world")
```

To then read the data, you can build the opposite chain in reverse.

```go
// Setup the decoding configs
key, _ := hex.DecodeString("6368616e676520746869732070617373")
aes := cipher.AESConfig{Key: key}
gzip := compress.GZIPConfig{}

f, _ := os.Open("data.gzip.encrypted")

r, _ := chain.ReadingFrom(f).
    Then(aes.Decrypt).
    Finally(gzip.Decompress)

// Closing `r` will automatically close `f`
defer r.Close()

io.Copy(os.Stdout, r)
// Output: hello world
```

### File system chain

This chaining system also has some simple file system support to build some more
complex chains.

Encode the text `hello world\n` in base64, write it to a file called `hello.txt`
in a zip folder, saved to the OS at `./example/archive.zip`

* example/
    * archive.zip
        * hello.txt
            * aGVsbG8gd29ybGQK (`hello world\n` in base64)

```go
// Setup the encoding configs
zip := archive.ZipConfig{}
b64 := encoding.Base64Config{Encoding: base64.RawStdEncoding}

w, _ := chain.NewWriterBuilder(b64.Encode).
    Create("hello.txt").
    InFS(zip.FSWriter).
    Create("archive.zip").
    WritingToFS(chain.OS{RootDir: "./example"})
defer w.Close()

io.WriteString(w, "hello world\n")
```

The opposite of the above, reading the file

```go
// Setup the decoding configs
zip := archive.ZipConfig{}
b64 := encoding.Base64Config{Encoding: base64.RawStdEncoding}

// Build the reader chain
r, _ := chain.ReadingFromFS(chain.OS{RootDir: "./example"}).
    Open("archive.zip").
    AsFS(zip.FSReader).
    Open("hello.txt").
    Finally(b64.Decode)
defer r.Close()

io.Copy(os.Stdout, r)
// Output: hello world
```

### Multi file chain

Encode the text `hello world\n` in base64, write it to a file called `hello.txt`
in a zip folder, saved to the OS at `./example/archive.zip`

* example/
    * archive.zip
        * hello.txt
            * aGVsbG8gd29ybGQK (`hello world\n` in base64)
        * goodbye.txt
            * Z29vZGJ5ZSB3b3JsZAo= (`goodbye world\n` in base64)

```go
// Setup the encoding configs
zip := archive.ZipConfig{}
b64 := encoding.Base64Config{Encoding: base64.RawStdEncoding}

fs, _ := chain.NewWriterBuilder(b64.Encode).
    IntoFS(zip.FSWriter).
    Create("archive.zip").
    WritingToFS(chain.OS{RootDir: "./example"})
defer fs.Close()

f1, _ := fs.Create("hello.txt")
io.WriteString(f1, "hello world\n")
f1.Close()

f2, _ := fs.Create("goodbye.txt")
io.WriteString(f2, "goodbye world\n")
f2.Close()
```

The opposite of the above, reading the file

```go
// Setup the decoding configs
zip := archive.ZipConfig{}
b64 := encoding.Base64Config{Encoding: base64.RawStdEncoding}

// Build the reader chain
fs, _ := chain.ReadingFromFS(chain.OS{RootDir: "./example"}).
    Open("archive.zip").
    AsFS(zip.FSReader).
    Finally(b64.Decode)
defer fs.Close()

f1, _ := fs.Open("hello.txt")
io.Copy(os.Stdout, f1)
// Output: hello world
f1.Close()

f2, _ := fs.Open("goodbye.txt")
io.Copy(os.Stdout, f2)
// Output: goodbye world
f2.Close()
```