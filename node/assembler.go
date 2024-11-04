package node

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

// ASSUMPTIONS:
// 1. The file is already chunked and stored in the Data folder
// 2. The location slice contains the paths of the chunks to be assembled
// 3. The ChunkInfo name does not contain the extension
// 4. The output file is stored in the assemble folder with the format "<filename> from <fromNodeID>.<extension>"

// Workflow
// 1. Use FindSuccessor to find the node that contains the chunk using the key
// 2. Make a RPC call to the node to get the chunk(Call jiggis function)
// 3. Store the chunk in the assemble folder
// 4. Repeat the process for all the chunks
// 5. Assemble the chunks to get the original file
// 6. OPTIONAL: Remove the chunks from the assemble folder done

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
	chunkTemplate, outputFileName, err := getFileNames(tempChunkFile, message.ID)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	err = n.getAllChunks(message.ChunkTransferParams.Chunks)
	if err != nil {
		fmt.Printf("Error collecting chunks: %v\n", err)
		return err
	}

	err = assembleChunks(outputFileName, chunkTemplate, message.ID, filepath.Ext(tempChunkFile))
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
func assembleChunks(outputFileName string, chunkName string, sourceID int, extension string) error {
	
	// Making the output file
	if err := os.MkdirAll(outputFolder, 0755); err != nil {
		return fmt.Errorf("error creating output folder: %v", err)
	}

	outputFilePath := filepath.Join(outputFolder, outputFileName)
	outFile, err := os.Create(outputFilePath)

	chunks, err := ioutil.ReadDir(assembleFolder)

	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer outFile.Close()

	for i, chunk := range chunks {
		// filename-chunk
		content, err := ioutil.ReadFile(filepath.Join(assembleFolder, chunkName) + "-" + strconv.Itoa(i+1) + "-" + strconv.Itoa(sourceID) + extension)
		if err != nil {
			return fmt.Errorf("error reading chunk %s-chunk%d.txt: %v", chunk.Name(), int(i+1), err)
		}

		_, err = outFile.Write(content)
		if err != nil {
			return fmt.Errorf("error writing chunk %s-chunk%d.txt to output file: %v", chunk, int(i+1), err)
		}
	}

	return nil
}

func getFileNames(chunkName string, senderID int) (string, string, error) {
	for i, v := range chunkName {
		if v == '-' && chunkName[i+1:i+6] == "chunk" {
			return chunkName[:i+6], chunkName[:i] + "_from_" + strconv.Itoa(senderID) + filepath.Ext(chunkName), nil
		}
	}
	return "", "", fmt.Errorf("error getting output file name")
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