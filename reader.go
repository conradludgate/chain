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

type ReadFSChain func(io.Reader) (ReadFS, error)
type ReaderFSBuilder struct {
	first *ReaderBuilder
	fs    ReadFSChain
	after []ReadChain
}

func ReadingFromFS(fs ReadFS) *ReaderFSBuilder {
	return &ReaderFSBuilder{
		first: &ReaderBuilder{},
		fs:    func(io.Reader) (ReadFS, error) { return fs, nil },
		after: nil,
	}
}

func (chain *ReaderBuilder) AsFS(next ReadFSChain) *ReaderFSBuilder {
	return &ReaderFSBuilder{
		first: chain,
		fs:    next,
		after: nil,
	}
}

func (chain *ReaderFSBuilder) Open(name string) *ReaderBuilder {
	fs, err := chain.build()
	if err != nil {
		return &ReaderBuilder{err: err}
	}
	r, err := fs.Open(name)
	if err != nil {
		fs.Close()
		return &ReaderBuilder{err: err}
	}
	return &ReaderBuilder{r: reader{
		Reader:     r,
		closeStack: closeStack{fs, r},
	}}
}

func (chain *ReaderFSBuilder) Then(next ReadChain) *ReaderFSBuilder {
	chain.after = append(chain.after, next)
	return chain
}

func (chain *ReaderFSBuilder) Finally(next ReadChain) (ReadFS, error) {
	return chain.Then(next).build()
}

func (chain *ReaderFSBuilder) build() (ReadFS, error) {
	if chain.first.err != nil {
		return nil, chain.first.err
	}

	fs, err := chain.fs(chain.first.r)
	if err != nil {
		return nil, err
	}

	return &readFS{
		Closer: chain.first.r,
		fs:     fs,
		after:  chain.after,
	}, nil
}

type readFS struct {
	io.Closer
	fs    ReadFS
	after []ReadChain
}

func (fs *readFS) Open(path string) (io.ReadCloser, error) {
	r, err := fs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	builder := ReadingFrom(r)
	for _, then := range fs.after {
		builder.Then(then)
	}
	if builder.err != nil {
		return nil, builder.err
	}

	return builder.r, nil
}

type ReadFS interface {
	Open(path string) (io.ReadCloser, error)
	io.Closer
}
