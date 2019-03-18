package protocol

import (
	"bytes"
	"fmt"
	"net"
	"reflect"
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

func TestJoinResponeRoundTrip(t *testing.T) {
	r := joinResponse{
		peers: []net.Addr{
			&net.IPAddr{IP: net.ParseIP("0.0.0.0")},
			&net.IPAddr{IP: net.ParseIP("155.10.29.128")},
			&net.IPAddr{IP: net.ParseIP("203.123.2.34")},
			&net.IPAddr{IP: net.ParseIP("111.2.4.45")},
		},
	}
	fmt.Println(len(r.MessageBytes()))
	roundTrip, err := ReadMessage(bytes.NewReader(r.MessageBytes()))
	if err != nil {
		t.Errorf("Parsing failed: %v", err)
		return
	}
	if !reflect.DeepEqual(roundTrip.(joinResponse), r) {
		t.Errorf("Expected %v got %v", r, roundTrip)
	}
}
