package main

import (
	"bytes"
	"crypto/rsa"
	"encoding/gob"
	"net"
	"strconv"
)

type Peer struct {
	ID     string
	IP     string
	Port   int
	Peers  []Peer
	RSA    RSAUtil
	Server net.UDPConn
}

func (p *Peer) StopListening() error {
	//Close UDPConn
	err := p.Server.Close()
	if err != nil {
		return err
	}
	return nil
}

func (p *Peer) StartListening() error {
	//Create UDPConn
	u, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP(p.ID), Port: p.Port})
	if err != nil {
		return err
	}
	//Assign UDPConn to Peer
	p.Server = *u

	//Listen to incoming messages
	go func() {

		defer p.Server.Close()

		for {
			buf := make([]byte, 2048)
			//Read data from connection
			n, _, err := p.Server.ReadFrom(buf)
			if err != nil {
				continue
			}
			//Pass message to handle message
			go p.HandleMessage(buf[:n])
		}
	}()
	return nil
}

func (p *Peer) HandleMessage(message []byte) error {
	//Fill bytes.Buffer with message
	messageBytes := bytes.Buffer{}
	_, err := messageBytes.Write(message)
	if err != nil {
		return err
	}
	//Create decoder for bytes.Buffer to Message{}
	decoder := gob.NewDecoder(&messageBytes)
	m := Message{}
	//Decode message
	err = decoder.Decode(&m)
	if err != nil {
		return err
	}
	//Verify message
	err = m.VerifyMessage(p.RSA)
	if err != nil {
		return err
	}

	return nil
}

func (p *Peer) InitializeRSAUtil(length int, Key *rsa.PrivateKey) error {
	//Create RSAUtil struct
	p.RSA = RSAUtil{}
	//Initialize source of crypto/rand bytes
	err := p.RSA.InitializeReader()
	if err != nil {
		return err
	}
	//If key is not set we create a new key
	if Key == nil {
		//Set the key length
		err = p.RSA.SetKeyLength(length)
		if err != nil {
			return err
		}
		//Generate the key
		err = p.RSA.GenerateKey()
		if err != nil {
			return err
		}
	} else {
		//Set peer's key to one passed in
		p.RSA.Key = *Key
	}
	return nil
}

func (p *Peer) SendMessage(p2 Peer, m Message) error {

	//Sign Message
	err := m.SignMessage(p.RSA)
	if err != nil {
		return err
	}
	//Encode Message
	messageBytes, err := m.Encode()
	if err != nil {
		return err
	}
	//Prepare UDP Connection
	conn, err := net.Dial("udp", p2.IP+":"+strconv.Itoa(p2.Port))
	if err != nil {
		return err
	}
	defer conn.Close()
	//Write message to conn
	_, err = conn.Write(messageBytes)
	if err != nil {
		return err
	}
	return nil
}
