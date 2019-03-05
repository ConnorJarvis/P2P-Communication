package main

import (
	"crypto/rand"
	"crypto/rsa"
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
