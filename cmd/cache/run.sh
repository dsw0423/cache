#!/bin/bash
trap "rm server;kill 0" EXIT

GRPC="-grpc=1"
API="-api=1"

go build -o server
./server -port=8001 -apiaddr=localhost:10001 ${GRPC} ${API} &
./server -port=8002 -apiaddr=localhost:10002 ${GRPC} ${API} &
./server -port=8003 -apiaddr=localhost:10003 ${GRPC} ${API} &

# sleep 2
# echo ">>> start test"
# curl "http://localhost:10001/api?key=dsw" &
# curl "http://localhost:10002/api?key=dsw" &
# curl "http://localhost:10003/api?key=dsw" &

wait