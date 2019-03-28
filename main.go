package main

import (
	"log"
	"net"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/cronokirby/ripple/internal/app"
	"github.com/cronokirby/ripple/internal/network"
	"github.com/cronokirby/ripple/internal/protocol"
)

func main() {
	logger := log.New(os.Stdout, "", log.Flags())
	switch kingpin.MustParse(app.App.Parse(os.Args[1:])) {
	case app.Start.FullCommand():
		me, err := net.ResolveTCPAddr("tcp", *app.StartAddr)
		if err != nil {
			logger.Fatalln("Failed to resolve own address: ", err)
		}
		logger.Println("Starting new swarm...")
		swarm, err := network.CreateSwarm(logger, me)
		if err != nil {
			logger.Fatalln("Failed to join swarm: ", err)
		}
		swarm.SetReceiver(protocol.PrintReceiver{})
		app.Interact(swarm)
	case app.Connect.FullCommand():
		me, err := net.ResolveTCPAddr("tcp", *app.ConnectListenAddr)
		if err != nil {
			logger.Fatalln("Failed to resolve own address: ", err)
		}
		them, err := net.ResolveTCPAddr("tcp", *app.ConnectAddr)
		if err != nil {
			logger.Fatalln("Failed to resolve peer address: ", err)
		}
		logger.Println("Joining swarm...")
		swarm, err := network.JoinSwarm(logger, me, them)
		if err != nil {
			logger.Fatalln("Failed to join swarm: ", err)
		}
		swarm.SetReceiver(protocol.PrintReceiver{})
		app.Interact(swarm)
	}
}
