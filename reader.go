package chain

import (
	"io"
)

type ReadChain struct {
	io.Reader
	closeStack
	err error
}

type ReadChainer func(io.Reader) (io.Reader, error)
type ReadChainEnder func() (io.Reader, error)

func ReadingFrom(r io.Reader) *ReadChain {
	return &ReadChain{Reader: r, closeStack: nil}
}

func (chain *ReadChain) Then(next ReadChainer) *ReadChain {
	chain.Reader, chain.err = next(chain.Reader)
	if rc, ok := chain.Reader.(io.ReadCloser); ok {
		chain.closeStack = append(chain.closeStack, rc.Close)
	}
	return chain
}

func (chain *ReadChain) Finally(next ReadChainer) (*ReadChain, error) {
	chain.Reader, chain.err = next(chain.Reader)
	if rc, ok := chain.Reader.(io.ReadCloser); ok {
		chain.closeStack = append(chain.closeStack, rc.Close)
	}
	return chain, chain.err
}
