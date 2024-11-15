package node

import (
	"distributed-chord/utils"
	"fmt"
	"log"
	"math"
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Pointer struct {
	ID int    // Node ID
	IP string // Node IP address with the port
}

type Node struct {
	ID            int
	IP            string
	Successor     Pointer
	Predecessor   Pointer
	FingerTable   []Pointer
	SuccessorList []Pointer
	Lock          sync.Mutex
}

type FileTransferRequest struct {
	SenderIP string
	FileName string
}

type NodeInfo struct {
	ID        int
	IP        string
	Successor Pointer
}

const (
	timeInterval = 5 // Time interval for stabilization and fixing fingers
	r            = 1 // Number of successors to keep in the successor list
)

// Starting the RPC server for the nodes
func (n *Node) StartRPCServer(ready chan<- bool) {
	rpc.Register(n)
	listener, err := net.Listen("tcp", n.IP)
	if err != nil {
		fmt.Printf("[NODE-%d] Error starting RPC server: %v\n", n.ID, err)
		ready <- false
		return
	}
	defer listener.Close()
	fmt.Printf("[NODE-%d] Listening on %s\n", n.ID, n.IP)

	// Signal that the server is ready
	ready <- true

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("[NODE-%d] accept error: %s\n", n.ID, err)
			return
		}
		go rpc.ServeConn(conn)
	}

}

func (n *Node) RequestFileTransfer(targetNodeID int, fileName string) error {
	message := Message{ID: targetNodeID}
	var reply Message
	err := n.FindSuccessor(message, &reply)
	if err != nil {
		return fmt.Errorf("failed to find successor: %v", err)
	}
	targetNodeIP := reply.IP
	fmt.Printf("Target node IP: %s\n", targetNodeIP)

	client, err := rpc.Dial("tcp", targetNodeIP)
	if err != nil {
		return fmt.Errorf("failed to connect to target node: %v", err)
	}
	defer client.Close()

	var response string
	request := FileTransferRequest{
		SenderIP: n.IP,
		FileName: fileName,
	}

	if request.SenderIP == targetNodeIP {
		fmt.Println("Cannot send file to the same node.")
		return nil
	}

	err = client.Call("Node.ConfirmFileTransfer", request, &response)
	if err != nil {
		return fmt.Errorf("failed to send file transfer request: %v", err)
	}

	// fmt.Printf("Response from target: %s\n", response)
	if (request.SenderIP != targetNodeIP) && (response == "es") {
		fmt.Println("Target accepted the file transfer. Initiating transfer...")
		// Add file transfer logic here (e.g., sending the file in chunks)
		n.Chunker(fileName, targetNodeIP)
	} else {
		fmt.Println("Target declined the file transfer.")
	}
	return nil
}

func (n *Node) ConfirmFileTransfer(request FileTransferRequest, reply *string) error {
	fmt.Printf("Received file transfer request from %s for file %s\n", request.SenderIP, request.FileName)
	var userResponse string
	fmt.Print("Do you want to receive the file? (yes/no):")
	fmt.Scanln(&userResponse)
	// fmt.Printf("User response: %s\n", userResponse)
	*reply = userResponse
	return nil
}

func (n *Node) FindSuccessor(message Message, reply *Message) error {
	// fmt.Printf("[NODE-%d] Finding successor for %d...\n", n.ID, message.ID)
	if utils.Between(message.ID, n.ID, n.Successor.ID, true) { // message.ID is between n.ID and n.Successor.ID (inclusive of Successor ID)
		*reply = Message{
			ID: n.Successor.ID,
			IP: n.Successor.IP,
		}
		// fmt.Printf("[NODE-%d] Successor found: %v\n", n.ID, reply.ID)
		return nil
	} else {
		closest := n.closestPrecedingNode(message.ID)
		if closest.ID == n.ID {
			*reply = Message{
				ID: n.ID,
				IP: n.IP,
			}
			// fmt.Printf("[NODE-%d] Successor is self: %v\n", n.ID, reply.ID)
			return nil
		}
		newReply, err := CallRPCMethod(closest.IP, "Node.FindSuccessor", message)
		if err != nil {
			return fmt.Errorf("[NODE-%d] Failed to find successor via RPC: %v", n.ID, err)
		}
		*reply = *newReply
		return nil
	}
}

