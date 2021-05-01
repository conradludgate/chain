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

func TestCloseStack(t *testing.T) {
	stack := closeStack{Closer0{}, Closer1{}, Closer2{}}
	err := stack.Close()
	assert.EqualError(t, err, "[Error 2 Error 1]")

	assert.ErrorIs(t, err, Err1)
	assert.ErrorIs(t, err, Err2)
	assert.NotErrorIs(t, err, Err3)

	var err1 ErrorString1
	assert.ErrorAs(t, err, &err1)
	assert.Equal(t, "Error 1", string(err1))

	var err2 ErrorString2
	assert.ErrorAs(t, err, &err2)
	assert.Equal(t, "Error 2", string(err2))

	var err3 ErrorString3
	assert.False(t, errors.As(err, &err3))
	assert.Equal(t, "", string(err3))
}

func TestReaderClose(t *testing.T) {
	var chainClosed bool
	var readerClosed bool

	r, err := ReadingFrom(readCloser{
		Reader: strings.NewReader("hello world"),
		closed: &readerClosed,
	}).
		Finally(func(r io.Reader) (io.Reader, error) {
			return readCloser{Reader: r, closed: &chainClosed}, nil
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
		Reader: strings.NewReader("hello world"),
	}).
		Finally(func(r io.Reader) (io.Reader, error) {
			return readCloser{Reader: r, closed: &chainClosed}, nil
		})
	require.Nil(t, err)
	err = r.Close()
	assert.EqualError(t, err, "[Error 2]")
	assert.True(t, chainClosed)
}

func TestReaderClose_ErrorChain(t *testing.T) {
	var readerClosed bool

	_, err := ReadingFrom(readCloser{
		Reader: strings.NewReader("hello world"),
		closed: &readerClosed,
	}).
		Finally(func(r io.Reader) (io.Reader, error) {
			return nil, Err1
		})
	assert.EqualError(t, err, "Error 1")
	assert.True(t, readerClosed)
}

func TestWriterClose(t *testing.T) {
	var chainClosed bool
	var writerClosed bool

	w, err := NewWriteBuilder(func(w io.Writer) (io.Writer, error) {
		return writeCloser{Writer: w, closed: &chainClosed}, nil
	}).WritingTo(writeCloser{
		Writer: bytes.NewBuffer(nil),
		closed: &writerClosed,
	})
	require.Nil(t, err)
	err = w.Close()
	require.Nil(t, err)
	assert.True(t, chainClosed)
	assert.True(t, writerClosed)
}

func TestWriterClose_ErrorChain(t *testing.T) {
	var writerClosed bool

	_, err := NewWriteBuilder(func(w io.Writer) (io.Writer, error) {
		return nil, Err1
	}).WritingTo(writeCloser{
		Writer: bytes.NewBuffer(nil),
		closed: &writerClosed,
	})
	assert.EqualError(t, err, "Error 1")
	assert.True(t, writerClosed)
}

func TestWriterClose_ErrorClosing(t *testing.T) {
	var chainClosed bool

	w, err := NewWriteBuilder(func(w io.Writer) (io.Writer, error) {
		return writeCloser{Writer: w, closed: &chainClosed}, nil
	}).WritingTo(writeCloser{
		Writer: bytes.NewBuffer(nil),
	})
	require.Nil(t, err)
	err = w.Close()
	assert.EqualError(t, err, "[Error 2]")
	assert.True(t, chainClosed)
}

type readCloser struct {
	io.Reader
	closed *bool
}

func (n readCloser) Close() error {
	if n.closed == nil {
		return Err2
	}
	*n.closed = true
	return nil
}

type writeCloser struct {
	io.Writer
	closed *bool
}

func (n writeCloser) Close() error {
	if n.closed == nil {
		return Err2
	}
	*n.closed = true
	return nil
}
