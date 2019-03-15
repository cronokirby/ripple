package protocol

// Ping represents a Ping message, that has to be sent from time to time, in
// order to keep a connection between 2 peers alive
type Ping struct{}

// JoinRequest represents a request from one peer to join the swarm
// chat the other peer is in. The peer receiving this message should
// respond with its own JoinResponse
type JoinRequest struct{}
