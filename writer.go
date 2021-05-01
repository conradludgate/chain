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
type WriteChain func(io.Writer) (io.Writer, error)

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

type writer struct {
	io.Writer
	closeStack
}

// WritingTo builds the chain. The resulting data from the chain is
// written to the io.Writer provided.
//
// If w is a Closer type too, calling the returned writer's Close function
// will also close w.
func (wc *WriterBuilder) WritingTo(w io.Writer) (io.WriteCloser, error) {
	var close closeStack = nil

	if wc, ok := w.(io.WriteCloser); ok {
		close = append(close, wc)
	}

	for i := len(wc.wcs) - 1; i >= 0; i-- {
		var err error
		w, err = wc.wcs[i](w)
		if err != nil {
			close.Close()
			return nil, err
		}

		if wc, ok := w.(io.WriteCloser); ok {
			close = append(close, wc)
		}
	}

	return &writer{
		Writer:     w,
		closeStack: close,
	}, nil
}

type WriteFSBuilder struct {
	first *WriterBuilder
	fs    WriteFSChain
	after *WriterBuilder
}

type WriteFSChain func(io.Writer) (WriteFS, error)

func (wc *WriterBuilder) IntoFS(next WriteFSChain) *WriteFSBuilder {
	return &WriteFSBuilder{
		first: wc,
		fs:    next,
		after: &WriterBuilder{},
	}
}

func (wc *WriterBuilder) WritingToFS(fs WriteFS) WriteFS {
	return &writeFs{
		first: wc,
		fs:    fs,
	}
}

func (wc *WriteFSBuilder) Then(next WriteChain) *WriteFSBuilder {
	wc.after.Then(next)
	return wc
}

func (wc *WriteFSBuilder) WritingTo(w io.Writer) (WriteFS, error) {
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
