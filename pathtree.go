package aa

import (
	"io"
	"strings"
)

type pathTree struct {
	a   Archive
	sub map[string]*pathTree
}

func (pt *pathTree) Close() error {
	if pt.a != nil {
		err := pt.a.Close()
		if err != nil {
			return err
		}
	}

	for p, sub := range pt.sub {
		err := sub.Close()
		if err != nil {
			return err
		}
		delete(pt.sub, p)
	}

	return nil
}

func (pt *pathTree) Add(mp string, a Archive) {
	parts := strings.SplitN(mp, "/", 2)
	if parts[0] == "" {
		if pt.a != nil {
			a = layeredArchive{
				a,
				pt.a,
			}
		}
		pt.a = a
		return
	}

	sub, ok := pt.sub[parts[0]]
	if !ok {
		sub = &pathTree{
			sub: make(map[string]*pathTree),
		}
		pt.sub[parts[0]] = sub
	}

	next := ""
	if len(parts) == 2 {
		next = parts[1]
	}

	sub.Add(next, a)
}

func (pt *pathTree) Open(p string) (io.ReadCloser, error) {
	parts := strings.SplitN(p, "/", 2)
	if len(parts) == 1 {
		return pt.a.Open(parts[0])
	}

	if sub, ok := pt.sub[parts[0]]; ok {
		return sub.Open(parts[1])
	}

	return pt.a.Open(p)
}
