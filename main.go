// main.go

package main

import (
    "distributed-chord/node"
    "flag"
    "fmt"
    "log"
)

func main() {
    // Parse command-line flags
    ip := flag.String("ip", "127.0.0.1", "IP address of the node")
    port := flag.Int("port", 8000, "Port number of the node")
    joinAddr := flag.String("join", "", "Address of an existing node to join (format: ip:port)")
    m := flag.Int("m", 5, "Key space size in bits")

    flag.Parse()

    // Create a new node instance
    n := node.NewNode(*ip, *port, *m)

    // Start the RPC server
    err := n.StartRPCServer()
    if err != nil {
        log.Fatalf("Failed to start RPC server: %v", err)
    }

    // Join the network or create a new one
    if *joinAddr != "" {
        fmt.Printf("Node %x joining the network via %s...\n", n.ID, *joinAddr)
        err = n.Join(*joinAddr)
        if err != nil {
            log.Fatalf("Failed to join network: %v", err)
        }
    } else {
        fmt.Printf("Node %x creating a new network...\n", n.ID)
        n.InitializeFingerTable()
    }

    // Start periodic tasks (finger table updates and stabilization)
    n.StartPeriodicTasks()

    // Keep the main function running
    select {}
}
