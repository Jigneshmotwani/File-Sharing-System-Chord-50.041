package node

import (
    "distributed-file-sharing-chord/network"
)

type Node struct {
    ID           []byte          // Node identifier (SHA-1 hash)
    IP           string          // IP address of the node
    Port         int             // Port number
    Successor    *Node           // Pointer to successor node
    Predecessor  *Node           // Pointer to predecessor node
    FingerTable  []*FingerEntry  // Slice of finger entries
}

