package fca

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ASSUMPTIONS:
// 1. The file is already chunked and stored in the Data folder
// 2. The location slice contains the paths of the chunks to be assembled
// 3. Node lookup functionality will be implemented in the future
// 4. Chunk location folders already exist and contain the ALL the required chunks
// 5. The ChunkInfo name does not contain the extension

// The type that recipient node will receive after the chunking is done.
type ChunkInfo struct {
	ChunkLocations []string
	Name string // name of the original file
}

const (
	dataDir = "Data"
	assembleFolder = "Data/Node1/Assembled_Chunks"
)

// Assembler is a function that assembles the chunks of a file
func Assembler(chunkInfo ChunkInfo) error{
	
	outputFile := filepath.Join(assembleFolder, "../assembled_file.txt")
	
	// Create the assemble folder if it doesn't exist
	if err := os.MkdirAll(assembleFolder, 0755); err != nil {
		return fmt.Errorf("error creating assemble folder: %v", err)
	}

	var err error = getAllChunks(chunkInfo)
	if err != nil {
		fmt.Printf("Error collecting chunks: %v\n", err)
		return err
	}

	err = assembleChunks(outputFile, chunkInfo.Name)
	if err != nil {
		fmt.Printf("Error assembling chunks: %v\n", err)
		return err
	}

	fmt.Printf("File assembled successfully: %s\n", outputFile)
	return nil
}

// Moves all the chunk files from the src to the destination folder.
func getAllChunks(chunkInfo ChunkInfo) error{

	for _, folder := range chunkInfo.ChunkLocations {
		files, err := ioutil.ReadDir(filepath.Join(dataDir, folder))
		if err != nil {
			return fmt.Errorf("error reading directory %s: %v", folder, err)
		}

		for _, file := range files {
			fileName := file.Name()
			// TODO: Check if the chunk verification part should be improved or not
			if (strings.Contains(fileName, "chunk")) && (fileName[:len(chunkInfo.Name)] == chunkInfo.Name){
				srcPath := filepath.Join(filepath.Join(dataDir, folder), fileName)
				destPath := filepath.Join(assembleFolder, fileName)
				
				if err := copyFile(srcPath, destPath); err != nil {
					return fmt.Errorf("error copying file %s: %v", srcPath, err)
				}
			}
		}
	}

	return nil
}

func copyFile(src string, dest string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func assembleChunks(outputFile string, chunkTemplate string) error {
	chunks, err := ioutil.ReadDir(assembleFolder)

	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer outFile.Close()

	for i, chunk := range chunks {
		content, err := ioutil.ReadFile(filepath.Join(assembleFolder, chunkTemplate) + "-chunk" + strconv.Itoa(i + 1) + ".txt")
		if err != nil {
			return fmt.Errorf("error reading chunk %s-chunk%d.txt: %v", chunk.Name(), int(i + 1), err)
		}

		_, err = outFile.Write(content)
		if err != nil {
			return fmt.Errorf("error writing chunk %s-chunk%d.txt to output file: %v", chunk, int(i + 1), err)
		}
	}

	return nil
}

func removeExtension(fileName string) string{
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}