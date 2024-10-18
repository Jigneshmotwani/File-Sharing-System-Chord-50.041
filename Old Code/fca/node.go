package fca

import (
	"crypto/sha1"
	"encoding/hex"
	"math/big"
	"sort"
)

const (
	chunkSize        = 1024                                                // 1KB size limit for each chunk, change if needed
	maxHashValue     = "1461501637330902918203684832716283019655932542975" // 2^160 - 1
	numFingerEntries = 160                                                 // Finger table size, based on 160-bit identifiers
)

type Node struct {
	FolderName  string
	HashValue   *big.Int
	FingerTable []FingerTableEntry
}

// FingerTableEntry represents an entry in the finger table
type FingerTableEntry struct {
	Start *big.Int
	Node  *Node
}

// findSuccessor finds the successor node for a given hash
func findSuccessor(hash *big.Int, nodes []*Node) *Node {
	for _, node := range nodes {
		if hash.Cmp(node.HashValue) <= 0 {
			return node
		}
	}
	return nodes[0] // Wrap around to the first node
}

// initializeFingerTables assigns hash values and initializes finger tables for each node
func initializeFingerTables(nodeFolders []string) []*Node {
	var nodes []*Node

	// Create Node structs with hashed values
	for _, folder := range nodeFolders {
		hashedFolderName := hashSHA1(folder)
		hashedFolderBigInt := hashToBigInt(hashedFolderName)
		node := &Node{FolderName: folder, HashValue: hashedFolderBigInt}
		nodes = append(nodes, node)
	}

	// Sort nodes by their hash value
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].HashValue.Cmp(nodes[j].HashValue) < 0
	})

	// Initialize finger tables for each node
	for _, node := range nodes {
		node.FingerTable = make([]FingerTableEntry, numFingerEntries)
		for i := 0; i < numFingerEntries; i++ {
			start := new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(i)), nil)
			start.Add(start, node.HashValue)
			maxHashInt, _ := new(big.Int).SetString(maxHashValue, 10)
			start.Mod(start, maxHashInt)

			successor := findSuccessor(start, nodes)
			node.FingerTable[i] = FingerTableEntry{Start: start, Node: successor}
		}

		// Print the finger table for each node
		// fmt.Printf("\nFinger Table for Node %s (Hash: %s):\n", node.FolderName, node.HashValue.String())
		// for i, entry := range node.FingerTable {
		// 	fmt.Printf("  Entry %d: Start: %s, Successor Node: %s (Hash: %s)\n",
		// 		i,
		// 		entry.Start.String(),
		// 		entry.Node.FolderName,
		// 		entry.Node.HashValue.String())
		// }
	}

	return nodes
}

// SHA-1 hashing
func hashSHA1(data string) string {
	hasher := sha1.New()
	hasher.Write([]byte(data))
	return hex.EncodeToString(hasher.Sum(nil))
}

// hashed to big integer
func hashToBigInt(hash string) *big.Int {
	hashedBytes, _ := hex.DecodeString(hash)
	return new(big.Int).SetBytes(hashedBytes)
}
