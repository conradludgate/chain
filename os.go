package chain

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type OS struct {
	RootDir string
}

func (o OS) Open(name string) (io.ReadCloser, error) {
	path := filepath.Join(o.RootDir, name)
	return os.Open(path)
}

func (o OS) Close() error { return nil }

func (o OS) Create(name string) (io.WriteCloser, error) {
	path := filepath.Join(o.RootDir, name)
	fmt.Println("open", path)
	return os.Create(path)
}
