package network

import (
	"net"
	"sync"
)

// nickMap provides a concurrent store over nicknames
type nickMap struct {
	// nicks is the underlying storage net.Addr -> string
	nicks sync.Map
}

func makeNickMap() *nickMap {
	return &nickMap{}
}

// set will set the nickname for a given node
func (nmap *nickMap) set(node net.Addr, name string) {
	nmap.nicks.Store(node.String(), name)
}

// get will return the address of the node as a string if no nick is present
func (nmap *nickMap) get(node net.Addr) string {
	res, ok := nmap.nicks.Load(node.String())
	if !ok {
		return node.String()
	}
	// we know this is safe because we control storage
	return res.(string)
}
