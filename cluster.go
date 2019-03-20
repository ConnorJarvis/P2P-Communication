package main

import (
	"crypto/rsa"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Cluster struct {
	Peers             map[string]*Peer
	LastSeenPeer      map[string]int64
	PeerIDs           []string
	LocalPeer         Peer
	Values            map[string]*Value
	PeersMutex        *sync.RWMutex
	LastSeenPeerMutex *sync.RWMutex
	ValuesMutex       *sync.RWMutex
	MaxConnections    int
}

type Value struct {
	Modified               int64
	ConflictResolutionMode int
	// File                   *filetransfer.File
	// Wanted bool
	Value map[string]interface{}
}

func (c *Cluster) Bootstrap(LocalIP, RemoteIP string, LocalPort, RemotePort int, Key rsa.PrivateKey, MaxConnections int) error {
	c.MaxConnections = MaxConnections
	c.Peers = make(map[string]*Peer)
	c.PeerIDs = make([]string, 0)
	c.LastSeenPeer = make(map[string]int64)
	c.Values = make(map[string]*Value)
	// c.DownloadQueue = make(chan ChunkRequest, 100000)
	c.PeersMutex = new(sync.RWMutex)
	c.LastSeenPeerMutex = new(sync.RWMutex)
	c.ValuesMutex = new(sync.RWMutex)
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

	// err = c.LocalPeer.StartDownloaders()
	// if err != nil {
	// 	return err
	// }

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

func (c *Cluster) Start(LocalIP string, LocalPort int, Key rsa.PrivateKey, MaxConnections int) error {
	c.MaxConnections = MaxConnections
	c.Peers = make(map[string]*Peer)
	c.PeerIDs = make([]string, 0)
	c.LastSeenPeer = make(map[string]int64)
	c.Values = make(map[string]*Value)
	// c.DownloadQueue = make(chan ChunkRequest, 100000)
	c.PeersMutex = new(sync.RWMutex)
	c.LastSeenPeerMutex = new(sync.RWMutex)
	c.ValuesMutex = new(sync.RWMutex)
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
	// err = c.LocalPeer.StartDownloaders()
	// if err != nil {
	// 	return err
	// }
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
	c.LastSeenPeerMutex.RLock()
	lastSeenPeer := c.LastSeenPeer
	c.LastSeenPeerMutex.RUnlock()
	for i := range lastSeenPeer {
		if time.Now().UTC().Unix()-lastSeenPeer[i] > 60 {
			c.LastSeenPeerMutex.Lock()
			delete(c.LastSeenPeer, i)
			c.LastSeenPeerMutex.Unlock()

			c.PeersMutex.Lock()
			delete(c.Peers, i)
			c.PeersMutex.Unlock()
			for index := 0; index < len(c.PeerIDs); index++ {
				if c.PeerIDs[index] == i {
					c.PeerIDs = append(c.PeerIDs[:index], c.PeerIDs[index+1:]...)
				}
			}
		}
	}
	return nil
}

func (c *Cluster) ParseNewValues(values map[string]*Value) error {
	c.ValuesMutex.Lock()
	for key := range values {
		value := values[key]
		if c.Values[key] == nil {
			// value.File.AvailableChunks = 0
			// value.File.ChunkAvailability = make([]bool, int(value.File.NumberOfChunks))
			// value.Wanted = false
			c.Values[key] = value
		} else if value.ConflictResolutionMode == 0 { //Use the newer value
			if value.Modified > c.Values[key].Modified {
				c.Values[key] = value
			}
		} else if value.ConflictResolutionMode == 1 { //Merge keeping newer values
			if value.Modified > c.Values[key].Modified {
				for subKey := range value.Value {
					if c.Values[key].Value[subKey] == nil {
						c.Values[key].Value[subKey] = value.Value[subKey]
					} else {
						c.Values[key].Value[subKey] = value.Value[subKey]
					}
				}
				c.Values[key].Modified = value.Modified
			}

		}
	}
	c.ValuesMutex.Unlock()
	return nil
}

// func (c *Cluster) AddFile(filePath string) (string, error) {
// 	file, err := filetransfer.InitializeFromFile(filePath)
// 	if err != nil {
// 		return "", err
// 	}
// 	value := make(map[string]interface{})
// 	value[c.LocalPeer.ID] = file.NumberOfChunks
// 	v := Value{Modified: time.Now().UnixNano(), ConflictResolutionMode: 1, File: file, Value: value}
// 	c.ValuesMutex.Lock()
// 	c.Values[file.ID] = &v
// 	c.ValuesMutex.Unlock()
// 	return file.ID, nil
// }

// func (c *Cluster) DownloadFile(id, filePath, temporaryPath string) error {
// 	c.ValuesMutex.RLock()
// 	if c.Values[id].Wanted == true || c.Values[id].File.AvailableChunks == c.Values[id].File.NumberOfChunks {
// 		c.ValuesMutex.RUnlock()
// 		return errors.New("File is already downloading or downloaded")
// 	}
// 	err := c.Values[id].File.InitializeFile(filePath, temporaryPath)
// 	if err != nil {
// 		return err
// 	}
// 	c.ValuesMutex.RUnlock()

// 	c.ValuesMutex.Lock()
// 	c.Values[id].Wanted = true
// 	for i := 0; i < c.Values[id].File.NumberOfChunks; i++ {
// 		c.DownloadQueue <- ChunkRequest{ID: id, Index: i}
// 	}
// 	c.ValuesMutex.Unlock()

// 	return nil
// }
