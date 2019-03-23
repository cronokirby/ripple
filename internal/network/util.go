package network

import (
	"io"
	"net"
	"sync"

	"github.com/cronokirby/ripple/internal/protocol"
)

// syncConn lets us manage access to an arriving peer
//
// the inner conn is safe to read, but not to modify
type syncConn struct {
	conn    net.Conn
	muEmpty sync.Mutex
	muFill  sync.Mutex
}

// makeSyncConn constructs a valid syncConn in the empty state
func makeSyncConn() *syncConn {
	conn := &syncConn{}
	conn.muEmpty.Lock()
	return conn
}

// String formats a syncConn by printing the undlerying connection
func (conn *syncConn) String() string {
	return conn.conn.RemoteAddr().String()
}

// empty will block until the connection is full
func (conn *syncConn) empty() {
	conn.muEmpty.Lock()
	conn.conn = nil
	conn.muFill.Unlock()
}

// fill will block until the connection is empty
func (conn *syncConn) fill(newConn net.Conn) {
	conn.muFill.Lock()
	conn.conn = newConn
	conn.muEmpty.Unlock()
}

// isEmpty checks if a syncConn contains nothing
func (conn *syncConn) isEmpty() bool {
	return conn.conn == nil
}

func sendMessage(w io.Writer, msg protocol.Message) error {
	//time.Sleep(1000 * time.Millisecond)
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
