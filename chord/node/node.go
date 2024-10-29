package node

// Node/Fingertable:
// node struct
// fingertable/fingerentry struct
// hashing function
// findSuccessor function (use another function to find closestPreceedingNode)
// initialize finger table function
// finger table filler function

import (
	"context"
	"crypto/sha1"
	pb "distributed-chord/pb"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	maxHashValue     = "1461501637330902918203684832716283019655932542975" // 2^160 - 1
	numFingerEntries = 160                                                 // Finger table size, based on 160-bit identifiers
)

type Node struct {
	pb.UnimplementedChordNodeServer
	id          *big.Int
	nodeName    string
	dataPath    string
	ip          string
	fingerTable []FingerTableEntry
	successor   *Node
	predecessor *Node
	grpcServer  *grpc.Server
	port        int
	mutex       sync.Mutex
}

type RemoteNode struct {
	ID   []byte
	IP   string
	Port int
}

// FingerTableEntry represents an entry in the finger table
type FingerTableEntry struct {
	slot *big.Int
	node *Node
}

// SHA-1 hashing
func SHA1Hash(data string) *big.Int {
	// SHA-1 hashing
	hasher := sha1.New()
	hasher.Write([]byte(data))
	hash := hex.EncodeToString(hasher.Sum(nil))

	// Convert hash to big integer
	hashedBytes, _ := hex.DecodeString(hash)
	bigIntHash := new(big.Int).SetBytes(hashedBytes)
	return bigIntHash
}

// findSuccessor finds the successor node for a given key.
func (n *Node) findSuccessor(id *big.Int) (*Node, error) {
	n.mutex.Lock()
	successor := n.successor
	defer n.mutex.Unlock()

	// Step 1: Check if id is between this node and its successor
	if isInInterval(id, n.getID(), n.successor.getID(), true) {
		return successor, nil
	}

	// Step 2: If not, ask the closest preceding node to find the successor
	closestNode := n.closestPrecedingNode(id)
	if closestNode == n { // If no closer node is found, we assume the current node is the closest
		return n, nil
	}

	// Recursive call to find the successor from the closest preceding node
	//return closestNode.findSuccessor(id)
	conn, err := grpc.NewClient(closestNode.Address(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pb.NewChordNodeClient(conn)
	req := &pb.FindSuccessorRequest{Id: id.Bytes()}
	resp, err := client.FindSuccessor(context.Background(), req)
	if err != nil {
		return nil, err
	}
	return &Node{
		id:   new(big.Int).SetBytes(resp.Successor.Id),
		ip:   resp.Successor.Ip,
		port: int(resp.Successor.Port),
	}, nil
}

func (n *Node) FindSuccessor(ctx context.Context, req *pb.FindSuccessorRequest) (*pb.FindSuccessorResponse, error) {
	successor, err := n.findSuccessor(new(big.Int).SetBytes(req.Id))
	if err != nil {
		return nil, err
	}
	return &pb.FindSuccessorResponse{
		Successor: &pb.NodeInfo{
			Id:   successor.id.Bytes(),
			Ip:   successor.ip,
			Port: int32(successor.port),
		},
	}, nil
}

// closestPrecedingNode returns the closest preceding node in the finger table for the given id.
func (n *Node) closestPrecedingNode(id *big.Int) *Node {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	// Traverse the finger table from the farthest entry
	for i := numFingerEntries - 1; i >= 0; i-- {
		if n.fingerTable[i].node != nil && isInInterval(n.fingerTable[i].node.getID(), n.getID(), id, false) {
			return n.fingerTable[i].node
		}
	}
	// If no closer node found, return the current node
	return n
}

// isInInterval checks if the id is in the interval (start, end] in the identifier space.
// If includeEnd is true, the interval becomes (start, end], otherwise it is (start, end).
func isInInterval(id, start, end *big.Int, includeEnd bool) bool {
	// Special case: wrap around the identifier space
	if start.Cmp(end) > 0 {
		return id.Cmp(start) > 0 || (id.Cmp(end) <= 0 && (includeEnd || id.Cmp(end) != 0))
	}
	// Regular case: no wrap-around
	if includeEnd {
		return id.Cmp(start) > 0 && id.Cmp(end) <= 0
	}
	return id.Cmp(start) > 0 && id.Cmp(end) < 0
}

// getID returns the node's identifier (hash).
func (n *Node) getID() *big.Int {
	return SHA1Hash(n.nodeName)
}

func (n *Node) Address() string {
	return n.ip + ":" + strconv.Itoa(n.port)
}

// initializeFingerTable initializes the finger table of a node
func (n *Node) initializeFingerTables() {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.fingerTable = make([]FingerTableEntry, numFingerEntries)
	maxHash := new(big.Int)
	maxHash.SetString(maxHashValue, 10)

	for i := 0; i < numFingerEntries; i++ {
		// Calculate the start of each finger entry: (n.getID() + 2^i) % 2^160
		slot := new(big.Int).Add(n.getID(), new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(i)), nil))
		slot.Mod(slot, maxHash)

		// Find the successor of the start value
		successor, err := n.findSuccessor(slot)
		if err != nil {
			fmt.Println("error finding successor:", err)
			return
		}

		// Fill the finger table entry
		n.fingerTable[i] = FingerTableEntry{
			slot: slot,
			node: successor,
		}
	}
}

// updateFingerTable periodically updates the node's finger table to reflect new nodes in the Chord ring.
func (n *Node) updateFingerTable() {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	maxHash := new(big.Int)
	maxHash.SetString(maxHashValue, 10)

	for i := 0; i < numFingerEntries; i++ {
		// Calculate the start of each finger entry: (n.getID() + 2^i) % 2^160
		start := new(big.Int).Add(n.getID(), new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(i)), nil))
		start.Mod(start, maxHash)

		// Find the successor of the start value
		successor, err := n.findSuccessor(start)
		if err != nil {
			fmt.Println("error finding successor:", err)
			return
		}

		// Only update if the successor has changed
		if n.fingerTable[i].node != successor {
			n.fingerTable[i].node = successor
		}
	}
}

func (n *Node) startFingerTableUpdater() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				n.updateFingerTable()
			}
		}
	}()
}
