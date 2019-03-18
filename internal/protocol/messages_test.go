package protocol

import (
	"bytes"
	"net"
	"testing"
)

func TestPingMessageBytes(t *testing.T) {
	var p ping
	result := p.MessageBytes()
	expected := []byte{1}
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %v got %v", expected, result)
	}
}

func TestJoinRequestMessageBytes(t *testing.T) {
	var r joinRequest
	result := r.MessageBytes()
	expected := []byte{2}
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %v got %v", expected, result)
	}
}

func TestJoinResponseMessageBytes(t *testing.T) {
	r := joinResponse{
		peers: []net.Addr{&net.IPAddr{IP: net.ParseIP("0.0.0.0")}},
	}
	result := r.MessageBytes()
	expected := []byte{3, 0, 0, 0, 1, 7, 48, 46, 48, 46, 48, 46, 48}
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %v got %v", expected, result)
	}
}

func TestNewMessageMessageBytes(t *testing.T) {
	r := newMessage{content: "AAA"}
	expected := []byte{4, 0, 0, 0, 3, 65, 65, 65}
	result := r.MessageBytes()
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %v got %v", expected, result)
	}
}
