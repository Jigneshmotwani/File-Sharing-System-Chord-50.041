package node

// import (
// 	"fmt"
// 	"io"
// 	"io/ioutil"
// 	"net/rpc"
// 	"os"
// 	"path/filepath"
// 	"strconv"
// 	"strings"
// )

// // ASSUMPTIONS:
// // 1. The file is already chunked and stored in the Data folder
// // 2. The location slice contains the paths of the chunks to be assembled
// // 3. The ChunkInfo name does not contain the extension
// // 4. The output file is stored in the assemble folder with the format "<filename> from <fromNodeID>.<extension>"

// // Workflow
// // 1. Use FindSuccessor to find the node that contains the chunk using the key
// // 2. Make a RPC call to the node to get the chunk(Call jiggis function)
// // 3. Store the chunk in the assemble folder
// // 4. Repeat the process for all the chunks
// // 5. Assemble the chunks to get the original file
// // 6. OPTIONAL: Remove the chunks from the assemble folder

// // The type that recipient node will receive after the chunking is done.
// // type ChunkInfo struct {
// // 	Key       int
// // 	ChunkName string
// // }

// type ChunkRequest struct {
// 	ChunkName string
// }

// const (
// 	dataDir        = "/shared"   // Directory where the chunks are stored
// 	assembleFolder = "/assemble" // Directory where all the chunks retrieved from the nodes are stored
// 	outputFolder   = "/output"   // Directory where the assembled file is stored
// )

// // Assembler is a function that assembles the chunks of a file
// func (n *Node) Assembler(message Message, reply *Message) error {

// 	// Get the output file name from the chunk name
// 	chunkTemplate, outputFileName, err := getFileNames(message.ChunkInfos[0].ChunkName, message.ID)
// 	if err != nil {
// 		return fmt.Errorf(err.Error())
// 	}

// 	// Created the output file path
// 	outputFile := filepath.Join(outputFolder, outputFileName)

// 	err = n.getAllChunks(message.ChunkInfos)
// 	if err != nil {
// 		fmt.Printf("Error collecting chunks: %v\n", err)
// 		return err
// 	}

// 	err = assembleChunks(outputFile, chunkTemplate)
// 	if err != nil {
// 		fmt.Printf("Error assembling chunks: %v\n", err)
// 		return err
// 	}

// 	fmt.Printf("File assembled successfully: %s\n", outputFile)
// 	return nil
// }

// // Moves all the chunk files from the src to the destination folder.
// func (n *Node) getAllChunks(chunkInfo []ChunkInfo) error {

// 	// Testing out locally
// 	// for _, folder := range chunkInfo.ChunkLocations {
// 	// 	files, err := ioutil.ReadDir(filepath.Join(dataDir, folder))
// 	// 	if err != nil {
// 	// 		return fmt.Errorf("error reading directory %s: %v", folder, err)
// 	// 	}

// 	// 	for _, file := range files {
// 	// 		fileName := file.Name()
// 	// 		// TODO: Check if the chunk verification part should be improved or not
// 	// 		if (strings.Contains(fileName, "chunk")) && (fileName[:len(chunkInfo.Name)] == chunkInfo.Name){
// 	// 			srcPath := filepath.Join(filepath.Join(dataDir, folder), fileName)
// 	// 			destPath := filepath.Join(assembleFolder, fileName)

// 	// 			if err := copyFile(srcPath, destPath); err != nil {
// 	// 				return fmt.Errorf("error copying file %s: %v", srcPath, err)
// 	// 			}
// 	// 		}
// 	// 	}
// 	// }

// 	// Production code

// 	// Create the assemble folder if it doesn't exist
// 	if err := os.MkdirAll(assembleFolder, 0755); err != nil {
// 		return fmt.Errorf("error creating assemble folder: %v", err)
// 	}

// 	var reply *Message
// 	for _, chunk := range chunkInfo {
// 		message := Message{
// 			ID: chunk.Key,
// 		}

// 		n.FindSuccessor(message, reply)

// 		targetNode := Pointer{reply.ID, reply.IP}

// 		// Call the TCP function to get the chunk from the target node
// 		// PLACEHOLDER FUNCTION
// 	}

// 	return nil
// }

// func copyFile(src string, dest string) error {
// 	sourceFile, err := os.Open(src)
// 	if err != nil {
// 		return err
// 	}
// 	defer sourceFile.Close()

