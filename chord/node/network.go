package node

import (
	"context"
	"math/big"

	pb "distributed-chord/pb"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (n *Node) StartGRPCServer() error {
	lis, err := net.Listen("tcp", n.Address())
	if err != nil {
		return err
	}
	n.grpcServer = grpc.NewServer()
	pb.RegisterChordNodeServer(n.grpcServer, n)
	fmt.Printf("Node %s listening on %s\n", n.Address())
	go n.grpcServer.Serve(lis)
	return nil
}

func (n *Node) JoinNetwork(existingNodeAddress string) error {
	// Connect to the existing node
	conn, err := grpc.NewClient(existingNodeAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to existing node: %v", err)
	}
	defer conn.Close()

	client := pb.NewChordNodeClient(conn)

	// Send join request
	req := &pb.JoinRequest{
		Node: &pb.NodeInfo{
			Id:   n.id.Bytes(),
			Ip:   n.ip,
			Port: int32(n.port),
		},
	}

	// Get successor and its predecessor information
	resp, err := client.Join(context.Background(), req)
	if err != nil {
		return fmt.Errorf("join request failed: %v", err)
	}

	// Create successor node reference
	nextNode := &Node{
		id:   new(big.Int).SetBytes(resp.Successor.Id),
		ip:   resp.Successor.Ip,
		port: int(resp.Successor.Port),
	}

	// Get successor's predecessor information
	succConn, err := grpc.NewClient(fmt.Sprintf("%s:%d", nextNode.ip, nextNode.port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to successor: %v", err)
	}
	defer succConn.Close()

	succClient := pb.NewChordNodeClient(succConn)
	predInfo, err := succClient.GetPredecessor(context.Background(), &pb.EmptyRequest{})
	if err != nil {
		return fmt.Errorf("failed to get predecessor info: %v", err)
	}

	// Create predecessor node reference
	prevNode := &Node{
		id:   new(big.Int).SetBytes(predInfo.Id),
		ip:   predInfo.Ip,
		port: int(predInfo.Port),
	}

	n.mutex.Lock()
	n.predecessor = prevNode
	n.successor = nextNode
	n.mutex.Unlock()

	// Update successor's predecessor pointer
	updateSuccReq := &pb.UpdatePredecessorRequest{
		Node: &pb.NodeInfo{
			Id:   n.id.Bytes(),
			Ip:   n.ip,
			Port: int32(n.port),
		},
	}

	_, err = succClient.UpdatePredecessor(context.Background(), updateSuccReq)
	if err != nil {
		return fmt.Errorf("failed to update successor's predecessor: %v", err)
	}

	// Update predecessor's successor pointer
	predConn, err := grpc.NewClient(fmt.Sprintf("%s:%d", prevNode.ip, prevNode.port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return fmt.Errorf("failed to connect to predecessor: %v", err)
	}
	defer predConn.Close()

	predClient := pb.NewChordNodeClient(predConn)
	updatePredReq := &pb.UpdateSuccessorRequest{
		Node: &pb.NodeInfo{
			Id:   n.id.Bytes(),
			Ip:   n.ip,
			Port: int32(n.port),
		},
	}

	_, err = predClient.UpdateSuccessor(context.Background(), updatePredReq)
	if err != nil {
		return fmt.Errorf("failed to update predecessor's successor: %v", err)
	}

	fmt.Printf("Node %x joined between %x and %x\n", n.id, prevNode.id, nextNode.id)
	n.initializeFingerTables()

	// update the finger table of the successor and predecessor
	go n.predecessor.updateFingerTable()
	go n.successor.updateFingerTable()
	return nil
}

// Join is the RPC handler for when a new node wants to join the network
func (n *Node) Join(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {
	successor, err := n.findSuccessor(new(big.Int).SetBytes(req.Node.Id))
	if err != nil {
		return nil, fmt.Errorf("failed to find successor: %v", err)
	}

	return &pb.JoinResponse{
		Successor: &pb.NodeInfo{
			Id:   successor.id.Bytes(),
			Ip:   successor.ip,
			Port: int32(successor.port),
		},
	}, nil
}

func (n *Node) UpdatePredecessor(ctx context.Context, req *pb.UpdatePredecessorRequest) (*pb.UpdatePredecessorResponse, error) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.predecessor = &Node{
		id:   new(big.Int).SetBytes(req.Node.Id),
		ip:   req.Node.Ip,
		port: int(req.Node.Port),
	}
	return &pb.UpdatePredecessorResponse{}, nil
}

func (n *Node) UpdateSuccessor(ctx context.Context, req *pb.UpdateSuccessorRequest) (*pb.UpdateSuccessorResponse, error) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.successor = &Node{
		id:   new(big.Int).SetBytes(req.Node.Id),
		ip:   req.Node.Ip,
		port: int(req.Node.Port),
	}
	return &pb.UpdateSuccessorResponse{}, nil
}

// GetPredecessorGrpc returns the node's predecessor
func (n *Node) GetPredecessor(ctx context.Context, _ *pb.EmptyRequest) (*pb.NodeInfo, error) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if n.predecessor != nil {
		return &pb.NodeInfo{
			Id:   n.predecessor.id.Bytes(),
			Ip:   n.predecessor.ip,
			Port: int32(n.predecessor.port),
		}, nil
	}
	return &pb.NodeInfo{}, nil
}

func (n *Node) SendFileInfo(filename string, chunkLocations []string, remoteAddress string) error {
	conn, err := grpc.NewClient(remoteAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to remote node %s: %v", remoteAddress, err)
	}
	defer conn.Close()

	client := pb.NewChordNodeClient(conn)
	req := &pb.FileChunkInfo{
		Filename:       filename,
		ChunkLocations: chunkLocations,
	}

	resp, err := client.ReceiveFileInfo(context.Background(), req)
	if err != nil {
		return fmt.Errorf("error in sending file info to node %s: %v", remoteAddress, err)
	}

	if resp.Success {
		fmt.Printf("Successfully sent file info for %s to node %s\n", filename, remoteAddress)
	} else {
		fmt.Printf("Failed to send file info for %s to node %s\n", filename, remoteAddress)
	}

	return nil
}

func (n *Node) ReceiveFileInfo(ctx context.Context, req *pb.FileChunkInfo) (*pb.FileChunkResponse, error) {
	fmt.Printf("Node %x received file info: %s with chunks at %v\n", n.id, req.Filename, req.ChunkLocations)

	// Handle the received chunk locations

	return &pb.FileChunkResponse{Success: true}, nil
}
