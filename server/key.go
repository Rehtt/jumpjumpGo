package server

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/mikesmitty/edkey"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/ssh"
	"io/fs"
	"os"
	"path/filepath"
)

func parseKey(path string) (keys []ssh.Signer, errs error) {
	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
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
				errs = errors.Join(errs, err)
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
func genKey(path string) {
	data := genRSA()
	os.WriteFile(filepath.Join(path, "rsa"), data, 644)
	data = genEd25519()
	os.WriteFile(filepath.Join(path, "ed25519"), data, 644)
}
func genRSA() []byte {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("Failed to generate RSA private key: %v", err)
		return nil
	}

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	var tmp bytes.Buffer

	err = pem.Encode(&tmp, privateKeyPEM)
	if err != nil {
		fmt.Printf("Failed to encode private key to PEM format: %v", err)
		return nil
	}
	return tmp.Bytes()
}
func genEd25519() []byte {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Printf("Failed to generate Ed25519 private key: %v", err)
		return nil
	}

	privateKeyPEM := &pem.Block{
		Type:  "OPENSSH PRIVATE KEY",
		Bytes: edkey.MarshalED25519PrivateKey(privateKey),
	}
	var tmp bytes.Buffer

	err = pem.Encode(&tmp, privateKeyPEM)
	if err != nil {
		fmt.Printf("Failed to encode private key to PEM format: %v", err)
		return nil
	}
	return tmp.Bytes()
}
