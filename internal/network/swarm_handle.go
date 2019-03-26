package network

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/cronokirby/ripple/internal/protocol"
)

// sameAddr checks if 2 nodes are the same, by string equality
func sameAddr(a net.Addr, b net.Addr) bool {
	return a.String() == b.String()
}

// clientState holds the state a client needs in normal operation
type clientState struct {
	mu sync.Mutex
	// me is the address of this client
	me net.Addr
	// newPred is the address of a node trying to become our Predecessor
	newPred net.Addr
	// pred is our Predecessor node
	pred peer
	// succ is our Successor node
	succ peer
	// latestSuccAddr is not nil once we know the latest is trying to become our Succ
	latestSuccAddr net.Addr
	// latestPredAddr is not nil once we know the latest is trying to become our Succ
	latestPredAddr net.Addr
}

// normalClient contains the information needed in normal operation
type normalClient struct {
	log *log.Logger
	// broadcaster lets us print the text messages
	receiver protocol.ContentReceiver
	// state represents the mutable state under a single lock
	state *clientState
	// pool holds the connection pool for our peers
	pool *peerPool
	// latest has its own locking mechanism
	latest *syncConn
}

func (client *normalClient) listenLoop(l net.Listener) {
	if l == nil {
		newL, err := net.Listen("tcp", client.state.me.String())
		if err != nil {
			log.Fatalln("Couldn't start listener ", err)
		}
		l = newL
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalln("Error accepting conn ", err)
		}
		client.latest.fill(conn)
		msg, err := protocol.ReadMessage(conn)
		if err != nil {
			log.Println("Error reading message ", err)
			conn.Close()
		}
		wrappedClient := client.withOrigin(newRole)
		if err := msg.PassToClient(wrappedClient); err != nil {
			log.Println(err)
			conn.Close()
		}
	}
}

func (client *normalClient) messageLoop() {
	for {
		select {
		case oMsg := <-client.pool.messages:
			wrappedClient := client.withOrigin(oMsg.origin)
			if err := oMsg.msg.PassToClient(wrappedClient); err != nil {
				log.Println(err)
			}
		case err := <-client.pool.errors:
			client.log.Println(err)
		}
	}
}

// originClient wraps a client with the origin of a message
type originClient struct {
	// origin holds the source of the message
	origin   int
	log      *log.Logger
	receiver protocol.ContentReceiver
	state    *clientState
	// pool is the underlying peer pool
	pool *peerPool
	// latest is the connection trying to join us
	latest *syncConn
}

// withOrigin embellishes a client with an origin
func (client *normalClient) withOrigin(origin int) *originClient {
	return &originClient{origin, client.log, client.receiver, client.state, client.pool, client.latest}
}

