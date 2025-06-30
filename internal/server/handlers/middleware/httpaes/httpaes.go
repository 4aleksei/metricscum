// Package httpgzip - middleware for compress/decompress http request/response
package httpaes

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"io"
	"os"

	"github.com/4aleksei/metricscum/internal/common/middleware/aescoder"
)

type aesReader struct {
	r  io.ReadCloser
	zr *aescoder.AesReader
}

func NewAesReader(r io.ReadCloser, privateKeyLoaded *rsa.PrivateKey, aesSkey string) (*aesReader, error) {
	key, err := hex.DecodeString(aesSkey)
	decryptedKey, err := rsa.DecryptPKCS1v15(rand.Reader, privateKeyLoaded, key)

	zr, err := aescoder.NewReader(r, decryptedKey)

	if err != nil {
		return nil, err
	}

	return &aesReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c *aesReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *aesReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func LoadKey(name string) (*rsa.PrivateKey, error) {
	privateKeyPEM, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, err
	}
	privateKeyLoaded, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return privateKeyLoaded, nil
}
