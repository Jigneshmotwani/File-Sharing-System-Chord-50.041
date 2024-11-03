package node

import (
	"distributed-chord/utils"
	"fmt"
	"io"
	"net/rpc"
	"os"
	"path/filepath"
	"strings"
)

type ChunkInfo struct {
	Key       int
	ChunkName string
}

// ChunkTransferRequest represents the data needed to transfer a chunk to another node
type ChunkTransferRequest struct {
	ChunkName string
	Data      []byte
}

type ReceiveChunkInfoRequest struct {
	Chunks []ChunkInfo
}

// file-chunk1-<node_id>.txt
// if dup: file-chunk1-<node_id>-1.txt and so on

func (n *Node) Chunker(fileName string, targetNodeIP string) []ChunkInfo {
	// Paths
	dataDir := "/local" // Change if needed
	const chunkSize = 1024
	var chunks []ChunkInfo

	// checking if the file exists in the loacl file path of the docker container
	filePath := filepath.Join(dataDir, fileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("File %s does not exist in directory %s\n", fileName, dataDir)
		return nil
	} else if err != nil {
		fmt.Printf("Error checking file existence: %v\n", err)
		return nil
	}

	// Open the source file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil
	}
	defer file.Close()

	ext := filepath.Ext(fileName)
	baseName := strings.TrimSuffix(fileName, ext)

	// fmt.Println("All Perfect till here")
	// fmt.Println("Ami tumala bhalubhashi")

	buffer := make([]byte, chunkSize)
	chunkNumber := 1

	for {
		bytesRead, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			fmt.Println("Error reading file:", err)
			return nil
		}
		if bytesRead == 0 {
			break
		}

		// Create the chunk file name by appending the chunk number at the end of the sanitized path without extension
		chunkFileName := fmt.Sprintf("%s-chunk-%d-%d%s", baseName, chunkNumber, n.ID, ext)
		chunkFilePath := filepath.Join(dataDir, chunkFileName)
		err = os.WriteFile(chunkFilePath, buffer[:bytesRead], 0644)
		if err != nil {
			fmt.Printf("Error writing chunk file %s: %v\n", chunkFileName, err)
			return nil
		}

		fmt.Printf("Chunk %d written: %s\n", chunkNumber, chunkFilePath)
		// hashedKey := hashChunkName(chunkFileName)
		hashedKey := utils.Hash(chunkFileName)
		chunks = append(chunks, ChunkInfo{
			Key:       hashedKey,
			ChunkName: chunkFileName,
		})

		fmt.Println("Chunks: %s", chunks)
		chunkNumber++
	}

	fmt.Println("Sending the chunks to the receiver folder of the target node ...")
	n.send(chunks, targetNodeIP)
	fmt.Println("Send done ...")

	// Cleanup loop to delete each chunk file after transfer
	for _, chunk := range chunks {
		chunkFilePath := filepath.Join(dataDir, chunk.ChunkName)
		err := os.Remove(chunkFilePath)
		if err != nil {
			fmt.Printf("Error deleting chunk file %s: %v\n", chunk.ChunkName, err)
		} else {
			fmt.Printf("Deleted chunk file %s from local storage.\n", chunk.ChunkName)
		}
	}

	return chunks
}

// ReceiveChunk handles receiving a chunk and saving it to the shared directory
func (n *Node) ReceiveChunk(request ChunkTransferRequest, reply *string) error {
	destinationPath := filepath.Join("/shared", request.ChunkName)

	// Write the chunk data to the shared directory
	err := os.WriteFile(destinationPath, request.Data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write chunk to %s: %v", destinationPath, err)
	}

	*reply = "Chunk received successfully"
	return nil
}

func (n *Node) ReceiveChunkInfo(request ReceiveChunkInfoRequest, reply *string) error {
	fmt.Printf("Received chunk info: %+v\n", request.Chunks)
	*reply = "Chunk info received successfully"
	return nil
}

func (n *Node) send(chunks []ChunkInfo, targetNodeIP string) {
	for _, chunk := range chunks {
		var key = chunk.Key
		var chunkName = chunk.ChunkName

		message := Message{ID: key}
		var reply Message
		err := n.FindSuccessor(message, &reply)
		if err != nil {
			fmt.Printf("Failed to find successor: %v\n", err)
			continue
		}
		sendToNodeIP := reply.IP
		fmt.Printf("Sending chunk %s to node IP: %s\n", chunkName, sendToNodeIP)

		// Read the chunk data from the local directory
		chunkPath := filepath.Join("/local", chunkName)
		data, err := os.ReadFile(chunkPath)
		if err != nil {
			fmt.Printf("Failed to read chunk %s: %v\n", chunkName, err)
			continue
		}

		// Establish RPC connection to the target node
		client, err := rpc.Dial("tcp", sendToNodeIP)
		if err != nil {
			fmt.Printf("Failed to connect to node %s: %v\n", sendToNodeIP, err)
			continue
		}
		defer client.Close()

		// Create the chunk transfer request
		request := ChunkTransferRequest{
			ChunkName: chunkName,
			Data:      data,
		}
		var response string

		// Call the ReceiveChunk method on the target node
		err = client.Call("Node.ReceiveChunk", request, &response)
		if err != nil {
			fmt.Printf("Failed to send chunk %s to node %s: %v\n", chunkName, sendToNodeIP, err)
			continue
		}

		fmt.Printf("Chunk %s sent successfully to node %s\n", chunkName, sendToNodeIP)
	}

	client, err := rpc.Dial("tcp", targetNodeIP)
	if err != nil {
		fmt.Printf("Failed to connect to target node %s to send chunk info: %v\n", targetNodeIP, err)
		return
	}
	defer client.Close()

	// Create the chunk info transfer request
	chunkInfoRequest := ReceiveChunkInfoRequest{Chunks: chunks}
	var response string

	// Call the ReceiveChunkInfo method on the target node
	err = client.Call("Node.ReceiveChunkInfo", chunkInfoRequest, &response)
	if err != nil {
		fmt.Printf("Failed to send chunk info to node %s: %v\n", targetNodeIP, err)
		return
	}

	fmt.Printf("Chunk info sent successfully to node %s\n", targetNodeIP)
}
