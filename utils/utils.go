package utils

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"
)

const (
	M = 5 // number of bits to consider for the final hash value or number of rows in the finger table
)

var (
	transferStartTime time.Time
	transferMutex     sync.RWMutex
)

// Add hash function here
func Hash(s string) int {
	h := sha1.New()
	h.Write([]byte(s))

	hashBytes := h.Sum(nil)

	unModdedID := binary.BigEndian.Uint64(hashBytes[:8]) // Gets the first 8 bytes of the hash
	moddedID := unModdedID % (1 << M)                    // Mod the ID by 2^m

	return int(moddedID)
}

func GetContainerIP() (string, error) {
	// Getting IP address of the container from eth0 interface
	iface, err := net.InterfaceByName("eth0")
	if err != nil {
		return "", fmt.Errorf("failed to get eth0 interface: %v", err)
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return "", fmt.Errorf("failed to get addresses for eth0: %v", err)
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if ipv4 := ipnet.IP.To4(); ipv4 != nil {
				return ipv4.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no IPv4 address found for eth0")
}

func Between(id int, a int, b int, equalsTo bool) bool {
	if a == b {
		return true
	}
	if a < b {
		if equalsTo {
			return id > a && id <= b
		}
		return id > a && id < b
	} else {
		if equalsTo {
			return id > a || id <= b
		}
		return id > a || id < b
	}
}

// StartTransferTimer sets the start time for file transfer
func StartTransferTimer() {
	transferMutex.Lock()
	defer transferMutex.Unlock()
	transferStartTime = time.Now()
}

// GetTransferDuration returns the duration since transfer started
func GetTransferDuration() time.Duration {
	transferMutex.RLock()
	defer transferMutex.RUnlock()
	return time.Since(transferStartTime)
}

// GetTransferStartTime returns the transfer start time
func GetTransferStartTime() time.Time {
	transferMutex.RLock()
	defer transferMutex.RUnlock()
	return transferStartTime
}
