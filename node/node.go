package node

import (
	"crypto/sha1"
	"math"
	"time"
)

type Node struct {
	ID          int
	IP          string
	Successor   *Node
	Predecessor *Node
	FingerTable []*FingerTableEntry
}

type FingerTableEntry struct {
	key  int
	node *Node
}

var m = 8

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

	for i := 0; i < m; i++ {
		node.FingerTable[i] = &FingerTableEntry{
			key:  (node.ID+2^i)%2 ^ m,
			node: nil,
		}
	}

	node.Successor = node

	// Initialize finger table entries to point to self
	for i := 0; i < m; i++ {
		node.FingerTable[i].node = node
	}

	return node
}

func (n *Node) FindSuccessor(id int) *Node {
	// Handle nil successor case
	if n.Successor == nil {
		return n
	}

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
			n.FingerTable[i].node = n
		}
	}
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

func (n *Node) CheckPredecessor() {
	for {
		time.Sleep(time.Second)
		if n.Predecessor != nil {
			n.Predecessor.Predecessor = nil
		}
	}
}
