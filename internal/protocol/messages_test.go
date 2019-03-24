package protocol

import (
	"bytes"
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

func TestJoinSwarmMessageBytes(t *testing.T) {
	r := JoinSwarm{
		Addr: &net.TCPAddr{IP: net.ParseIP("100.0.0.0"), Port: 2002},
	}
	addrString := "100.0.0.0:2002"
	expected := []byte{2, byte(len(addrString))}
	expected = append(expected, []byte(addrString)...)
	result := r.MessageBytes()
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %v got %v", expected, result)
	}
}

func TestJoinSwarmRoundTrip(t *testing.T) {
	r := JoinSwarm{
		Addr: &net.TCPAddr{IP: net.ParseIP("128.125.44.20"), Port: 8008},
	}
	expected, err := ReadMessage(bytes.NewReader(r.MessageBytes()))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(r, expected) {
		t.Errorf("Expected %v got %v", expected, r)
	}
}

func TestReferralMessageBytes(t *testing.T) {
	r := Referral{
		Addr: &net.TCPAddr{IP: net.ParseIP("100.0.0.0"), Port: 2002},
	}
	addrString := "100.0.0.0:2002"
	expected := []byte{3, byte(len(addrString))}
	expected = append(expected, []byte(addrString)...)
	result := r.MessageBytes()
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %v got %v", expected, result)
	}
}

func TestReferralRoundTrip(t *testing.T) {
	r := Referral{
		Addr: &net.TCPAddr{IP: net.ParseIP("128.125.44.20"), Port: 8008},
	}
	expected, err := ReadMessage(bytes.NewReader(r.MessageBytes()))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(r, expected) {
		t.Errorf("Expected %v got %v", expected, r)
	}
}

func TestNewPredecessorMessageBytes(t *testing.T) {
	r := NewPredecessor{
		Addr: &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: 99},
	}
	addrString := "0.0.0.0:99"
	expected := []byte{4, byte(len(addrString))}
	expected = append(expected, []byte(addrString)...)
	result := r.MessageBytes()
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %v got %v", expected, result)
	}
}

func TestNewPredecessorRoundTrip(t *testing.T) {
	r := NewPredecessor{
		Addr: &net.TCPAddr{IP: net.ParseIP("128.125.44.20"), Port: 8008},
	}
	expected, err := ReadMessage(bytes.NewReader(r.MessageBytes()))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(r, expected) {
		t.Errorf("Expected %v got %v", expected, r)
	}
}

func TestConfirmPredecessorMessageBytes(t *testing.T) {
	r := ConfirmPredecessor{}
	expected := []byte{5}
	result := r.MessageBytes()
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %v got %v", expected, result)
	}
}

func TestConfirmReferralMessageBytes(t *testing.T) {
	r := ConfirmReferral{}
	expected := []byte{6}
	result := r.MessageBytes()
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %v got %v", expected, result)
	}
}

func TestNewMessageMessageBytes(t *testing.T) {
	r := NewMessage{
		Sender:  &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234},
		Content: "Hello World!",
	}
	addrString := "127.0.0.1:1234"
	content := "Hello World!"
	expected := []byte{7, byte(len(addrString))}
	expected = append(expected, []byte(addrString)...)
	contentLen := len(content)
	expected = append(
		expected,
		byte(contentLen>>24),
		byte(contentLen>>16),
		byte(contentLen>>8),
		byte(contentLen),
	)
	expected = append(expected, []byte(content)...)
	result := r.MessageBytes()
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %v got %v", expected, result)
	}
}

func TestNewMessageRoundTrip(t *testing.T) {
	r := NewMessage{
		Sender:  &net.TCPAddr{IP: net.ParseIP("127.0.120.1"), Port: 8090},
		Content: "Round Trip!",
	}
	expected, err := ReadMessage(bytes.NewReader(r.MessageBytes()))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(r, expected) {
		t.Errorf("Expected %v got %v", expected, r)
	}
}
