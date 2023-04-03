package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"golang.org/x/crypto/bcrypt"
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

func RandBytes(c int) []byte {
	out := make([]byte, c)
	rand.Read(out)
	return out
}

func Bcrypt(str string) string {
	out, err := bcrypt.GenerateFromPassword([]byte(str), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(out)
}
func CheckBcrypt(hashStr, str string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashStr), []byte(str)) == nil
}

func SHA256(data []byte) []byte {
	s := sha256.New()
	s.Write(data)
	return s.Sum(nil)
}
func AESCbcEncrypt(str, password string) []byte {
	s := SHA256([]byte(password))
	iv := SHA256(s[16:])[16:]
	key := SHA256(iv)[16:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil
	}
	ciphertext := make([]byte, len(str))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, []byte(str))
	return ciphertext
}

func AESCbcDecrypt(ciphertext []byte, password string) []byte {
	s := SHA256([]byte(password))
	iv := SHA256(s[16:])[16:]
	key := SHA256(iv)[16:]
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil
	}

	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, ciphertext)

	return plaintext
}
