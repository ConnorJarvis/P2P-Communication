package main

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/gob"
	"errors"
)

type Message struct {
	From          string
	Body          string
	FromSignature []byte
	BodySignature []byte
}

func (m *Message) EncodeMessage() ([]byte, error) {
	//Encode message to bytes
	bytes := bytes.Buffer{}
	err := gob.NewEncoder(&bytes).Encode(m)
	if err != nil {
		return nil, err
	}
	return bytes.Bytes(), nil
}

func (m *Message) SignMessage(RSA RSAUtil) error {
	//Calculate hash of body
	bodyHasher := sha256.New()
	bodyHasher.Write([]byte(m.Body))
	bodyHash := bodyHasher.Sum(nil)
	//Calculate hash of from
	fromHasher := sha256.New()
	fromHasher.Write([]byte(m.From))
	fromHash := fromHasher.Sum(nil)
	//Calculate signature of body
	bodySignature, err := rsa.SignPKCS1v15(RSA.Reader, &RSA.Key, crypto.SHA256, bodyHash)
	if err != nil {
		return err
	}
	//Calculate signature of from
	fromSignature, err := rsa.SignPKCS1v15(RSA.Reader, &RSA.Key, crypto.SHA256, fromHash)
	if err != nil {
		return err
	}

	m.BodySignature = bodySignature
	m.FromSignature = fromSignature
	return nil
}

func (m *Message) VerifyMessage(RSA RSAUtil) error {
	//Calculate hash of body
	bodyHasher := sha256.New()
	bodyHasher.Write([]byte(m.Body))
	bodyHash := bodyHasher.Sum(nil)
	//Calculate hash of from
	fromHasher := sha256.New()
	fromHasher.Write([]byte(m.From))
	fromHash := fromHasher.Sum(nil)
	//Verify body signature
	err := rsa.VerifyPKCS1v15(&RSA.Key.PublicKey, crypto.SHA256, bodyHash, m.BodySignature)
	if err != nil {
		return errors.New("Invalid Body Signature")
	}
	//Verify from signature
	err = rsa.VerifyPKCS1v15(&RSA.Key.PublicKey, crypto.SHA256, fromHash, m.FromSignature)
	if err != nil {
		return errors.New("Invalid From Signature")
	}
	return nil
}
