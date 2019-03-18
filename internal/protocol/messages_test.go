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

func TestPingParsing(t *testing.T) {
	msg, err := ReadMessage(bytes.NewReader([]byte{1}))
	if err != nil {
		t.Fatalf("Parsing failed: %v", err)
	}
	expected := ping{}
	if msg.(ping) != expected {
		t.Errorf("Expected %v got %v", expected, msg)
	}
}

func TestJoinRequestParsing(t *testing.T) {
	msg, err := ReadMessage(bytes.NewReader([]byte{2}))
	if err != nil {
		t.Fatalf("Parsing failed: %v", err)
	}
	expected := joinRequest{}
	if msg.(joinRequest) != expected {
		t.Errorf("Expected %v got %v", expected, msg)
	}
}

func TestJoinResponseRoundTrip(t *testing.T) {
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
		t.Fatalf("Parsing failed: %v", err)
	}
	if !reflect.DeepEqual(roundTrip.(joinResponse), r) {
		t.Errorf("Expected %v got %v", r, roundTrip)
	}
}

func TestJoinResponseRoundTripBig(t *testing.T) {
	ip := net.ParseIP("101.101.101.101")
	peers := make([]net.Addr, 100)
	for i := 0; i < 100; i++ {
		peers[i] = &net.IPAddr{IP: ip}
	}
	r := joinResponse{peers}
	roundTrip, err := ReadMessage(bytes.NewReader(r.MessageBytes()))
	if err != nil {
		t.Fatalf("Parsing failed: %v", err)
	}
	if !reflect.DeepEqual(roundTrip.(joinResponse), r) {
		t.Errorf("Expected %v got %v", r, roundTrip)
	}
}

func TestNewMessageRoundTrip(t *testing.T) {
	r := newMessage{content: "Hello World!"}
	roundTrip, err := ReadMessage(bytes.NewReader(r.MessageBytes()))
	if err != nil {
		t.Fatalf("Parsing failed: %v", err)
	}
	if !reflect.DeepEqual(roundTrip.(newMessage), r) {
		t.Errorf("Expected %v got %v", r, roundTrip)
	}
}
