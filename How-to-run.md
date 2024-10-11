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

