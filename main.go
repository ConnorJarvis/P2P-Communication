package main

import "encoding/gob"

func init() {
	gob.Register(Message{})
}

func main() {

}
