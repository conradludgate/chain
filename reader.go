package chain

import (
	"io"
)

// ReaderBuilder lets you build a chain of io.Readers
// in a more natural way
type ReaderBuilder struct {
	r   reader
	err error
}

// ReadChain represents a common pattern in go packages.
// For examples, see:
// https://pkg.go.dev/compress/gzip#NewReader
//
// While this type only expects an io.Reader to be returned, it will
// detect if the result is a Closer and make sure the
// types Close function is run.
type ReadChain func(io.Reader) (io.Reader, error)

// ReadingFrom creates a new ReaderBuilder where the
// contents of the chain comes from the given io.Reader
//
// If r is an io.ReadCloser, the resulting Reader
// after building the chain will call r.Close for you
func ReadingFrom(r io.Reader) *ReaderBuilder {
	var close closeStack
	if rc, ok := r.(io.ReadCloser); ok {
		close = append(close, rc)
	}

	return &ReaderBuilder{r: reader{Reader: r, closeStack: close}}
}

// Then adds the next ReadChain to the current builder chain
func (chain *ReaderBuilder) Then(next ReadChain) *ReaderBuilder {
	if chain.err == nil {
		chain.r.Reader, chain.err = next(chain.r.Reader)
		if rc, ok := chain.r.Reader.(io.ReadCloser); ok {
			chain.r.closeStack = append(chain.r.closeStack, rc)
		}
	}
	if chain.err != nil {
		chain.r.closeStack.Close()
	}
	return chain
}

type reader struct {
	io.Reader
	closeStack
}

// Finally adds the last ReadChain to the current builder chain,
// then builds it into an io.ReadCloser
func (chain *ReaderBuilder) Finally(next ReadChain) (io.ReadCloser, error) {
	chain.Then(next)
	return chain.r, chain.err
}
