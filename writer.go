package chain

import (
	"fmt"
	"io"
)

type WriteChain struct {
	wcs []WriteChainer
}

// WriteChainer abstract away a pattern
// that's often used in go code where you supply a function
// with an io.Writer `a` and it also returns an io.Writer `b`.
// Then, by writing to `b`, modified data gets written to `a`.
// This is the chain.
// type Writer func(io.Writer) error
type WriteChainer func(io.Writer) (io.Writer, error)

func Chain(first WriteChainer) *WriteChain {
	return &WriteChain{
		wcs: []WriteChainer{first},
	}
}

// Then returns a new Chain which will call the supplied ChainWriter
func (wc *WriteChain) Then(next WriteChainer) *WriteChain {
	wc.wcs = append(wc.wcs, next)
	return wc
}

type ChainWriteCloser struct {
	io.Writer
	close []func() error
}

// WriteToAndClose writes the data from the chain into the provided io.WriteCloser
// as well as closing it
func (wc *WriteChain) WritingTo(w io.Writer) (*ChainWriteCloser, error) {
	var close []func() error = nil

	if wc, ok := w.(io.WriteCloser); ok {
		close = append(close, wc.Close)
	}

	for i := len(wc.wcs) - 1; i >= 0; i-- {
		var err error
		w, err = wc.wcs[i](w)
		if err != nil {
			return nil, err
		}

		if wc, ok := w.(io.WriteCloser); ok {
			close = append(close, wc.Close)
		}
	}

	return &ChainWriteCloser{
		Writer: w,
		close:  close,
	}, nil
}

type ChainWriteCloseError []error

func (cwce ChainWriteCloseError) Error() string {
	return fmt.Sprint([]error(cwce))
}

func (cwc *ChainWriteCloser) Close() error {
	var errors []error

	for i := len(cwc.close) - 1; i >= 0; i-- {
		err := cwc.close[i]()
		if err != nil {
			errors = append(errors, err)
		}
	}

	if errors != nil {
		return ChainWriteCloseError(errors)
	}

	return nil
}
