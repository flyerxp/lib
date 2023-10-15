package safe

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

type AesL struct {
	key []byte //16位
	iv  []byte
}

func CreateAesL(key []byte, iv []byte) *AesL {
	t := new(AesL)
	t.key = key
	t.iv = iv
	return t
}
func (a *AesL) Encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return "", err
	}
	padLen := 16 - len(plaintext)%16
	if padLen != 16 {
		for i := 0; i < padLen; i++ {
			plaintext = append(plaintext, 0x00)
		}
	}
	ciphertext := make([]byte, len(plaintext))
	stream := cipher.NewCBCEncrypter(block, a.iv)
	stream.CryptBlocks(ciphertext, plaintext)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (a *AesL) Decrypt(ciphertext string) (string, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return "", err
	}
	decodedCiphertext, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	stream := cipher.NewCBCDecrypter(block, a.iv)
	stream.CryptBlocks(decodedCiphertext, decodedCiphertext)
	return string(decodedCiphertext), nil
}
