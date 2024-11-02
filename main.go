package main

import (
	"distributed-chord/node"
	"distributed-chord/utils"
	"fmt"
	"log"
	"os"
	"time"
)

func showmenu() {
	red := "\033[31m"  // ANSI code for red text
	reset := "\033[0m" // ANSI code to reset color

	fmt.Println(red + "--------------------------------" + reset)
	fmt.Println(red + "\t\tMENU" + reset)
	fmt.Println(red + "Press 1 to see the fingertable" + reset)
	fmt.Println(red + "Press 2 to see the successor and predecessor" + reset)
	fmt.Println(red + "Press 3 to do the file transfer" + reset)
	fmt.Println(red + "Press 4 to see the menu" + reset)
	fmt.Println(red + "--------------------------------" + reset)
}

func main() {
	joinAddr := os.Getenv("BOOTSTRAP_ADDR")
	chordPort := os.Getenv("CHORD_PORT")

	// fca.Chunker()
	// fca.Assembler(fca.ChunkInfo{[]string{"Node2", "Node4", "Node1"}, "C__Users_saran_Documents_GitHub_File-Sharing-System-Chord-50.041_Data_Node1_file"})

	// chunkInfo := fca.Chunker()
	// fmt.Printf("%v", *chunkInfo)

	// Print the name of the original file and the node locations for each chunk
	// if chunkInfo != nil {
	// 	fmt.Printf("%v", chunkInfo)
	// } else {
	// 	fmt.Println("No chunks were created.")
	// }

	containerIP, err := utils.GetContainerIP()

	fmt.Printf("Container IP: %s\n", containerIP)
	if err != nil {
		log.Fatalf("Failed to get container IP: %v", err)
	}
	n := node.CreateNode(containerIP + ":" + chordPort)

	fmt.Printf("Node %d created\n", n.ID)

	// fca.Assembler(fca.ChunkInfo{[]string{"Node2", "Node4", "Node1"}, "C__Users_saran_Documents_GitHub_File-Sharing-System-Chord-50.041_Data_Node1_file"})
	// Start the RPC server
	go n.StartRPCServer()
	// Check if the node is a bootstrap node or not

	if joinAddr != "" {
		// Join the network
		n.Join(joinAddr)
	}
	// else {
	// 	// Startup the bootstrap node
	// 	n.StartBootstrap()
	// }

	go n.Stabilize()
	go n.FixFingers()

	showmenu()

	for {
		var choice int
		fmt.Print("Enter choice:")
		fmt.Scan(&choice)
		time.Sleep(1 * time.Second)

		switch choice {
		case 1:
			fmt.Println("Finger Table:", n.FingerTable)
		case 2:
			fmt.Printf("Successor: %v, Predecessor: %v\n", n.Successor, n.Predecessor)
		case 3:
			var targetNodeIP, fileName string
			fmt.Print("Enter the IP address of the target node: ")
			fmt.Scan(&targetNodeIP)
			// time.Sleep(5 * time.Second)
			fmt.Print("Enter the file name to transfer: ")
			fmt.Scan(&fileName)
			// time.Sleep(5 * time.Second)
			fmt.Printf("File transfer initiated successfully.\n")
			fmt.Printf("File Name: %s, Target Node IP: %s\n", fileName, targetNodeIP)
			// time.Sleep(5 * time.Second)

			// Call a function to handle the file transfer (implement this function in node package)
			// err := n.TransferFile(targetNodeIP, fileName)
			// if err != nil {
			// 	fmt.Printf("File transfer failed: %v\n", err)
			// } else {
			// 	fmt.Println("File transfer initiated successfully.")
			// }
		case 4:
			showmenu()
		default:
			fmt.Println("Invalid choice")
		}
		time.Sleep(5 * time.Second)
	}
	// // Join the network or create a new one
	// if *joinAddr != "" {
	// 	fmt.Printf("Node %x joining the network via %s...\n", n.ID, *joinAddr)
	// 	err = n.JoinNetwork(*joinAddr)
	// 	if err != nil {
	// 		log.Fatalf("Failed to join network: %v", err)
	// 	}
	// } else {
	// 	fmt.Printf("Node %x creating a new network...\n", n.ID)
	// 	n.InitializeFingerTable()
	// }

	// // Start periodic tasks (finger table updates and stabilization)

	// Periodically start tasks every 30 seconds
	// go func() {
	// 	ticker := time.NewTicker(30 * time.Second)
	// 	defer ticker.Stop()

	// 	for {
	// 		select {
	// 		case <-ticker.C:
	// 			n.StartPeriodicTasks()
	// 			// n.startFingerTableUpdater()
	// 			// Add other periodic tasks here if needed
	// 		}
	// 	}
	// }()

	// go func() {
	// 	for {
	// 		var command, key, value string
	// 		fmt.Println("Enter command (put/get/exit):")
	// 		fmt.Scanln(&command)
	// 		switch command {
	// 		case "put":
	// 			fmt.Println("Enter key:")
	// 			fmt.Scanln(&key)
	// 			fmt.Println("Enter value:")
	// 			fmt.Scanln(&value)
	// 			err := n.PutKey(key, value)
	// 			if err != nil {
	// 				fmt.Printf("Error putting key: %v\n", err)
	// 			} else {
	// 				fmt.Println("Key stored successfully.")
	// 			}
	// 		case "get":
	// 			fmt.Println("Enter key:")
	// 			fmt.Scanln(&key)
	// 			value, err := n.GetKey(key)
	// 			if err != nil {
	// 				fmt.Printf("Error getting key: %v\n", err)
	// 			} else {
	// 				fmt.Printf("Retrieved value: %s\n", value)
	// 			}
	// 		case "exit":
	// 			fmt.Println("Exiting...")
	// 			os.Exit(0)
	// 		default:
	// 			fmt.Println("Unknown command.")
	// 		}
	// 	}
	// }()

	// Keep the main function running
	// select {}
}

