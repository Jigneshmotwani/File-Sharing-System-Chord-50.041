package node

import (
	"context"
	pb "distributed-chord/chord"
	"errors"
	"fmt"
	"math/big"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// FingerEntry represents an entry in the finger table
type FingerEntry struct {
	Start    []byte
	Interval [2][]byte
	Node     *RemoteNode
}

// Args and Reply structures for FindSuccessorRPC
type FindSuccessorArgs struct {
	ID []byte
}

type FindSuccessorReply struct {
	Successor *RemoteNode
}

// InitializeFingerTable initializes the finger table for the node
func (n *Node) InitializeFingerTable() {
	m := n.KeySize // Use the node's key size
	n.FingerTable = make([]*FingerEntry, m)

	for i := 0; i < m; i++ {
		start := calculateStart(n.ID, i, m)
		n.FingerTable[i] = &FingerEntry{
			Start:    start,
			Interval: [2][]byte{start, calculateStart(n.ID, i+1, m)},
			Node:     n.Successor, // Initially point to successor
		}
	}
	n.PrintFingerTable()
}

// UpdateFingerTable updates the finger table entries periodically
func (n *Node) UpdateFingerTable() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		n.FixFingerTable()
	}
}

// FixFingerTable refreshes the finger table entries
func (n *Node) FixFingerTable() {
	m := len(n.FingerTable)
	for i := 0; i < m; i++ {
		start := n.FingerTable[i].Start
		successor, err := n.findSuccessor(start)
		if err == nil {
			n.mutex.Lock()
			n.FingerTable[i].Node = successor
			n.mutex.Unlock()
			fmt.Printf("Node %x: Finger[%d] updated to node %x\n", n.ID, i, successor.ID)
		} else {
			fmt.Printf("Node %x: Error finding successor for finger[%d]: %v\n", n.ID, i, err)
		}
	}
	n.PrintFingerTable()
}

func (n *Node) findSuccessor(id []byte) (*RemoteNode, error) {
	n.mutex.Lock()
	successor := n.Successor
	n.mutex.Unlock()

	if betweenRightInclusive(n.ID, successor.ID, id, n.KeySize) {
		fmt.Printf("Node %s: findSuccessor(%s) -> Successor %s\n", formatID(n.ID, n.KeySize), formatID(id, n.KeySize), formatID(successor.ID, n.KeySize))
		return successor, nil
	} else {
		closestNode := n.closestPrecedingNode(id)
		if closestNode == nil {
			fmt.Printf("Node %x: No closest preceding node found for %x\n", n.ID, id)
			return nil, errors.New("No closest preceding node found")
		}
		fmt.Printf("Node %s: findSuccessor(%s) delegating to node %s\n", formatID(n.ID, n.KeySize), formatID(id, n.KeySize), formatID(closestNode.ID, n.KeySize))

		conn, err := grpc.NewClient(closestNode.Address(), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		defer conn.Close()

		client := pb.NewChordNodeClient(conn)
		req := &pb.FindSuccessorRequest{Id: id}
		resp, err := client.FindSuccessor(context.Background(), req)
		if err != nil {
			return nil, err
		}
		return &RemoteNode{
			ID:   resp.Successor.Id,
			IP:   resp.Successor.Ip,
			Port: int(resp.Successor.Port),
		}, nil
	}
}

func (n *Node) FindSuccessorGrpc(ctx context.Context, req *pb.FindSuccessorRequest) (*pb.FindSuccessorResponse, error) {
	successor, err := n.findSuccessor(req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.FindSuccessorResponse{
		Successor: &pb.NodeInfo{
			Id:   successor.ID,
			Ip:   successor.IP,
			Port: int32(successor.Port),
		},
	}, nil
}

// closestPrecedingNode finds the closest preceding node for a given ID
func (n *Node) closestPrecedingNode(id []byte) *RemoteNode {
	for i := len(n.FingerTable) - 1; i >= 0; i-- {
		n.mutex.Lock()
		fingerNode := n.FingerTable[i].Node
		n.mutex.Unlock()
		if fingerNode == nil {
			continue
		}
		if between(n.ID, id, fingerNode.ID, n.KeySize) {
			fmt.Printf("Node %x: closestPrecedingNode(%x) -> %x\n", n.ID, id, fingerNode.ID)
			return fingerNode
		}
	}
	return n.Successor
}

// Helper functions

// calculateStart calculates (n + 2^(i)) mod (2^m)
func calculateStart(nodeID []byte, i int, m int) []byte {
	two := big.NewInt(2)
	exponent := big.NewInt(int64(i))
	power := new(big.Int).Exp(two, exponent, nil) // 2^i

	nodeInt := new(big.Int).SetBytes(nodeID)
	modulo := new(big.Int).Lsh(big.NewInt(1), uint(m)) // 2^m

	sum := new(big.Int).Add(nodeInt, power)
	result := new(big.Int).Mod(sum, modulo)

	return result.Bytes()
}

// between checks if id is between start and end (excluding start and end)
func between(start, end, id []byte, m int) bool {
	startInt := new(big.Int).SetBytes(start)
	endInt := new(big.Int).SetBytes(end)
	idInt := new(big.Int).SetBytes(id)

	// Adjust for ring overflow
	if startInt.Cmp(endInt) < 0 {
		return idInt.Cmp(startInt) > 0 && idInt.Cmp(endInt) < 0
	}
	return idInt.Cmp(startInt) > 0 || idInt.Cmp(endInt) < 0
}

func betweenRightInclusive(start, end, id []byte, m int) bool {
	startInt := new(big.Int).SetBytes(start)
	endInt := new(big.Int).SetBytes(end)
	idInt := new(big.Int).SetBytes(id)

	// Adjust for ring overflow
	if startInt.Cmp(endInt) < 0 {
		return idInt.Cmp(startInt) > 0 && idInt.Cmp(endInt) <= 0
	}
	return idInt.Cmp(startInt) > 0 || idInt.Cmp(endInt) <= 0
}

// PrintFingerTable prints the node's finger table
func (n *Node) PrintFingerTable() {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	fmt.Printf("Finger table for node %x:\n", n.ID)
	for i, finger := range n.FingerTable {
		if finger.Node != nil {
			fmt.Printf("Entry %d: start %x, node %x\n", i, finger.Start, finger.Node.ID)
		} else {
			fmt.Printf("Entry %d: start %x, node nil\n", i, finger.Start)
		}
	}
}

// Helper function to format IDs consistently
func formatID(id []byte, m int) string {
	idInt := new(big.Int).SetBytes(id)
	// Calculate the number of hex digits based on m
	hexDigits := (m + 3) / 4 // Each hex digit represents 4 bits
	return fmt.Sprintf("%0*x", hexDigits, idInt)
}
