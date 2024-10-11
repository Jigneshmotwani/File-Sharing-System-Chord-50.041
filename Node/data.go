package node

import (
    "crypto/sha1"
    "errors"
    "fmt"
    "math/big"
    "net/rpc"
)


func (n *Node) Put(key string, value string) error {
    keyHash := HashKey(key, n.KeySize)
    keyStr := KeyToString(keyHash, n.KeySize)
    successor, err := n.findSuccessor(keyHash)
    if err != nil {
        return err
    }

    if bytesEqual(successor.ID, n.ID) {
        // This node is responsible for the key
        n.mutex.Lock()
        n.Data[keyStr] = value
        n.mutex.Unlock()
        fmt.Printf("Node %s stored key %s locally\n", formatID(n.ID, n.KeySize), key)
    } else {
        // Send RPC to successor to store the key
        client, err := rpc.Dial("tcp", successor.Address())
        if err != nil {
            return err
        }
        defer client.Close()

        args := &PutArgs{
            Key:   keyHash,
            Value: value,
        }
        var reply PutReply
        err = client.Call("Node.PutRPC", args, &reply)
        if err != nil {
            return err
        }
    }
    return nil
}


func (n *Node) Get(key string) (string, error) {
    keyHash := HashKey(key, n.KeySize)
    keyStr := KeyToString(keyHash, n.KeySize)
    successor, err := n.findSuccessor(keyHash)
    if err != nil {
        return "", err
    }

    if bytesEqual(successor.ID, n.ID) {
        // This node is responsible for the key
        n.mutex.Lock()
        value, exists := n.Data[keyStr]
        n.mutex.Unlock()
        if exists {
            return value, nil
        } else {
            return "", errors.New("Key not found")
        }
    } else {
        // Send RPC to successor to get the key
        client, err := rpc.Dial("tcp", successor.Address())
        if err != nil {
            return "", err
        }
        defer client.Close()

        args := &GetArgs{
            Key: keyHash,
        }
        var reply GetReply
        err = client.Call("Node.GetRPC", args, &reply)
        if err != nil {
            return "", err
        }
        if reply.Found {
            return reply.Value, nil
        } else {
            return "", errors.New("Key not found")
        }
    }
}


// RPC method for Put
func (n *Node) PutRPC(args *PutArgs, reply *PutReply) error {
    keyStr := KeyToString(args.Key, n.KeySize)
    n.mutex.Lock()
    n.Data[keyStr] = args.Value
    n.mutex.Unlock()
    fmt.Printf("Node %s stored key %s via RPC\n", formatID(n.ID, n.KeySize), keyStr)
    reply.Success = true
    return nil
}

// RPC method for Get
func (n *Node) GetRPC(args *GetArgs, reply *GetReply) error {
    keyStr := KeyToString(args.Key, n.KeySize)
    n.mutex.Lock()
    value, exists := n.Data[keyStr]
    n.mutex.Unlock()
    if exists {
        reply.Value = value
        reply.Found = true
        fmt.Printf("Node %s retrieved key %s via RPC\n", formatID(n.ID, n.KeySize), keyStr)
    } else {
        reply.Found = false
    }
    return nil
}

// Helper functions
func HashKey(key string, m int) []byte {
    hash := sha1.Sum([]byte(key))
    hashInt := new(big.Int).SetBytes(hash[:])
    modulo := new(big.Int).Lsh(big.NewInt(1), uint(m)) // 2^m
    idInt := new(big.Int).Mod(hashInt, modulo)
    return idInt.Bytes()
}

func KeyToString(key []byte, m int) string {
    // For simplicity, we'll represent the key as a hex string
    return formatID(key, m)
}

func bytesEqual(a, b []byte) bool {
    return new(big.Int).SetBytes(a).Cmp(new(big.Int).SetBytes(b)) == 0
}
