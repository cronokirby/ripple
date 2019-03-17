package protocol

import "net"

// ping represents a Ping message, that has to be sent from time to time, in
// order to keep a connection between 2 peers alive
type ping struct{}

// joinRequest represents a request from one peer to join the swarm
// chat the other peer is in. The peer receiving this message should
// respond with its own JoinResponse
type joinRequest struct{}

// joinResponse represents the response from a peer after a joinRequest
type joinResponse struct {
	// A list of peers we can connect to, and that may try and connect
	// with us
	peers []net.Addr
}

// message represents a message sent to a swarm by a peer
type message struct {
	content string
}
