package main

import (
	"testing"
	"time"
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
	C3 := Cluster{}
	err = C3.Bootstrap("127.0.0.1", "127.0.0.1", 8082, 8081, RSA.Key)
	if err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second * 2)
	C.Shutdown()
	C2.Shutdown()
	C3.Shutdown()
}
