package protocol

import (
	"bytes"
	"net"
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
	r := JoinSwarm{}
	expected := []byte{2}
	result := r.MessageBytes()
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected %v got %v", expected, result)
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

func TestConfirmPredecessorMessageBytes(t *testing.T) {
	r := ConfirmPredecessor{}
	expected := []byte{5}
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
	expected := []byte{6, byte(len(addrString))}
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
