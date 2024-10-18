package node

import (
	"context"
	pb "distributed-chord/chord"
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
	fmt.Printf("Node %s listening on %s\n", formatID(n.ID, n.KeySize), n.Address())
	go n.grpcServer.Serve(lis)
	return nil
}

func (n *Node) JoinNetwork(existingNodeAddress string) error {
	conn, err := grpc.NewClient(existingNodeAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pb.NewChordNodeClient(conn)
	req := &pb.JoinRequest{
		Node: &pb.NodeInfo{
			Id:   n.ID,
			Ip:   n.IP,
			Port: int32(n.Port),
		},
	}

	resp, err := client.Join(context.Background(), req)
	if err != nil {
		return err
	}

	n.Successor = &RemoteNode{
		ID:   resp.Successor.Id,
		IP:   resp.Successor.Ip,
		Port: int(resp.Successor.Port),
	}
	fmt.Printf("Node %x set its successor to %x\n", n.ID, n.Successor.ID)

	n.InitializeFingerTable()

	return nil
}

// Join funciton for Grpc
func (n *Node) Join(ctx context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {
	successor, err := n.findSuccessor(req.Node.Id)
	if err != nil {
		return nil, err
	}
	return &pb.JoinResponse{
		Successor: &pb.NodeInfo{
			Id:   successor.ID,
			Ip:   successor.IP,
			Port: int32(successor.Port),
		},
	}, nil
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
	fmt.Printf("Node %x received file info: %s with chunks at %v\n", n.ID, req.Filename, req.ChunkLocations)

	// Handle the received chunk locations

	return &pb.FileChunkResponse{Success: true}, nil
}
