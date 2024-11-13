package node

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

const (
	dataFolder        = "/shared"   // Directory where the chunks are stored
	assembleFolder = "/assemble" // Directory where all the chunks retrieved from the nodes are stored
	outputFolder   = "/output"   // Directory where the assembled file is stored
)

// Assembler is a function that assembles the chunks of a file
func (n *Node) Assembler(message Message, reply *Message) error {

	if message.ChunkTransferParams.Chunks == nil || len(message.ChunkTransferParams.Chunks) == 0 {
		return fmt.Errorf("no chunks to assemble")
	}

	// Get the name of the first chunk to decipher the output file name and chunk template
	tempChunkFile := message.ChunkTransferParams.Chunks[0].ChunkName

	// Get the output file name from the chunk name
	outputFileName, err := getFileNames(tempChunkFile, message.ID)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	err = n.getAllChunks(message.ChunkTransferParams.Chunks)
	if err != nil {
		fmt.Printf("Error collecting chunks: %v\n", err)
		return err
	}

	err = assembleChunks(outputFileName, message.ChunkTransferParams.Chunks)
	if err != nil {
		fmt.Printf("Error assembling chunks: %v\n", err)
		fmt.Printf("Aborting assembling...\n")
		return err
	}

	fmt.Printf("File %s assembled successfully\n", outputFileName)

	// Clean up the assemble folder 
	removeChunksFromLocal(assembleFolder, message.ChunkTransferParams.Chunks)
	return nil
}

// Gets all the chunks from the nodes and compiles them into the /assemble folder.
func (n *Node) getAllChunks(chunkInfo []ChunkInfo) error {

	// Create the assemble folder if it doesn't exist
	if err := os.MkdirAll(assembleFolder, 0755); err != nil {
		return fmt.Errorf("error creating assemble folder: %v", err)
	}

	var reply Message
	for _, chunk := range chunkInfo {
		message := Message{
			ID: chunk.Key,
			ChunkTransferParams: ChunkTransferRequest{
				ChunkName: chunk.ChunkName,
			},
		}

		n.FindSuccessor(message, &reply)

		targetNode := Pointer{ID: reply.ID, IP: reply.IP}

		// Call the TCP function to get the chunk from the target node
		reply, err := CallRPCMethod(targetNode.IP, "Node.SendChunk", message)
		if err != nil {
			return fmt.Errorf("error receiving chunk from node %d during assembly: %v", targetNode.ID, err)
		}

		// Save the chunk data in the assemble directory
		destinationPath := filepath.Join(assembleFolder, chunk.ChunkName)
		err = os.WriteFile(destinationPath, reply.ChunkTransferParams.Data, 0644)
		if err != nil {
			return fmt.Errorf("error writing chunk %s to %s: %v", chunk.ChunkName, destinationPath, err)
		}
	}
	return nil
}

// Function to assemble all the chunks from the assemble folder
func assembleChunks(outputFileName string, chunks []ChunkInfo) error {
	
	// Making the output file
	if err := os.MkdirAll(outputFolder, 0755); err != nil {
		return fmt.Errorf("error creating output folder: %v", err)
	}

	outputFilePath := filepath.Join(outputFolder, outputFileName)
	outFile, err := os.Create(outputFilePath)

	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer outFile.Close()

	for i, chunk := range chunks {
		// filename-chunk
		content, err := ioutil.ReadFile(filepath.Join(assembleFolder, chunk.ChunkName))
		if err != nil {
			return fmt.Errorf("error reading chunk %s-chunk%d.txt: %v", chunk.ChunkName, int(i+1), err)
		}

		_, err = outFile.Write(content)
		if err != nil {
			return fmt.Errorf("error writing chunk %s-chunk%d.txt to output file: %v", chunk, int(i+1), err)
		}
	}

	return nil
}

func getFileNames(chunkName string, senderID int) (string, error) {
	for i, v := range chunkName {
		if v == '-' && chunkName[i+1:i+6] == "chunk" {
			return chunkName[:i] + "_from_" + strconv.Itoa(senderID) + filepath.Ext(chunkName), nil
		}
	}
	return "", fmt.Errorf("error getting output file name")
}

// SendChunk handles sending a chunk to a requesting node
func (n *Node) SendChunk(request Message, reply *Message) error {
	sourcePath := filepath.Join(dataFolder, request.ChunkTransferParams.ChunkName)

	// Read the chunk data from the shared directory
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read chunk from %s: %v", sourcePath, err)
	}

	// Send the chunk data as the reply
	*reply = Message{ChunkTransferParams: ChunkTransferRequest{
		Data: data,
	}}
	return nil
}