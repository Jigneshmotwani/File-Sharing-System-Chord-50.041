package node

type Message struct {
	Type                string
	ID                  int
	IP                  string
	SuccessorList       []Pointer
	DataDir             string
	FileName string
	ChunkTransferParams ChunkTransferRequest
}

type FileTransferRequest struct {
	SenderIP string
	FileName string
}

// Struct to hold the chunk transfer request
type ChunkTransferRequest struct {
	ChunkName string
	Data      []byte
	Chunks    []ChunkInfo
}
