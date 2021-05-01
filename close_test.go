package chain

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ErrorString1 string
type ErrorString2 string
type ErrorString3 string

func (s ErrorString1) Error() string { return string(s) }
func (s ErrorString2) Error() string { return string(s) }
func (s ErrorString3) Error() string { return string(s) }

var (
	Err1 = ErrorString1("Error 1")
	Err2 = ErrorString2("Error 2")
	Err3 = errors.New("Error 3")
)

type Closer0 struct{}
type Closer1 struct{}
type Closer2 struct{}
type Closer3 struct{}

func (Closer0) Close() error { return nil }
func (Closer1) Close() error { return Err1 }
func (Closer2) Close() error { return Err2 }
func (Closer3) Close() error { return Err3 }

func TestReaderClose(t *testing.T) {
	var chainClosed bool
	var readerClosed bool

	r, err := ReadingFrom(readCloser{
		ReadCloser: io.NopCloser(strings.NewReader("hello world")),
		closed:     &readerClosed,
	}).
		Finally(func(r io.ReadCloser) (io.ReadCloser, error) {
			return readCloser{ReadCloser: r, closed: &chainClosed}, nil
		})
	require.Nil(t, err)
	err = r.Close()
	require.Nil(t, err)
	assert.True(t, chainClosed)
	assert.True(t, readerClosed)
}

func TestReaderClose_ErrorClosing(t *testing.T) {
	var chainClosed bool

	r, err := ReadingFrom(readCloser{
		ReadCloser: io.NopCloser(strings.NewReader("hello world")),
	}).
		Finally(func(r io.ReadCloser) (io.ReadCloser, error) {
			return readCloser{ReadCloser: r, closed: &chainClosed}, nil
		})
	require.Nil(t, err)
	err = r.Close()
	assert.EqualError(t, err, "Error 2")
	assert.True(t, chainClosed)
}

func TestReaderClose_ErrorChain(t *testing.T) {
	var readerClosed bool

	_, err := ReadingFrom(readCloser{
		ReadCloser: io.NopCloser(strings.NewReader("hello world")),
		closed:     &readerClosed,
	}).
		Finally(func(r io.ReadCloser) (io.ReadCloser, error) {
			return nil, Err1
		})
	assert.EqualError(t, err, "Error 1")
	assert.True(t, readerClosed)
}

func TestWriterClose(t *testing.T) {
	var chainClosed bool
	var writerClosed bool

	w, err := NewWriteBuilder(func(w io.WriteCloser) (io.WriteCloser, error) {
		return writeCloser{WriteCloser: w, closed: &chainClosed}, nil
	}).WritingTo(writeCloser{
		WriteCloser: NopWriteCloser{Writer: bytes.NewBuffer(nil)},
		closed:      &writerClosed,
	})
	require.Nil(t, err)
	err = w.Close()
	require.Nil(t, err)
	assert.True(t, chainClosed)
	assert.True(t, writerClosed)
}

func TestWriterClose_ErrorChain(t *testing.T) {
	var writerClosed bool

	_, err := NewWriteBuilder(func(w io.WriteCloser) (io.WriteCloser, error) {
		return nil, Err1
	}).WritingTo(writeCloser{
		WriteCloser: NopWriteCloser{Writer: bytes.NewBuffer(nil)},
		closed:      &writerClosed,
	})
	assert.EqualError(t, err, "Error 1")
	assert.True(t, writerClosed)
}

func TestWriterClose_ErrorClosing(t *testing.T) {
	var chainClosed bool

	w, err := NewWriteBuilder(func(w io.WriteCloser) (io.WriteCloser, error) {
		return writeCloser{WriteCloser: w, closed: &chainClosed}, nil
	}).WritingTo(writeCloser{
		WriteCloser: NopWriteCloser{Writer: bytes.NewBuffer(nil)},
	})
	require.Nil(t, err)
	err = w.Close()
	assert.EqualError(t, err, "Error 2")
	assert.True(t, chainClosed)
}

type readCloser struct {
	io.ReadCloser
	closed *bool
}

func (n readCloser) Close() error {
	err1 := n.ReadCloser.Close()
	if n.closed != nil {
		*n.closed = true
	} else if err1 == nil {
		return Err2
	}
	return err1
}

type writeCloser struct {
	io.WriteCloser
	closed *bool
}

func (n writeCloser) Close() error {
	err1 := n.WriteCloser.Close()
	if n.closed != nil {
		*n.closed = true
	} else if err1 == nil {
		return Err2
	}
	return err1
}
