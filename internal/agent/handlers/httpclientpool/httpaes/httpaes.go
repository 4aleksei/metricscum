// Package httpaes - AES encrypt
package httpaes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"io"
	"os"
)

var (
	ErrNoPublic = errors.New("не удалось декодировать публичный ключ")
	ErrNoRSA    = errors.New("не удалось привести к *rsa.PublicKey")
)

func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

type AesWriter struct {
	w      io.Writer
	nonce  []byte
	aesgcm cipher.AEAD
	key    string
}

func NewWriter(w io.Writer, pub *rsa.PublicKey) (*AesWriter, error) {
	key, err := generateRandom(2 * aes.BlockSize)
	if err != nil {
		return nil, err
	}
	cipherKeyLoaded, err := rsa.EncryptPKCS1v15(rand.Reader, pub, key)

	if err != nil {
		return nil, err
	}
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, err
	}

	return &AesWriter{
		w:      w,
		aesgcm: aesgcm,
		nonce:  key[len(key)-aesgcm.NonceSize():],
		key:    hex.EncodeToString(cipherKeyLoaded),
	}, nil
}

func (a *AesWriter) Write(p []byte) (int, error) {
	if p == nil {
		return 0, nil
	}
	_ = a.aesgcm.Seal(p[0:], a.nonce, p, nil)
	return a.w.Write(p)
}

func (a *AesWriter) GetKey() string {
	return a.key
}

func LoadPublicKey(name string) (*rsa.PublicKey, error) {
	publicKeyPEM, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, ErrNoPublic
	}
	publicKeyLoaded, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaPublicKeyLoaded, ok := publicKeyLoaded.(*rsa.PublicKey)
	if !ok {
		return nil, ErrNoRSA
	}
	return rsaPublicKeyLoaded, nil
}
