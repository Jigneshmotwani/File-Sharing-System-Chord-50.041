package main

import (
	"distributed-chord/node"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

var n *node.Node

func main() {
	joinAddr := os.Getenv("BOOTSTRAP_ADDR")
	ip, err := GetContainerIP()

	if err != nil {
		log.Fatalf("Failed to get container IP: %v", err)
	}

	log.Printf("Node starting at %s", ip)

	n = node.NewNode(ip)

	if joinAddr != "" {
		for i := 0; i < 5; i++ {
			err := attemptJoin(joinAddr)
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

	http.HandleFunc("/join", handleJoin)
	http.HandleFunc("/notify", handleNotify)
	http.HandleFunc("/successor", handleSuccessor)
	http.HandleFunc("/predecessor", handlePredecessor)

	log.Printf("Node %d starting HTTP server at %s", n.ID, ip)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:8080", ip), nil))
}

func handleJoin(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	// Find successor and return it as response
	successor := n.FindSuccessor(id)
	json.NewEncoder(w).Encode(successor)
}

func handleNotify(w http.ResponseWriter, r *http.Request) {
	var node node.Node
	json.NewDecoder(r.Body).Decode(&node)
	n.Notify(&node)
	w.WriteHeader(http.StatusOK)
}

func handleSuccessor(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(n.Successor)
}

func handlePredecessor(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(n.Predecessor)
}

func attemptJoin(joinAddr string) error {
	resp, err := http.Get(fmt.Sprintf("http://%s/join?id=%d", joinAddr, n.ID))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to join, status: %s", resp.Status)
	}
	return nil
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