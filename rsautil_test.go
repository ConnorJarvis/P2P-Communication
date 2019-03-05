package main

import (
	"errors"
	"testing"
)

func TestGenerateKey(t *testing.T) {
	RSA := RSAUtil{}
	RSA.InitializeReader()
	RSA.SetKeyLength(2048)
	err := RSA.GenerateKey()
	if err != nil {
		t.Error(err)
	}

}

func TestSetKeyLength(t *testing.T) {
	RSA := RSAUtil{}
	length := 2048
	err := RSA.SetKeyLength(length)
	if err != nil {
		t.Error(err)
	}
	if RSA.Length != length {
		t.Error(errors.New("Length not the same"))
	}
}

func TestInitalizeReader(t *testing.T) {
	RSA := RSAUtil{}
	err := RSA.InitializeReader()
	if err != nil {
		t.Error(err)
	}
}
