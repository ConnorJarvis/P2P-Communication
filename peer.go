package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/gob"
	"math/big"
	"net"
	"strconv"
)

type Peer struct {
	ID            string
	IP            string
	Port          int
	RSA           *RSAUtil
	server        *net.UDPConn
	parentCluster *Cluster
}

func (p *Peer) StopListening() error {
	//Close UDPConn
	err := p.server.Close()
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
	p.server = u

	//Listen to incoming messages
	go func() {

		defer p.server.Close()

		for {
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
	//Create decoder for bytes.Buffer to Message{}
	decoder := gob.NewDecoder(&messageBytes)
	m := Message{}
	//Decode message
	err = decoder.Decode(&m)
	if err != nil {
		return err
	}
	//Verify message
	err = m.VerifyMessage(*p.RSA)
	if err != nil {
		return err
	}

	switch m.Header.ID {
	case 0:
		p.HandleBootstrap(m)
	case 1:
		p.HandleNewPeers(m)
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
