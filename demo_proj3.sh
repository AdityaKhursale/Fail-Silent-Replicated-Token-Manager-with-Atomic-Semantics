#!/bin/bash

# Apologies for some hardcoding in this file
# And relative paths

# setup output directories and files
rm -rf output
mkdir -p output/scene_1_system_demo
mkdir -p output/scene_2_unauthorized_demo
mkdir -p output/scene_3_fail_silent_demo
mkdir -p output/scene_4_unavialable_token_demo

# Demonstration 1
echo
echo "-----------------"
echo "Demonstration 1: Replication"
echo "-->  Write through one node, read throuh different nodes"
echo "-->  Also shows writer can be reader"
echo "-----------------"
echo "Launching token management system"
set -x
go run tokenmanager_launcher.go > output/scene_1_system_demo/tokenmanager_launcher.txt 2>&1 &
{ set +x; } 2>/dev/null
LAUNCHER_PID=$!
# echo "Sleeping for some time to be on safe side"
sleep 15
# echo "Launched token management system"


echo "Sending Client requests"
set -x 
go run client.go -write -id 1 -name abc -low 1 -mid 5 -high 17 -host localhost -port 50051
{ set +x; } 2>/dev/null
sleep 1
set -x 
go run client.go -read -id 1 -host localhost -port 50053
{ set +x; } 2>/dev/null
sleep 1
set -x 
go run client.go -read -id 1 -host localhost -port 50056
{ set +x; } 2>/dev/null
sleep 1
set -x 
go run client.go -read -id 1 -host localhost -port 50051
{ set +x; } 2>/dev/null
sleep 1

echo "Moving server logs inside scene_1_system_demo"
mv output/server* output/scene_1_system_demo/

# echo "Closing token management system"
kill -2 $LAUNCHER_PID
sleep 5
# last minute rush fix
ps aux | grep server | grep 500 | awk '{print $2}' | xargs kill -9 $1
echo "Closed token management system"

# Demonstration 2
echo
echo "-----------------"
echo "Demonstration 2: Unauthorized Reader/Writer"
echo "-->  Demonstrates correct failure in case of unauthorized read/write requests"
echo "-----------------"
echo "Launching token management system"
set -x
go run tokenmanager_launcher.go > output/scene_2_unauthorized_demo/tokenmanager_launcher.txt 2>&1 &
{ set +x; } 2>/dev/null
LAUNCHER_PID=$!
# echo "Sleeping for some time to be on safe side"
sleep 15
# echo "Launched token management system"


echo "Sending Client requests"
set -x 
go run client.go -write -id 1 -name abc -low 1 -mid 5 -high 25 -host localhost -port 50052
{ set +x; } 2>/dev/null
sleep 1
set -x 
go run client.go -read -id 3 -host localhost -port 50051
{ set +x; } 2>/dev/null
sleep 1

echo "Moving server logs inside scene_2_unauthorized_demo"
mv output/server* output/scene_2_unauthorized_demo

# echo "Closing token management system"
kill -2 $LAUNCHER_PID
sleep 5
# last minute rush fix
ps aux | grep server | grep 500 | awk '{print $2}' | xargs kill -9 $1
echo "Closed token management system"

# Demonstration 3
echo
echo "-----------------"
echo "Demonstration 3: Fail Silent Behavior"
echo "-->  Demonstrates system behaves well even if some nodes crashes"
echo "-----------------"
echo "Launching token management system"
set -x
go run tokenmanager_launcher.go > output/scene_3_fail_silent_demo/tokenmanager_launcher.txt 2>&1 &
{ set +x; } 2>/dev/null
LAUNCHER_PID=$!
# echo "Sleeping for some time to be on safe side"
sleep 15
# echo "Launched token management system"

echo "Sending Client requests"
set -x 
go run client.go -write -id 1 -name abc -low 1 -mid 5 -high 13 -host localhost -port 50051
{ set +x; } 2>/dev/null
sleep 0.5
set -x 
go run client.go -read -id 1 -host localhost -port 50053
{ set +x; } 2>/dev/null
sleep 0.5
set -x 
ps aux | grep "server" | grep 50053
{ set +x; } 2>/dev/null
echo "Killing 50053 node"
ps aux | grep "server" | grep 50053 | awk '{print $2}' | xargs kill -9 $1
{ set +x; } 2>/dev/null
echo "Killed 50053 nodes"
set -x 
ps aux | grep "server" | grep 50053
{ set +x; } 2>/dev/null 
echo "Reading now from different node"
set -x 
go run client.go -read -id 1 -host localhost -port 50054
{ set +x; } 2>/dev/null
sleep 0.5

echo "Moving server logs inside scene_3_fail_silent_demo"
mv output/server* output/scene_3_fail_silent_demo

# echo "Closing token management system"
kill -2 $LAUNCHER_PID
sleep 5
# last minute rush fix
ps aux | grep server | grep 500 | awk '{print $2}' | xargs kill -9 $1
echo "Closed token management system"

# Demonstration 4
echo
echo "-----------------"
echo "Demonstration 4: Token not available"
echo "-->  Demonstrates when token is not available in data store"
echo "-----------------"
echo "Launching token management system"
set -x
go run tokenmanager_launcher.go > output/scene_4_unavialable_token_demo/tokenmanager_launcher.txt 2>&1 &
{ set +x; } 2>/dev/null
LAUNCHER_PID=$!
# echo "Sleeping for some time to be on safe side"
sleep 15
# echo "Launched token management system"


echo "Sending Client requests"
set -x 
go run client.go -write -id 3 -name abc -low 1 -mid 5 -high 14 -host localhost -port 50051
{ set +x; } 2>/dev/null
sleep 1
set -x 
go run client.go -read -id 3 -host localhost -port 50053
{ set +x; } 2>/dev/null
sleep 1

echo "Moving server logs inside scene_4_unavialable_token_demo"
mv output/server* output/scene_4_unavialable_token_demo

# echo "Closing token management system"
kill -2 $LAUNCHER_PID
sleep 5
# last minute rush fix
ps aux | grep server | grep 500 | awk '{print $2}' | xargs kill -9 $1
echo "Closed token management system"

echo
echo
echo "**** Check outputs at 'output' directory ****"