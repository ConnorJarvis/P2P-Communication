package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"io"
)

type RSAUtil struct {
	Reader io.Reader
	Key    rsa.PrivateKey
	Length int
}

func (r *RSAUtil) InitializeReader() error {
	r.Reader = rand.Reader
	return nil
}

func (r *RSAUtil) SetKeyLength(length int) error {
	r.Length = length
	return nil
}

func (r *RSAUtil) GenerateKey() error {
	privateKey, err := rsa.GenerateKey(r.Reader, r.Length)
	if err != nil {
		return err
	}
	r.Key = *privateKey
	return nil
}

func (r *RSAUtil) Encrypt(data []byte) ([]byte, error) {

	ciphertext, err := rsa.EncryptOAEP(sha256.New(), r.Reader, &r.Key.PublicKey, data, nil)
	if err != nil {
		return nil, err
	}
	return ciphertext, nil
}

func (r *RSAUtil) Decrypt(data []byte) ([]byte, error) {

	plaintext, err := rsa.DecryptOAEP(sha256.New(), r.Reader, &r.Key, data, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
