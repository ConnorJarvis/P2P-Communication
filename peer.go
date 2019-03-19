package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/gob"
	"fmt"
	"io"
	"math/big"
	"net"
	"strconv"
	"time"
)

type Peer struct {
	ID              string
	IP              string
	Port            int
	RSA             *RSAUtil
	server          *net.UDPConn
	fileServer      *net.TCPListener
	parentCluster   *Cluster
	activeDownloads int
	Stopped         bool
}

func (p *Peer) StopListening() error {
	//Close UDPConn
	err := p.server.Close()
	if err != nil {
		return err
	}
	//Close TCPListener
	err = p.fileServer.Close()
	if err != nil {
		return err
	}
	p.Stopped = true
	return nil
}

func (p *Peer) StartListening() error {
	//Create UDPConn
	u, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP(p.IP), Port: p.Port})
	if err != nil {
		return err
	}
	//Assign UDPConn to Peer
	p.server = u

	//Listen to incoming messages
	go func() {

		for {
			if p.Stopped == true {
				break
			}
			buf := make([]byte, 65507)
			//Read data from connection
			n, _, err := p.server.ReadFrom(buf)
			if err != nil {
				continue
			}
			//Pass message to handle message
			go p.HandleMessage(buf[:n])
		}
	}()

	t, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP(p.IP), Port: p.Port + 1})
	if err != nil {
		return err
	}
	p.fileServer = t

	go func() {
		defer p.fileServer.Close()
		for {
			if p.Stopped == true {
				break
			}

			conn, err := p.fileServer.Accept()
			if err != nil {
				continue
			}
			go p.HandleFileMessage(conn)
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
	p.parentCluster.LastSeenPeerMutex.Lock()
	p.parentCluster.LastSeenPeer[decryptedMessage.Header.From] = time.Now().UTC().Unix()
	p.parentCluster.LastSeenPeerMutex.Unlock()

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
		_, err = conn.Write(messageBytes)

	}()

	return nil
}

func (p *Peer) HandleBootstrap(m Message) error {
	newPeer := m.Body.Content.(Peer)
	p.parentCluster.PeersMutex.Lock()
	p.parentCluster.Peers[newPeer.ID] = &newPeer
	p.parentCluster.PeersMutex.Unlock()
	p.parentCluster.PeerIDs = append(p.parentCluster.PeerIDs, newPeer.ID)
	peers := make([]Peer, 0)
	p.parentCluster.PeersMutex.RLock()
	for _, peer := range p.parentCluster.Peers {
		peers = append(peers, Peer{ID: peer.ID, IP: peer.IP, Port: peer.Port})
	}
	p.parentCluster.PeersMutex.RUnlock()
	p.parentCluster.ValuesMutex.RLock()
	M := Message{Header: Header{ID: 1, From: p.ID}, Body: Body{Content: Gossip{Peers: peers, Values: p.parentCluster.Values}}}
	p.parentCluster.ValuesMutex.RUnlock()
	p.SendMessage(newPeer, M)
	return nil
}

func (p *Peer) HandleNewPeers(m Message) error {
	gossip := m.Body.Content.(Gossip)
	newPeers := gossip.Peers
	changed := false
	err := p.parentCluster.ParseNewValues(gossip.Values)
	if err != nil {
		fmt.Println(err)
		return err
	}
	p.parentCluster.PeersMutex.Lock()
	for i := 0; i < len(newPeers); i++ {
		if p.parentCluster.Peers[newPeers[i].ID] == nil {
			p.parentCluster.Peers[newPeers[i].ID] = &newPeers[i]
			p.parentCluster.PeerIDs = append(p.parentCluster.PeerIDs, newPeers[i].ID)
			changed = true
		}
	}
	p.parentCluster.PeersMutex.Unlock()
	if m.Header.ID == 3 {
		return nil
	}
	if m.Header.ID == 2 {
		peers := make([]Peer, 0)
		p.parentCluster.PeersMutex.RLock()
		for _, peer := range p.parentCluster.Peers {
			peers = append(peers, Peer{ID: peer.ID, IP: peer.IP, Port: peer.Port})
		}
		p.parentCluster.ValuesMutex.RLock()
		M := Message{Header: Header{ID: 3, From: p.ID}, Body: Body{Content: Gossip{Peers: peers, Values: p.parentCluster.Values}}}
		p.parentCluster.ValuesMutex.RUnlock()
		p.SendMessage(*p.parentCluster.Peers[m.Header.From], M)
		p.parentCluster.PeersMutex.RUnlock()
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
			p.parentCluster.PeersMutex.RLock()
			for _, peer := range p.parentCluster.Peers {
				peers = append(peers, Peer{ID: peer.ID, IP: peer.IP, Port: peer.Port})
			}
			p.parentCluster.ValuesMutex.RLock()
			M := Message{Header: Header{ID: 1, From: p.ID}, Body: Body{Content: Gossip{Peers: peers, Values: p.parentCluster.Values}}}
			p.parentCluster.ValuesMutex.RUnlock()
			p.SendMessage(*p.parentCluster.Peers[p.parentCluster.PeerIDs[indexID]], M)
			p.parentCluster.PeersMutex.RUnlock()

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
				p.parentCluster.PeersMutex.RLock()
				for _, peer := range p.parentCluster.Peers {
					peers = append(peers, Peer{ID: peer.ID, IP: peer.IP, Port: peer.Port})
				}
				p.parentCluster.PeersMutex.RUnlock()
				p.parentCluster.ValuesMutex.Lock()
				M := Message{Header: Header{ID: 2, From: p.ID}, Body: Body{Content: Gossip{Peers: peers, Values: p.parentCluster.Values}}}
				p.parentCluster.ValuesMutex.Unlock()
				p.parentCluster.PeersMutex.RLock()
				p.SendMessage(*p.parentCluster.Peers[p.parentCluster.PeerIDs[indexID]], M)
				p.parentCluster.PeersMutex.RUnlock()

			}
			p.parentCluster.AgeOutPeers()

			time.Sleep(time.Millisecond * 500)
		}
	}()
	return nil
}