// var n *node.Node

// func main() {
// 	joinAddr := os.Getenv("BOOTSTRAP_ADDR")
// 	log.Printf("Join address: %s", joinAddr)
// 	ip, err := GetContainerIP()

// 	if err != nil {
// 		log.Fatalf("Failed to get container IP: %v", err)
// 	}

// 	log.Printf("Node starting at %s", ip)

// 	n = node.NewNode(ip)

// 	log.Printf("Node %d created", n.ID)

// 	if joinAddr != "" {
// 		for i := 0; i < 5; i++ {
// 			// err := attemptJoin(joinAddr)
// 			if err == nil {
// 				log.Printf("Node %d joined the network via %s", n.ID, joinAddr)
// 				break
// 			}
// 			log.Printf("Retrying join after failure: %v", err)
// 			time.Sleep(time.Duration(i+1) * time.Second) // Exponential backoff
// 		}
// 	} else {
// 		n.Join(nil)
// 	}

// 	go n.Stabilize()
// 	go n.FixFingers()
// 	go n.CheckPredecessor()

// 	log.Printf("Node %d started", n.ID)

// 	log.Printf("Node %d starting HTTP server at %s", n.ID, ip)
// 	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:8080", ip), nil))
// }

// func GetContainerIP() (string, error) {
// 	// Getting IP address of the container from eth0 interface
// 	iface, err := net.InterfaceByName("eth0")
// 	if err != nil {
// 		return "", fmt.Errorf("failed to get eth0 interface: %v", err)
// 	}

// 	addrs, err := iface.Addrs()
// 	if err != nil {
// 		return "", fmt.Errorf("failed to get addresses for eth0: %v", err)
// 	}

// 	for _, addr := range addrs {
// 		if ipnet, ok := addr.(*net.IPNet); ok {
// 			if ipv4 := ipnet.IP.To4(); ipv4 != nil {
// 				return ipv4.String(), nil
// 			}
// 		}
// 	}

// 	return "", fmt.Errorf("no IPv4 address found for eth0")
// }
