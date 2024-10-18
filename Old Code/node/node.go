package node

import (
	"crypto/sha1"
	pb "distributed-chord/chord"
	"math/big"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
)

type Node struct {
	pb.UnimplementedChordNodeServer
	ID          []byte
	IP          string
	Port        int
	Successor   *RemoteNode
	Predecessor *RemoteNode
	FingerTable []*FingerEntry
	Data        map[string]string
	KeySize     int
	mutex       sync.Mutex
	grpcServer  *grpc.Server
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
}

func (n *Node) DisplayFingerTablePeriodically() {
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		n.PrintFingerTable()
	}
}
