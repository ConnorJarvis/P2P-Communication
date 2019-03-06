package main

import (
	"testing"
)

func TestStartListening(t *testing.T) {
	P := Peer{IP: "", Port: 8080, ID: "1"}
	err := P.StartListening()
	if err != nil {
		t.Error(err)
	}
	P.StopListening()
}

func TestStopListening(t *testing.T) {
	P := Peer{IP: "", Port: 8080, ID: "1"}
	P.StartListening()
	err := P.StopListening()
	if err != nil {
		t.Error(err)
	}
}

func TestInitializeRSAUtil(t *testing.T) {
	P := Peer{IP: "", Port: 8080, ID: "1"}
	err := P.InitializeRSAUtil(2048, nil)
	if err != nil {
		t.Error(err)
	}
	P2 := Peer{IP: "", Port: 8080, ID: "1"}
	RSA := RSAUtil{}
	RSA.InitializeReader()
	RSA.SetKeyLength(2048)
	RSA.GenerateKey()
	err = P2.InitializeRSAUtil(2048, &RSA.Key)
	if err != nil {
		t.Error(err)
	}

}
