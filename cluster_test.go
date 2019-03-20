package main

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestStart(t *testing.T) {

	RSA := RSAUtil{}
	RSA.InitializeReader()
	RSA.SetKeyLength(2048)
	RSA.GenerateKey()
	C := Cluster{}
	err := C.Start("127.0.0.1", 8080, RSA.Key, 1)
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
	C.Start("127.0.0.1", 8080, RSA.Key, 1)
	// id, _ := C.AddFile("./go.mod")
	time.Sleep(time.Second * 1)
	C2 := Cluster{}
	err := C2.Bootstrap("127.0.0.1", "127.0.0.1", 8082, 8080, RSA.Key, 1)
	if err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second * 1)
	C3 := Cluster{}
	err = C3.Bootstrap("127.0.0.1", "127.0.0.1", 8084, 8082, RSA.Key, 1)
	if err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second * 5)
	// err = C3.DownloadFile(id, "./go2.mod", "./tmp")
	// if err != nil {
	// 	t.Error(errors.New("Failed to download"))
	// }
	// file1, _ := os.Open("./go.mod")
	// h1 := sha1.New()
	// io.Copy(h1, file1)

	// file2, _ := os.Open("./go2.mod")
	// h2 := sha1.New()
	// io.Copy(h2, file2)
	// os.Remove("./go2.mod")
	// if !reflect.DeepEqual(h1.Sum(nil), h2.Sum(nil)) {
	// 	t.Error(errors.New("File did not transfer properly"))
	// }
	if !reflect.DeepEqual(C.Values, C2.Values) {
		t.Error(errors.New("Values did not propagate"))
	}
	if !reflect.DeepEqual(C.Values, C3.Values) {
		t.Error(errors.New("Values did not propagate"))
	}

	if !reflect.DeepEqual(C.Peers, C2.Peers) {
		t.Error(errors.New("Peers did not propagate"))
	}
	if !reflect.DeepEqual(C3.Peers, C2.Peers) {
		t.Error(errors.New("Peers did not propagate"))
	}
	C.Shutdown()
	C2.Shutdown()
	C3.Shutdown()
}
