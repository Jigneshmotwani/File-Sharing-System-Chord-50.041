package node

import (
	"crypto/sha1"
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

	return node
}

func (n *Node) FindSuccessor(id int) *Node {
	if n.ID < n.Successor.ID && id > n.ID && id <= n.Successor.ID {
		return n.Successor
	} else {
		closest := n.closestPrecedingNode(id)
		return closest.FindSuccessor(id)
	}
}

func (n *Node) closestPrecedingNode(id int) *Node {
	for i := m - 1; i >= 0; i-- {
		if n.FingerTable[i].node.ID > n.ID && n.FingerTable[i].node.ID < id {
			return n.FingerTable[i].node
		}
	}
	return n
}

func (n *Node) Join(existingNode *Node) {
	if existingNode != nil {
		n.Predecessor = nil
		n.Successor = existingNode.FindSuccessor(n.ID)
	} else {
		n.Predecessor = n
		n.Successor = n
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
		n.FingerTable[next].node = n.FindSuccessor(n.ID + 2 ^ next)
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
