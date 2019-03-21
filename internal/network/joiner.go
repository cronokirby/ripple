package network

import (
	"errors"
	"fmt"
	"net"

	"github.com/cronokirby/ripple/internal/protocol"
)

// Joiner represents a client that tries and join a new swarm
type Joiner struct {
	peers       *peerList
	broadcaster *broadcaster
	receiver    protocol.ContentReceiver
}

// NewJoiner creates a valid joiner, since the zero value isn't
func NewJoiner(receiever protocol.ContentReceiver) *Joiner {
	return &Joiner{peers: &peerList{}, broadcaster: makeBroadcaster(), receiver: receiever}
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
	j.peers.addPeers(resp.Peers...)
	return nil
}

// HandleNewMessage exits the connection, because we only expect JoinResponse
func (j *Joiner) HandleNewMessage(msg protocol.NewMessage) error {
	return errors.New("Unexpected NewMessage when joining swarm")
}

// Start blocks the current thread to run a joiner given a certain address,
// which it will use to enter the swarm
// This function will exit if it encounters an error before entering the swarm.
// After entering the swarm, it will become self sufficient, and restart
// the different parts it manages if necessary
//
// myAddr is the address to listen to after first connecting to remoteAddr
func (j *Joiner) Start(myAddr string, remoteAddr net.Addr) error {
	conn, err := net.Dial(remoteAddr.Network(), remoteAddr.String())
	if err != nil {
		return err
	}
	// We can't defer the conn closing because
	err = sendMessage(conn, protocol.JoinRequest{})
	if err != nil {
		conn.Close()
		return err
	}
	msg, err := protocol.ReadMessage(conn)
	if err != nil {
		conn.Close()
		return err
	}
	err = msg.PassToClient(j)
	if err != nil {
		conn.Close()
		return err
	}
	go j.broadcaster.loop()
	go func() {
		// Connect to every peer we've been given
		j.broadcaster.sendConn(conn)
		firstClient := &normalClient{receiver: j.receiver}
		go firstClient.justLoop(conn)
		for _, addr := range j.peers.getPeers() {
			client := &normalClient{}
			go client.connectAndLoop(addr, j.broadcaster)
		}
		// We add the first peer now, not wanting to loop over it
		j.peers.addPeers(conn.RemoteAddr())
		for {
			err = j.innerListen(myAddr)
			fmt.Println(err)
		}
	}()
	return nil
}

func (j *Joiner) innerListen(myAddr string) error {
	l, err := net.Listen("tcp", myAddr)
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
		receiver := j.receiver
		client := &acceptingClient{j.broadcaster, conn, peers, receiver}
		go client.accept()
	}
}

// Listen won't join an existing swarm, but create its own swarm
func (j *Joiner) Listen(myAddr string) error {
	go j.broadcaster.loop()
	return j.innerListen(myAddr)
}

// SendContent sends a new piece of a content the swarm we're connected to
//
// This function will block briefly, but not until the message is completely
// transmitted.
func (j *Joiner) SendContent(content string) {
	j.broadcaster.sendContent(content)
}

// acceptingClient tries and connect with a newly accepted peer
type acceptingClient struct {
	// newConns is a channel we can use to broadcast the new connection
	// if it's client that behaves correctly, and we enter a normal
	// relationship in it
	broadcaster *broadcaster
	conn        net.Conn
	peers       *peerList
	receiever   protocol.ContentReceiver
}

// normalise makes the relationship into a normal client one
func (client *acceptingClient) normalise() error {
	client.broadcaster.sendConn(client.conn)
	client.peers.addPeers(client.conn.RemoteAddr())
	newClient := &normalClient{receiver: client.receiever}
	return newClient.innerLoop(client.conn)
}

func (client *acceptingClient) HandlePing() error {
	return errors.New("Unexpected Ping in acceptingClient")
}

func (client *acceptingClient) HandleJoinRequest() error {
	resp := protocol.JoinResponse{Peers: client.peers.getPeers()}
	err := sendMessage(client.conn, resp)
	if err != nil {
		return err
	}
	return client.normalise()
}

func (client *acceptingClient) HandleJoinResponse(_ protocol.JoinResponse) error {
	return errors.New("Unexpected JoinResponse in acceptingClient")
}

func (client *acceptingClient) HandleNewMessage(msg protocol.NewMessage) error {
	// In this case, the client doesn't want to join the swarm, just connect
	client.receiever.ReceiveContent(msg.Content)
	return client.normalise()
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

type normalClient struct {
	receiver protocol.ContentReceiver
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
	client.receiver.ReceiveContent(msg.Content)
	return nil
}

// connectAndLoop starts a normal client with a new address to connect to,
// and returns an error whenever something fatal occurrs
func (client *normalClient) connectAndLoop(addr net.Addr, bc *broadcaster) error {
	conn, err := net.Dial(addr.Network(), addr.String())
	if err != nil {
		return err
	}
	defer conn.Close()
	bc.sendConn(conn)
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
