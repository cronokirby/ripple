package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/cronokirby/ripple/internal/network"
)

func interact(swarm *network.SwarmHandle) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		text := scanner.Text()
		swarm.SendContent(text)
	}
}

func main() {
	res, _ := net.ResolveTCPAddr("tcp", "localhost:8080")
	fmt.Println(res)
	os.Exit(1)
	args := os.Args
	argLen := len(args)
	if argLen < 2 {
		fmt.Println("Insufficient arguments")
		os.Exit(1)
	}
	writer := bufio.NewWriter(os.Stdout)
	logger := log.New(writer, "", log.Flags())
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
		interact(swarm)
	default:
		fmt.Println("Unkown command")
	}
}
