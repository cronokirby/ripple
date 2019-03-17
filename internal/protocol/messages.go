package protocol

import "net"

// message represents some object we can serialize and be understood
// by a peer
type message interface {
	messageBytes() []byte
}

// ping represents a Ping message, that has to be sent from time to time, in
// order to keep a connection between 2 peers alive
type ping struct{}

func (p ping) messageBytes() []byte {
	return []byte{1}
}

// joinRequest represents a request from one peer to join the swarm
// chat the other peer is in. The peer receiving this message should
// respond with its own JoinResponse
type joinRequest struct{}

func (r joinRequest) messageBytes() []byte {
	return []byte{2}
}

// joinResponse represents the response from a peer after a joinRequest
type joinResponse struct {
	// A list of peers we can connect to, and that may try and connect
	// with us
	peers []net.Addr
}

func (resp joinResponse) messageBytes() []byte {
	l := len(resp.peers)
	acc := []byte{3, byte(l >> 24), byte(l >> 16), byte(l >> 8), byte(l)}
	for _, peer := range resp.peers {
		str := peer.String()
		acc = append(acc, byte(len(str)))
		acc = append(acc, []byte(str)...)
	}
	return acc
}

// message represents a message sent to a swarm by a peer
type newMessage struct {
	content string
}

func (req newMessage) messageBytes() []byte {
	l := len(req.content)
	acc := []byte{4, byte(l >> 24), byte(l >> 16), byte(l >> 8), byte(l)}
	return append(acc, []byte(req.content)...)
}
