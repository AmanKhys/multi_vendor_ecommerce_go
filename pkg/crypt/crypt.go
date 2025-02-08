package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"os"

	env "github.com/joho/godotenv"
)

var envM, err = env.Read(".crypt_env")

// Read secret key from environment variable
var secretKeyString = envM["CRYPT_SECRET_KEY"]

// Read IV (Initialization Vector) from environment variable (must be 16 bytes)
var bytesString = os.Getenv("AES_IV")

// Convert to byte slices
var secretKey = []byte(secretKeyString)
var bytes = []byte(bytesString)

func encode(b []byte) string {
	return hex.EncodeToString(b)
}

func decode(s string) ([]byte, error) {
	data, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func Encrypt(text string) (string, error) {
	if len(secretKey) != 16 {
		return "", errors.New("invalid key length: must be 16 bytes")
	}
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}
	plainText := []byte(text)
	cfb := cipher.NewCFBEncrypter(block, bytes)
	cipherText := make([]byte, len(plainText))
	cfb.XORKeyStream(cipherText, plainText)
	return encode(cipherText), nil
}

func Decrypt(text string) (string, error) {
	if len(secretKey) != 16 {
		return "", errors.New("invalid key length: must be 16 bytes")
	}
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}
	cipherText, err := decode(text)
	if err != nil {
		return "", err
	}
	cfb := cipher.NewCFBDecrypter(block, bytes)
	plainText := make([]byte, len(cipherText))
	cfb.XORKeyStream(plainText, cipherText)
	return string(plainText), nil
}
