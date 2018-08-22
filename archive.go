package aa

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
)

// Archive is the common interface shared by all archive types, such
// as zip files and directories.
type Archive interface {
	// Open opens the file at p inside the archive for reading.
	Open(p string) (io.ReadCloser, error)

	io.Closer
}

// layeredArchive is a used when multiple archives are mounted to the
// same point. It calls methods on each of its members in turn.
type layeredArchive []Archive

func (a layeredArchive) Open(p string) (io.ReadCloser, error) {
	for _, sub := range a {
		r, err := sub.Open(p)
		if (err == nil) || !os.IsNotExist(err) {
			return r, err
		}
	}

	return nil, os.ErrNotExist
}

func (a layeredArchive) Close() error {
	for _, sub := range a {
		err := sub.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

type knownArchive struct {
	magic []byte
	open  func(string, io.Reader) (Archive, error)
}

var knownArchives = []knownArchive{
	{
		open: openDir,
	},
}

// RegisterArchive registers an archive type. This is intended to
// primarily be called from package init() functions, much like how
// the image package handles archive types.
//
// The open function argument is passed both the path to the archive
// on disk and the already opened file. It may use either one, or
// both, of these in order to open the archive. If open returns an
// error, that error will be returned from the attempted mount that
// called it.
func RegisterArchive(magic []byte, open func(string, io.Reader) (Archive, error)) {
	knownArchives = append(knownArchives, knownArchive{
		magic: magic,
		open:  open,
	})
}

func openArchive(p string) (Archive, error) {
	file, err := os.Open(p) // nolint
	if err != nil {
		return nil, err
	}
	defer file.Close() // nolint
	r := bufio.NewReader(file)

	for _, a := range knownArchives {
		if len(a.magic) > 0 {
			buf, err := r.Peek(len(a.magic))
			if (err != nil) && (err != io.ErrUnexpectedEOF) {
				return nil, err
			}

			if !bytes.Equal(buf, a.magic) {
				continue
			}
		}

		o, err := a.open(p, r)
		if (o == nil) && (err == nil) {
			continue
		}
		return o, err
	}

	return nil, errors.New("unknown archive type")
}

type dirArchive string

func openDir(p string, r io.Reader) (Archive, error) {
	p = filepath.FromSlash(p)

	fi, err := os.Stat(p)
	if err != nil {
		return nil, err
	}

	if !fi.IsDir() {
		return nil, nil
	}

	p, err = filepath.Abs(p)
	if err != nil {
		return nil, err
	}

	return dirArchive(p), nil
}

func (a dirArchive) Open(p string) (io.ReadCloser, error) {
	// TODO: Should this be changed to always return a nil io.ReadCloser
	// if a nil *os.File is returned?
	return os.Open(filepath.Join(string(a), filepath.FromSlash(p)))
}

func (a dirArchive) Close() error {
	return nil
}
