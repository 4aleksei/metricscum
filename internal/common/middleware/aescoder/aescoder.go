// Package aescoder
package aescoder

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
)

type AesReader struct {
	r      io.ReadCloser
	nonce  []byte
	aesgcm cipher.AEAD
}

func NewReader(r io.ReadCloser, key []byte) (*AesReader, error) {
	aesblock, err := aes.NewCipher(key)

	if err != nil {
		fmt.Printf("error: %v\n", err)
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return nil, err
	}

	return &AesReader{
		r:      r,
		nonce:  key[len(key)-aesgcm.NonceSize():],
		aesgcm: aesgcm,
	}, nil
}

func (h *AesReader) Read(p []byte) (int, error) {
	n, err := h.r.Read(p)
	if n > 0 {
		_, err := h.aesgcm.Open(p[0:], h.nonce, p, nil)
		return len(p), err
	}
	return n, err
}

func (h *AesReader) Close() error {
	return nil
}
