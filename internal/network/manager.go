package network

import (
	"fmt"
	"log"
	"net"
	"sync"
)

const (
	fromPred = iota
	fromSucc
	fromNew
)

// sameAddr checks if 2 nodes are the same, by string equality
func sameAddr(a net.Addr, b net.Addr) bool {
	return a.String() == b.String()
}

// clientState holds the state a client needs in normal operation
type clientState struct {
	log *log.Logger
	mu  sync.Mutex
	// newPred is the address of a node trying to become our Predecessor
	newPred net.Addr
	// newPred is the connection we have with that node
	newPredConn net.Conn
	// predConn is the connection we have with the current Predecessor
	predConn net.Conn
	// succConn is the connection in front of us
	succConn net.Conn
	// latestIsSucc is true if the latest conn is trying to replace the Successor
	latestIsSucc bool
}

// client contains the information needed in normal operation
type client struct {
	// state represents the mutable state under a single lock
	state *clientState
	// latestConn has its own locking mechanism
	latestConn *syncConn
}

// originClient wraps a client with the origin of a message
type originClient struct {
	// origin holds the source of the message
	origin     int
	state      *clientState
	latestConn *syncConn
}

// withOrigin embellishes a client with an origin
func (client client) withOrigin(origin int) originClient {
	state := client.state
	latestConn := client.latestConn
	return originClient{origin, state, latestConn}
}

// fmtOrigin is mainly useful for debugging purposes
func (client *originClient) fmtOrigin() string {
	switch client.origin {
	case fromPred:
		return fmt.Sprintf("fromPred: %v", client.state.predConn)
	case fromSucc:
		return fmt.Sprintf("fromSucc: %v", client.state.succConn)
	case fromNew:
		return fmt.Sprintf("fromNew: %v", client.latestConn)
	default:
		return "unknown"
	}
}

// HandlePing does nothing at the moment, but could be used for keep alives
func (client *originClient) HandlePing() error {
	// TODO: Implement keep alive
	return nil
}

// HandleJoinSwarm just ignores the message.
//
// We could quit the client instead of just logging, but it's
// a bit more resilient to keep chugging along.
func (client *originClient) HandleJoinSwarm() error {
	if client.origin != fromNew {
		return fmt.Errorf(
			"Unexpected JoinSwarm message %s",
			client.fmtOrigin(),
		)
	}
	client.state.mu.Lock()
	defer client.state.mu.Unlock()
	client.state.latestIsSucc = true
	return nil
}
