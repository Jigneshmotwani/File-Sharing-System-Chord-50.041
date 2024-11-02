# Distributed P2P File Transfer System

## How to run the code:

Make sure you have docker installed on your machine. If not, you can download it from [here](https://docs.docker.com/get-docker/). After installing docker, you can run the following commands to run the code:

1. Start the bootstrap node handling the peer joining and leaving:
```bash
docker compose up -d bootstrap --build
```
2. Once the bootstrap node is up and running, you can start the peer nodes: 
```bash 
docker compose up -d --scale peer=1 --build
```
If you want to scale the number of peer nodes, you can run the following command:
```bash
docker compose up -d --scale peer=5 --build
```
This command will start 5 peer nodes at once. If there is already a peer node running, this command will start 4 more peer nodes.
3. Once the bootstrap and peer nodes running, you can find the container id of each of the nodes by running the following:
```bash
docker container ls
```
Once you obtain the container ID of the bootstrap and peer nodes, you can run the following command to initiate the chord network and see the logs of the nodes:
```bash 
docker exec -it <container_id> bash
fts
```
4. Once you are done with the execution, you can stop the containers by running the following command:
```bash
docker compose down
```

