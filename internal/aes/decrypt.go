package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
)

// Decrypt decrypts an array of bytes with crypto/aes.
func Decrypt(data []byte, b64key string) ([]byte, error) {
	if len(data) < aes.BlockSize {
		return nil, errors.New("data < aes.BlockSize")
	}

	key, err := base64.StdEncoding.DecodeString(b64key)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(data, data)

	decoded, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, err
	}

	return decoded, nil
}
