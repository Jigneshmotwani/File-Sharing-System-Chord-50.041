package node

type Message struct {
	Type       string
	ID         int
	IP         string
	ChunkInfos []ChunkInfo
	Payload    []byte
}
