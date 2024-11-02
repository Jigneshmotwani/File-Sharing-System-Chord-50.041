package utils

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"net"
)

const (
	m = 5 // number of bits to consider for the final hash value
)

// Add hash function here
func Hash(s string) int {
	h := sha1.New()
	h.Write([]byte(s))

	hashBytes := h.Sum(nil)

	unModdedID := binary.BigEndian.Uint64(hashBytes[:8]) // Gets the first 8 bytes of the hash
	moddedID := unModdedID % (1 << m)                    // Mod the ID by 2^m

	return int(moddedID) % m
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
