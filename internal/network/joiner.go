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
	firstClient := &normalClient{}
	go firstClient.justLoop(conn)
	for _, addr := range j.peers {
		client := &normalClient{}
		go client.connectAndLoop(addr, newConns)
	}
	l, err := net.Listen("tcp", "localhost:3000")
	if err != nil {
		return err
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		peers := j.peers
		client := &acceptingClient{newConns, conn, peers}
		go client.accept()
	}
}

// acceptingClient tries and connect with a newly accepted peer
type acceptingClient struct {
	// newConns is a channel we can use to broadcast the new connection
	// if it's client that behaves correctly, and we enter a normal
	// relationship in it
	newConns chan net.Conn
	conn     net.Conn
	peers    []net.Addr
}

func (client *acceptingClient) HandlePing() error {
	return errors.New("Unexpected Ping in acceptingClient")
}

func (client *acceptingClient) HandleJoinRequest() error {
	resp := protocol.JoinResponse{Peers: client.peers}
	err := sendMessage(client.conn, resp)
	if err != nil {
		return err
	}
	client.newConns <- client.conn
	newClient := &normalClient{}
	return newClient.innerLoop(client.conn)
}

func (client *acceptingClient) HandleJoinResponse(_ protocol.JoinResponse) error {
	return errors.New("Unexpected JoinResponse in acceptingClient")
}

func (client *acceptingClient) HandleNewMessage(_ protocol.NewMessage) error {
	return errors.New("Unexpected NewMessage in acceptingClient")
}

func (client *acceptingClient) accept() error {
	defer client.conn.Close()
	msg, err := protocol.ReadMessage(client.conn)
	if err != nil {
		fmt.Println(err)
		return err
	}
	// this will loop forever in the new state
	err = msg.PassToClient(client)
	fmt.Println(err)
	return err
}

type normalClient struct{}

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
func (client *normalClient) connectAndLoop(addr net.Addr, newConns chan net.Conn) error {
	conn, err := net.Dial(addr.Network(), addr.String())
	if err != nil {
		return err
	}
	defer conn.Close()
	newConns <- conn
	err = client.innerLoop(conn)
	fmt.Println(err)
	return err
}

func (client *normalClient) justLoop(conn net.Conn) error {
	defer conn.Close()
	err := client.innerLoop(conn)
	fmt.Println(err)
	return err
}

// this will not close the connection, and is meant to be called by other things
func (client *normalClient) innerLoop(conn net.Conn) error {
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