// fmtOrigin is mainly useful for debugging purposes
func (client *originClient) fmtOrigin() string {
	return originString(client.origin)
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
func (client *originClient) HandleJoinSwarm(msg protocol.JoinSwarm) error {
	if client.origin != newRole {
		return fmt.Errorf(
			"Unexpected JoinSwarm message %s",
			client.fmtOrigin(),
		)
	}
	client.state.mu.Lock()
	defer client.state.mu.Unlock()
	client.state.latestSuccAddr = msg.Addr
	referral := protocol.Referral{Addr: client.state.succ.addr}
	if err := sendMessage(client.latest.conn, referral); err != nil {
		return err
	}
	newPred := protocol.NewPredecessor{Addr: msg.Addr}
	if err := sendMessage(client.state.succ.conn, newPred); err != nil {
		return err
	}
	return nil
}

// HandleReferral is always ignored, because we're not joining a swarm
func (client *originClient) HandleReferral(msg protocol.Referral) error {
	return fmt.Errorf(
		"Unexpected Referral message %v %s",
		msg,
		client.fmtOrigin(),
	)
}

func (client *originClient) clearLatest() {
	client.latest.empty()
	client.state.latestPredAddr = nil
	client.state.latestSuccAddr = nil
}

func (client *originClient) swapPredecessorsIfReady() error {
	if !client.latest.isEmpty() && client.state.latestPredAddr != nil && client.state.newPred != nil {
		latestAddr := client.state.latestPredAddr
		announceAddr := client.state.newPred
		if !sameAddr(latestAddr, announceAddr) {
			return fmt.Errorf(
				"Mismatched Predecessors; announced: %v; connected: %v",
				announceAddr,
				latestAddr,
			)
		}
		confirm := protocol.ConfirmReferral{}
		if err := sendMessage(client.state.pred.conn, confirm); err != nil {
			return err
		}
		client.pool.remove(client.state.pred, true)
		client.state.pred = peer{
			addr: client.state.latestPredAddr,
			conn: client.latest.conn,
		}
		client.pool.submit(client.state.pred, true)
		client.clearLatest()
	}
	return nil
}

// HandleNewPredecessor is handled from our Predecessor
//
// If we have receieved a ConfirmPredecessor already, we can finalise
// the replacement of our Predecessor.
func (client *originClient) HandleNewPredecessor(msg protocol.NewPredecessor) error {
	if !isPredRole(client.origin) {
		return fmt.Errorf(
			"Unexpected NewPredecessor message %v %s",
			msg,
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
func (client *originClient) HandleConfirmPredecessor(msg protocol.ConfirmPredecessor) error {
	if client.origin != newRole {
		return fmt.Errorf(
			"Unexpected ConfirmPredecessor message %s",
			client.fmtOrigin(),
		)
	}
	client.state.mu.Lock()
	defer client.state.mu.Unlock()
	client.state.latestPredAddr = msg.Addr
	return client.swapPredecessorsIfReady()
}

// HandleConfirmReferral allows us to replace our Successor
func (client *originClient) HandleConfirmReferral() error {
	if !isSuccRole(client.origin) || client.state.latestSuccAddr == nil {
		return fmt.Errorf(
			"Unexpected ConfirmReferral message %s",
			client.fmtOrigin(),
		)
	}
	client.state.mu.Lock()
	defer client.state.mu.Unlock()
	client.pool.remove(client.state.succ, false)
	client.state.succ = peer{
		addr: client.state.latestSuccAddr,
		conn: client.latest.conn,
	}
	client.pool.submit(client.state.succ, false)
	client.clearLatest()
	return nil
}

// HandleNewMessage allows us to handle text messages
func (client *originClient) HandleNewMessage(msg protocol.NewMessage) error {
	if !isPredRole(client.origin) {
		return fmt.Errorf(
			"Unexpected NewMessage %v %s",
			msg,
			client.fmtOrigin(),
		)
	}
	if sameAddr(client.state.me, msg.Sender) {
		return nil
	}
	client.receiver.ReceiveContent(msg.Content)
	return sendMessage(client.state.succ.conn, msg)
}

// joiningClient is a client trying to join a swarm
type joiningClient struct {
	referral net.Addr
}

func (client *joiningClient) HandlePing() error {
	return errors.New("Unexpected Ping message")
}

func (client *joiningClient) HandleJoinSwarm(msg protocol.JoinSwarm) error {
	return fmt.Errorf("Unexpected JoinSwarm message: %v", msg)
}

func (client *joiningClient) HandleReferral(msg protocol.Referral) error {
	client.referral = msg.Addr
	return nil
}

func (client *joiningClient) HandleNewPredecessor(msg protocol.NewPredecessor) error {
	return fmt.Errorf("Unexpected NewPredecessor message: %v", msg)
}

func (client *joiningClient) HandleConfirmPredecessor(msg protocol.ConfirmPredecessor) error {
	return fmt.Errorf("Unexpected ConfirmPredecessor message: %v", msg)
}

func (client *joiningClient) HandleConfirmReferral() error {
	return errors.New("Unexpected ConfirmReferral message")
}

func (client *joiningClient) HandleNewMessage(msg protocol.NewMessage) error {
	return fmt.Errorf("Unexpected NewMessage: %v", msg)
}

// joinSwarm can't and won't complete the logging and receiever fields of client
func (client *joiningClient) joinSwarm(log *log.Logger, start, me net.Addr) (*normalClient, error) {
	predConn, err := net.Dial(start.Network(), start.String())
	if err != nil {
		return nil, err
	}
	if err := sendMessage(predConn, protocol.JoinSwarm{Addr: me}); err != nil {
		return nil, err
	}
	msg, err := protocol.ReadMessage(predConn)
	if err != nil {
		return nil, err
	}
	if err := msg.PassToClient(client); err != nil {
		return nil, err
	}
	succAddr := client.referral
	succConn := predConn
	// this is usually the case
	if !sameAddr(succAddr, start) {
		conn, err := net.Dial(succAddr.Network(), succAddr.String())
		if err != nil {
			return nil, err
		}
		succConn = conn
	}
	confirmPredecessor := protocol.ConfirmPredecessor{Addr: me}
	if err := sendMessage(succConn, confirmPredecessor); err != nil {
		return nil, err
	}
	predPeer := peer{addr: start, conn: predConn}
	succPeer := peer{addr: succAddr, conn: succConn}
	state := &clientState{me: me, pred: predPeer, succ: succPeer}
	normal := &normalClient{
		log:      log,
		receiver: protocol.NilReceiver{},
		pool:     makePeerPool(),
		state:    state,
		latest:   makeSyncConn(),
	}
	normal.log.Println("Starting loops...")
	normal.pool.submit(normal.state.pred, true)
	normal.pool.submit(normal.state.succ, false)
	go normal.listenLoop(nil)
	go normal.messageLoop()
	return normal, nil
}

// lonelyClient is a client starting a new swarm, with no peers
//
// we need to treat this case slightly differently from a normalClient.
// The case with just 2 peers should "just work" with the normalClient code.
type lonelyClient struct {
	// me is the address of this client
	me net.Addr
	// nil indicates no joinSwarm message yet
	firstAddr net.Addr
	// first starts off nil, and becomes filled as we try and get our first peer
	first net.Conn
	log   *log.Logger
}

// HandlePing is unexpected
func (client *lonelyClient) HandlePing() error {
	return fmt.Errorf("Unexpected Ping in lonelyClient")
}

// HandleJoinSwarm allows us to start accepting our first peer
//
// If we've already receieved this once though, we can't continue
func (client *lonelyClient) HandleJoinSwarm(msg protocol.JoinSwarm) error {
	if client.firstAddr != nil {
		return fmt.Errorf("Unexpected JoinSwarm in lonelyClient (already received)")
	}
	referral := protocol.Referral{Addr: client.me}
	if err := sendMessage(client.first, referral); err != nil {
		return err
	}
	client.firstAddr = msg.Addr
	return nil
}

// HandleReferral isn't expected at this point
func (client *lonelyClient) HandleReferral(protocol.Referral) error {
	return fmt.Errorf("Unexpected Referral in lonelyClient")
}

// HandleNewPredecessor is unexpected at this point
func (client *lonelyClient) HandleNewPredecessor(protocol.NewPredecessor) error {
	return fmt.Errorf("Unexpected NewPredecessor in lonelyClient")
}

// HandleConfirmPredecessor allows us to continue and finish accepting our first peer
//
// We must have already received a JoinSwarm message to be able to accept this
func (client *lonelyClient) HandleConfirmPredecessor(msg protocol.ConfirmPredecessor) error {
	if client.firstAddr == nil {
		return fmt.Errorf("Unexpected ConfirmPredecessor in lonelyClient (no joinSwarm)")
	}
	if !sameAddr(client.firstAddr, msg.Addr) {
		return fmt.Errorf(
			"Mismatched peer addresses. joined: %v confirmed %v",
			client.firstAddr,
			msg.Addr,
		)
	}
	return nil
}

// ConfirmReferral is unexpected at this time
func (client *lonelyClient) HandleConfirmReferral() error {
	return fmt.Errorf("Unexpected ConfirmReferral in lonelyClient")
}

// HandleNewMessage is unexpected at this time
func (client *lonelyClient) HandleNewMessage(protocol.NewMessage) error {
	return fmt.Errorf("Unexpected NewMessage in lonelyClient")
}

func (client *lonelyClient) receiveMsg(conn net.Conn) error {
	msg, err := protocol.ReadMessage(conn)
	if err != nil {
		return err
	}
	if err := msg.PassToClient(client); err != nil {
		return err
	}
	return nil
}

// startSwarm starts a new swarm
//
// make sure to reuse the listener we set in lonelyClient after this though
func (client *lonelyClient) startSwarm() (*normalClient, error) {
	l, err := net.Listen(client.me.Network(), client.me.String())
	if err != nil {
		return nil, err
	}
	for client.first == nil {
		client.firstAddr = nil
		conn, err := l.Accept()
		if err != nil {
			client.log.Println(err)
			continue
		}
		client.first = conn
		if err := client.receiveMsg(conn); err != nil {
			client.log.Println(err)
			client.first = nil
			continue
		}
		sendMessage(conn, protocol.Referral{Addr: client.me})
		if err := client.receiveMsg(conn); err != nil {
			client.log.Println(err)
			client.first = nil
			continue
		}
	}
	peer := peer{addr: client.firstAddr, conn: client.first}
	state := &clientState{me: client.me, pred: peer, succ: peer}
	normal := &normalClient{
		log:      client.log,
		receiver: protocol.NilReceiver{},
		pool:     makePeerPool(),
		state:    state,
		latest:   makeSyncConn(),
	}
	normal.log.Println("Starting loops...")
	normal.pool.submit(peer, true)
	normal.pool.submit(peer, false)
	go normal.listenLoop(l)
	go normal.messageLoop()
	return normal, nil
}

// SwarmHandle allows us to interact with a swarm
//
// The main ways of creating one are to join an existing one, or create
// a new swarm by acting as the first node.
type SwarmHandle struct {
	client *normalClient
}

// JoinSwarm creates a new SwarmHandle by joining an existing swarm
//
// It takes a node to enter the swarm with, and an address to listen on
// after joining.
func JoinSwarm(log *log.Logger, you, start net.Addr) (*SwarmHandle, error) {
	joining := &joiningClient{}
	normal, err := joining.joinSwarm(log, start, you)
	if err != nil {
		return nil, err
	}
	return &SwarmHandle{normal}, nil
}

// CreateSwarm starts a new swarm by listening at an address
//
// This will block until the first peer joins the swarm.
func CreateSwarm(log *log.Logger, you net.Addr) (*SwarmHandle, error) {
	lonely := &lonelyClient{me: you, log: log}
	normal, err := lonely.startSwarm()
	if err != nil {
		return nil, err
	}
	return &SwarmHandle{normal}, nil
}

// SetReceiver changes the receiever in a swarm handle to do something useful
func (swarm *SwarmHandle) SetReceiver(receiver protocol.ContentReceiver) {
	swarm.client.receiver = receiver
}

// SendContent allows us to send a piece of text to the rest of the swarm
func (swarm *SwarmHandle) SendContent(content string) {
	msg := protocol.NewMessage{Sender: swarm.client.state.me, Content: content}
	// ignore errors
	sendMessage(swarm.client.state.succ.conn, msg)
}
