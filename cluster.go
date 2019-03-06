package main

import (
	"crypto/rsa"
	"time"

	"github.com/google/uuid"
)

type Cluster struct {
	Peers        map[string]*Peer
	LastSeenPeer map[string]int64
	PeerIDs      []string
	LocalPeer    Peer
	Values       map[string]string
}

func (c *Cluster) Bootstrap(LocalIP, RemoteIP string, LocalPort, RemotePort int, Key rsa.PrivateKey) error {
	c.Peers = make(map[string]*Peer)
	c.PeerIDs = make([]string, 0)
	c.LastSeenPeer = make(map[string]int64)
	uuid, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	c.LocalPeer = Peer{IP: LocalIP, Port: LocalPort, ID: uuid.String(), parentCluster: c}
	c.Peers[c.LocalPeer.ID] = &Peer{IP: LocalIP, Port: LocalPort, ID: uuid.String(), Stopped: false}
	c.PeerIDs = append(c.PeerIDs, uuid.String())
	err = c.LocalPeer.InitializeRSAUtil(2048, &Key)
	if err != nil {
		return err
	}
	err = c.LocalPeer.StartListening()
	if err != nil {
		return err
	}
	err = c.LocalPeer.StartGossip()
	if err != nil {
		return err
	}

	RemotePeer := Peer{IP: RemoteIP, Port: RemotePort}
	RemotePeer.InitializeRSAUtil(2048, &Key)
	if err != nil {
		return err
	}
	M := Message{Header: Header{ID: 0, From: c.LocalPeer.ID}, Body: Body{Content: Peer{IP: LocalIP, Port: LocalPort, ID: c.LocalPeer.ID}}}
	err = c.LocalPeer.SendMessage(RemotePeer, M)
	if err != nil {
		return err
	}
	return nil
}

func (c *Cluster) Start(LocalIP string, LocalPort int, Key rsa.PrivateKey) error {
	c.Peers = make(map[string]*Peer)
	c.PeerIDs = make([]string, 0)
	c.LastSeenPeer = make(map[string]int64)
	uuid, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	c.LocalPeer = Peer{IP: LocalIP, Port: LocalPort, ID: uuid.String(), parentCluster: c}
	c.Peers[c.LocalPeer.ID] = &Peer{IP: LocalIP, Port: LocalPort, ID: uuid.String(), Stopped: false}
	c.PeerIDs = append(c.PeerIDs, uuid.String())
	err = c.LocalPeer.InitializeRSAUtil(2048, &Key)
	if err != nil {
		return err
	}
	err = c.LocalPeer.StartListening()
	if err != nil {
		return err
	}
	err = c.LocalPeer.StartGossip()
	if err != nil {
		return err
	}
	return nil
}

func (c *Cluster) Shutdown() error {
	err := c.LocalPeer.StopListening()
	if err != nil {
		return err
	}
	return nil
}

func (c *Cluster) AgeOutPeers() error {
	for i := range c.LastSeenPeer {
		if time.Now().UTC().Unix()-c.LastSeenPeer[i] > 60 {
			delete(c.LastSeenPeer, i)
			delete(c.Peers, i)
			for index := 0; index < len(c.PeerIDs); index++ {
				if c.PeerIDs[index] == i {
					c.PeerIDs = append(c.PeerIDs[:index], c.PeerIDs[index+1:]...)
				}
			}
		}
	}
	return nil
}
