package node

import (
	"net"
	"net/rpc"
)

type NodeRPC struct {
	Node *Node
}

func (rpcNode *NodeRPC) FindSuccessor(id int, reply *Node) error {
	*reply = *rpcNode.Node.FindSuccessor(id)
	return nil
}

func (rpcNode *NodeRPC) Notify(pred *Node, reply *bool) error {
	rpcNode.Node.Notify(pred)
	*reply = true
	return nil
}

func (n *Node) StartNodeServer(node *Node) error {
	rpcNode := &NodeRPC{Node: node}
	server := rpc.NewServer()
	server.Register(rpcNode)

	listener, err := net.Listen("tcp", node.IP)
	if err != nil {
		return err
	}
	go server.Accept(listener)
	return nil
}
