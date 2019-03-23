package network

import (
	"io"
	"net"
	"sync"

	"github.com/cronokirby/ripple/internal/protocol"
)

// syncPeer lets us manage access to an arriving peer
type syncConn struct {
	conn    net.Conn
	muEmpty sync.Mutex
	muFill  sync.Mutex
}

// makeSyncConn constructs a valid syncConn in the empty state
func makeSyncConn() *syncConn {
	conn := &syncConn{}
	conn.muEmpty.Lock()
	return conn
}

// empty will block until the connection is full
func (conn *syncConn) empty() {
	conn.muEmpty.Lock()
	conn.conn = nil
	conn.muFill.Unlock()
}

// fill will block until the connection is empty
func (conn syncConn) fill(newConn net.Conn) {
	conn.muFill.Lock()
	conn.conn = newConn
	conn.muEmpty.Unlock()
}

// peerList is a list of peers that can be updated concurrently.
// the zero value of peerList is valid
type peerList struct {
	peers []net.Addr
	lock  sync.Mutex
}

// addPeers safely adds a list of peers to a peerList
// this function can be used safely concurrently, but will block
func (peers *peerList) addPeers(newPeers ...net.Addr) {
	peers.lock.Lock()
	defer peers.lock.Unlock()
	peers.peers = append(peers.peers, newPeers...)
}

// This function is safe so long as the slice isn't mutated
func (peers *peerList) getPeers() []net.Addr {
	return peers.peers
}

func sendMessage(w io.Writer, msg protocol.Message) error {
	data := msg.MessageBytes()
	for len(data) > 0 {
		written, err := w.Write(data)
		if err != nil {
			return err
		}
		data = data[written:]
	}
	return nil
}

// Broadcaster represents an object we can send new messages to, and
// have it broadcast to other peers
type broadcaster struct {
	conns      []net.Conn
	newConns   chan net.Conn
	newContent chan string
}

func makeBroadcaster() *broadcaster {
	return &broadcaster{
		newConns:   make(chan net.Conn),
		newContent: make(chan string),
	}
}

func (bc *broadcaster) loop() {
	for {
		select {
		case content := <-bc.newContent:
			for _, conn := range bc.conns {
				// we don't care about errors
				sendMessage(conn, protocol.NewMessage{Content: content})
			}
		case newConn := <-bc.newConns:
			bc.conns = append(bc.conns, newConn)
		}
	}
}

func (bc *broadcaster) sendConn(conn net.Conn) {
	bc.newConns <- conn
}

func (bc *broadcaster) sendContent(content string) {
	bc.newContent <- content
}
