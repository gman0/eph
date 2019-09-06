package diriter

import (
	"io"
	"os"
)

type Iter struct {
	base    string
	d       *os.File
	current []os.FileInfo
	err     error
}

func NewIter(directory string) (*Iter, error) {
	d, err := os.Open(directory)
	if err != nil {
		return nil, err
	}

	i := &Iter{base: directory, d: d}
	i.Increment()

	return i, nil
}

func (i *Iter) Increment() {
	i.current, i.err = i.d.Readdir(1)
}

func (i *Iter) AtEnd() bool {
	return i.err != nil
}

func (i *Iter) Err() error {
	if i.err == io.EOF {
		return nil
	}
	return i.err
}

func (i *Iter) FileInfo() os.FileInfo {
	return i.current[0]
}

func (i *Iter) Close() error {
	return i.d.Close()
}

func (i *Iter) Base() string {
	return i.base
}
