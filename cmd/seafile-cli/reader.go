package main

import "io"

type reader struct {
	r     io.Reader
	total int
	ch    chan int
}

func (r reader) Read(b []byte) (int, error) {
	n, err := r.r.Read(b)
	r.ch <- n
	r.total -= n
	if n == 0 {
		close(r.ch)
	}
	return n, err
}
