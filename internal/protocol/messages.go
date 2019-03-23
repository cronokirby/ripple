package protocol

import (
	"fmt"
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

// JoinSwarm represents a request from one peer to join the swarm
type JoinSwarm struct{}

// MessageBytes serializes a JoinSwarm into a byte slice
func (r JoinSwarm) MessageBytes() []byte {
	return []byte{2}
}

// PassToClient implements the visitor pattern for JoinSwarm
func (r JoinSwarm) PassToClient(client Client) error {
	return client.HandleJoinSwarm()
}

// Referral represents a redirection to another node.
//
// After sending a JoinSwarm message, a peer receives this in response.
// It will then contact the address in this node with a ConfirmPredecessor message.
type Referral struct {
	// Addr is the address of the node we're being referred to
	Addr net.Addr
}

// MessageBytes serializes a Refferal into a byte slice
func (r Referral) MessageBytes() []byte {
	addrString := r.Addr.String()
	header := []byte{3, byte(len(addrString))}
	return append(header, []byte(addrString)...)
}

// PassToClient implements the visitor pattern for Refferal
func (r Referral) PassToClient(client Client) error {
	return client.HandleReferral(r)
}

// NewPredecessor is used to alert a sucessor node of a new replacement.
//
// After being contacted by a new node interested in joining the swarm,
// a node will send this message to its Successor, allowing that Successor
// to be aware of that new node. When the new node contacts that Successor,
// the Successor can replace this node with the newly joined one.
type NewPredecessor struct {
	// Addr is the node we're being made aware of
	Addr net.Addr
}

// MessageBytes serializes a NewPredecessor
func (r NewPredecessor) MessageBytes() []byte {
	addrString := r.Addr.String()
	header := []byte{4, byte(len(addrString))}
	return append(header, []byte(addrString)...)
}

// PassToClient implements the visitor pattern for NewPredecessor
func (r NewPredecessor) PassToClient(client Client) error {
	return client.HandleNewPredecessor(r)
}

// ConfirmPredecessor is used by a joining need to finalize that process
//
// It contacts the node contained in Referral using this message, and
// once that node receieves a NewPredecessor message with the same address
// as this joining node, then it becomes the new Predecessor.
type ConfirmPredecessor struct{}

// MessageBytes serializes a ConfirmPredecessor
func (r ConfirmPredecessor) MessageBytes() []byte {
	return []byte{5}
}

// PassToClient implements the visitor pattern for ConfirmPredecessor
func (r ConfirmPredecessor) PassToClient(client Client) error {
	return client.HandleConfirmPredecessor()
}

// ConfirmReferral is used by a node to confirm replacement of its Predecessor
//
// After receiving this from its Successor, a node can replace it.
type ConfirmReferral struct{}

// MessageBytes serializes a ConfirmReferral
func (r ConfirmReferral) MessageBytes() []byte {
	return []byte{6}
}

// PassToClient implements the visitor pattern for ConfirmReferral
func (r ConfirmReferral) PassToClient(client Client) error {
	return client.HandleConfirmReferral()
}

// NewMessage allows us to send new text messages across the swarm
type NewMessage struct {
	// Sender is the node that sent this message
	Sender net.Addr
	// Content is the actual text content of the message
	Content string
}

// MessageBytes serializes a NewMessage
func (r NewMessage) MessageBytes() []byte {
	senderString := r.Sender.String()
	bytes := []byte{7, byte(len(senderString))}
	bytes = append(bytes, []byte(senderString)...)
	cLen := len(r.Content)
	bytes = append(bytes, byte(cLen>>24), byte(cLen>>16), byte(cLen>>8), byte(cLen))
	bytes = append(bytes, []byte(r.Content)...)
	return bytes
}

// PassToClient implements the visitor pattern for NewMessage
func (r NewMessage) PassToClient(client Client) error {
	return client.HandleNewMessage(r)
}

func readAddr(r io.Reader, addrLen byte, slice []byte, buf []byte) (net.Addr, error) {
	addrBuf := make([]byte, 0, addrLen)
	addrBuf = append(addrBuf, slice[:addrLen]...)
	for byte(len(addrBuf)) < addrLen {
		amount, err := r.Read(buf)
		if err != nil {
			return nil, err
		}
		addrBuf = append(addrBuf, buf[:amount]...)
	}
	addrString := string(addrBuf[:addrLen])
	addr, err := net.ResolveTCPAddr("tcp", addrString)
	if err != nil {
		return nil, err
	}
	return addr, nil
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
		res = JoinSwarm{}
	case 3:
		addrLen := slice[1]
		slice = slice[2:]
		addr, err := readAddr(r, addrLen, slice, buf)
		if err != nil {
			return nil, err
		}
		res = Referral{Addr: addr}
	case 4:
		addrLen := slice[1]
		slice = slice[2:]
		addr, err := readAddr(r, addrLen, slice, buf)
		if err != nil {
			return nil, err
		}
		res = NewPredecessor{Addr: addr}
	case 5:
		res = ConfirmPredecessor{}
	case 6:
		res = ConfirmReferral{}
	case 7:
		addrLen := slice[1]
		slice = slice[2:]
		// We can't reuse the function because we need to modify slice
		addrBuf := make([]byte, 0, addrLen)
		addrBuf = append(addrBuf, slice[:addrLen]...)
		slice = slice[addrLen:]
		overwrite := byte(len(addrBuf)) < addrLen
		for byte(len(addrBuf)) < addrLen {
			amount, err := r.Read(buf)
			if err != nil {
				return nil, err
			}
			addrBuf = append(addrBuf, buf[:amount]...)
		}
		addrString := string(addrBuf[:addrLen])
		if overwrite {
			slice = addrBuf[addrLen:]
		}
		addr, err := net.ResolveTCPAddr("tcp", addrString)
		if err != nil {
			return nil, err
		}
		length := uint(slice[0]) << 24
		length |= uint(slice[1]) << 16
		length |= uint(slice[2]) << 8
		length |= uint(slice[3])
		slice = slice[4:]
		stringBuf := make([]byte, 0, length)
		stringBuf = append(stringBuf, slice...)
		for uint(len(stringBuf)) < length {
			amount, err := r.Read(buf)
			if err != nil {
				return nil, err
			}
			stringBuf = append(stringBuf, buf[:amount]...)
		}
		content := string(stringBuf[:length])
		res = NewMessage{Sender: addr, Content: content}
	}
	return res, nil
}

// ContentReceiver is some type that can do something when new content arrives
//
// This is useful in testing, as it allows us to define tests that check
// if certain content was received. In normal usage, this allows us to
// update a gui or terminal based client.
type ContentReceiver interface {
	// ReceiveContent allows this object to react to some new content
	ReceiveContent(string)
}

// PrintReceiver is a ContentReceiver that just prints the received content
type PrintReceiver struct{}

// ReceiveContent just prints the new content we've received
func (p PrintReceiver) ReceiveContent(content string) {
	fmt.Println(content)
}

// NilReceiver simply does nothing on receieving content
type NilReceiver struct{}

// ReceiveContent does nothing
func (NilReceiver) ReceiveContent(content string) {}
