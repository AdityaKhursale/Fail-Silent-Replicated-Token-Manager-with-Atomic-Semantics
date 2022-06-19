#!/bin/bash

# To check requests (what and sequence) 
grep -rnw "Request" output/server_op.txt

# To check how requests are processed and token state after request
grep -rnw "Processed" output/server_op.txt

# To check token store's status after each processed request
grep -irnw "tokenstore" output/server_op.txt 

# To check updated partial values
grep -irnw "partial" output/client_*.txt

# To check updated final values
grep -irnw "final" output/client_*.txt

# To check created sequence of tokens
grep -irnw "created" output/client_*.txt

# To check dropped sequence of tokens
grep -irnw "dropped" output/client_*.txt

# Read server output
cat output/server_op.txt

# Get processed sequence (filtered only request numbers)
grep -rnw "Processed" output/server_op.txt | cut -w -f 4 | xargs echo $1

# Check if there is any error on the server side
grep -irnw "error" output/server_op.txt

# Check if there is any error on the client side
grep -irnw "error" output/client_*.txt