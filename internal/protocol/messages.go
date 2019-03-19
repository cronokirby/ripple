package protocol

import (
	"io"
	"net"
)

// Message represents some object we can serialize and be understood
// by a peer
type Message interface {
	MessageBytes() []byte
	PassToClient(Client) error
}

// Ping represents a Ping message, that has to be sent from time to time, in
// order to keep a connection between 2 peers alive
type Ping struct{}

// MessageBytes serializes a ping message
func (p Ping) MessageBytes() []byte {
	return []byte{1}
}

// PassToClient implements the visitor pattern for Ping
func (p Ping) PassToClient(c Client) error {
	return c.HandlePing()
}

// JoinRequest represents a request from one peer to join the swarm
// chat the other peer is in. The peer receiving this message should
// respond with its own JoinResponse
type JoinRequest struct{}

// MessageBytes serializes JoinRequest
func (r JoinRequest) MessageBytes() []byte {
	return []byte{2}
}

// PassToClient just implements the visitor pattern
func (r JoinRequest) PassToClient(c Client) error {
	return c.HandleJoinRequest()
}

// JoinResponse represents the response from a peer after a joinRequest
type JoinResponse struct {
	// A list of peers we can connect to, and that may try and connect
	// with us
	peers []net.Addr
}

// MessageBytes serializes a JoinResponse
func (resp JoinResponse) MessageBytes() []byte {
	l := len(resp.peers)
	acc := []byte{3, byte(l >> 24), byte(l >> 16), byte(l >> 8), byte(l)}
	for _, peer := range resp.peers {
		str := peer.String()
		acc = append(acc, byte(len(str)))
		acc = append(acc, []byte(str)...)
	}
	return acc
}

// PassToClient implements the visitor pattern for JoinResponse
func (resp JoinResponse) PassToClient(c Client) error {
	return c.HandleJoinResponse(resp)
}

// NewMessage represents a message sent to a swarm by a peer
type NewMessage struct {
	content string
}

// MessageBytes serializes a NewMessage
func (req NewMessage) MessageBytes() []byte {
	l := len(req.content)
	acc := []byte{4, byte(l >> 24), byte(l >> 16), byte(l >> 8), byte(l)}
	return append(acc, []byte(req.content)...)
}

// PassToClient implements the visitor pattern for NewMessage
func (req NewMessage) PassToClient(c Client) error {
	return c.HandleNewMessage(req)
}

// ReadMessage reads bytes into a Message
// It does the opposite of MessageBytes.
// If the byte slice is misformatted, or not long enough, this will fail
func ReadMessage(r io.Reader) (Message, error) {
	buf := make([]byte, 1024)
	amount, err := r.Read(buf)
	if err != nil {
		return nil, err
	}
	slice := buf[:amount]
	var res Message
	switch slice[0] {
	case 1:
		res = Ping{}
	case 2:
		res = JoinRequest{}
	case 3:
		peerCount := uint(slice[1]) << 24
		peerCount |= uint(slice[2]) << 16
		peerCount |= uint(slice[3]) << 8
		peerCount |= uint(slice[4])
		slice = slice[5:]
		var peers []net.Addr
		for uint(len(peers)) < peerCount {
			length := slice[0]
			slice = slice[1:]
			for len(slice) < 256 && byte(len(slice)) < length {
				newBuf := make([]byte, len(slice))
				copy(newBuf, slice)
				slice = newBuf
				amount, err := r.Read(buf)
				if err != nil {
					return nil, err
				}
				slice = append(slice, buf[:amount]...)
			}
			addrString := string(slice[:length])
			peers = append(peers, &net.IPAddr{IP: net.ParseIP(addrString)})
			slice = slice[length:]
		}
		res = JoinResponse{peers}
	case 4:
		length := uint(slice[1]) << 24
		length |= uint(slice[2]) << 16
		length |= uint(slice[3]) << 8
		length |= uint(slice[4])
		slice = slice[5:]
		stringBuf := make([]byte, 0, length)
		stringBuf = append(stringBuf, slice...)
		for uint(len(stringBuf)) < length {
			amount, err := r.Read(buf)
			if err != nil {
				return nil, err
			}
			stringBuf = append(stringBuf, buf[:amount]...)
		}
		res = NewMessage{content: string(stringBuf)}
	}
	return res, nil
}
