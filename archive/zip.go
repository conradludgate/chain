package archive

import (
	"archive/zip"
	"bytes"
	"io"
	"io/fs"

	"github.com/conradludgate/chain"
)

type ZipConfig struct {
	Comment    string
	Offset     int64
	Compressor zip.Compressor
}

func (cfg ZipConfig) FSWriter(w io.Writer) (chain.WriteFS, error) {
	zipW := zip.NewWriter(w)
	zipW.SetOffset(cfg.Offset)
	if cfg.Compressor != nil {
		zipW.RegisterCompressor(zip.Deflate, cfg.Compressor)
	}

	return zipFSWriter{zipW: zipW, comment: cfg.Comment}, nil
}

type zipFSWriter struct {
	zipW    *zip.Writer
	comment string
}

func (fs zipFSWriter) Create(name string) (io.WriteCloser, error) {
	f, err := fs.zipW.Create(name)
	if err != nil {
		return nil, err
	}
	return nopWriteCloser{f}, nil
}

func (fs zipFSWriter) Close() error {
	err1 := fs.zipW.SetComment(fs.comment)
	err2 := fs.zipW.Close()
	if err2 != nil {
		return err2
	}
	return err1
}

type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error { return nil }

func (cfg ZipConfig) FSReader(r io.Reader) (chain.ReadFS, error) {
	var ra io.ReaderAt
	var size int64
	if rs, ok := r.(readerStat); ok {
		ra = rs
		fi, err := rs.Stat()
		if err != nil {
			return nil, err
		}
		size = fi.Size()
	} else {
		buf := bytes.NewBuffer(nil)
		_, err := io.Copy(buf, r)
		if err != nil {
			return nil, err
		}
		ra = bytes.NewReader(buf.Bytes())
		size = int64(buf.Len())
	}

	zipR, err := zip.NewReader(ra, size)
	if err != nil {
		return nil, err
	}
	return zipFSReader{zipR: zipR}, nil
}

type zipFSReader struct {
	zipR *zip.Reader
}

func (z zipFSReader) Open(name string) (io.ReadCloser, error) {
	return z.zipR.Open(name)
}

func (z zipFSReader) Close() error {
	return nil
}

type readerStat interface {
	io.ReaderAt
	Stat() (fs.FileInfo, error)
}
