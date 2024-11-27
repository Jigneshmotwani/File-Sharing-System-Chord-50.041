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

### 1. Node Joining During Chunking or Assembly

- **Scenario:** A new node joins the system while other nodes are chunking or assembling files. The new node may not yet have the required file chunks, or it may not be able to retrieve them during the assembly process.
- **Resolution:**
  - When a new node joins, it may not immediately have the required chunks if those chunks have already been distributed to other nodes. **To resolve this:** During file assembly, if the new node does not have the required chunk, it will use its successor list to search for the missing chunks across its successor nodes, continuing the search until the chunk is found or the list is exhausted. This ensures that even as new nodes join the network, the system remains resilient and can still locate and assemble files, even if they haven't yet received all the file chunks.

### 2. Node Failure

a. **During find_successor or chunking/assembly (not the target or the sender node)**

- **Scenario:** A node fails during the file chunking process (before the chunk is sent) or during file assembly (when chunks are being reconstructed).
- **Resolution:**
  - **Successor List:** If a node fails during the find_successor operation (i.e., while trying to locate the next node for storing or retrieving a chunk), the system relies on its successor list to continue the search for the correct successor node.
  - **During chunking and assembly**, if a node fails, the system checks the successor list and moves on to the next available node in the ring to locate the file chunks.

b. **Node Receiving/Holding Chunk Fails**

- **Scenario:** A node holding or receiving a file chunk fails before, during, or after the chunk is sent.
- **Resolution:**
  - **Before Chunking:** The system is designed to handle this failure by using the successor list. If the node fails before chunking starts, the system simply checks the successor list and moves the chunk to the next node in the list.
  - **During Assembly:** If a node fails during assembly (i.e., when reconstructing the file from chunks), the system retries the chunk retrieval from the successor list or asks for the replica from another node, ensuring the file is reassembled without failure.

c. **Sender Node Fails**

- **Scenario:** The sender node fails during/ before assembly processes.
- **Resolution:**
  - Before assembly, a new function is introduced to receive chunk locations from the sender and trigger the assembly process in a go routine, ensuring the receiver can proceed with the assembly even if the sender node fails.
  - During assembly, the system retries the chunk retrieval using the successor list to continue the process even if the sender node is unavailable.

d. **Target Node Failure**

- **Scenario:** The Target node fails during/ before assembly processes.
- **Resolution:**
  - If the target node fails before or during chunking, the system retries the operation several times (up to a set number of retries). If the node remains unreachable, the system declares the node as dead and moves the file transfer to the next available node.
  - During assembly, if the target node fails, the file transfer is stopped, and the system deletes any partial chunks from all nodes involved. The failure is communicated via RPC calls to ensure consistency and avoid orphaned chunks.

### 3. Multiple Nodes Fail

- **Scenario:** Multiple nodes fail at the same time, disrupting the system's ability to retrieve or store file chunks.
- **Resolution (Can handle upto r node failure):**
  The system continues using its successor list to find available nodes, ensuring file chunks can still be retrieved or stored despite multiple node failures. The successor list helps maintain the integrity of the network by finding the next closest node to handle the file operations.

### 4. Sending to Non-Existent Node

- **Scenario:** A file chunk is mistakenly sent to a node that does not exist or has failed.
- **Resolution:** The system checks if this particular node exists in the system using the Find Successor before attempting to send data. This check is enforced during file sending operation, ensuring no data is sent to a non-existent node.
