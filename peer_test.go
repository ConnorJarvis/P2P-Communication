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

// func TestHandleMessage(t *testing.T) {
// 	P := Peer{IP: "", Port: 8080, ID: "1"}
// 	P.InitializeRSAUtil(2048, nil)
// 	m := Message{Header: Header{ID: 0, From: "1"}, Body: Body{Content: "Hello"}}
// 	m.SignMessage(P.RSA)
// 	messageBytes, _ := m.Encode()
// 	err := P.HandleMessage(messageBytes)
// 	if err != nil {
// 		t.Error(err)
// 	}
// }

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

// func TestSendMessage(t *testing.T) {
// 	R := RSAUtil{}
// 	R.InitializeReader()
// 	R.SetKeyLength(2048)
// 	R.GenerateKey()

// 	P := Peer{IP: "", Port: 8080, ID: "1"}
// 	P.InitializeRSAUtil(2048, &R.Key)
// 	P.StartListening()

// 	P2 := Peer{IP: "", Port: 8081, ID: "2"}
// 	P2.InitializeRSAUtil(2048, &R.Key)
// 	m := Message{Header: Header{ID: 0, From: "1"}, Body: Body{Content: "Hello"}}
// 	err := P2.SendMessage(P, m)
// 	if err != nil {
// 		t.Error(err)
// 	}
// }
