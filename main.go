package main

import "encoding/gob"

func init() {
	gob.Register(Message{})
	gob.Register(Header{})
	gob.Register(Body{})
}

func main() {

}
