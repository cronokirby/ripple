*version 0.0*
# Message Format
This doc provides the binary specification of the messages used in the ripple protocol.

## Ping
The Ping message contains no information, so it only has a type tag.

| Field | Length | Description   |
| ----- | ------ | ------------- |
| Type  | 1      | 0x01 for Ping |
| **Total** | **1** ||

## JoinRequest
Like Ping, the join request just has a type tag

| Field | Length | Description          |
| ----- | ------ | -------------------- |
| Type  | 1      | 0x02 for JoinRequest |

## JoinResponse
For encoding network addresses, we sacrifice some compactness for convenience
by encoding them as strings. This allows us to have a uniform encoding
for both IPv4 and IPv6 addresses. We only use a byte to encode the length
of the string for each address, which is sufficient for the standard encoding
of both types. The string for each of the address should be sufficient to 
contact the peer, i.e. should include a port as well.

| Field     | Length | Description           |
| --------- | ------ | --------------------- |
| Type      | 1      | 0x03 for JoinResponse |
| PeerCount | 4      | Unsigned 32 bit integer, denoting the number of peers |
| Length i  | 1      | Unsigned byte, how long the following field is |
| Addr i    | Length i | A UTF-8 string containing the ith peer's address |

## NewMessage
Once again we prefix the message string with its length

| Field     | Length | Description           |
| --------- | ------ | --------------------- |
| Type      | 1      | 0x04 for NewMessage      |
| Length    | 4      | Unsigned 32 bit integer, length of following field |
| Content   | Length | UTF-8 string with message content |