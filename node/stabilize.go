package node

import (
	"context"
	pb "distributed-chord/chord"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (n *Node) StartStabilization() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		n.Stabilize()
	}
}

func (n *Node) Stabilize() {
	n.mutex.Lock()
	successor := n.Successor
	n.mutex.Unlock()

	conn, err := grpc.NewClient(successor.Address(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("Error dialing successor %s - %v\n", successor.Address(), err)
		return
	}
	defer conn.Close()

	client := pb.NewChordNodeClient(conn)
	resp, err := client.GetPredecessor(context.Background(), &pb.EmptyRequest{})
	if err == nil && resp.Id != nil {
		x := RemoteNode{ID: resp.Id, IP: resp.Ip, Port: int(resp.Port)}
		n.mutex.Lock()
		if between(n.ID, successor.ID, x.ID, n.KeySize) {
			n.Successor = &x
			fmt.Printf("Node %x updated successor to %x\n", n.ID, n.Successor.ID)
		}
		n.mutex.Unlock()
	}

	// Notify successor
	_, err = client.Notify(context.Background(), &pb.NodeInfo{
		Id:   n.ID,
		Ip:   n.IP,
		Port: int32(n.Port),
	})
	if err != nil {
		fmt.Printf("Error notifying successor %s - %v\n", successor.Address(), err)
	}
}

// GetPredecessorGrpc returns the node's predecessor
func (n *Node) GetPredecessorß(ctx context.Context, _ *pb.EmptyRequest) (*pb.NodeInfo, error) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if n.Predecessor != nil {
		return &pb.NodeInfo{
			Id:   n.Predecessor.ID,
			Ip:   n.Predecessor.IP,
			Port: int32(n.Predecessor.Port),
		}, nil
	}
	return &pb.NodeInfo{}, nil
}

// NotifyGrpc is called by other nodes to potentially update the predecessor
func (n *Node) Notify(ctx context.Context, nodeInfo *pb.NodeInfo) (*pb.EmptyResponse, error) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if n.Predecessor == nil || between(n.Predecessor.ID, n.ID, nodeInfo.Id, n.KeySize) {
		n.Predecessor = &RemoteNode{
			ID:   nodeInfo.Id,
			IP:   nodeInfo.Ip,
			Port: int(nodeInfo.Port),
		}
		fmt.Printf("Node %x updated predecessor to %x\n", n.ID, n.Predecessor.ID)
	}
	return &pb.EmptyResponse{}, nil
}
