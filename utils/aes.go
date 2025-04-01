package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"fmt"
)

var encryptionPassword = fmt.Sprintf("%x", md5.Sum([]byte("next-terminal")))

func MustDecrypt(encrypted string) string {
	if encrypted == "" || encrypted == "-" {
		return ""
	}

	decrypt, _ := Decrypt(encrypted)
	return decrypt
}

func Decrypt(encrypted string) (string, error) {
	origData, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}
	decrypt, err := AesDecryptCBC(origData, []byte(encryptionPassword))
	if err != nil {
		return "", err
	}
	return string(decrypt), nil
}

func AesDecryptCBC(encrypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(encrypted))
	blockMode.CryptBlocks(origData, encrypted)
	origData = PKCS5UnPadding(origData)
	return origData, nil
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unPadding := int(origData[length-1])
	return origData[:(length - unPadding)]
}
