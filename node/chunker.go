package node

import (
	"distributed-chord/utils"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ChunkInfo struct {
	Key       int
	ChunkName string
}

func (n *Node) Chunker(fileName string, targetNodeIP string) []ChunkInfo {
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
		hashedKey := utils.Hash(chunkFileName)
		chunks = append(chunks, ChunkInfo{
			Key:       hashedKey,
			ChunkName: chunkFileName,
		})

		fmt.Printf("Chunks: %v\n", chunks)
		chunkNumber++
	}

	fmt.Println("Sending the chunks to the receiver folder of the target node ...")
	n.send(chunks, targetNodeIP)

	fmt.Printf("Chunk info sent to the target node at %s\n", targetNodeIP)

	// Send the chunk info to the target node for assembling
	message := Message{
		ID: n.ID,
		ChunkTransferParams: ChunkTransferRequest{
			Chunks: chunks,
		},
	}

	_, err = CallRPCMethod(targetNodeIP, "Node.Assembler", message)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Printf("Chunks have been successfully assembled at the target node\n")

	// Cleanup loop to delete each chunk file after transfer
	removeChunksFromLocal(dataDir, chunks)

	return chunks
}

// ReceiveChunk handles receiving a chunk and saving it to the shared directory
func (n *Node) ReceiveChunk(request Message, reply *string) error {
	destinationPath := filepath.Join("/shared", request.ChunkTransferParams.ChunkName)

	// Write the chunk data to the shared directory
	err := os.WriteFile(destinationPath, request.ChunkTransferParams.Data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write chunk to %s: %v", destinationPath, err)
	}

	*reply = "Chunk received successfully"
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

		// Create the chunk transfer request
		request := Message{
			Type: "CHUNK_TRANSFER",
			ChunkTransferParams: ChunkTransferRequest{
				ChunkName: chunkName,
				Data:      data,
			},
		}

		_, err = CallRPCMethod(sendToNodeIP, "Node.ReceiveChunk", request)
		if err != nil {
			fmt.Printf("Failed to send chunk %s to node %s: %v\n", chunkName, sendToNodeIP, err)
			continue
		}

		fmt.Printf("Chunk %s sent successfully to node %s\n", chunkName, sendToNodeIP)
	}

	fmt.Printf("Chunk info sent successfully to node %s\n", targetNodeIP)
}
