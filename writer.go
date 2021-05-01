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
