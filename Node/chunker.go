package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	chunkSize    = 1024                                                // 1KB size limit for each chunk, change if needed
	maxHashValue = "1461501637330902918203684832716283019655932542975" // 2^160 - 1
)

type Node struct {
	FolderName string
	HashValue  *big.Int
}

// NodeInterval represents the interval assigned to a node
type NodeInterval struct {
	Node       Node
	StartValue *big.Int
	EndValue   *big.Int
}

func main() {
	// Paths
	dataDir := "../Data" // Change if needed
	var sourceFolder, sourceFile, absPath string

	// Loop to allow user to switch nodes or select a file
	for {
		// Select source node folder
		sourceFolder = promptUserForFolder(dataDir)
		if sourceFolder == "" {
			fmt.Println("No valid folder selected.")
			return
		}

		sourceFile, absPath = promptUserForFile(dataDir, sourceFolder)
		if sourceFile == "-1" { // If user enters -1, switch the node
			continue
		}
		if sourceFile == "" {
			fmt.Println("No valid file selected.")
			return
		}

		break // Exit the loop after a valid file is selected
	}

	// Open the source file
	file, err := os.Open(sourceFile)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Get all the node folders and assign intervals
	nodeFolders, err := getNodeFolders(dataDir, sourceFolder)
	if err != nil {
		fmt.Println("Error retrieving folders:", err)
		return
	}

	if len(nodeFolders) == 0 {
		fmt.Println("No other folders found for distributing chunks.")
		return
	}

	// Create and assign intervals to nodes based on hashed folder names
	nodeIntervals := assignNodeIntervals(nodeFolders)

	buffer := make([]byte, chunkSize)
	chunkNumber := 1

	// Replace special characters in the absolute path to make it valid as a file name
	absPathForFileName := sanitizeFileName(absPath)

	// Remove the file extension from the absolute path
	absPathWithoutExtension := removeFileExtension(absPathForFileName)

	for {
		bytesRead, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			fmt.Println("Error reading file:", err)
			return
		}
		if bytesRead == 0 {
			break
		}

		// Create the chunk file name by appending the chunk number at the end of the sanitized path without extension
		chunkFileName := fmt.Sprintf("%s-chunk%d.txt", absPathWithoutExtension, chunkNumber)

		// Hash the chunk file name using SHA-1 and convert it to a big integer
		hashedChunkFileName := hashSHA1(chunkFileName)
		hashedChunkBigInt := hashToBigInt(hashedChunkFileName)
		fmt.Printf("Chunk File Name: %s\nSHA-1 Hash: %s\nBigInt: %s\n", chunkFileName, hashedChunkFileName, hashedChunkBigInt)

		// Find the appropriate node based on the chunk's big integer value
		assignedNode := findAssignedNode(hashedChunkBigInt, nodeIntervals)

		if assignedNode != nil {
			// Save the chunk in the assigned node's folder
			destinationFolder := filepath.Join(dataDir, assignedNode.Node.FolderName)
			chunkPath := filepath.Join(destinationFolder, chunkFileName)
			err = writeChunk(chunkPath, buffer[:bytesRead])
			if err != nil {
				fmt.Println("Error writing chunk:", err)
				return
			}
			fmt.Printf("Chunk %d assigned to node: %s\n", chunkNumber, assignedNode.Node.FolderName)
		} else {
			fmt.Println("No node found for chunk:", chunkFileName)
		}

		chunkNumber++
	}
}

