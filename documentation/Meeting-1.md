# Project Meeting 1

## Distributed Systems: Distributed File Sharing System using Chord

**Team Members:**

- Chandrasekar Akash (1006228)
- Dorishetti Kaushik Varma (1006012)
- Jash Jignesh Veragiwala (1006185)
- Jignesh Motwani (1006178)
- Sarang Nambiar (1006181)
  
---

## Problem Description

Traditional centralized file sharing systems often struggle with issues such as single points of failure, bottlenecks, and scalability limitations. These challenges can lead to decreased system reliability, slower access times, and increased operational costs, especially as the number of users and shared files grows.

To address these issues, this project aims to develop a **Distributed File Sharing System** leveraging the **Chord protocol**, which provides a peer to peer architecture where each participant contributes to the network, distributing files in a way that reduces dependency on any single node. The primary goal of this project is to develop a system capable of locating, retrieving, and storing files across a decentralized network while ensuring rapid access and high availability, even with large numbers of files and users.

Distributed systems are crucial in overcoming the limitations of centralized systems by allowing data and processing tasks to be spread across multiple nodes. In this project, the **Chord protocol** acts as the backbone of a **Distributed Hash Table (DHT)**, providing a structured way of locating and accessing files based on their unique keys. The distributed nature of Chord ensures that file sharing capabilities remain resilient, scalable, and fault tolerant.

---

## Key Features

The Distributed File Sharing System using Chord will implement several key features:

- **Correctness**: Ensures that file operations, such as storage and retrieval, are accurately executed while maintaining data integrity.
- **Scalability**: The Chord DHT protocol facilitates scalable and efficient file sharing, making it well-suited for accommodating a large number of users without compromising performance.
- **Fault Tolerance**: The system will be designed to manage node failures effectively.
- **Decentralized**: Chord based file sharing eliminates reliance on a centralized server, enabling users to exchange files directly with one another in a decentralized manner.
- **File Chunking and Distribution**: Shared files are divided into smaller chunks, which are then distributed across multiple nodes in the network. The system assembles these chunks when a node requests a file, reconstructing the complete file.

---

## Implementation Plan

In this project, we will implement the Chord protocol using the **Go programming language**. We will develop the protocol from scratch, adapting it to our specific use case of File Sharing. We will utilize the following packages from Go:

- **Crypto Package**: Provides access to the built in **SHA-1 algorithm**, which enables consistent hashing for the Chord protocol.
- **Net Package**: Facilitates file sharing among various nodes using TCP/IP sockets.

---

## Validation Plan

### Application Use Cases

1. **File Storage**:  
   This use case will validate our system’s ability to effectively store files within the Chord network. When a user uploads a file, the system will:

   - Split the file into smaller chunks.
   - Hash the ID of each chunk using the **SHA-1 algorithm**.
   - Determine their storage locations using consistent hashing.
     Validation will confirm that these chunks are correctly distributed across multiple nodes, and all stored files can be accurately located and retrieved.

2. **File Retrieval**:  
   This use case will showcase the system’s ability to retrieve files from the Chord network. When a user requests a file:

   - The system will perform a lookup using the hashed key to identify the corresponding chunks.
   - Nodes will communicate to gather the required chunks and reassemble them to reconstruct the file.  
     Validation will ensure reliable file retrieval, even in the presence of node failures.

3. **Node Joining and Leaving**:  
   This use case will simulate scenarios where nodes join or leave the Chord network. We will demonstrate how the system adapts dynamically:
   - When a new node joins, it registers itself, updates the Chord ring, and takes responsibility for a portion of the keyspace.
   - When a node leaves, we will showcase how the remaining nodes redistribute its responsibilities.  
     Validation will confirm that the system maintains data integrity and accessibility even with dynamic network changes.

---

### If Time Permits: Additional Use Case

4. **Virtual Node Creation**:  
   This use case addresses the issue of uneven data distribution in the Chord network. By implementing virtual nodes, we will:
   - Simulate multiple smaller nodes within a single physical node.
   - Distribute these virtual nodes strategically across the keyspace to balance the load.  
     Validation will demonstrate that virtual nodes help prevent data hotspots and optimize resource utilization across the Chord ring.

---

## Simulation of Distributed Nodes

Distributed nodes in the Chord Network are simulated using **Docker containers**, with each container representing an individual node. This approach enables easy scalability, allowing new nodes to be added by spinning up additional containers and removed by stopping them as needed.
