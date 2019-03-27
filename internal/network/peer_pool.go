package network

import (
	"fmt"
	"net"
	"sync"

	"github.com/cronokirby/ripple/internal/protocol"
)

const (
	newRole  = 0
	predRole = 1 << iota
	succRole
)

func isNewRole(role int) bool {
	return role == 0
}

func isSuccRole(role int) bool {
	return (role & succRole) != 0
}

func isPredRole(role int) bool {
	return (role & predRole) != 0
}

func isUselessRole(role int) bool {
	return !isSuccRole(role) && !isPredRole(role)
}

func originString(role int) string {
	switch role {
	case newRole:
		return "new"
	case predRole | succRole:
		return "pred/succ"
	case predRole:
		return "pred"
	case succRole:
		return "succ"
	default:
		return "unknown"
	}

}

// peer holds an address and a connection so other peers can connect as well
type peer struct {
	// addr is the address this peer is known by
	addr net.Addr
	// conn is the connection we currently have with it
	conn net.Conn
}

// originMessage wrapes a message with an origin
type originMessage struct {
	origin int
	msg    protocol.Message
}

// originError wraps an error with an origin
type originError struct {
	origin int
	err    error
}

func (err originError) Error() string {
	return fmt.Sprintf("Error reading from %s: %v", originString(err.origin), err.err)
}

// peerPool allows us to manage connections sensibly.
//
// We change which connections represent what client
// often in the protocol. For example, we need to replace the
// current Predecessor with a new connection. We also need
// to manage the fact that a single connection may be serving
// multiple roles. Instead of having a loop over each role,
// and the connection in that loop changing. We can instead
// maintain one loop per connection, and have the roles change.
// This allows us to react to changing roles in a more robust way.
// One problem with the "loop per role" way, is that we block
// in the loop while waiting for the next message,
// and there's no good way of replacing the conn while this is
// happening. By instead figuring out where to push the message
// after receiving this, we alleviate this problem
type peerPool struct {
	// roles maps the identity of each connection to its role set
	roles    map[string]int
	messages chan originMessage
	errors   chan error
	mu       sync.RWMutex
}

func makePeerPool() *peerPool {
	return &peerPool{
		make(map[string]int),
		make(chan originMessage),
		make(chan error),
		sync.RWMutex{},
	}
}

func (pool *peerPool) submit(peer peer, pred bool) {
	var role int
	if pred {
		role = predRole
	} else {
		role = succRole
	}
	newlyInserted := false
	pool.mu.Lock()
	oldRole, ok := pool.roles[peer.addr.String()]
	newlyInserted = !ok
	pool.roles[peer.addr.String()] = role | oldRole
	pool.mu.Unlock()
	if !newlyInserted {
		return
	}
	go poolLoop(pool, peer)
}

// remove safely removes a peer from the pool, closing
// the connection if it no longer holds any useful roles
func (pool *peerPool) remove(peer peer, pred bool) {
	var role int
	if pred {
		role = predRole
	} else {
		role = succRole
	}
	shouldClose := false
	pool.mu.Lock()
	newRole := pool.roles[peer.addr.String()] &^ role
	if isUselessRole(newRole) {
		delete(pool.roles, peer.addr.String())
		// closing the connection may be slow
		shouldClose = true
	} else {
		pool.roles[peer.addr.String()] = newRole
	}
	pool.mu.Unlock()
	if shouldClose {
		peer.conn.Close()
	}
}

func (pool *peerPool) getRole(peer peer) int {
	pool.mu.RLock()
	defer pool.mu.RUnlock()
	return pool.roles[peer.addr.String()]
}

func poolLoop(pool *peerPool, peer peer) {
	for {
		msg, err := protocol.ReadMessage(peer.conn)
		// an error can also indicate a closed connection, our signal to die
		if err != nil {
			role := pool.getRole(peer)
			// we're no longer needed, and receieved a close connection error
			if isUselessRole(role) {
				break
			}
			pool.errors <- originError{origin: role, err: err}
			continue
		}
		role := pool.getRole(peer)
		pool.messages <- originMessage{origin: role, msg: msg}
	}
}
