package chord

// import (
// 	"context"
// 	"crypto/sha1"
// 	pb "distributed-chord/pb"
// 	"errors"
// 	"fmt"
// 	"math/big"
// 	"distributed-chord/chord/node"
// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/credentials/insecure"
// )

// func (n *node.Node) PutKey(key string, value string) error {
// 	keyHash := HashKey(key, n.KeySize)
// 	keyStr := KeyToString(keyHash, n.KeySize)
// 	successor, err := n.findSuccessor(keyHash)
// 	if err != nil {
// 		return err
// 	}

// 	if bytesEqual(successor.ID, n.ID) {
// 		n.mutex.Lock()
// 		n.Data[keyStr] = value
// 		n.mutex.Unlock()
// 		fmt.Printf("node.Node %s stored key %s locally\n", node.(, n.KeySize), key)
// 	} else {
// 		conn, err := grpc.NewClient(successor.Address(), grpc.WithTransportCredentials(insecure.NewCredentials()))
// 		if err != nil {
// 			return err
// 		}
// 		defer conn.Close()

// 		client := pb.NewChordnode.NodeClient(conn)
// 		req := &pb.PutRequest{
// 			Key:   keyHash,
// 			Value: value,
// 		}
// 		_, err = client.Put(context.Background(), req)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func (n *node.Node) GetKey(key string) (string, error) {
// 	keyHash := HashKey(key, n.KeySize)
// 	keyStr := KeyToString(keyHash, n.KeySize)
// 	successor, err := n.findSuccessor(keyHash)
// 	if err != nil {
// 		return "", err
// 	}

// 	if bytesEqual(successor.ID, n.ID) {
// 		n.mutex.Lock()
// 		value, exists := n.Data[keyStr]
// 		n.mutex.Unlock()
// 		if exists {
// 			return value, nil
// 		} else {
// 			return "", errors.New("Key not found")
// 		}
// 	} else {
// 		conn, err := grpc.Dial(successor.Address(), grpc.WithInsecure())
// 		if err != nil {
// 			return "", err
// 		}
// 		defer conn.Close()

// 		client := pb.NewChordnode.NodeClient(conn)
// 		req := &pb.GetRequest{
// 			Key: keyHash,
// 		}
// 		resp, err := client.Get(context.Background(), req)
// 		if err != nil {
// 			return "", err
// 		}
// 		if resp.Found {
// 			return resp.Value, nil
// 		} else {
// 			return "", errors.New("Key not found")
// 		}
// 	}
// }

// // Put funtion for Grpc
// func (n *node.Node) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
// 	keyStr := KeyToString(req.Key, n.KeySize)
// 	n.mutex.Lock()
// 	n.Data[keyStr] = req.Value
// 	n.mutex.Unlock()
// 	fmt.Printf("node.Node %s stored key %s via gRPC\n", formatID(n.ID, n.KeySize), keyStr)
// 	return &pb.PutResponse{Success: true}, nil
// }

// // Get funtion for Grpc
// func (n *node.Node) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
// 	keyStr := KeyToString(req.Key, n.KeySize)
// 	n.mutex.Lock()
// 	value, exists := n.Data[keyStr]
// 	n.mutex.Unlock()
// 	if exists {
// 		fmt.Printf("node.Node %s retrieved key %s via gRPC\n", formatID(n.ID, n.KeySize), keyStr)
// 		return &pb.GetResponse{Value: value, Found: true}, nil
// 	} else {
// 		return &pb.GetResponse{Found: false}, nil
// 	}
// }

// // Helper functions
// func HashKey(key string, m int) []byte {
// 	hash := sha1.Sum([]byte(key))
// 	hashInt := new(big.Int).SetBytes(hash[:])
// 	modulo := new(big.Int).Lsh(big.NewInt(1), uint(m)) // 2^m
// 	idInt := new(big.Int).Mod(hashInt, modulo)
// 	return idInt.Bytes()
// }

// func KeyToString(key []byte, m int) string {
// 	// For simplicity, we'll represent the key as a hex string
// 	return formatID(key, m)
// }

// func bytesEqual(a, b []byte) bool {
// 	return new(big.Int).SetBytes(a).Cmp(new(big.Int).SetBytes(b)) == 0
// }
