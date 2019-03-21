## Sending a message
Each node has a Sucessor, and Predecessor. To send a text-message,
a node sends a **NewMessage** message to its Sucessor.
When a node receieves a **NewMessage** from its Predecessor that it didn't
create, it forwards it to its Sucessor. This means that a message will eventually round-trip back to its sender, closing the loop.

## Connecting to a swarm
Connecting to a swarm happens in 5 steps:

1) A peer wanting to a join a swarm starts with a single client,
and sends them a **JoinSwarm** message.
2) After receiving that message, the node sends a **NewPredecessor** message
to its Successor node, and replies to the peer joining the swarm with a 
**Referral** message, directing it to that Successor node.
3) Using the node given in the **Referral** message, the joining peer
sends a **ConfirmPredecessor** message to that node.
4) After the Successor node has receieved both a **NewPredecessor**
message mentioning a peer, and a **ConfirmPredecessor** message from the same peer, it sends a **ConfirmReferral** message back to its Predecessor, and the joining peer replaces its Predecessor.
5) After receiving the **ConfirmReferral** message, the first node now replaces its Sucessor with the joining peer

The peer has now joined the swarm, and operations can happen normally.