package server

import (
	"bytes"
	"golang.org/x/crypto/ssh"
	"io/fs"
	"os"
	"path/filepath"
)

func parseKey(path string) (keys []ssh.Signer, err error) {
	err = filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if bytes.Contains(data, []byte("PRIVATE KEY")) {
			k, err := ssh.ParsePrivateKey(data)
			if err != nil {
				return nil
			}
			keys = append(keys, k)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return
}
