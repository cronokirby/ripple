package network

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/cronokirby/ripple/internal/protocol"
)

const (
	fromPred = iota
	fromSucc
	fromNew
	fromNewPred
	fromNewSucc
)

// sameAddr checks if 2 nodes are the same, by string equality
func sameAddr(a net.Addr, b net.Addr) bool {
	return a.String() == b.String()
}

// clientState holds the state a client needs in normal operation
type clientState struct {
	mu sync.Mutex
	// newPred is the address of a node trying to become our Predecessor
	newPred net.Addr
	// predConn is the connection we have with the current Predecessor
	pred net.Conn
	// succConn is the connection in front of us
	succ net.Conn
	// latestIsSucc is true if the latest conn is trying to replace the Successor
	latestIsSucc bool
	// latestIsPred is true if the latest conn is trying to replace the Predecessor
	//
	// this is mutually exclusive with latestIsSucc, but both fields can be false
	latestIsPred bool
}

// client contains the information needed in normal operation
type client struct {
	log *log.Logger
	// state represents the mutable state under a single lock
	state *clientState
	// latest has its own locking mechanism
	latest *syncConn
}

// originClient wraps a client with the origin of a message
type originClient struct {
	// origin holds the source of the message
	origin int
	log    *log.Logger
	state  *clientState
	// latest is the connection trying to join us
	latest *syncConn
}

// withOrigin embellishes a client with an origin
func (client client) withOrigin(origin int) originClient {
	state := client.state
	latest := client.latest
	log := client.log
	return originClient{origin, log, state, latest}
}

// fmtOrigin is mainly useful for debugging purposes
func (client *originClient) fmtOrigin() string {
	switch client.origin {
	case fromPred:
		return fmt.Sprintf("fromPred: %v", client.state.pred)
	case fromSucc:
		return fmt.Sprintf("fromSucc: %v", client.state.succ)
	case fromNew:
		return fmt.Sprintf("fromNew: %v", client.latest)
	case fromNewPred:
		return fmt.Sprintf("fromNewPred: %v", client.latest)
	case fromNewSucc:
		return fmt.Sprintf("fromNewSucc: %v", client.latest)
	default:
		return "unknown"
	}
}

// HandlePing does nothing at the moment, but could be used for keep alives
func (client *originClient) HandlePing() error {
	// TODO: Implement keep alive
	return nil
}

// HandleJoinSwarm should be accepted when it's coming from a new connection
//
// We then promote the new peer to a node trying to replace our Successor,
// and send a NewPredecessor message to that Successor, as well as a Referral
// back to the new peer.
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
	referral := protocol.Referral{Addr: client.state.succ.RemoteAddr()}
	if err := sendMessage(client.latest.conn, referral); err != nil {
		return err
	}
	newPred := protocol.NewPredecessor{Addr: client.latest.conn.RemoteAddr()}
	if err := sendMessage(client.state.succ, newPred); err != nil {
		return err
	}
	return nil
}

// HandleReferral is always ignored, because we're not joining a swarm
func (client *originClient) HandleReferral(msg protocol.Referral) error {
	return fmt.Errorf(
		"Unexpected Referral message %s",
		client.fmtOrigin(),
	)
}

func (client *originClient) swapPredecessorsIfReady() error {
	if !client.latest.isEmpty() && client.state.latestIsPred {
		latestAddr := client.latest.conn.RemoteAddr()
		announceAddr := client.state.newPred
		if !sameAddr(latestAddr, announceAddr) {
			return fmt.Errorf(
				"Mismatched Predecessors; announced: %v; connected: %v",
				announceAddr,
				latestAddr,
			)
		}
		confirm := protocol.ConfirmReferral{}
		if err := sendMessage(client.state.pred, confirm); err != nil {
			return err
		}
		client.state.pred = client.latest.conn
		client.latest.empty()
		client.state.latestIsPred = false
		client.state.latestIsSucc = false
	}
	return nil
}

// HandleNewPredecessor is handled from our Predecessor
//
// If we have receieved a ConfirmPredecessor already, we can finalise
// the replacement of our Predecessor.
func (client *originClient) HandleNewPredecessor(msg protocol.NewPredecessor) error {
	if client.origin != fromPred {
		return fmt.Errorf(
			"Unexpected NewPredecessor message %s",
			client.fmtOrigin(),
		)
	}
	if client.state.newPred != nil {
		client.log.Printf(
			"Replacing newPred; existing: %v; new: %v\n",
			client.state.newPred, msg.Addr,
		)
	}
	client.state.mu.Lock()
	defer client.state.mu.Unlock()
	client.state.newPred = msg.Addr
	return client.swapPredecessorsIfReady()
}

// HandleConfirmPredecessor acts as a twin to NewPredecessor
//
// The difference between the 2 is who they expect messages from, and what
// state they affect. This will set latestIsPred to true, but
// HandleNewPredecessor will instead set newPred to the announced addr
func (client *originClient) HandleConfirmPredecessor() error {
	if client.origin != fromNewPred {
		return fmt.Errorf(
			"Unexpected ConfirmPredecessor message %s",
			client.fmtOrigin(),
		)
	}
	client.state.mu.Lock()
	defer client.state.mu.Unlock()
	client.state.latestIsPred = true
	return client.swapPredecessorsIfReady()
}
