package chain

import (
	"io"
)

type WriteChain struct {
	wcs []WriteChainer
}

type WriteChainer func(io.Writer) (io.Writer, error)
type WriteChainEnder func() (io.Writer, error)

func W(first WriteChainer) *WriteChain {
	return &WriteChain{
		wcs: []WriteChainer{first},
	}
}

func (wc *WriteChain) Then(next WriteChainer) *WriteChain {
	wc.wcs = append(wc.wcs, next)
	return wc
}

type ChainWriteCloser struct {
	io.Writer
	closeStack
}

func (wc *WriteChain) WritingTo(w io.Writer) (*ChainWriteCloser, error) {
	var close closeStack = nil

	if wc, ok := w.(io.WriteCloser); ok {
		close = append(close, wc.Close)
	}

	for i := len(wc.wcs) - 1; i >= 0; i-- {
		var err error
		w, err = wc.wcs[i](w)
		if err != nil {
			close.Close()
			return nil, err
		}

		if wc, ok := w.(io.WriteCloser); ok {
			close = append(close, wc.Close)
		}
	}

	return &ChainWriteCloser{
		Writer:     w,
		closeStack: close,
	}, nil
}
