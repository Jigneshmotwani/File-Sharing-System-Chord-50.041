#!/bin/bash

# Function to start the bootstrap node
start_bootstrap() {
    echo "Starting the bootstrap node..."
    docker compose up -d bootstrap --build
}

# Function to start peer nodes
start_peers() {
    read -p "Enter the number of peer nodes to create: " num_peers
    echo "Starting $num_peers peer nodes..."
    docker compose up -d --scale peer=$num_peers --build
}

# Function to stop all nodes
stop_all() {
    echo "Stopping all nodes..."
    docker compose down
}

# Main function to start the system
ft_start() {
    start_bootstrap
    start_peers
}

# Check if the first argument is "start"
if [ "$1" == "start" ]; then
    ft_start
elif [ "$1" == "close" ]; then
    stop_all
else
    echo "Usage: ft start | ft close"
fi