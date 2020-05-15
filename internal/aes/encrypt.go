package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

// Encrypt encrypts an array of bytes with crypto/aes. Returns: cipher, key.
func Encrypt(data []byte) ([]byte, string, error) {
	key, err := generateKey()
	if err != nil {
		return nil, "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, "", err
	}

	base64encoded := base64.StdEncoding.EncodeToString(data)
	cipherText := make([]byte, aes.BlockSize+len(base64encoded))
	iv := cipherText[:aes.BlockSize]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, "", err
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(cipherText[aes.BlockSize:], []byte(base64encoded))

	return cipherText, base64.StdEncoding.EncodeToString(key), nil
}

// generateKey generates a 32-bit key.
func generateKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}

	return key, nil
}
