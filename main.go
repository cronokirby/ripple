package main

import (
	"fmt"
	"net"
	"os"
)

/**
func interact(j *network.Joiner) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		text := scanner.Text()
		j.SendContent(text)
	}
}
**/

func main() {
	res, _ := net.ResolveTCPAddr("tcp", "localhost:8080")
	fmt.Println(res)
	os.Exit(1)
	/**
	args := os.Args
	argLen := len(args)
	if argLen < 2 {
		fmt.Println("Insufficient arguments")
		os.Exit(1)
	}
	joiner := network.NewJoiner(&protocol.PrintReceiver{})
	switch args[1] {
	case "listen":
		if argLen < 3 {
			fmt.Println("Insufficient arguments")
			os.Exit(1)
		}
		fmt.Println("Starting new swarm...")
		go joiner.Listen(args[2])
		interact(joiner)
	case "connect":
		if argLen < 4 {
			fmt.Println("Insufficient arguments")
			os.Exit(1)
		}
		remote, err := net.ResolveTCPAddr("tcp", args[3])
		if err != nil {
			fmt.Println("Failed to join swarm: ", err)
			os.Exit(1)
		}
		err = joiner.Start(args[2], remote)
		fmt.Println("Succesfully joined swarm!")
		interact(joiner)
	default:
		fmt.Println("Unkown command")
	}
	*/
}
