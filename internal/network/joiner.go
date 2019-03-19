package network

import (
	"errors"
	"fmt"
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
type Joiner struct {
	peers []net.Addr
}

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
	j.peers = resp.Peers
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
	if err != nil {
		return err
	}
	defer conn.Close()
	err = sendMessage(conn, protocol.JoinRequest{})
	if err != nil {
		return err
	}
	msg, err := protocol.ReadMessage(conn)
	if err != nil {
		return err
	}
	// This will loop if the client sends us a JoinResponse
	err = msg.PassToClient(j)
	if err != nil {
		return err
	}
	// Connect to every peer we've been given
	newConns := make(chan net.Conn)
	newConns <- conn
	firstClient := &normalClient{newConns}
	go firstClient.loop(conn)
	for _, addr := range j.peers {
		client := &normalClient{newConns}
		go client.connectAndLoop(addr)
	}
	// Listen and loop with new connections
	return nil
}

type normalClient struct {
	// newConn is a channel used to notify the broadcast messanger of
	// new connections it will need to sending to
	// This is really only used at the start of the lifecylce of a normalClient,
	// after it manages to connect to the peer it is attached to
	newConns chan net.Conn
}

func (client *normalClient) HandlePing() error {
	// TODO: Handle timeouts
	return nil
}

func (client *normalClient) HandleJoinRequest() error {
	return errors.New("Unexpected JoinRequest in normalClient")
}

func (client *normalClient) HandleJoinResponse(_ protocol.JoinResponse) error {
	return errors.New("Unexpected JoinResponse in normalClient")
}

func (client *normalClient) HandleNewMessage(msg protocol.NewMessage) error {
	fmt.Println(msg.Content)
	return nil
}

// connectAndLoop starts a normal client with a new address to connect to,
// and returns an error whenever something fatal occurrs
func (client *normalClient) connectAndLoop(addr net.Addr) error {
	conn, err := net.Dial(addr.Network(), addr.String())
	if err != nil {
		return err
	}
	return client.loop(conn)
}

func (client *normalClient) loop(conn net.Conn) error {
	defer conn.Close()
	client.newConns <- conn
	for {
		msg, err := protocol.ReadMessage(conn)
		if err != nil {
			return err
		}
		err = msg.PassToClient(client)
		if err != nil {
			return err
		}
	}
}
