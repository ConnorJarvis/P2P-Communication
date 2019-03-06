package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/gob"
	"math/big"
	"net"
	"strconv"
	"time"
)

type Peer struct {
	ID            string
	IP            string
	Port          int
	RSA           *RSAUtil
	server        *net.UDPConn
	parentCluster *Cluster
	Stopped       bool
}

func (p *Peer) StopListening() error {
	//Close UDPConn
	err := p.server.Close()
	if err != nil {
		return err
	}
	p.Stopped = true
	return nil
}

func (p *Peer) StartListening() error {
	//Create UDPConn
	u, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP(p.ID), Port: p.Port})
	if err != nil {
		return err
	}
	//Assign UDPConn to Peer
	p.server = u

	//Listen to incoming messages
	go func() {

		defer p.server.Close()

		for {
			if p.Stopped == true {
				break
			}
			buf := make([]byte, 2048)
			//Read data from connection
			n, _, err := p.server.ReadFrom(buf)
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
	//Create decoder for bytes.Buffer to EncryptedMessage{}
	decoder := gob.NewDecoder(&messageBytes)
	m := EncryptedMessage{}

	//Decode message
	err = decoder.Decode(&m)
	if err != nil {
		return err
	}

	//Decrypt message
	decryptedMessage, err := m.Decrypt(*p.RSA)
	if err != nil {
		return err
	}

	//Verify message
	err = decryptedMessage.VerifyMessage(*p.RSA)
	if err != nil {
		return err
	}
	p.parentCluster.LastSeenPeer[decryptedMessage.Header.From] = time.Now().UTC().Unix()

	switch decryptedMessage.Header.ID {
	case 0:
		p.HandleBootstrap(*decryptedMessage)
	case 1:
		p.HandleNewPeers(*decryptedMessage)
	case 2:
		p.HandleNewPeers(*decryptedMessage)
	case 3:
		p.HandleNewPeers(*decryptedMessage)
	}
	return nil
}

func (p *Peer) InitializeRSAUtil(length int, Key *rsa.PrivateKey) error {
	//Create RSAUtil struct
	p.RSA = &RSAUtil{}
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
	err := m.SignMessage(*p.RSA)
	if err != nil {
		return err
	}

	//Encrypt Message
	encryptedMessage, err := m.Encrypt(*p.RSA)
	if err != nil {
		return err
	}

	//Encode Message
	messageBytes, err := encryptedMessage.Encode()
	if err != nil {
		return err
	}
	go func() {
		//Prepare UDP Connection
		conn, _ := net.Dial("udp", p2.IP+":"+strconv.Itoa(p2.Port))

		defer conn.Close()
		//Write message to conn
		_, _ = conn.Write(messageBytes)

	}()

	return nil
}

func (p *Peer) HandleBootstrap(m Message) error {
	newPeer := m.Body.Content.(Peer)
	p.parentCluster.Peers[newPeer.ID] = &newPeer
	p.parentCluster.PeerIDs = append(p.parentCluster.PeerIDs, newPeer.ID)
	peers := make([]Peer, 0)
	for _, peer := range p.parentCluster.Peers {
		peers = append(peers, Peer{ID: peer.ID, IP: peer.IP, Port: peer.Port})
	}
	M := Message{Header: Header{ID: 1, From: p.ID}, Body: Body{Content: peers}}
	p.SendMessage(newPeer, M)
	sentIDs := make(map[int]bool)
	peersToShare := len(p.parentCluster.PeerIDs) - 1
	if peersToShare > 5 {
		peersToShare = 5
	}
	for i := 0; i < peersToShare; i++ {
		indexID := 0
		for {
			index, err := rand.Int(rand.Reader, big.NewInt(int64(len(p.parentCluster.PeerIDs)-1)))
			indexID = int(index.Int64())
			if err != nil {
				return err
			}
			if p.parentCluster.PeerIDs[indexID] != p.ID && sentIDs[indexID] == false {
				sentIDs[indexID] = true
				break
			}
		}
		peers := make([]Peer, 0)
		for _, peer := range p.parentCluster.Peers {
			peers = append(peers, Peer{ID: peer.ID, IP: peer.IP, Port: peer.Port})
		}
		M := Message{Header: Header{ID: 1, From: p.ID}, Body: Body{Content: peers}}
		p.SendMessage(*p.parentCluster.Peers[p.parentCluster.PeerIDs[indexID]], M)

	}
	return nil
}

func (p *Peer) HandleNewPeers(m Message) error {
	newPeers := m.Body.Content.([]Peer)
	changed := false
	for i := 0; i < len(newPeers); i++ {
		if p.parentCluster.Peers[newPeers[i].ID] == nil {
			p.parentCluster.Peers[newPeers[i].ID] = &newPeers[i]
			p.parentCluster.PeerIDs = append(p.parentCluster.PeerIDs, newPeers[i].ID)
			changed = true
		}
	}
	if m.Header.ID == 3 {
		return nil
	}
	if m.Header.ID == 2 {
		peers := make([]Peer, 0)
		for _, peer := range p.parentCluster.Peers {
			peers = append(peers, Peer{ID: peer.ID, IP: peer.IP, Port: peer.Port})
		}
		M := Message{Header: Header{ID: 3, From: p.ID}, Body: Body{Content: peers}}
		p.SendMessage(*p.parentCluster.Peers[m.Header.From], M)
		return nil
	}
	if changed {
		sentIDs := make(map[int]bool)
		peersToShare := len(p.parentCluster.PeerIDs) - 1
		if peersToShare > 5 {
			peersToShare = 5
		}
		for i := 0; i < peersToShare; i++ {
			indexID := 0
			for {
				index, err := rand.Int(rand.Reader, big.NewInt(int64(len(p.parentCluster.PeerIDs)-1)))
				indexID = int(index.Int64())
				if err != nil {
					return err
				}
				if p.parentCluster.PeerIDs[indexID] != p.ID && sentIDs[indexID] == false {
					sentIDs[indexID] = true
					break
				}
			}
			peers := make([]Peer, 0)
			for _, peer := range p.parentCluster.Peers {
				peers = append(peers, Peer{ID: peer.ID, IP: peer.IP, Port: peer.Port})
			}
			M := Message{Header: Header{ID: 1, From: p.ID}, Body: Body{Content: peers}}
			p.SendMessage(*p.parentCluster.Peers[p.parentCluster.PeerIDs[indexID]], M)

		}
	}
	return nil
}

func (p *Peer) StartGossip() error {
	go func() {
		for {
			if p.Stopped == true {
				break
			}
			if len(p.parentCluster.PeerIDs) != 1 {
				indexID := 0
				for {

					index, _ := rand.Int(rand.Reader, big.NewInt(int64(len(p.parentCluster.PeerIDs)-1)))
					indexID = int(index.Int64())

					if p.parentCluster.PeerIDs[indexID] != p.ID {
						break
					}
				}
				peers := make([]Peer, 0)
				for _, peer := range p.parentCluster.Peers {
					peers = append(peers, Peer{ID: peer.ID, IP: peer.IP, Port: peer.Port})
				}
				M := Message{Header: Header{ID: 2, From: p.ID}, Body: Body{Content: peers}}
				p.SendMessage(*p.parentCluster.Peers[p.parentCluster.PeerIDs[indexID]], M)
			}
			p.parentCluster.AgeOutPeers()

			time.Sleep(time.Millisecond * 500)
		}
	}()
	return nil
}
