#!/bin/bash
trap "rm server;kill 0" EXIT

GRPC=-grpc=1

go build -o server
./server -port=8001 ${GRPC} &
./server -port=8002 ${GRPC} &
./server -port=8003 ${GRPC} -api=1 &

sleep 2
echo ">>> start test"
curl "http://localhost:9999/api?key=dsw" &
curl "http://localhost:9999/api?key=dsw" &
curl "http://localhost:9999/api?key=dsw" &

wait