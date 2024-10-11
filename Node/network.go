package node

import (
    "fmt"
    "net"
    "net/rpc"
    // "strconv"
)

type JoinArgs struct {
    Node *RemoteNode
}

type JoinReply struct {
    Successor *RemoteNode
}

type GetArgs struct {
    Key []byte
}

type GetReply struct {
    Value string
    Found bool
}

type PutArgs struct {
    Key   []byte
    Value string
}

type PutReply struct {
    Success bool
}



// StartRPCServer starts the RPC server for the node
func (n *Node) StartRPCServer() error {
    rpc.Register(n)
    listener, err := net.Listen("tcp", n.Address())
    if err != nil {
        return err
    }
    fmt.Printf("Node %s listening on %s\n", formatID(n.ID, n.KeySize), n.Address())
    go rpc.Accept(listener)
    return nil
}


func (n *Node) Join(existingNodeAddress string) error {
    fmt.Printf("Node %x attempting to join the network via %s\n", n.ID, existingNodeAddress)
    client, err := rpc.Dial("tcp", existingNodeAddress)
    if err != nil {
        return err
    }
    defer client.Close()

    args := &JoinArgs{
        Node: &RemoteNode{
            ID:   n.ID,
            IP:   n.IP,
            Port: n.Port,
        },
    }
    var reply JoinReply
    err = client.Call("Node.JoinRPC", args, &reply)
    if err != nil {
        return err
    }

    n.Successor = reply.Successor
    fmt.Printf("Node %x set its successor to %x\n", n.ID, n.Successor.ID)

    // Initialize finger table after setting the successor
    n.InitializeFingerTable()

    return nil
}

// JoinRPC handles a join request from a new node
func (n *Node) JoinRPC(args *JoinArgs, reply *JoinReply) error {
    successor, err := n.findSuccessor(args.Node.ID)
    if err != nil {
        return err
    }
    reply.Successor = successor
    return nil
}

