# Distributed P2P File Transfer System

## Building the Project

From the root of your project directory (`distributed-chord/`), run:

```bash
go build -o chordnode main.go
```

This command compiles your application and produces an executable named `chordnode`.

---

## Running the Nodes

To simulate a distributed network, you'll run multiple instances of your node program, each in a separate terminal window or background process.

### 1. Start the First Node (Bootstrap Node)

In your first terminal window:

```bash
./chordnode -ip=127.0.0.1 -port=8000 -m=5
```

- **Parameters**:
  - `-ip`: The IP address the node listens on (use `127.0.0.1` for local testing).
  - `-port`: The port number (ensure it's unique for each node).
  - `-m`: Key space size in bits (e.g., `5` for a small key space suitable for testing).

This node creates a new Chord ring.

### 2. Start Additional Nodes

In additional terminal windows, start more nodes to join the network via the bootstrap node.

#### Second Node

```bash
./chordnode -ip=127.0.0.1 -port=8001 -join=127.0.0.1:8000 -m=5
```

#### Third Node

```bash
./chordnode -ip=127.0.0.1 -port=8002 -join=127.0.0.1:8000 -m=5
```

- **Parameter `-join`**: Specifies the address of an existing node to join (format: `ip:port`).

### 3. Observe Node Behavior

Each node should output logs indicating their activities, such as:

- Node creation or joining the network.
- Finger table initialization and updates.
- Stabilization processes.
- Successor and predecessor updates.

---

### 4. Use the `put` and `get` Commands

You can now store and retrieve key-value pairs in the network using the `put` and `get` commands.

#### Store a Key-Value Pair

In **Terminal 1**, enter:

```plaintext
Enter command (put/get/exit):
put
Enter key:
apple
Enter value:
red
Key stored successfully.
```

The node will hash the key and store the value on the appropriate node in the network.

#### Retrieve a Key-Value Pair

In **Terminal 2**, enter:

```plaintext
Enter command (put/get/exit):
get
Enter key:
apple
Retrieved value: red
```

Even though you stored the key in Terminal 1, you can retrieve it from Terminal 2.

This demonstrates the node lookup functionality across the network.

#### Store Another Key-Value Pair

In **Terminal 2**, enter:

```plaintext
Enter command (put/get/exit):
put
Enter key:
banana
Enter value:
yellow
Key stored successfully.
```

#### Retrieve the New Key-Value Pair

In **Terminal 3**, enter:

```plaintext
Enter command (put/get/exit):
get
Enter key:
banana
Retrieved value: yellow
```

### 5. Exit the Nodes

To gracefully exit a node, enter:

```plaintext
Enter command (put/get/exit):
exit
Exiting...
```

Repeat this in each terminal when you're done testing.

---

Understanding the Node Lookup Process:

- **Hashing Keys**: When you `put` or `get` a key, the key is hashed using SHA-1 and mapped into the key space `[0, 2^m)`.
- **Finding the Responsible Node**: The node uses the `findSuccessor` function to locate the node responsible for the key's hashed ID.
- **Data Storage and Retrieval**: The key-value pair is stored on or retrieved from the node responsible for that key.