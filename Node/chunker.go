package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	chunkSize = 1024 // 1KB size limit for each chunk, change if needed
)

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

	nodeFolders, err := getNodeFolders(dataDir, sourceFolder)
	if err != nil {
		fmt.Println("Error retrieving folders:", err)
		return
	}

	if len(nodeFolders) == 0 {
		fmt.Println("No other folders found for distributing chunks.")
		return
	}

	buffer := make([]byte, chunkSize)
	chunkNumber := 1
	nodeIndex := 0

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

		destinationFolder := nodeFolders[nodeIndex]
		nodeIndex = (nodeIndex + 1) % len(nodeFolders)

		// Create the chunk file name by appending the chunk number at the end of the sanitized path without extension
		chunkFileName := fmt.Sprintf("%s-chunk%d.txt", absPathWithoutExtension, chunkNumber)
		chunkPath := filepath.Join(destinationFolder, chunkFileName)

		// Write chunk to the destination folder
		err = writeChunk(chunkPath, buffer[:bytesRead])
		if err != nil {
			fmt.Println("Error writing chunk:", err)
			return
		}

		fmt.Printf("Created %s in %s\n", chunkFileName, destinationFolder)
		chunkNumber++
	}
}

// Function to sanitize the file name, replacing invalid characters like `:` and `\`
func sanitizeFileName(path string) string {
	// Replace backslashes and colons with underscores
	replacer := strings.NewReplacer("\\", "_", ":", "_")
	return replacer.Replace(path)
}

// Function to remove the file extension from the sanitized absolute path
func removeFileExtension(path string) string {
	// Remove the extension using filepath.Ext to get the extension and strings.TrimSuffix
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
		if _, err := os.Stat(filePath); err == nil {
			// Print the full path of the selected file
			absPath, _ := filepath.Abs(filePath) // Get the absolute path
			fmt.Printf("Full path of the selected file: %s\n", absPath)
			return filePath, absPath
		} else {
			fmt.Println("File not found in the selected folder. Please try again.")
		}
	}
}

func getNodeFolders(dataDir, excludeFolder string) ([]string, error) {
	var nodeFolders []string

	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != excludeFolder {
			nodeFolders = append(nodeFolders, filepath.Join(dataDir, entry.Name()))
		}
	}

	return nodeFolders, nil
}

func writeChunk(path string, data []byte) error {
	chunkFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer chunkFile.Close()
	_, err = chunkFile.Write(data)
	return err
}