func (n *Node) closestPrecedingNode(id int) Pointer {
	for i := utils.M - 1; i >= 0; i-- {
		if utils.Between(n.FingerTable[i].ID, n.ID, id, false) {
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
		IP:   n.IP,
	}

	reply, err := CallRPCMethod(joinIP, "Node.FindSuccessor", message)

	if err != nil {
		log.Fatalf("[NODE-%d] Failed to join network: %v", n.ID, err)
	}

	// fmt.Printf("[NODE-%d] Joining network with successor: %v\n", n.ID, reply.ID)
	n.Predecessor = Pointer{}
	n.Successor = Pointer{ID: reply.ID, IP: reply.IP}

	// Notify the successor of the new predecessor
	message = Message{
		Type: "NOTIFY",
		ID:   n.ID,
		IP:   n.IP,
	}

	_, err = CallRPCMethod(n.Successor.IP, "Node.Notify", message)
	if err != nil {
		log.Fatalf("[NODE-%d] Failed to notify successor: %v", n.ID, err)
	}
}

func (n *Node) Stabilize() {
	for {
		time.Sleep(timeInterval * time.Second)
		// fmt.Printf("[NODE-%d] Stabilizing...\n", n.ID)

		reply, err := CallRPCMethod(n.Successor.IP, "Node.GetPredecessor", Message{})
		if err != nil {
			nextSuccessor := n.findNextAlive()
			if nextSuccessor == (Pointer{}) {
				fmt.Printf("[NODE-%d] No Successor from the successor list is alive.", n.ID)
				fmt.Printf("[NODE-%d] Failed to get successor's predecessor: %v\n", n.ID, err)
				continue
			}
			n.Successor = nextSuccessor
		} else {
			successorPredecessor := Pointer{ID: reply.ID, IP: reply.IP}
			if successorPredecessor != (Pointer{}) && utils.Between(successorPredecessor.ID, n.ID, n.Successor.ID, false) {
				n.Successor = successorPredecessor
				// fmt.Printf("[NODE-%d] Successor updated to %d\n", n.ID, n.Successor.ID)
			}
		}

		// Notify the successor of the new predecessor
		message := Message{
			Type: "NOTIFY",
			ID:   n.ID,
			IP:   n.IP,
		}
		_, err = CallRPCMethod(n.Successor.IP, "Node.Notify", message)

		if err != nil {
			fmt.Printf("[NODE-%d] Failed to notify successor: %v\n", n.ID, err)
		}

		// Update the successor list
		n.updateSuccessorList()
	}
}

func (n *Node) GetSuccessor(message Message, reply *Message) error {
	*reply = Message{
		ID: n.Successor.ID,
		IP: n.Successor.IP,
	}
	return nil
}

func (n *Node) GetSuccessorList(message Message, reply *Message) error {
	*reply = Message{
		ID:            n.Successor.ID,
		SuccessorList: n.SuccessorList,
	}
	return nil
}

func (n *Node) GetPredecessor(message Message, reply *Message) error {
	*reply = Message{
		ID: n.Predecessor.ID,
		IP: n.Predecessor.IP,
	}
	return nil
}

func (n *Node) Notify(message Message, reply *Message) error {
	// fmt.Printf("[NODE-%d] Notified by node %d...\n", n.ID, message.ID)
	if n.Predecessor == (Pointer{}) || utils.Between(message.ID, n.Predecessor.ID, n.ID, false) {
		n.Predecessor = Pointer{ID: message.ID, IP: message.IP}
		// fmt.Printf("[NODE-%d] Predecessor updated to %d\n", n.ID, n.Predecessor.ID)
	}
	return nil
}

func (n *Node) FixFingers() {
	for {
		time.Sleep((timeInterval + 2) * time.Second)

		for next := 0; next < utils.M; next++ {
			// Calculate the start of the finger interval
			start := (n.ID + int(math.Pow(2, float64(next)))) % int(math.Pow(2, float64(utils.M)))

			// fmt.Printf("[NODE-%d] Fixing finger %d for key %d\n", n.ID, next, start)
			// Find and update successor for this finger
			message := Message{ID: start}
			var reply Message
			err := n.FindSuccessor(message, &reply)
			if err != nil {
				fmt.Printf("[NODE-%d] Failed to find successor for finger %d: %v\n", n.ID, next, err)
				continue
			}
			// fmt.Printf("[NODE-%d] Found successor for key %d: %v\n", n.ID, start, reply.ID)

			n.Lock.Lock()
			n.FingerTable[next] = Pointer{ID: reply.ID, IP: reply.IP}
			n.Lock.Unlock()
		}
	}
}

// Add the GetNodeInfo method here
func (n *Node) GetNodeInfo(args struct{}, reply *NodeInfo) error {
	n.Lock.Lock()
	defer n.Lock.Unlock()
	reply.ID = n.ID
	reply.IP = n.IP
	reply.Successor = n.Successor
	return nil
}

// Potential failure: When the find successor function is called, it should check if the find successor is alive or not
// If the find successor is not alive, it should keeping checking the next successor until it finds an alive one(?)
func (n *Node) updateSuccessorList() {
	next := Pointer{n.ID, n.IP}
	n.SuccessorList = []Pointer{}
	for i := 0; i < r; i++ {
		successorInfo, err := CallRPCMethod(next.IP, "Node.GetSuccessor", Message{})
		if err != nil {
			fmt.Printf("[NODE-%d] Failed to get successor %d: %v\n", n.ID, i, err)
			break
		}
		next = Pointer{ID: successorInfo.ID, IP: successorInfo.IP}
		n.SuccessorList = append(n.SuccessorList, next)
	}
}

// func (n *Node) updateSuccessorList() {
// 	n.Lock.Lock()
// 	defer n.Lock.Unlock()

// 	// Join condition: If the successor list is empty,
// 	if len(n.SuccessorList) == 0 {
// 		n.createSuccessorList()
// 		return
// 	}

// 	next := n.Successor
// 	successorInfo, err := CallRPCMethod(next.IP, "Node.GetSuccessorList", Message{})
// 	if err != nil {
// 		fmt.Printf("[NODE-%d] Failed to get successor list: %v\n", n.ID, err)
// 		fmt.Printf("[NODE-%d] Looking through the successor list for the next alive successor...\n", n.ID)
// 		next = n.findNextAlive()
// 		if next == (Pointer{}) {
// 			fmt.Printf("[NODE-%d] No Successor from the successor list is alive.", n.ID)
// 		}
// 		successorInfo, err = CallRPCMethod(next.IP, "Node.GetSuccessorList", Message{})
// 	}
// 	n.SuccessorList = successorInfo.SuccessorList[:r - 1] // Removing the last element
// 	n.SuccessorList = append([]Pointer{next}, n.SuccessorList...) // prepending the successor ID

// 	fmt.Printf("[NODE-%d] Updated successor list: %v\n", n.ID, n.SuccessorList)
// }

func (n *Node) findNextAlive() Pointer {
	for i := 1; i < r; i++ {
		reply, err := CallRPCMethod(n.SuccessorList[i].IP, "Node.Ping", Message{})
		if err == nil && reply != nil {
			return n.SuccessorList[i]
		}
	}
	return Pointer{}
}

func CreateNode(ip string) *Node {
	id := utils.Hash(ip) % int(math.Pow(2, float64(utils.M))) // Ensure ID is within [0, 2^m - 1]

	node := &Node{
		ID:            id,
		IP:            ip,
		Successor:     Pointer{ID: id, IP: ip},
		Predecessor:   Pointer{},
		FingerTable:   make([]Pointer, utils.M),
		SuccessorList: make([]Pointer, 0),
		Lock:          sync.Mutex{},
	}

	// Initialize finger table with self to prevent nil entries
	for i := 0; i < utils.M; i++ {
		node.FingerTable[i] = Pointer{ID: node.ID, IP: node.IP}
	}

	return node
}

func CallRPCMethod(ip string, method string, message Message) (*Message, error) {
	client, err := rpc.Dial("tcp", ip)
	if err != nil {
		return &Message{}, fmt.Errorf("[NODE-%d] Failed to connect to node at %s: %v", message.ID, ip, err)
	}
	defer client.Close()

	var reply Message
	err = client.Call(method, message, &reply)
	if err != nil {
		return &Message{}, fmt.Errorf("[NODE-%d] Failed to call method %s: %v", message.ID, method, err)
	}

	return &reply, nil
}

func removeChunksFromLocal(dataDir string, chunks []ChunkInfo) {
	for _, chunk := range chunks {
		chunkFilePath := filepath.Join(dataDir, chunk.ChunkName)
		err := os.Remove(chunkFilePath)
		if err != nil {
			fmt.Printf("Error deleting chunk file %s: %v\n", chunk.ChunkName, err)
		}
		//else {
		// 	fmt.Printf("Deleted chunk file %s from local storage.\n", chunk.ChunkName)
		// }
	}
	fmt.Printf("Deleted chunk files from local storage.\n")
}

func (n *Node) CheckPredecessor() {
	for {
		time.Sleep((timeInterval - 4) * time.Second)
		if n.Predecessor != (Pointer{}) {
			// Try to ping the predecessor
			_, err := CallRPCMethod(n.Predecessor.IP, "Node.Ping", Message{})
			if err != nil {
				fmt.Printf("[NODE-%d] Predecessor (Node-%d) appears to be down: %v\n", n.ID, n.Predecessor.ID, err)

				// Clear predecessor pointer
				n.Lock.Lock()
				n.Predecessor = Pointer{}
				n.Lock.Unlock()

				fmt.Printf("[NODE-%d] Predecessor pointer cleared\n", n.ID)
			}
		}
	}
}

func (n *Node) Ping(message Message, reply *Message) error {
	*reply = Message{
		ID: n.ID,
		IP: n.IP,
	}
	return nil
}
