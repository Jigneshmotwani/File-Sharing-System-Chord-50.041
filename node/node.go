package node

import (
	"distributed-chord/utils"
	"fmt"
	"log"
	"math"
	"net"
	"net/rpc"
	"time"
)

type Pointer struct {
	ID int // Node ID
	IP string // Node IP address with the port
}

type Node struct {
	ID          int
	IP          string
	Successor   Pointer
	Predecessor Pointer
	FingerTable []Pointer
}

type FingerTableEntry struct {
	key  int
	node *Node
}

const (
	m = 32
)

// Starting the RPC server for the nodes
func (n *Node) StartRPCServer() {
	// Start the net RPC server
	rpc.Register(n)

	listener, err := net.Listen("tcp", n.IP)

	if err != nil {
		fmt.Printf("Error starting RPC server: %v\n", err)
		return
	}
	
	defer listener.Close()

	fmt.Printf("Listening on %s\n", n.IP)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("[NODE-%d] accept error: %s\n", n.ID, err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}


func (n *Node) StartBootstrap() {
	go n.Stabilize()
	go n.FixFingers()
}

func (n *Node) ReceiveMessage(message string, reply *string) error {
	fmt.Printf("[NODE-%d] Received message: %s\n", n.ID, message)
	*reply = "Message received"
	return nil
}

func (n *Node) JoinNetwork(message Message, reply *Message) error {
	n.FindSuccessor(message.ID)
}

func (n *Node) FindSuccessor(id int) Pointer {
	if n.ID < n.Successor.ID && id > n.ID && id <= n.Successor.ID {
		return n.Successor
	} else {
		closest := n.closestPrecedingNode(id)
		if closest == n {
			return n
		}
		return closest.FindSuccessor(id)
	}
}

func (n *Node) closestPrecedingNode(id int) *Node {
	for i := m - 1; i >= 0; i-- {
		if n.FingerTable[i] != nil && n.FingerTable[i].node != nil {
			if n.FingerTable[i].node.ID > n.ID && n.FingerTable[i].node.ID < id {
				return n.FingerTable[i].node
			}
		}
	}
	return n
}

// Handled by the bootstrap node
func (n *Node) Join(joinIP string) {
		// Joining the network

		client, err := rpc.Dial("tcp", joinIP)
		if err != nil {
			log.Fatalf("Failed to connect to bootstrap node: %v", err)
		}

		message := Message{
			Type: "Join",
			ID: n.ID,
		}

		var reply Message
		client.Call("Node.JoinNetwork", message, &reply)
		n.Predecessor = Pointer{}
		n.Successor = Pointer{ID: reply.ID, IP: reply.IP}
}

func (n *Node) initFingerTable(existingNode *Node) {
	n.FingerTable[0].node = existingNode.FindSuccessor(n.FingerTable[0].key)
	n.Successor = n.FingerTable[0].node

	for i := 0; i < m-1; i++ {
		if n.FingerTable[i+1].key > n.ID && n.FingerTable[i+1].key < n.FingerTable[i].node.ID {
			n.FingerTable[i+1].node = n.FingerTable[i].node
		} else {
			n.FingerTable[i+1].node = existingNode.FindSuccessor(n.FingerTable[i+1].key)
		}
	}
}

func (n *Node) Stabilize() {
	for {
		time.Sleep(time.Second)
		if n.Successor != nil && n.Successor != n {
			successorPredecessor := n.Successor.Predecessor
			if successorPredecessor != nil && successorPredecessor.ID > n.ID && successorPredecessor.ID < n.Successor.ID {
				n.Successor = successorPredecessor
			}
			n.Successor.Notify(n)
		}
	}
}

func (n *Node) Notify(existingNode *Node) {
	if n.Predecessor == nil || (existingNode.ID > n.Predecessor.ID && existingNode.ID < n.ID) {
		n.Predecessor = existingNode
	}
}

func (n *Node) FixFingers() {
	next := 0
	for {
		time.Sleep(time.Second)
		next = (next + 1) % m

		// Safely calculate the start of finger interval
		start := (n.ID + int(math.Pow(2, float64(next)))) % int(math.Pow(2, float64(m)))

		// Find and update successor for this finger
		successor := n.FindSuccessor(start)
		if successor != nil {
			n.FingerTable[next].node = successor
		}
	}
}

// Fault tolerance
func (n *Node) CheckPredecessor() {
	for {
		time.Sleep(time.Second)
		if n.Predecessor != nil {
			n.Predecessor.Predecessor = nil
		}
	}
}

func CreateNode(ip string) *Node {
	id := utils.Hash(ip)

	node := &Node{
		ID:          id,
		IP:          ip,
		Successor:  Pointer{ID: id, IP: ip},
		FingerTable: make([]Pointer, m),
	}

	return node
}
