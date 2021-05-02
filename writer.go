package chain

import (
	"io"
)

// WriterBuilder lets you build a chain of io.Writers
// in a more natural way
type WriterBuilder struct {
	wcs []WriteChain
}

// WriteChain represents a common pattern in go packages.
// For examples, see:
// https://pkg.go.dev/compress/gzip#NewWriter
// https://pkg.go.dev/golang.org/x/crypto/openpgp#Encrypt
//
// While this type only expects an io.Writer to be returned, it will
// detect if the result is a Closer and make sure the
// types Close function is run.
type WriteChain func(io.WriteCloser) (io.WriteCloser, error)

// NewWriteBuilder creates a new WriteBuilder with the
// given WriteChain being the first in the chain
func NewWriteBuilder(first WriteChain) *WriterBuilder {
	return &WriterBuilder{
		wcs: []WriteChain{first},
	}
}

// Then adds the next WriteChain to the current builder chain.
// Returns self
func (wc *WriterBuilder) Then(next WriteChain) *WriterBuilder {
	wc.wcs = append(wc.wcs, next)
	return wc
}

// WritingTo builds the chain. The resulting data from the chain is
// written to the io.Writer provided.
//
// If w is a Closer type too, calling the returned writer's Close function
// will also close w.
func (wc *WriterBuilder) WritingTo(w io.WriteCloser) (io.WriteCloser, error) {
	for i := len(wc.wcs) - 1; i >= 0; i-- {
		newW, err := wc.wcs[i](w)
		if err != nil {
			w.Close()
			return nil, err
		}

		w = newW
	}

	return w, nil
}

type WriterFileBuilder struct {
	builder *WriterBuilder
	name    string
}

func (wc *WriterBuilder) Create(name string) *WriterFileBuilder {
	return &WriterFileBuilder{
		builder: wc,
		name:    name,
	}
}

func (builder *WriterFileBuilder) InFS(next WriteFSChain) *WriterBuilder {
	return builder.builder.Then(func(w io.WriteCloser) (io.WriteCloser, error) {
		fs, err := next(w)
		if err != nil {
			return nil, err
		}

		f, err := fs.Create(builder.name)
		if err != nil {
			fs.Close()
			return nil, err
		}

		return WriteCloser2{
			WriteCloser: f,
			Closer:      fs,
		}, nil
	})
}
func (builder *WriterFileBuilder) WritingToFS(fs WriteFS) (io.WriteCloser, error) {
	f, err := fs.Create(builder.name)
	if err != nil {
		fs.Close()
		return nil, err
	}

	return builder.builder.WritingTo(WriteCloser2{
		WriteCloser: f,
		Closer:      fs,
	})
}

type WriteFSBuilder struct {
	first *WriterBuilder
	fs    WriteFSChain
	after *WriterBuilder
}

type WriteFSChain func(io.WriteCloser) (WriteFS, error)

func (wc *WriterBuilder) IntoFS(next WriteFSChain) *WriteFSBuilder {
	return &WriteFSBuilder{
		first: wc,
		fs:    next,
		after: &WriterBuilder{},
	}
}

func (wc *WriterBuilder) WritingToFS(fs WriteFS) WriteFS {
	return &writeFs{
		close: NopWriteCloser{},
		first: wc,
		fs:    fs,
	}
}

func (wc *WriteFSBuilder) Then(next WriteChain) *WriteFSBuilder {
	wc.after.Then(next)
	return wc
}

func (wc *WriteFSBuilder) WritingTo(w io.WriteCloser) (WriteFS, error) {
	after, err := wc.after.WritingTo(w)
	if err != nil {
		return nil, err
	}

	fs, err := wc.fs(after)
	if err != nil {
		after.Close()
		return nil, err
	}

	return &writeFs{
		first: wc.first,
		fs:    fs,
		close: after,
	}, nil
}

type writeFs struct {
	close io.Closer
	fs    WriteFS
	first *WriterBuilder
}

func (fs *writeFs) Create(path string) (io.WriteCloser, error) {
	w, err := fs.fs.Create(path)
	if err != nil {
		return nil, err
	}

	return fs.first.WritingTo(w)
}

func (fs *writeFs) Close() error {
	err1 := fs.fs.Close()
	err2 := fs.close.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

type WriteFS interface {
	Create(path string) (io.WriteCloser, error)
	io.Closer
}

type NopWriteCloser struct {
	io.Writer
}

func (NopWriteCloser) Close() error { return nil }

type WriteCloser2 struct {
	io.WriteCloser
	Closer io.Closer
}

func (wc WriteCloser2) Close() error {
	err1 := wc.WriteCloser.Close()
	err2 := wc.Closer.Close()
	if err1 != nil {
		return err1
	}
	return err2
}
