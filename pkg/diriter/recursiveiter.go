package diriter

import (
	"os"
	"path"
)

type RecursiveIter struct {
	err       error
	iterStack []*Iter
}

func NewRecursiveIter(directory string) (*RecursiveIter, error) {
	iter, err := NewIter(directory)
	if err != nil {
		return nil, err
	}

	r := &RecursiveIter{}
	r.err = iter.err

	if r.err == nil {
		r.iterStack = append(r.iterStack, iter)
	} else {
		iter.Close()
	}

	return r, nil
}

func (r *RecursiveIter) Close() {
	for _, iter := range r.iterStack {
		iter.Close()
	}
}

func (r *RecursiveIter) top() *Iter {
	return r.iterStack[len(r.iterStack)-1]
}

func (r *RecursiveIter) pop() error {
	if err := r.top().d.Close(); err != nil {
		return err
	}

	r.iterStack = r.iterStack[:len(r.iterStack)-1]
	return nil
}

func (r *RecursiveIter) empty() bool {
	return len(r.iterStack) == 0
}

func (r *RecursiveIter) Increment() {
	if r.empty() {
		return
	}

	i := r.top()

	if i.FileInfo().IsDir() {
		subIter, err := NewIter(path.Join(i.base, i.FileInfo().Name()))
		if err != nil {
			r.err = err
			return
		}

		r.iterStack = append(r.iterStack, subIter)
	} else {
		i.Increment()
		r.err = i.err
	}

	r.resolveEmptyDirectories()
}

func (r *RecursiveIter) OrthogonalIncrement() {
	if !r.empty() {
		r.top().Increment()
		r.err = r.top().err
		r.resolveEmptyDirectories()
	}
}

func (r *RecursiveIter) Err() error {
	return r.err
}

func (r *RecursiveIter) FileInfo() os.FileInfo {
	return r.top().FileInfo()
}

func (r *RecursiveIter) AtEnd() bool {
	return r.empty()
}

func (r *RecursiveIter) Base() string {
	return r.top().base
}

func (r *RecursiveIter) resolveEmptyDirectories() {
	for r.top().AtEnd() {
		r.err = r.pop()
		if r.err != nil {
			return
		}

		if r.empty() {
			break
		}

		r.top().Increment()
		r.err = r.top().err
	}
}
