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
	ID int    // Node ID
	IP string // Node IP address with the port
}

type Node struct {
	ID          int
	IP          string
	Successor   Pointer
	Predecessor Pointer
	FingerTable []Pointer
}

const (
	timeInterval = 10 * time.Second
	m            = 5
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

func CallRPCMethod(ip string, method string, message Message) (*Message, error) {
	client, err := rpc.Dial("tcp", ip)
	if err != nil {
		return &Message{}, fmt.Errorf("Failed to connect to node: %v", err)
	}
	defer client.Close()

	var reply *Message
	err = client.Call(method, message, &reply)
	if err != nil {
		return &Message{}, fmt.Errorf("Failed to call method: %v", err)
	}

	return reply, nil
}

func (n *Node) StartBootstrap() {
	go n.Stabilize()
	go n.FixFingers()
}

func (n *Node) FindSuccessor(message Message, reply *Message) error {
	fmt.Printf("Finding successor for %d...\n", message.ID)
	if message.ID > n.ID && message.ID <= n.Successor.ID {
		*reply = Message{
			ID: n.Successor.ID,
			IP: n.Successor.IP,
		}
		return nil
	} else {
		closest := n.closestPrecedingNode(message.ID)
		if closest.ID == n.ID {
			*reply = Message{
				ID: n.ID,
				IP: n.IP,
			}
			return nil
		}
		newReply, err := CallRPCMethod(closest.IP, "Node.FindSuccessor", message)
		if err != nil {
			fmt.Printf("Failed to call FindSuccessor: %v\n", err)
		}
		reply = newReply
		return nil
	}
}

func (n *Node) closestPrecedingNode(id int) Pointer {
	for i := m - 1; i >= 0; i-- {
		if n.FingerTable[i].ID > n.ID && n.FingerTable[i].ID < id {
			return n.FingerTable[i]
		}
	}
	return Pointer{ID: n.ID, IP: n.IP}
}

// Handled by the bootstrap node
func (n *Node) Join(joinIP string) {
	// Joining the network
	message := Message{
		Type: "Join",
		ID:   n.ID,
	}

	reply, err := CallRPCMethod(joinIP, "Node.FindSuccessor", message)

	if err != nil {
		log.Fatalf("Failed to join network: %v", err)
	}
	n.Predecessor = Pointer{}
	n.Successor = Pointer{ID: reply.ID, IP: reply.IP}

	// Notify the node of the new successor
	message = Message{
		Type: "Notify",
		ID:   n.ID,
		IP:   n.IP,
	}

	_, err = CallRPCMethod(reply.IP, "Node.Notify", message)
	if err != nil {
		log.Fatalf("Failed to notify successor: %v", err)
	}

	go n.Stabilize()
	go n.FixFingers()

}

// func (n *Node) initFingerTable(existingNode *Node) {
// 	n.FingerTable[0].node = existingNode.FindSuccessor(n.FingerTable[0].key)
// 	n.Successor = n.FingerTable[0].node

// 	for i := 0; i < m-1; i++ {
// 		if n.FingerTable[i+1].key > n.ID && n.FingerTable[i+1].key < n.FingerTable[i].node.ID {
// 			n.FingerTable[i+1].node = n.FingerTable[i].node
// 		} else {
// 			n.FingerTable[i+1].node = existingNode.FindSuccessor(n.FingerTable[i+1].key)
// 		}
// 	}
// }

func (n *Node) Stabilize() {
	for {
		fmt.Printf("Stabilizing...\n")
		time.Sleep(timeInterval)
		reply, err := CallRPCMethod(n.Successor.IP, "Node.GetPredecessor", Message{})
		if err != nil {
			fmt.Printf("Failed to get predecessor: %v\n", err)
		}

		successorPredecessor := Pointer{ID: reply.ID, IP: reply.IP}
		if successorPredecessor != (Pointer{}) && successorPredecessor.ID > n.ID && successorPredecessor.ID < n.Successor.ID {
			n.Successor = successorPredecessor
		}

		// Notify the successor of the new predecessor
		_, err = CallRPCMethod(n.Successor.IP, "Node.Notify", Message{ID: n.ID, IP: n.IP})

		if err != nil {
			fmt.Printf("Failed to notify successor: %v\n", err)
		}

	}
}

func (n *Node) GetPredecessor(message Message, reply *Message) error {
	*reply = Message{
		ID: n.Predecessor.ID,
		IP: n.Predecessor.IP,
	}
	return nil
}

func (n *Node) Notify(message Message, reply *Message) error {
	fmt.Printf("Getting notified by node %d...\n", message.ID)
	if n.Predecessor == (Pointer{}) || (message.ID > n.Predecessor.ID && message.ID < n.ID) {
		n.Predecessor = Pointer{ID: message.ID, IP: message.IP}
	}
	return nil
}

func (n *Node) FixFingers() {
	next := 0
	for {
		fmt.Printf("Fixing fingers...\n")
		time.Sleep(timeInterval)

		// Safely calculate the start of finger interval
		start := (n.ID + int(math.Pow(2, float64(next)))) % int(math.Pow(2, float64(m)))

		// Find and update successor for this finger
		message := Message{ID: start}
		var reply Message
		err := n.FindSuccessor(message, &reply)

		if err != nil {
			fmt.Printf("Failed to find successor: %v\n", err)
			continue
		}
		n.FingerTable[next] = Pointer{ID: reply.ID, IP: reply.IP}

		next = (next + 1) % m
	}
}

// // Fault tolerance
// func (n *Node) CheckPredecessor() {
// 	for {
// 		time.Sleep(time.Second)
// 		if n.Predecessor != nil {
// 			n.Predecessor.Predecessor = nil
// 		}
// 	}
// }

func CreateNode(ip string) *Node {
	id := utils.Hash(ip)

	node := &Node{
		ID:          id,
		IP:          ip,
		Successor:   Pointer{ID: id, IP: ip},
		FingerTable: make([]Pointer, m),
	}

	return node
}
