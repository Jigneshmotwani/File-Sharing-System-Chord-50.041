package main

import (
	"distributed-chord/node"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

var n *node.Node

func main() {
	joinAddr := os.Getenv("BOOTSTRAP_ADDR")
	log.Printf("Join address: %s", joinAddr)
	ip, err := GetContainerIP()

	if err != nil {
		log.Fatalf("Failed to get container IP: %v", err)
	}

	log.Printf("Node starting at %s", ip)

	n = node.NewNode(ip)

	log.Printf("Node %d created", n.ID)

	if joinAddr != "" {
		for i := 0; i < 5; i++ {
			// err := attemptJoin(joinAddr)
			if err == nil {
				log.Printf("Node %d joined the network via %s", n.ID, joinAddr)
				break
			}
			log.Printf("Retrying join after failure: %v", err)
			time.Sleep(time.Duration(i+1) * time.Second) // Exponential backoff
		}
	} else {
		n.Join(nil)
	}

	go n.Stabilize()
	go n.FixFingers()
	go n.CheckPredecessor()

	log.Printf("Node %d started", n.ID)

	log.Printf("Node %d starting HTTP server at %s", n.ID, ip)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:8080", ip), nil))
}

func GetContainerIP() (string, error) {
	// Getting IP address of the container from eth0 interface
	iface, err := net.InterfaceByName("eth0")
	if err != nil {
		return "", fmt.Errorf("failed to get eth0 interface: %v", err)
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return "", fmt.Errorf("failed to get addresses for eth0: %v", err)
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if ipv4 := ipnet.IP.To4(); ipv4 != nil {
				return ipv4.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no IPv4 address found for eth0")
}
