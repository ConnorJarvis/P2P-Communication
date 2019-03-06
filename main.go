package main

import (
	"encoding/gob"
)

func init() {
	gob.Register(Message{})
	gob.Register(Header{})
	gob.Register([]Peer{})
	gob.Register(Peer{})
	gob.Register(Body{})
	gob.Register(EncryptedMessage{})
}

func main() {
}
