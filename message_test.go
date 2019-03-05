package main

import (
	"testing"
)

func TestEncode(t *testing.T) {
	m := Message{Header: Header{ID: 0, From: "1"}, Body: Body{Content: "Hello"}}
	_, err := m.Encode()
	if err != nil {
		t.Error(err)
	}
	_, err = m.Body.Encode()
	if err != nil {
		t.Error(err)
	}
	_, err = m.Header.Encode()
	if err != nil {
		t.Error(err)
	}
}
func TestSignMessage(t *testing.T) {
	RSA := RSAUtil{}
	RSA.InitializeReader()
	RSA.SetKeyLength(2048)
	RSA.GenerateKey()
	m := Message{Header: Header{ID: 0, From: "1"}, Body: Body{Content: "Hello"}}
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
	m := Message{Header: Header{ID: 0, From: "1"}, Body: Body{Content: "Hello"}}
	m.SignMessage(RSA)
	err := m.VerifyMessage(RSA)
	if err != nil {
		t.Error(err)
	}
}
