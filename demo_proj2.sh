#!/bin/bash

PORT=$1
HOST=localhost

# setup output directories and files
mkdir -p output
rm -rf output/server_op.txt
touch output/server_op.txt

# launch server
echo "Lauching Server"
set -x
go run server.go -port $PORT > output/server_op.txt 2>&1 &
{ set +x; } 2>/dev/null
SERVER_PID=$!


echo "Server Port: $PORT"
echo "Server PID: $SERVER_PID"
sleep 5
echo "Launched Sever"

# launch client requests
echo "Launching Client Requests"
i=1
while [ $i -le 10 ]
do
    rm -rf output/client_$i.txt
    touch output/client_$i.txt
    echo "Create Request: id=$i" > output/client_$i.txt
    set -x
    go run client.go -create -id $i -host $HOST -port $PORT >> output/client_$i.txt 2>&1 &
    { set +x; } 2>/dev/null
    sleep 0.5
    echo "Write Request: id=$i, -name=a$i, low=0, mid=100000, high=5000000" >> output/client_$i.txt
    set -x
    go run client.go -write -id $i -name a$i -low 0 -mid 100000 -high 5000000 -host $HOST -port $PORT >> output/client_$i.txt 2>&1 &
    { set +x; } 2>/dev/null
    sleep 0.5
    echo "Read Request: id=$i" >> output/client_$i.txt
    set -x
    go run client.go -read -id $i -host $HOST -port $PORT >> output/client_$i.txt 2>&1 &
    { set +x; } 2>/dev/null
    sleep 0.5
    echo "Drop Request: id=$i" >> output/client_$i.txt
    set -x
    go run client.go -drop -id $i -host $HOST -port $PORT >> output/client_$i.txt 2>&1 &
    { set +x; } 2>/dev/null
    let i++
done
echo "Launched Client Requests"

# wait till all 40 requests are serviced
j=1
requests=$(( 4*i-4 ))
while true
do
    PROCESSED_REQ="$(grep "Processed request number" output/server_op.txt | tail -n 1 | cut -w -f 4)"
    if [[ $PROCESSED_REQ =~ $( printf '%d' $requests ) || $j == 30 ]]; then
        echo "All requests processed or timeout reached, existing"
        break
    else
        echo "Waiting for all requests to get processed..."
        sleep 5
        let j++
    fi
done

# close server
echo "Closing Server"
sleep 2
pkill -9 -P $SERVER_PID
echo "Server closed"

echo "**** Check outputs at 'output' directory ****"