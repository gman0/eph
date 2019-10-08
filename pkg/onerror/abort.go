package onerror

type Abort struct {
	err error
}

func (o *Abort) Try(f func() error) *Abort {
	if o.err == nil {
		o.err = f()
	}

	return o
}

func (o *Abort) Err() error {
	return o.err
}
