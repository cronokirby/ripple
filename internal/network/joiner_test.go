package network

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"
)

type nullReceiver struct{}

func (r nullReceiver) ReceiveContent(_ string) {}

type accReceiver struct {
	msgsLock sync.Mutex
	msgs     []string
}

func (r *accReceiver) ReceiveContent(content string) {
	fmt.Println("Receiving content: ", content)
	r.msgsLock.Lock()
	defer r.msgsLock.Unlock()
	r.msgs = append(r.msgs, content)
}

func joinAndSend(content string, myAddr string, addr string) error {
	remote, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}
	joiner := NewJoiner(&nullReceiver{})
	if err := joiner.Start(myAddr, remote); err != nil {
		fmt.Println(err)
	}
	time.Sleep(25 * time.Millisecond)
	joiner.SendContent(content)
	return nil
}

func TestSwarmIntegration(t *testing.T) {
	acc := &accReceiver{}
	if testing.Short() {
		t.Skip()
	}
	addr1 := "localhost:2001"
	addr2 := "localhost:2002"
	addr3 := "localhost:2003"
	mainJoiner := NewJoiner(acc)
	go mainJoiner.Listen(addr1)
	if err := joinAndSend("Hewwo 1", addr2, addr1); err != nil {
		t.Fatalf("Error joining swarm: %v", err)
	}
	if err := joinAndSend("Hello 2", addr3, addr2); err != nil {
		t.Fatalf("Error joining swarm: %v", err)
	}
	expected := make(map[string]bool)
	time.Sleep(50 * time.Millisecond)
	for _, val := range acc.msgs {
		fmt.Println(val)
		expected[val] = true
	}
	if !expected["Hewwo 1"] || !expected["Hello 2"] {
		t.Fatalf("Messages not receieved, got: %v", expected)
	}
}
