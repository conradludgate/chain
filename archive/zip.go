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

func (cfg ZipConfig) FSWriter(w io.WriteCloser) (chain.WriteFS, error) {
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

func (zipfs zipFSWriter) Create(name string) (io.WriteCloser, error) {
	f, err := zipfs.zipW.Create(name)
	if err != nil {
		return nil, err
	}
	return chain.NopWriteCloser{Writer: f}, nil
}

func (zipfs zipFSWriter) Close() error {
	err1 := zipfs.zipW.SetComment(zipfs.comment)
	err2 := zipfs.zipW.Close()
	if err2 != nil {
		return err2
	}
	return err1
}

func (cfg ZipConfig) FSReader(r io.ReadCloser) (chain.ReadFS, error) {
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