// 	destFile, err := os.Create(dest)
// 	if err != nil {
// 		return err
// 	}
// 	defer destFile.Close()

// 	_, err = io.Copy(destFile, sourceFile)
// 	return err
// }

// func assembleChunks(outputFile string, chunkName string) error {
// 	chunks, err := ioutil.ReadDir(assembleFolder)

// 	// Create the output file
// 	outFile, err := os.Create(outputFile)
// 	if err != nil {
// 		return fmt.Errorf("error creating output file: %v", err)
// 	}
// 	defer outFile.Close()

// 	for i, chunk := range chunks {
// 		content, err := ioutil.ReadFile(filepath.Join(assembleFolder, chunkName) + " - "strconv.Itoa(i+1) + ".txt")
// 		if err != nil {
// 			return fmt.Errorf("error reading chunk %s-chunk%d.txt: %v", chunk.Name(), int(i+1), err)
// 		}

// 		_, err = outFile.Write(content)
// 		if err != nil {
// 			return fmt.Errorf("error writing chunk %s-chunk%d.txt to output file: %v", chunk, int(i+1), err)
// 		}
// 	}

// 	return nil
// }

// func RetrieveChunk()

// func removeExtension(fileName string) string {
// 	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
// }

// func getFileNames(chunkName string, senderID int) (string, string, error) {
// 	for i, v := range chunkName {
// 		if v == '-' && chunkName[i+1:i+6] == "chunk" {
// 			return chunkName[:i+6], chunkName[:i] + " from " + strconv.Itoa(senderID) + filepath.Ext(chunkName), nil
// 		}
// 	}
// 	return "", "", fmt.Errorf("error getting output file name")
// }

// // Removes the chunks from the assemble folder
// // PROTOTYPE: This function is not final implementation
// func removeChunks() error {
// 	err := os.RemoveAll(assembleFolder)
// 	if err != nil {
// 		return fmt.Errorf("error removing assemble folder: %v", err)
// 	}
// 	return nil
// }

// // SendChunk handles sending a chunk to a requesting node
// func (n *Node) SendChunk(request ChunkRequest, reply *[]byte) error {
// 	sourcePath := filepath.Join("usr/src/app/shared", request.ChunkName)

// 	// Read the chunk data from the shared directory
// 	data, err := os.ReadFile(sourcePath)
// 	if err != nil {
// 		return fmt.Errorf("failed to read chunk from %s: %v", sourcePath, err)
// 	}

// 	// Send the chunk data as the reply
// 	*reply = data
// 	return nil
// }

// func (n *Node) receive(chunks []ChunkInfo) {
// 	for _, chunk := range chunks {
// 		var key = chunk.Key
// 		var chunkName = chunk.ChunkName

// 		message := Message{ID: key}
// 		var reply Message
// 		err := n.FindSuccessor(message, &reply)
// 		if err != nil {
// 			fmt.Printf("Failed to find successor: %v\n", err)
// 			continue
// 		}
// 		receiveFromNodeIP := reply.IP
// 		fmt.Printf("Receiving chunk %s from node IP: %s\n", chunkName, receiveFromNodeIP)

// 		// Establish RPC connection to the source node
// 		client, err := rpc.Dial("tcp", receiveFromNodeIP)
// 		if err != nil {
// 			fmt.Printf("Failed to connect to node %s: %v\n", receiveFromNodeIP, err)
// 			continue
// 		}
// 		defer client.Close()

// 		// Request the chunk data
// 		request := ChunkRequest{ChunkName: chunkName}
// 		var chunkData []byte
// 		err = client.Call("Node.SendChunk", request, &chunkData)
// 		if err != nil {
// 			fmt.Printf("Failed to receive chunk %s from node %s: %v\n", chunkName, receiveFromNodeIP, err)
// 			continue
// 		}

// 		// Save the chunk data in the assemble directory
// 		destinationPath := filepath.Join("/assemble", chunkName)
// 		err = os.WriteFile(destinationPath, chunkData, 0644)
// 		if err != nil {
// 			fmt.Printf("Failed to write chunk %s to %s: %v\n", chunkName, destinationPath, err)
// 			continue
// 		}

// 		fmt.Printf("Chunk %s received and saved to %s\n", chunkName, destinationPath)
// 	}
// }
