package main

import (
	"encoding/gob"
	"time"

	"github.com/davecgh/go-spew/spew"
)

func init() {
	gob.Register(Message{})
	gob.Register(Header{})
	gob.Register([]Peer{})
	gob.Register(Peer{})
	gob.Register(Body{})
	gob.Register(EncryptedMessage{})
	gob.Register(map[string]*Peer{})
	gob.Register(map[string]Value{})
	gob.Register(Gossip{})
}

func main() {
	RSA := RSAUtil{}
	RSA.InitializeReader()
	RSA.SetKeyLength(2048)
	RSA.GenerateKey()

	C := Cluster{}
	C.Start("127.0.0.1", 8080, RSA.Key)
	value := make(map[string]interface{})
	value["1"] = "Test"
	C.Values["Test"] = &Value{Modified: time.Now().UnixNano(), ConflictResolutionMode: 2, Value: value}
	time.Sleep(time.Second * 2)
	C2 := Cluster{}
	C2.Bootstrap("127.0.0.1", "127.0.0.1", 8081, 8080, RSA.Key)
	time.Sleep(time.Second * 1)
	C3 := Cluster{}
	C3.Bootstrap("127.0.0.1", "127.0.0.1", 8082, 8081, RSA.Key)
	time.Sleep(time.Second * 10)
	C3.Shutdown()
	time.Sleep(time.Second * 10)
	spew.Dump(C.Peers)
	spew.Dump(C2.Peers)
	time.Sleep(time.Second * 70)
	spew.Dump(&C.Peers)
	spew.Dump(C2.Peers)
	for {

	}
}
