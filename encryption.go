package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
)

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890=_"

func RandomString(length int) string {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	ret := ""
	for i := 0; i < length; i++ {
		ret += string(chars[random.Intn(len(chars))])
	}
	return ret
}
func HexStringToBytes(hexStr string) []byte {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(err)
	}
	return bytes
}
func EncryptByteArray(data []byte, encKey string, iv string) ([]byte, error) {
	key := []byte(encKey)
	ivBytes := []byte(iv)
	if data == nil || len(data) <= 0 {
		return nil, fmt.Errorf("data is empty")
	}
	if key == nil || len(key) <= 0 {
		return nil, fmt.Errorf("Key is empty")
	}
	if ivBytes == nil || len(ivBytes) <= 0 {
		return nil, fmt.Errorf("IV is empty")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	paddedData := addPadding(data, block.BlockSize())
	cipherData := make([]byte, len(paddedData))

	mode := cipher.NewCBCEncrypter(block, ivBytes)
	mode.CryptBlocks(cipherData, paddedData)
	return cipherData, nil
}
func DecryptByteArray(data []byte, encKey string, iv string) ([]byte, error) {
	key := []byte(encKey)
	ivBytes := []byte(iv)
	if data == nil || len(data) <= 0 {
		return nil, fmt.Errorf("data is empty")
	}
	if key == nil || len(key) <= 0 {
		return nil, fmt.Errorf("Key is empty")
	}
	if ivBytes == nil || len(ivBytes) <= 0 {
		return nil, fmt.Errorf("IV is empty")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(data)%block.BlockSize() != 0 {
		return nil, fmt.Errorf("data size is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, ivBytes)
	decryptedData := make([]byte, len(data))
	mode.CryptBlocks(decryptedData, data)

	return removePadding(decryptedData), nil
}
func removePadding(data []byte) []byte {
	padding := int(data[len(data)-1])
	return data[:len(data)-padding]
}
func addPadding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}
