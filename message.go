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
	Header          Header
	Body            Body
	HeaderSignature []byte
	BodySignature   []byte
}

type EncryptedMessage struct {
	Message []byte
	Key     []byte
	Nonce   []byte
}

type Header struct {
	ID   int
	From string
}

type Body struct {
	Content interface{}
}

type Gossip struct {
	Peers  []Peer
	Values map[string]*Value
}

func (h *Header) Encode() ([]byte, error) {
	//Encode header to bytes
	bytes := bytes.Buffer{}
	encoder := gob.NewEncoder(&bytes)
	err := encoder.Encode(h)
	if err != nil {
		return nil, err
	}
	return bytes.Bytes(), nil
}

func (b *Body) Encode() ([]byte, error) {
	//Encode body to bytes
	bytes := bytes.Buffer{}

	encoder := gob.NewEncoder(&bytes)
	err := encoder.Encode(b)
	if err != nil {
		return nil, err
	}
	return bytes.Bytes(), nil
}
func (m *Message) Encode() ([]byte, error) {
	//Encode message to bytes
	bytes := bytes.Buffer{}
	encoder := gob.NewEncoder(&bytes)
	err := encoder.Encode(m)
	if err != nil {
		return nil, err
	}
	return bytes.Bytes(), nil
}

func (em *EncryptedMessage) Encode() ([]byte, error) {
	//Encode message to bytes
	bytes := bytes.Buffer{}
	encoder := gob.NewEncoder(&bytes)
	err := encoder.Encode(em)
	if err != nil {
		return nil, err
	}
	return bytes.Bytes(), nil
}

func (m *Message) SignMessage(RSA RSAUtil) error {

	//Calculate hash of body
	bodyHasher := sha256.New()
	//Encode Body struct to bytes
	bodyBytes, err := m.Body.Encode()
	if err != nil {
		return err
	}
	bodyHasher.Write(bodyBytes)
	bodyHash := bodyHasher.Sum(nil)

	//Calculate hash of from
	headerHasher := sha256.New()
	//Encode Header struct to bytes
	headerBytes, err := m.Header.Encode()
	if err != nil {
		return err
	}
	headerHasher.Write(headerBytes)
	headerHash := headerHasher.Sum(nil)
	//Calculate signature of body
	bodySignature, err := rsa.SignPKCS1v15(RSA.Reader, &RSA.Key, crypto.SHA256, bodyHash)
	if err != nil {
		return err
	}
	//Calculate signature of header
	headerSignature, err := rsa.SignPKCS1v15(RSA.Reader, &RSA.Key, crypto.SHA256, headerHash)
	if err != nil {
		return err
	}

	m.BodySignature = bodySignature
	m.HeaderSignature = headerSignature
	return nil
}

func (m *Message) VerifyMessage(RSA RSAUtil) error {
	//Calculate hash of body
	bodyHasher := sha256.New()
	//Encode Body struct to bytes
	bodyBytes, err := m.Body.Encode()
	if err != nil {
		return err
	}
	bodyHasher.Write(bodyBytes)
	bodyHash := bodyHasher.Sum(nil)

	//Calculate hash of header
	headerHasher := sha256.New()
	//Encode Header struct to bytes
	headerBytes, err := m.Header.Encode()
	if err != nil {
		return err
	}
	headerHasher.Write(headerBytes)
	headerHash := headerHasher.Sum(nil)
	//Verify body signature
	err = rsa.VerifyPKCS1v15(&RSA.Key.PublicKey, crypto.SHA256, bodyHash, m.BodySignature)
	if err != nil {
		return errors.New("Invalid Body Signature")
	}
	//Verify header signature
	err = rsa.VerifyPKCS1v15(&RSA.Key.PublicKey, crypto.SHA256, headerHash, m.HeaderSignature)
	if err != nil {
		return errors.New("Invalid From Signature")
	}
	return nil
}

func (m *EncryptedMessage) Decrypt(RSA RSAUtil) (*Message, error) {
	key, err := RSA.Decrypt(m.Key)
	if err != nil {
		return nil, err
	}
	nonce, err := RSA.Decrypt(m.Nonce)
	if err != nil {
		return nil, err
	}
	plaintext, err := Decrypt(m.Message, key, nonce)
	if err != nil {
		return nil, err
	}

	encryptedMessageBytes := bytes.Buffer{}
	encryptedMessageBytes.Write(plaintext)
	decoder := gob.NewDecoder(&encryptedMessageBytes)

	encryptedMessage := Message{}
	err = decoder.Decode(&encryptedMessage)
	if err != nil {
		return nil, err
	}

	return &encryptedMessage, nil

}

func (m *Message) Encrypt(RSA RSAUtil) (*EncryptedMessage, error) {
	key, err := GenerateAESKey(RSA.Reader)
	if err != nil {
		return nil, err
	}
	nonce, err := GenerateAESNonce(RSA.Reader)
	if err != nil {
		return nil, err
	}

	encryptedKey, err := RSA.Encrypt(key)
	if err != nil {
		return nil, err
	}

	encryptedNonce, err := RSA.Encrypt(nonce)
	if err != nil {
		return nil, err
	}

	messageBytes := bytes.Buffer{}
	encoder := gob.NewEncoder(&messageBytes)
	err = encoder.Encode(m)
	if err != nil {
		return nil, err
	}

	encryptedMessage, err := Encrypt(messageBytes.Bytes(), key, nonce)
	if err != nil {
		return nil, err
	}
	return &EncryptedMessage{Message: encryptedMessage, Key: encryptedKey, Nonce: encryptedNonce}, nil

}
