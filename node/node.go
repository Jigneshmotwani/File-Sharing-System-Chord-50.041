// package node

// import (
// 	"crypto/sha1"
// 	"time"
// )

// type Node struct {
// 	ID          int
// 	IP          string
// 	Successor   *Node
// 	Predecessor *Node
// 	FingerTable []*FingerTableEntry
// }

// type FingerTableEntry struct {
// 	key  int
// 	node *Node
// }

// var m = 8

// func NewNode(ip string) *Node {
// 	h := sha1.New()
// 	h.Write([]byte(ip))
// 	id := int(h.Sum(nil)[0])

// 	node := &Node{
// 		ID:          id,
// 		IP:          ip,
// 		Successor:   nil,
// 		Predecessor: nil,
// 		FingerTable: make([]*FingerTableEntry, m),
// 	}

// 	for i := 0; i < m; i++ {
// 		node.FingerTable[i] = &FingerTableEntry{
// 			key:  (node.ID+2^i)%2 ^ m,
// 			node: nil,
// 		}
// 	}

// 	node.Successor = node

// 	return node
// }

// func (n *Node) FindSuccessor(id int) *Node {
// 	if n.ID < n.Successor.ID && id > n.ID && id <= n.Successor.ID {
// 		return n.Successor
// 	} else {
// 		closest := n.closestPrecedingNode(id)
// 		return closest.FindSuccessor(id)
// 	}
// }

// func (n *Node) closestPrecedingNode(id int) *Node {
// 	for i := m - 1; i >= 0; i-- {
// 		if n.FingerTable[i].node.ID > n.ID && n.FingerTable[i].node.ID < id {
// 			return n.FingerTable[i].node
// 		}
// 	}
// 	return n
// }

// func (n *Node) Join(existingNode *Node) {
// 	if existingNode != nil {
// 		n.Predecessor = nil
// 		n.Successor = existingNode.FindSuccessor(n.ID)
// 	} else {
// 		n.Predecessor = n
// 		n.Successor = n
// 	}
// }

// func (n *Node) Stabilize() {
// 	for {
// 		time.Sleep(time.Second)
// 		if n.Successor != nil && n.Successor != n {
// 			successorPredecessor := n.Successor.Predecessor
// 			if successorPredecessor != nil && successorPredecessor.ID > n.ID && successorPredecessor.ID < n.Successor.ID {
// 				n.Successor = successorPredecessor
// 			}
// 			n.Successor.Notify(n)
// 		}
// 	}
// }

// func (n *Node) Notify(existingNode *Node) {
// 	if n.Predecessor == nil || (existingNode.ID > n.Predecessor.ID && existingNode.ID < n.ID) {
// 		n.Predecessor = existingNode
// 	}
// }

// func (n *Node) FixFingers() {
// 	next := 0
// 	for {
// 		time.Sleep(time.Second)
// 		next = (next + 1) % m
// 		n.FingerTable[next].node = n.FindSuccessor(n.ID + 2 ^ next)
// 	}
// }

// func (n *Node) CheckPredecessor() {
// 	for {
// 		time.Sleep(time.Second)
// 		if n.Predecessor != nil {
// 			n.Predecessor.Predecessor = nil
// 		}
// 	}
// }

package node

import (
	"crypto/sha1"
	"math"
	"time"

	"google.golang.org/grpc"
)

type Node struct {
	ID          int
	IP          string
	Successor   *Node
	Predecessor *Node
	FingerTable []*FingerTableEntry
	grpcServer  *grpc.Server
}

type FingerTableEntry struct {
	Start int    // Start of finger interval
	Node  *Node  // Successor node for this interval
}

const m = 8 // Size of identifier space: 2^m

func NewNode(ip string) *Node {
	h := sha1.New()
	h.Write([]byte(ip))
	id := int(h.Sum(nil)[0])

	node := &Node{
		ID:          id,
		IP:          ip,
		Successor:   nil,
		Predecessor: nil,
		FingerTable: make([]*FingerTableEntry, m),
	}

	// Initialize finger table with correct start intervals
	for i := 0; i < m; i++ {
		node.FingerTable[i] = &FingerTableEntry{
			Start: (node.ID + int(math.Pow(2, float64(i)))) % int(math.Pow(2, float64(m))),
			Node:  nil,
		}
	}

	// Set self as successor initially
	node.Successor = node
	
	// Initialize finger table entries to point to self
	for i := 0; i < m; i++ {
		node.FingerTable[i].Node = node
	}

	return node
}

func (n *Node) FindSuccessor(id int) *Node {
	// Handle nil successor case
	if n.Successor == nil {
		return n
	}

	// Check if id is between current node and its successor
	if between(n.ID, id, n.Successor.ID, m) {
		return n.Successor
	}

	// Find closest preceding node and forward the query
	nprime := n.closestPrecedingNode(id)
	if nprime == n {
		return n
	}
	return nprime.FindSuccessor(id)
}

func (n *Node) closestPrecedingNode(id int) *Node {
	// Iterate through finger table backwards
	for i := m - 1; i >= 0; i-- {
		if n.FingerTable[i] != nil && n.FingerTable[i].Node != nil {
			fingerID := n.FingerTable[i].Node.ID
			// Check if finger node is between current node and target id
			if between(n.ID, fingerID, id, m) {
				return n.FingerTable[i].Node
			}
		}
	}
	return n
}

func (n *Node) Join(existingNode *Node) {
	if existingNode != nil {
		n.Predecessor = nil
		n.Successor = existingNode.FindSuccessor(n.ID)
		
		// Initialize finger table
		n.initFingerTable(existingNode)
	} else {
		n.Predecessor = n
		n.Successor = n
		
		// Initialize all fingers to point to self
		for i := 0; i < m; i++ {
			n.FingerTable[i].Node = n
		}
	}
}

func (n *Node) initFingerTable(existingNode *Node) {
	// Initialize first finger
	n.FingerTable[0].Node = existingNode.FindSuccessor(n.FingerTable[0].Start)
	n.Successor = n.FingerTable[0].Node

	// Initialize the rest of the finger table
	for i := 0; i < m-1; i++ {
		if between(n.ID, n.FingerTable[i+1].Start, n.FingerTable[i].Node.ID, m) {
			n.FingerTable[i+1].Node = n.FingerTable[i].Node
		} else {
			n.FingerTable[i+1].Node = existingNode.FindSuccessor(n.FingerTable[i+1].Start)
		}
	}
}

func (n *Node) Stabilize() {
	for {
		time.Sleep(time.Second)
		if n.Successor != nil && n.Successor != n {
			successorPredecessor := n.Successor.Predecessor
			if successorPredecessor != nil && 
			   between(n.ID, successorPredecessor.ID, n.Successor.ID, m) {
				n.Successor = successorPredecessor
			}
			n.Successor.Notify(n)
		}
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
			n.FingerTable[next].Node = successor
		}
	}
}

func (n *Node) Notify(node *Node) {
	if n.Predecessor == nil || 
	   between(n.Predecessor.ID, node.ID, n.ID, m) {
		n.Predecessor = node
	}
}

func (n *Node) CheckPredecessor() {
	for {
		time.Sleep(time.Second)
		if n.Predecessor != nil {
			// Add actual predecessor failure detection here
			// Current implementation is oversimplified
		}
	}
}

// Helper function to check if x is between a and b in the ring
func between(a, x, b, m int) bool {
	if a < b {
		return a < x && x <= b
	}
	return a < x || x <= b
}
