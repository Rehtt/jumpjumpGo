package util

import (
	"io"
)

func IoCopy(dst io.Writer, src io.Reader) error {
	buf := make([]byte, 1024)

	for {
		n, err := src.Read(buf)
		if err != nil {
			return err
		}
		_, err = dst.Write(buf[:n])
		if err != nil {
			return err
		}
	}
}
func String(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}
