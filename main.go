package main

import (
	"distributed-chord/node"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var n *node.Node

func main() {
	ip := os.Getenv("NODE_IP")
	if ip == "" {
		log.Fatal("NODE_IP environment variable not set")
	}

	n = node.NewNode(ip)

	// If there's a known node to join, attempt to connect with retries
	joinAddr := os.Getenv("JOIN_ADDR")
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
