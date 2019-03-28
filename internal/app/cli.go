package app

import (
	"bufio"
	"fmt"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/cronokirby/ripple/internal/network"
)

var (
	// App provides the starting point for command parsing
	App = kingpin.New("ripple", "A decentralized chat application")

	// Start handles the start command
	Start = App.Command("start", "Start a new swarm")
	// StartAddr is the address we need to start the swarm on
	StartAddr = Start.Arg("addr", "The address to listen on").Required().String()

	// Connect is the command for joining an existing swarm
	Connect = App.Command("connect", "Connect to an existing swarm")
	// ConnectListenAddr is the address on which to listen after joining
	ConnectListenAddr = Connect.Arg("listen-addr", "The address to listen on once connected").Required().String()
	// ConnectAddr is the address to connect to
	ConnectAddr = Connect.Arg("connect-addr", "The address to connect to").Required().String()
)

// Interact allows us to interact in a terminal way with a SwarmHandle
func Interact(swarm *network.SwarmHandle) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		text := scanner.Text()
		var name string
		_, err := fmt.Sscanf(text, "!nick %s", &name)
		if err == nil {
			swarm.ChangeNickname(name)
		} else {
			swarm.SendContent(text)
		}
	}
}
