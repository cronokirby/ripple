package protocol

// Client represents a client capable of responding to different messages
// This is basically the visitor pattern to emulate sum types
type Client interface {
	// Handle a ping message from the peer
	HandlePing()
	// Handle a request to join a swarm
	HandleJoinRequest()
	// Handle the response to a join request
	HandleJoinResponse()
	// Handle a new message from a peer
	HandleNewMessage()
}
