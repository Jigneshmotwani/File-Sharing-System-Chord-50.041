package node

import (
    "fmt"
    "net/rpc"
    // "strconv"
    "time"
)

// RPC argument and reply structures
type NotifyArgs struct {
    Node *RemoteNode
}

type EmptyArgs struct{}
type EmptyReply struct{}

// StartStabilization starts the stabilization routine
func (n *Node) StartStabilization() {
    ticker := time.NewTicker(5 * time.Second)
    for range ticker.C {
        n.Stabilize()
    }
}

// Stabilize updates the successor and notifies the successor of this node
func (n *Node) Stabilize() {
    n.mutex.Lock()
    successor := n.Successor
    n.mutex.Unlock()

    client, err := rpc.Dial("tcp", successor.Address())
    if err != nil {
        fmt.Printf("Error dialing successor %s - %v\n", successor.Address(), err)
        return
    }
    defer client.Close()

    var x RemoteNode
    err = client.Call("Node.GetPredecessorRPC", &EmptyArgs{}, &x)
    if err == nil && x.ID != nil {
        n.mutex.Lock()
        if between(n.ID, successor.ID, x.ID, n.KeySize) {
            n.Successor = &x
            fmt.Printf("Node %x updated successor to %x\n", n.ID, n.Successor.ID)
        }
        n.mutex.Unlock()
    }

    // Notify successor
    args := &NotifyArgs{Node: &RemoteNode{ID: n.ID, IP: n.IP, Port: n.Port}}
    var reply EmptyReply
    err = client.Call("Node.NotifyRPC", args, &reply)
    if err != nil {
        fmt.Printf("Error notifying successor %s - %v\n", successor.Address(), err)
    }
}

// GetPredecessorRPC returns the node's predecessor
func (n *Node) GetPredecessorRPC(args *EmptyArgs, reply *RemoteNode) error {
    n.mutex.Lock()
    defer n.mutex.Unlock()
    if n.Predecessor != nil {
        *reply = *n.Predecessor
    } else {
        reply.ID = nil
    }
    return nil
}

// NotifyRPC is called by other nodes to potentially update the predecessor
func (n *Node) NotifyRPC(args *NotifyArgs, reply *EmptyReply) error {
    n.mutex.Lock()
    defer n.mutex.Unlock()
    if n.Predecessor == nil || between(n.Predecessor.ID, n.ID, args.Node.ID, n.KeySize) {
        n.Predecessor = args.Node
        fmt.Printf("Node %x updated predecessor to %x\n", n.ID, n.Predecessor.ID)
    }
    return nil
}