func (p *Peer) HandleFileMessage(c net.Conn) error {
	message := make([]byte, 2048)
	_, err := c.Read(message)
	if err != nil {
		fmt.Println(err)
		return err
	}
	//Fill bytes.Buffer with message
	messageBytes := bytes.Buffer{}
	_, err = messageBytes.Write(message)
	if err != nil {
		fmt.Println(err)
		return err
	}
	//Create decoder for bytes.Buffer to EncryptedMessage{}
	decoder := gob.NewDecoder(&messageBytes)
	m := EncryptedMessage{}

	//Decode message
	err = decoder.Decode(&m)
	if err != nil {
		fmt.Println(err)
		return err
	}

	//Decrypt message
	decryptedMessage, err := m.Decrypt(*p.RSA)
	if err != nil {
		fmt.Println(err)
		return err
	}

	//Verify message
	err = decryptedMessage.VerifyMessage(*p.RSA)
	if err != nil {
		fmt.Println(err)
		return err
	}

	p.parentCluster.LastSeenPeerMutex.Lock()
	p.parentCluster.LastSeenPeer[decryptedMessage.Header.From] = time.Now().UTC().Unix()
	p.parentCluster.LastSeenPeerMutex.Unlock()
	chunkRequest := decryptedMessage.Body.Content.(ChunkRequest)
	fileID := chunkRequest.ID
	chunkIndex := chunkRequest.Index

	chunkBytes, err := p.parentCluster.Values[fileID].File.ReadChunk(chunkIndex)
	if err != nil {
		fmt.Println(err)
		return err
	}

	writer := io.Writer(c)
	_, err = writer.Write(chunkBytes)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (p *Peer) StartDownloaders() error {
	for i := 0; i < p.parentCluster.MaxConnections; i++ {
		go p.Downloader()
	}
	return nil
}

func (p *Peer) Downloader() error {
	for {
		request := <-p.parentCluster.DownloadQueue
		M := Message{Header: Header{ID: 4, From: p.ID}, Body: Body{Content: request}}

		p.parentCluster.ValuesMutex.RLock()
		peers := p.parentCluster.Values[request.ID].Value
		p.parentCluster.ValuesMutex.RUnlock()
		lowestActiveDownloads := 0
		var lowestActivePeer Peer

		for key := range peers {
			p.parentCluster.PeersMutex.RLock()
			if p.parentCluster.Peers[key].activeDownloads <= lowestActiveDownloads {
				lowestActivePeer = *p.parentCluster.Peers[key]
			}
			p.parentCluster.PeersMutex.RUnlock()
		}
		lowestActivePeer.activeDownloads++

		p.parentCluster.PeersMutex.Lock()
		p.parentCluster.Peers[lowestActivePeer.ID] = &lowestActivePeer
		p.parentCluster.PeersMutex.Unlock()
		err := p.SendFileMessage(lowestActivePeer, M)
		if err != nil {
			fmt.Println(err)
		}
		p.parentCluster.PeersMutex.Lock()
		peer := p.parentCluster.Peers[lowestActivePeer.ID]
		peer.activeDownloads--
		p.parentCluster.Peers[lowestActivePeer.ID] = peer
		p.parentCluster.PeersMutex.Unlock()
	}
	return nil
}

func (p *Peer) SendFileMessage(p2 Peer, m Message) error {

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

	//Prepare UDP Connection
	conn, err := net.Dial("tcp", p2.IP+":"+strconv.Itoa(p2.Port+1))
	if err != nil {
		return err
	}

	//Write message to conn
	_, err = conn.Write(messageBytes)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	if _, err := io.Copy(&b, conn); err != nil {
		return err
	}
	p.parentCluster.ValuesMutex.Lock()
	chunkRequest := m.Body.Content.(ChunkRequest)
	err = p.parentCluster.Values[chunkRequest.ID].File.WriteChunk(chunkRequest.Index, b.Bytes())
	if err != nil {
		return err
	}
	if p.parentCluster.Values[chunkRequest.ID].File.AvailableChunks == p.parentCluster.Values[chunkRequest.ID].File.NumberOfChunks {
		p.parentCluster.Values[chunkRequest.ID].Value[p.ID] = p.parentCluster.Values[chunkRequest.ID].File.NumberOfChunks
		p.parentCluster.Values[chunkRequest.ID].Modified = time.Now().UnixNano()
	}
	p.parentCluster.ValuesMutex.Unlock()

	return nil
}
