package protocol

import (
	"bytes"
	"fmt"
	"net"
	"reflect"
	"testing"
)

func TestPingMessageBytes(t *testing.T) {
	var p Ping
	result := p.MessageBytes()
	expected := []byte{1}
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %v got %v", expected, result)
	}
}

func TestJoinRequestMessageBytes(t *testing.T) {
	var r JoinRequest
	result := r.MessageBytes()
	expected := []byte{2}
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %v got %v", expected, result)
	}
}

func TestJoinResponseMessageBytes(t *testing.T) {
	r := JoinResponse{
		Peers: []net.Addr{&net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: 99}},
	}
	result := r.MessageBytes()
	expected := []byte{3, 0, 0, 0, 1, 10}
	expected = append(expected, []byte("0.0.0.0:99")...)
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %v got %v", expected, result)
	}
}

func TestNewMessageMessageBytes(t *testing.T) {
	r := NewMessage{Content: "AAA"}
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
	expected := Ping{}
	if msg.(Ping) != expected {
		t.Errorf("Expected %v got %v", expected, msg)
	}
}

func TestJoinRequestParsing(t *testing.T) {
	msg, err := ReadMessage(bytes.NewReader([]byte{2}))
	if err != nil {
		t.Fatalf("Parsing failed: %v", err)
	}
	expected := JoinRequest{}
	if msg.(JoinRequest) != expected {
		t.Errorf("Expected %v got %v", expected, msg)
	}
}

func TestJoinResponseRoundTrip(t *testing.T) {
	r := JoinResponse{
		Peers: []net.Addr{
			&net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: 90},
			&net.TCPAddr{IP: net.ParseIP("155.10.29.128"), Port: 8000},
			&net.TCPAddr{IP: net.ParseIP("203.123.2.34"), Port: 1227},
			&net.TCPAddr{IP: net.ParseIP("111.2.4.45"), Port: 1234},
		},
	}
	fmt.Println(len(r.MessageBytes()))
	roundTrip, err := ReadMessage(bytes.NewReader(r.MessageBytes()))
	if err != nil {
		t.Fatalf("Parsing failed: %v", err)
	}
	if !reflect.DeepEqual(roundTrip.(JoinResponse), r) {
		t.Errorf("Expected %v got %v", r, roundTrip)
	}
}

func TestJoinResponseRoundTripBig(t *testing.T) {
	ip := net.ParseIP("101.101.101.101")
	peers := make([]net.Addr, 100)
	for i := 0; i < 100; i++ {
		peers[i] = &net.TCPAddr{IP: ip, Port: 400}
	}
	r := JoinResponse{peers}
	roundTrip, err := ReadMessage(bytes.NewReader(r.MessageBytes()))
	if err != nil {
		t.Fatalf("Parsing failed: %v", err)
	}
	if !reflect.DeepEqual(roundTrip.(JoinResponse), r) {
		t.Errorf("Expected %v got %v", r, roundTrip)
	}
}

func TestNewMessageRoundTrip(t *testing.T) {
	r := NewMessage{Content: "Hello World!"}
	roundTrip, err := ReadMessage(bytes.NewReader(r.MessageBytes()))
	if err != nil {
		t.Fatalf("Parsing failed: %v", err)
	}
	if !reflect.DeepEqual(roundTrip.(NewMessage), r) {
		t.Errorf("Expected %v got %v", r, roundTrip)
	}
}
