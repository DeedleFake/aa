package aa

import (
	"io"
	"path"
	"path/filepath"
)

// FS represents a virtual filesystem, potentially made up of multiple
// directories and archives mounted into it.
//
// The filesystem presented by FS starts with a single archive mounted
// into it at the root. When accessing the contents of the virtual
// filesystem, paths may be given either as absolute paths or relative
// paths; relative paths will be considered relative to the root. All
// paths are slash-separated.
type FS struct {
	paths *pathTree
}

// New returns a new FS with the archive at root as the root of the
// filesystem.
func New(root string) (*FS, error) {
	a, err := openArchive(root)
	if err != nil {
		return nil, err
	}

	return &FS{
		paths: &pathTree{
			a:   a,
			sub: make(map[string]*pathTree),
		},
	}, nil
}

// Close closes the FS, closing all of the mounted archives. Once this
// has been called the FS should not be used anymore.
func (fs *FS) Close() error {
	return fs.paths.Close()
}

// Mount mounts the archive at src to the virtual directory at dst.
// dst does not need to already exist in the filesystem. For example,
// given a zip file, example.zip, containing a file at inner/text.txt,
// after calling
//
//    fs.Mount("mnt/sub", "example.zip")
//
// that file will be accessible at mnt/sub/inner/text.txt in the
// virtual filesystem.
func (fs *FS) Mount(dst, src string) error {
	a, err := openArchive(src)
	if err != nil {
		return err
	}

	fs.paths.Add(cleanPath(dst), a)
	return nil
}

// Open opens the file in the virtual filesystem located at p, relative
// to the root.
func (fs *FS) Open(p string) (io.ReadCloser, error) {
	return fs.paths.Open(cleanPath(p))
}

func cleanPath(p string) string {
	p = path.Clean(filepath.ToSlash(p))
	if path.IsAbs(p) {
		p = p[1:]
	}

	return p
}
