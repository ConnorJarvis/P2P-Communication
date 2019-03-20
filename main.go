package main

import (
	"encoding/gob"
	"time"
)

func init() {
	gob.Register(Message{})
	gob.Register(Header{})
	gob.Register([]Peer{})
	gob.Register(Peer{})
	gob.Register(Body{})
	gob.Register(EncryptedMessage{})
	gob.Register(ChunkRequest{})
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
	C.Start("127.0.0.1", 8080, RSA.Key, 1)
	time.Sleep(time.Second * 2)
	C2 := Cluster{}
	C2.Bootstrap("127.0.0.1", "127.0.0.1", 8082, 8080, RSA.Key, 1)
	time.Sleep(time.Second * 1)
	C3 := Cluster{}
	C3.Bootstrap("127.0.0.1", "127.0.0.1", 8084, 8082, RSA.Key, 1)
	time.Sleep(time.Second * 5)
	time.Sleep(time.Second * 500)
	for {

	}
}
