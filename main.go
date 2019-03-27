package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/cronokirby/ripple/internal/network"
	"github.com/cronokirby/ripple/internal/protocol"
)

func interact(swarm *network.SwarmHandle) {
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

func main() {
	args := os.Args
	argLen := len(args)
	if argLen < 2 {
		fmt.Println("Insufficient arguments")
		os.Exit(1)
	}
	logger := log.New(os.Stdout, "", log.Flags())
	switch args[1] {
	case "listen":
		if argLen < 3 {
			fmt.Println("Insufficient arguments")
			os.Exit(1)
		}
		me, err := net.ResolveTCPAddr("tcp", args[2])
		if err != nil {
			logger.Fatalln("Failed to resolve own address: ", err)
		}
		logger.Println("Starting new swarm...")
		swarm, err := network.CreateSwarm(logger, me)
		if err != nil {
			logger.Fatalln("Failed to join swarm: ", err)
		}
		swarm.SetReceiver(protocol.PrintReceiver{})
		interact(swarm)
	case "connect":
		if argLen < 4 {
			fmt.Println("Insufficient arguments")
			os.Exit(1)
		}
		me, err := net.ResolveTCPAddr("tcp", args[2])
		if err != nil {
			logger.Fatalln("Failed to resolve own address: ", err)
		}
		them, err := net.ResolveTCPAddr("tcp", args[3])
		if err != nil {
			logger.Fatalln("Failed to resolve peer address: ", err)
		}
		logger.Println("Joining swarm...")
		swarm, err := network.JoinSwarm(logger, me, them)
		if err != nil {
			logger.Fatalln("Failed to join swarm: ", err)
		}
		swarm.SetReceiver(protocol.PrintReceiver{})
		interact(swarm)
	default:
		fmt.Println("Unknown command")
	}
}
