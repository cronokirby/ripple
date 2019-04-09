*version 0.0*
# Message Format
This doc provides the binary specification of the messages used in the ripple protocol.

All integers are in network order.

For encoding network addresses, we sacrifice some compactness for convenience
by encoding them as strings. This allows us to have a uniform encoding
for both IPv4 and IPv6 addresses. We only use a byte to encode the length
of the string for each address, which is sufficient for the standard encoding
of both types. The string for each of the address should be sufficient to 
contact the peer, i.e. should include a port as well.

## Ping
The Ping message contains no information, so it only has a type tag.

| Field | Length | Description   |
| ----- | ------ | ------------- |
| Type  | 1      | 0x01 for Ping |
| **Total** | **1** ||

## JoinSwarm
| Field | Length | Description          |
| ----- | ------ | -------------------- |
| Type  | 1      | 0x02 for JoinSwarm |
| Length    | 1      | Unsigned byte, how long the following field is  |
| Addr      | Length | A UTF-8 string containing the address of this node |

## Referral
Referral contains the address of the node to send a **ConfirmPredecessor** to

| Field     | Length | Description           |
| --------- | ------ | --------------------- |
| Type      | 1      | 0x03 for Referral     |
| Length    | 1      | Unsigned byte, how long the following field is  |
| Addr      | Length | A UTF-8 string containing the address of a node |

## NewPredecessor
| Field     | Length | Description             |
| --------- | ------ | ----------------------- |
| Type      | 1      | 0x04 for NewPredecessor |
| Length    | 1      | Unsigned byte, how long the following field is  |
| Addr      | Length | A UTF-8 string containing the address of a node |

## ConfirmPredecessor
| Field     | Length | Description             |
| --------- | ------ | ----------------------- |
| Type      | 1      | 0x05 for ConfirmPredecessor |
| Length    | 1      | Unsigned byte, how long the following field is  |
| Addr      | Length | A UTF-8 string containing the address of this node |

## ConfirmReferral
| Field     | Length | Description             |
| --------- | ------ | ----------------------- |
| Type      | 1      | 0x06 for ConfirmReferral |

## NewMessage
| Field      | Length | Description           |
| ---------- | ------ | --------------------- |
| Type       | 1      | 0x07 for NewMessage   |
| AddrLength | 1      | Unsigned byte, how long the following field is |
| Addr       | AddrLength | The address for this node |
| Length     | 4      | Unsigned 32 bit integer, length of following field |
| Content    | Length | UTF-8 string with message content |

## Nickname
| Field      | Length | Description           |
| ---------- | ------ | --------------------- |
| Type       | 1      | 0x08 for Nickname   |
| AddrLength | 1      | Unsigned byte, how long the following field is |
| Addr       | AddrLength | The address for this node |
| Length     | 4      | Unsigned 32 bit integer, length of following field |
| Name    | Length | UTF-8 string with the new name |
