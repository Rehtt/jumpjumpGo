package util

import "io"

type RW struct {
	r io.Reader
	w io.Writer
}

func (r *RW) Write(p []byte) (int, error) {
	return r.w.Write(p)
}
func (r *RW) Read(p []byte) (int, error) {
	return r.r.Read(p)
}
func NewRW(r io.Reader, w io.Writer) *RW {
	return &RW{
		r: r,
		w: w,
	}
}
