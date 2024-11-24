package node

import (
	"distributed-chord/utils"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
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
	fmt.Println("Waiting for 30 seconds before chunking. You can now kill the sender node.")
	time.Sleep(30 * time.Second)
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
		os.Setenv("TZ", "Asia/Singapore")
		timestamp := time.Now().In(time.Local).Format("02012006_150405")
		chunkFileName := fmt.Sprintf("%s-chunk-%d-%d-%s%s", baseName, chunkNumber, n.ID, timestamp, ext)
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

	// fmt.Printf("Sending the chunk info to target node")
	// time.Sleep(5 * time.Second)

	// _, err = CallRPCMethod(targetNodeIP, "Node.Assembler", message)
	_, err = CallRPCMethod(targetNodeIP, "Node.ChunkLocationReceiver", message)
	if err != nil {
		fmt.Println(err.Error()) // Print out more beutifully
	}

	fmt.Printf("Chunk location information sent. Sender can now disconnect.\n")

	// Cleanup loop to delete each chunk file after transfer
	// removeChunksFromLocal(dataDir, chunks)

	return chunks
}

// ReceiveChunk handles receiving a chunk and saving it to the shared directory
func (n *Node) ReceiveChunk(request Message, reply *Message) error {
	destinationPath := filepath.Join("/shared", request.ChunkTransferParams.ChunkName)

	// Write the chunk data to the shared directory
	err := os.WriteFile(destinationPath, request.ChunkTransferParams.Data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write chunk to %s: %v", destinationPath, err)
	}

	*reply = Message{Type: "CHUNK_TRANSFER", ChunkTransferParams: request.ChunkTransferParams}
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

		// Get the successor list of the node
		successorReply, err := CallRPCMethod(sendToNodeIP, "Node.GetSuccessorList", Message{})
		if err != nil {
			fmt.Printf("Failed to get successor list: %v\n", err)
			continue
		}
		successorList := successorReply.SuccessorList
		fmt.Printf("Successor list: %v\n", successorList)

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

		for i := 0; i < len(successorList); i++ {
			request2 := Message{
				Type: "CHUNK_TRANSFER",
				ChunkTransferParams: ChunkTransferRequest{
					ChunkName: chunkName,
					Data:      data,
				},
			}

			successor := successorList[i]
			_, err = CallRPCMethod(successor.IP, "Node.ReceiveChunk", request2)
			fmt.Printf("Sending chunk to successor node with IP %v\n", successor.IP)
			if err != nil {
				fmt.Printf("Failed to send chunk %s to node %s: %v\n", chunkName, successor.IP, err)
				continue
			}
			fmt.Printf("Chunk %s sent successfully to node %s\n", chunkName, successor.IP)
		}

	}

	fmt.Printf("Chunk info sent successfully to node %s\n", targetNodeIP)
}

// func (n *Node) GetSuccessorList(args *Message, reply *Message) error {
// 	successorIPs := []string{}
// 	for _, successor := range n.SuccessorList {
// 		successorIPs = append(successorIPs, successor.IP)
// 	}
// 	reply.SuccessorList = successorIPs
// 	return nil
// }

func (n *Node) ChunkLocationReceiver(message Message, reply *Message) error {
	// Validate chunk information
	if message.ChunkTransferParams.Chunks == nil || len(message.ChunkTransferParams.Chunks) == 0 {
		return fmt.Errorf("no chunks to process")
	}

	// Create a copy of the chunks to pass to the goroutine
	chunksCopy := make([]ChunkInfo, len(message.ChunkTransferParams.Chunks))
	copy(chunksCopy, message.ChunkTransferParams.Chunks)
	time.Sleep(30 * time.Second)
	// Trigger assembler as a goroutine
	go func() {
		// Simulate a pause to allow demonstration of node disconnection
		// fmt.Println("Waiting for 30 seconds before assembly. You can now kill the sender node.")

		// Create a new message for the assembler
		assemblerMessage := Message{
			ID: message.ID,
			ChunkTransferParams: ChunkTransferRequest{
				Chunks: chunksCopy,
			},
		}

		var assemblerReply Message
		err := n.Assembler(assemblerMessage, &assemblerReply)
		if err != nil {
			fmt.Printf("Assembler failed: %v\n", err)
		}
	}()

	// Immediately return to allow sender to disconnect
	*reply = Message{
		Type: "CHUNK_LOCATIONS_RECEIVED",
	}
	return nil
}