// assign intervals to nodes based on hashed folder names
func assignNodeIntervals(nodeFolders []string) []NodeInterval {
	var nodes []Node

	// Create Node structs with hashed values
	for _, folder := range nodeFolders {
		hashedFolderName := hashSHA1(folder)
		hashedFolderBigInt := hashToBigInt(hashedFolderName)
		nodes = append(nodes, Node{FolderName: folder, HashValue: hashedFolderBigInt})
		// fmt.Printf("Node: %s", nodes)
	}

	// Sort nodes by their hash value
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].HashValue.Cmp(nodes[j].HashValue) < 0
	})

	var nodeIntervals []NodeInterval
	maxHashBigInt := new(big.Int)
	maxHashBigInt.SetString(maxHashValue, 10)

	for i := 0; i < len(nodes); i++ {
		start := nodes[i].HashValue
		var end *big.Int

		if i == len(nodes)-1 {
			// Last node wraps around to include [nodes[i] -> 2^160-1] and [0 -> first node]
			end = new(big.Int).Sub(maxHashBigInt, big.NewInt(1))
		} else {
			// Set end to the next node's value - 1
			end = new(big.Int).Sub(nodes[i+1].HashValue, big.NewInt(1))
		}
		nodeIntervals = append(nodeIntervals, NodeInterval{
			Node:       nodes[i],
			StartValue: start,
			EndValue:   end,
		})
		// Print the interval for all nodes, including the selected node
		fmt.Printf("Node: %s, Interval: [%s, %s]\n", nodes[i].FolderName, start.String(), end.String())
	}
	// Assignment for last node
	if len(nodes) > 0 {
		firstNodeStartValue := nodeIntervals[0].StartValue
		nodeIntervals[len(nodeIntervals)-1].EndValue = maxHashBigInt

		nodeIntervals = append(nodeIntervals, NodeInterval{
			Node:       nodes[0],
			StartValue: big.NewInt(0),
			EndValue:   firstNodeStartValue,
		})
	}
	return nodeIntervals
}

func findAssignedNode(chunkBigInt *big.Int, nodeIntervals []NodeInterval) *NodeInterval {
	for _, interval := range nodeIntervals {
		if isInInterval(chunkBigInt, interval.StartValue, interval.EndValue) {
			return &interval
		}
	}
	return nil
}

func isInInterval(value, start, end *big.Int) bool {
	if start.Cmp(end) <= 0 {
		return value.Cmp(start) >= 0 && value.Cmp(end) <= 0
	} else {
		return value.Cmp(start) >= 0 || value.Cmp(end) <= 0
	}
}

func sanitizeFileName(path string) string {
	replacer := strings.NewReplacer("\\", "_", ":", "_")
	return replacer.Replace(path)
}

func removeFileExtension(path string) string {
	ext := filepath.Ext(path)
	return strings.TrimSuffix(path, ext)
}

func promptUserForFolder(dataDir string) string {
	folders, err := getNodeFolders(dataDir, "")
	if err != nil {
		fmt.Println("Error retrieving folders:", err)
		return ""
	}

	for {
		fmt.Println("Select the source node folder:")
		for i, folder := range folders {
			fmt.Printf("[%d] %s\n", i+1, filepath.Base(folder))
		}

		var choice int
		fmt.Print("Enter your choice: ")
		fmt.Scan(&choice)

		if choice >= 1 && choice <= len(folders) {
			return filepath.Base(folders[choice-1])
		}

		fmt.Println("Invalid choice. Please select a valid node.")
	}
}
func promptUserForFile(dataDir, folder string) (string, string) {
	for {
		var fileName string
		fmt.Print("Enter the file name (with extension) to share (or enter -1 to switch node): ")
		fmt.Scan(&fileName)

		if fileName == "-1" {
			return "-1", ""
		}
		filePath := filepath.Join(dataDir, folder, fileName)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Println("Invalid file name. Please try again.")
		} else {
			absPath, _ := filepath.Abs(filePath)
			return filePath, absPath
		}
	}
}

// Function to get all node folders in the data directory
func getNodeFolders(dataDir, sourceFolder string) ([]string, error) {
	var nodeFolders []string
	err := filepath.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != dataDir {
			nodeFolders = append(nodeFolders, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return nodeFolders, nil
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

// write chunk data to a file
func writeChunk(filePath string, data []byte) error {
	return os.WriteFile(filePath, data, 0644)
}
