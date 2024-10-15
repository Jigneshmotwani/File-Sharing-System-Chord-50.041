package node

import (
    "crypto/sha1"
    // "fmt"
    "strconv"
    "sync"
	"time"
	"math/big"
)

type Node struct {
    ID          []byte         // Node ID (SHA-1 hash)
    IP          string         // IP address
    Port        int            // Port number
    Successor   *RemoteNode    // Successor node
    Predecessor *RemoteNode    // Predecessor node
    FingerTable []*FingerEntry // Finger table entries
    Data        map[string]string // Key-value store for the node
    KeySize     int            // Key space size (m)

    mutex sync.Mutex // Mutex to protect shared resources
}

type RemoteNode struct {
    ID   []byte
    IP   string
    Port int
}

func (rn *RemoteNode) Address() string {
    return rn.IP + ":" + strconv.Itoa(rn.Port)
}

func NewNode(ip string, port int, m int) *Node {
    address := ip + ":" + strconv.Itoa(port)
    hash := sha1.Sum([]byte(address))

    // Map the hash to the key space [0, 2^m)
    hashInt := new(big.Int).SetBytes(hash[:])
    modulo := new(big.Int).Lsh(big.NewInt(1), uint(m)) // 2^m
    idInt := new(big.Int).Mod(hashInt, modulo)
    idBytes := idInt.Bytes()

    node := &Node{
        ID:          idBytes,
        IP:          ip,
        Port:        port,
        Successor:   nil,
        Predecessor: nil,
        Data:        make(map[string]string),
        KeySize:     m,
    }

    // Initially, the node's successor and predecessor are itself
    node.Successor = &RemoteNode{
        ID:   node.ID,
        IP:   node.IP,
        Port: node.Port,
    }

    node.Predecessor = &RemoteNode{
        ID:   node.ID,
        IP:   node.IP,
        Port: node.Port,
    }

    return node
}


func (n *Node) Address() string {
    return n.IP + ":" + strconv.Itoa(n.Port)
}

func (n *Node) StartPeriodicTasks() {
    go n.UpdateFingerTable()
    go n.StartStabilization()
    // go n.DisplayFingerTablePeriodically()
}

func (n *Node) DisplayFingerTablePeriodically() {
    ticker := time.NewTicker(30 * time.Second)
    for range ticker.C {
        n.PrintFingerTable()
    }
}

