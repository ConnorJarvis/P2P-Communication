package main

import (
	"testing"
)

func TestStart(t *testing.T) {

	RSA := RSAUtil{}
	RSA.InitializeReader()
	RSA.SetKeyLength(2048)
	RSA.GenerateKey()
	C := Cluster{}
	err := C.Start("127.0.0.1", 8080, RSA.Key)
	if err != nil {
		t.Error(err)
	}
	C.Shutdown()
}

func TestBootstrap(t *testing.T) {
	RSA := RSAUtil{}
	RSA.InitializeReader()
	RSA.SetKeyLength(2048)
	RSA.GenerateKey()
	C := Cluster{}
	C.Start("127.0.0.1", 8080, RSA.Key)
	C2 := Cluster{}
	err := C2.Bootstrap("127.0.0.1", "127.0.0.1", 8081, 8080, RSA.Key)
	if err != nil {
		t.Error(err)
	}
	C.Shutdown()
	C2.Shutdown()
}
