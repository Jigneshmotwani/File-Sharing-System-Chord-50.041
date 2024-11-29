# Project Meeting 3

## Distributed Systems: Distributed File Sharing System using Chord

**Team Members:**

- Chandrasekar Akash (1006228)
- Dorishetti Kaushik Varma (1006012)
- Jash Jignesh Veragiwala (1006185)
- Jignesh Motwani (1006178)
- Sarang Nambiar (1006181)

---

## Fault Tolerance Scenarios and Resolutions

In the Distributed File Sharing System using the Chord protocol, fault tolerance is crucial to maintaining the system's reliability and ensuring file transfer operations continue smoothly, even in the presence of node failures, network issues, or system crashes. Below are the various fault tolerance scenarios identified, along with the steps taken to resolve them:

### 1. Chord Fault Tolerance Handling

- **Scenario:** A node in the Chord network abruptly fails, disrupting the system's ability to locate keys or retrieve file chunks.
- **Resolution:**
  To address this issue, we implemented the following measures:

  - **Successor List Implementation:** Each node maintains a successor list containing the next r direct successors in the Chord ring. In our implementation, r is set to 3.
  - **Enhanced FindSuccessor Functionality:** The FindSuccessor function was modified to locate the next available (alive) successor from the successor list if the intended successor for the given key is unresponsive or has failed.
  - **Dynamic Successor List Updates:** The Stabilize function was enhanced to update the successor list during every stabilization process, ensuring that the list reflects the most recent network state.

**Data Replication for Fault Tolerance:**

The successor list is also leveraged to replicate file chunk data across nodes in the Chord network. The size of the successor list, denoted as r, corresponds to the number of replicas created for each file chunk. For simplicity in our case, we assign r equal to the successor list size (r=k).

Each chunk is replicated to the nodes in the successor list of the original holder node. In the event of a node failure:

- The replicas stored on the successors can be used to retrieve the data seamlessly.
- The next destination for a file chunk or replica can be quickly determined by iterating through the successor list, simplifying fault recovery and enhancing system reliability.

### 2. Node Joining During Chunking or Assembly

- **Scenario:** A new node joins the system while other nodes are chunking or assembling files. The new node may not yet have the required file chunks, or it may not be able to retrieve them during the assembly process.
- **Resolution:**
  - When a new node joins, it may not immediately have the required chunks if those chunks have already been distributed to other nodes. **To resolve this:** During file assembly, if the new node does not have the required chunk, it will use its successor list to search for the missing chunks across its successor nodes, continuing the search until the chunk is found or the list is exhausted. This ensures that even as new nodes join the network, the system remains resilient and can still locate and assemble files, even if they haven't yet received all the file chunks.

### 3. Node Failure

a. **During find_successor or chunking/assembly (not the target or the sender node)**

- **Scenario:** A node fails during the file chunking process (before the chunk is sent) or during file assembly (when chunks are being reconstructed).
- **Resolution:**
  - **Successor List:** If a node fails during the find_successor operation (i.e., while trying to locate the next node for storing or retrieving a chunk), the system relies on its successor list to continue the search for the correct successor node.
  - **During chunking and assembly**, if a node fails, the system checks the successor list and moves on to the next available node in the ring to locate the file chunks.

b. **Arbitrary Node That Might Hold Chunk Fails**

- **Scenario:** An Arbitrary node that might hold a chunk file fails before, during, or after the chunk is sent.
- **Resolution:**
  - **Before Chunking:** The system is designed to handle this failure by using the successor list. If the node fails before chunking starts, the system simply checks the successor list and moves the chunk to the next node in the list.
  - **During Assembly:** If a node fails during assembly (i.e., when reconstructing the file from chunks), the system retries the chunk retrieval from the successor list or asks for the replica from another node, ensuring the file is reassembled without failure.

c. **Sender Node Fails**

- **Scenario:** The sender node fails before chunking or before/during assembly processes.
- **Resolution:**
  - Before chunking, since chunking never got completed, chunks didn't get distributed to different node, and also the Chunk Info List didn't get delivered to the target node, there is nothing that can be done other than we can make the target node wait for some amount of time. If the target node doesn't receive it within this time limit then we delare the sender node to be disconnected from the network and the file transfer process is exited.
  - Before/during assembly, This is not really a problem since the chunk location information has already reached the target node and it can retrieve all the chunks. The system retries the chunk retrieval using the successor list to continue the process even if the sender node stores any chunks in its shared folder.

d. **Target Node Failure**

- **Scenario:** The Target node fails during/ before assembly processes.
- **Resolution:**
  - If the target node fails before or during chunking, the system retries the operation several times (up to a set number of retries). If the node remains unreachable, the system declares the node as dead and moves the file transfer to the next available node.
  - During assembly, if the target node fails, the file transfer is stopped, and the system deletes any partial chunks from all nodes involved. The failure is communicated via RPC calls to ensure consistency and avoid orphaned chunks.

### 4. Multiple Nodes Fail

- **Scenario:** Multiple nodes fail at the same time, disrupting the system's ability to retrieve or store file chunks.
- **Resolution (Can handle upto r node failure):**
  The system continues using its successor list to find available nodes, ensuring file chunks can still be retrieved or stored despite multiple node failures. The successor list helps maintain the integrity of the network by finding the next closest node to handle the file operations.

### 5. Sending to Non-Existent Node

- **Scenario:** A file is initiated to be sent to a node that does not exist in the network or has failed.
- **Resolution:** The system checks if this particular node exists in the system using the Find Successor before attempting to send data. This check is enforced during file sending operation, ensuring no data is sent to a non-existent node.

### 6. Other Safety Measures

a. **Timestamping File**

- **Scenario:** When a user sends multiple files with the same filename, there is a risk of conflict during the file assembly process, as the system may not differentiate between files with identical names.
- **Resolution:** To prevent such collisions, each file sent by the user is assigned a unique timestamp. By appending this timestamp to the filename, the system ensures that no two files share the same name. This measure eliminates the risk of assembly conflicts and maintains the integrity of the file-sharing process.

b. **Deleting Chunks from Local and Shared Folders**

- **Scenario:** Over time, nodes can become congested with unnecessary file chunks that are no longer needed, such as chunks left behind after file distribution or assembly. This can lead to inefficient use of storage and reduced system performance.
- **Resolution:** To mitigate this issue, the system automatically deletes file chunks from local and shared folders once the following conditions are met:
  - The file has been successfully chunked and distributed to all relevant nodes.
  - The target node has fully assembled the file.
  - If the target node dies: If the Remote Procedure Call (RPC) to the target node fails, the sender node initiates a cleanup process. This process ensures that all file chunks distributed to various nodes are deleted, preventing the accumulation of orphaned chunks in the network.

This cleanup process helps optimize storage, prevents node flooding, and ensures efficient resource utilization within the network.
