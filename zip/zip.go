package zip

import (
	"archive/zip"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/DeedleFake/aa"
)

type zipArchive struct {
	r *zip.ReadCloser
}

func openZip(p string, r io.Reader) (aa.Archive, error) {
	zr, err := zip.OpenReader(p)
	if err != nil {
		return nil, err
	}
	return zipArchive{
		r: zr,
	}, nil
}

func (a zipArchive) Open(p string) (io.ReadCloser, error) {
	for _, file := range a.r.File {
		name := path.Clean(filepath.ToSlash(file.Name))
		println(name)
		if name == p {
			return file.Open()
		}
	}

	return nil, os.ErrNotExist
}

func (a zipArchive) Close() error {
	return a.r.Close()
}

func init() {
	aa.RegisterArchive([]byte{0x50, 0x4B, 0x03, 0x04}, openZip)
}
