package main

import (
	"testing"
)

func TestEncodeMessage(t *testing.T) {
	m := Message{From: "Test", Body: "Hello"}
	_, err := m.EncodeMessage()
	if err != nil {
		t.Error(err)
	}
}
func TestSignMessage(t *testing.T) {
	RSA := RSAUtil{}
	RSA.InitializeReader()
	RSA.SetKeyLength(2048)
	RSA.GenerateKey()
	m := Message{From: "Test", Body: "Hello"}
	err := m.SignMessage(RSA)
	if err != nil {
		t.Error(err)
	}
}

func TestVerifyMessage(t *testing.T) {
	RSA := RSAUtil{}
	RSA.InitializeReader()
	RSA.SetKeyLength(2048)
	RSA.GenerateKey()
	m := Message{From: "Test", Body: "Hello"}
	m.SignMessage(RSA)
	err := m.VerifyMessage(RSA)
	if err != nil {
		t.Error(err)
	}
}
