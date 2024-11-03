package node

import (
	"distributed-chord/utils"
	"fmt"
	"log"
	"math"
	"net"
	"net/rpc"
	"sync"
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
	Lock        sync.Mutex
}

type FileTransferRequest struct {
	SenderIP string
	FileName string
}

const (
	timeInterval = 5 // Time interval for stabilization and fixing fingers
)

// Starting the RPC server for the nodes
func (n *Node) StartRPCServer() {
	// Start the net RPC server
	rpc.Register(n)

	listener, err := net.Listen("tcp", n.IP)

	if err != nil {
		fmt.Printf("[NODE-%d] Error starting RPC server: %v\n", n.ID, err)
		return
	}

	defer listener.Close()

	fmt.Printf("[NODE-%d] Listening on %s\n", n.ID, n.IP)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("[NODE-%d] accept error: %s\n", n.ID, err)
			return
		}
		go rpc.ServeConn(conn)
	}
}

func (n *Node) RequestFileTransfer(targetNodeIP, fileName string) error {
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

	err = client.Call("Node.ConfirmFileTransfer", request, &response)
	if err != nil {
		return fmt.Errorf("failed to send file transfer request: %v", err)
	}

	// fmt.Printf("Response from target: %s\n", response)
	if (request.SenderIP != targetNodeIP) && (response == "es") {
		fmt.Println("Target accepted the file transfer. Initiating transfer...")
		// Add file transfer logic here (e.g., sending the file in chunks)
	} else if (request.SenderIP == targetNodeIP) && (response == "yes") {
		fmt.Println("Target accepted the file transfer. Initiating transfer...")
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
	fmt.Printf("[NODE-%d] Finding successor for %d...\n", n.ID, message.ID)
	if utils.Between(message.ID, n.ID, n.Successor.ID, true) {
		*reply = Message{
			ID: n.Successor.ID,
			IP: n.Successor.IP,
		}
		fmt.Printf("[NODE-%d] Successor found: %v\n", n.ID, reply.ID)
		return nil
	} else {
		closest := n.closestPrecedingNode(message.ID)
		if closest.ID == n.ID {
			*reply = Message{
				ID: n.ID,
				IP: n.IP,
			}
			fmt.Printf("[NODE-%d] Successor found: %v\n", n.ID, reply.ID)
			return nil
		}
		newReply, err := CallRPCMethod(closest.IP, "Node.FindSuccessor", message)
		if err != nil {
			return fmt.Errorf("[NODE-%d] Failed to call FindSuccessor: %v\n", n.ID, err)
		}
		*reply = Message{
			ID: newReply.ID,
			IP: newReply.IP,
		}
		fmt.Printf("[NODE-%d] Successor found: %v\n", n.ID, reply.ID)
		return nil
	}
}

func (n *Node) closestPrecedingNode(id int) Pointer {
	for i := utils.M - 1; i >= 0; i-- {
		if utils.Between(n.FingerTable[i].ID, n.ID, id, true) {
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

	fmt.Printf("[NODE-%d] Joining network with successor: %v\n", n.ID, reply.ID)
	n.Predecessor = Pointer{}
	n.Successor = Pointer{ID: reply.ID, IP: reply.IP}

	// Notify the node of the new successor
	message = Message{
		Type: "NOTIFY",
		ID:   n.ID,
		IP:   n.IP,
	}

	_, err = CallRPCMethod(reply.IP, "Node.Notify", message)
	if err != nil {
		log.Fatalf("[NODE-%d] Failed to notify successor: %v", n.ID, err)
	}
}

func (n *Node) Stabilize() {
	for {
		time.Sleep(timeInterval * time.Second)
		fmt.Printf("[NODE-%d] Stabilizing...\n", n.ID)
		reply, err := CallRPCMethod(n.Successor.IP, "Node.GetPredecessor", Message{})
		if err != nil {
			fmt.Printf("[NODE-%d] Failed to get predecessor: %v\n", n.ID, err)
			continue
		}

		successorPredecessor := Pointer{ID: reply.ID, IP: reply.IP}
		if (successorPredecessor != Pointer{} && utils.Between(successorPredecessor.ID, n.ID, n.Successor.ID, false)) {
			n.Successor = successorPredecessor
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
	fmt.Printf("[NODE-%d] Getting notified by node %d...\n", n.ID, message.ID)
	if (n.Predecessor == Pointer{} || utils.Between(message.ID, n.Predecessor.ID, n.ID, false)) {
		n.Predecessor = Pointer{ID: message.ID, IP: message.IP}
		fmt.Printf("[NODE-%d] Predecessor updated to %d\n", n.ID, n.Predecessor.ID)
	}
	return nil
}

func (n *Node) FixFingers() {
	// next := 0
	for {
		time.Sleep((timeInterval + 2) * time.Second)

		for next := 0; next < utils.M; next++ {
			// Safely calculate the start of finger interval
			start := (n.ID + int(math.Pow(2, float64(next)))) % int(math.Pow(2, float64(utils.M)))

			fmt.Printf("[NODE-%d] Fixing finger for key %d \n", n.ID, start)
			// Find and update successor for this finger
			message := Message{ID: start}
			var reply Message
			err := n.FindSuccessor(message, &reply)
			fmt.Printf("[NODE-%d] Found successor for key %d: %v\n", n.ID, start, reply.ID)

			if err != nil {
				fmt.Printf("", n.ID, err)
				continue
			}

			n.Lock.Lock()
			n.FingerTable[next] = Pointer{ID: reply.ID, IP: reply.IP}
			n.Lock.Unlock()

			// n.Lock.Lock()
			// n.FingerTable[next] = Pointer{ID: reply.ID, IP: reply.IP}
			// next = (next + 1) % m
			// n.Lock.Unlock()
		}
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
		Predecessor: Pointer{},
		FingerTable: make([]Pointer, utils.M),
		Lock:        sync.Mutex{},
	}

	return node
}

func CallRPCMethod(ip string, method string, message Message) (*Message, error) {
	client, err := rpc.Dial("tcp", ip)
	if err != nil {
		return &Message{}, fmt.Errorf("[NODE-%d] Failed to connect to node: %v", message.ID, err)
	}
	defer client.Close()

	var reply *Message
	err = client.Call(method, message, &reply)
	if err != nil {
		return &Message{}, fmt.Errorf("[NODE-%d] Failed to call method: %v", message.ID, err)
	}

	return reply, nil
}
