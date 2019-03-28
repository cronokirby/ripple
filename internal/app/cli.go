package app

import "github.com/alecthomas/kingpin"

var (
	App = kingpin.New("ripple", "A decentralized chat application")

	Start     = App.Command("start", "Start a new swarm")
	StartAddr = Start.Arg("addr", "The address to listen on").Required().String()

	Connect           = App.Command("connect", "Connect to an existing swarm")
	ConnectListenAddr = Connect.Arg("listen-addr", "The address to listen on once connected").Required().String()
	ConnectAddr       = Connect.Arg("connect-addr", "The address to connect to").Required().String()
)
