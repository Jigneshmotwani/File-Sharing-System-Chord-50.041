package node

type Message struct {
	Type       string
	ID         int
	IP         string
	SuccessorList []Pointer
	ChunkTransferParams ChunkTransferRequest
}

// Struct to hold the chunk transfer request
type ChunkTransferRequest struct {
	ChunkName string
	Data      []byte
	Chunks   []ChunkInfo
}
