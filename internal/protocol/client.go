package protocol

// Client represents a client capable of responding to different messages
// This is basically the visitor pattern to emulate sum types
type Client interface {
	// Handle a ping message from the peer
	HandlePing() error
	// Handle a JoinSwarm message
	HandleJoinSwarm(JoinSwarm) error
	// Handle a Referral message
	HandleReferral(Referral) error
	// Handle a NewPredecessor message
	HandleNewPredecessor(NewPredecessor) error
	// Handle a ConfirmPredecessor message
	HandleConfirmPredecessor(ConfirmPredecessor) error
	// Handle a ConfirmReferral message
	HandleConfirmReferral() error
	// Handle a NewMessage message
	HandleNewMessage(NewMessage) error
}
