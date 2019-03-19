package main

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
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
	value2 := make(map[string]interface{})
	value2["1"] = "Test"
	modified1 := time.Now().UnixNano()
	value1 := &Value{Modified: modified1, ConflictResolutionMode: 1, Value: value2}
	C.Values["Test"] = value1
	spew.Dump()
	time.Sleep(time.Second * 1)
	C2 := Cluster{}
	err := C2.Bootstrap("127.0.0.1", "127.0.0.1", 8081, 8080, RSA.Key)
	if err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second * 1)
	C3 := Cluster{}
	err = C3.Bootstrap("127.0.0.1", "127.0.0.1", 8082, 8081, RSA.Key)
	value3 := make(map[string]interface{})
	value3["1"] = "Test2"
	modified2 := time.Now().UnixNano()
	value4 := &Value{Modified: modified2, ConflictResolutionMode: 1, Value: value3}
	C3.Values["Test"] = value4
	if err != nil {
		t.Error(err)
	}

	time.Sleep(time.Second * 3)
	if !reflect.DeepEqual(C.Values, C2.Values) {
		t.Error(errors.New("Values did not propagate"))
	}
	if !reflect.DeepEqual(C2.Values, C3.Values) {
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
