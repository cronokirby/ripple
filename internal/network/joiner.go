package network

import (
	"errors"
	"io"
	"net"

	"github.com/cronokirby/ripple/internal/protocol"
)

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

// Joiner represents a client that tries and join a new swarm
type Joiner struct{}

// HandlePing exits the connection, because we only expect JoinResponse
func (j *Joiner) HandlePing() error {
	return errors.New("Unexpected Ping message when joining swarm")
}

// HandleJoinRequest exits the connection, because we only expect JoinResponse
func (j *Joiner) HandleJoinRequest() error {
	return errors.New("Unexpected JoinRequest when joining swarm")
}

// HandleJoinResponse allows us to enter the swarm completely.
// This function will loop, and perform the normal swarm operations
func (j *Joiner) HandleJoinResponse(resp protocol.JoinResponse) error {
	return nil
}

// HandleNewMessage exits the connection, because we only expect JoinResponse
func (j *Joiner) HandleNewMessage(msg protocol.NewMessage) error {
	return errors.New("Unexpected NewMessage when joining swarm")
}

// Run blocks the current thread to run a joiner given a certain address,
// which it will use to enter the swarm
// This function will exit if it encounters an error
func (j *Joiner) Run(addr net.Addr) error {
	conn, err := net.Dial(addr.Network(), addr.String())
	defer conn.Close()
	if err != nil {
		return err
	}
	err = sendMessage(conn, protocol.JoinRequest{})
	if err != nil {
		return err
	}
	msg, err := protocol.ReadMessage(conn)
	if err != nil {
		return err
	}
	// This will loop if the client sends us a JoinResponse
	return msg.PassToClient(j)
}
