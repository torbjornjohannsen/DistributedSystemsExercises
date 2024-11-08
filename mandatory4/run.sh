#!/bin/bash


if [ -z "$1" ]; then 
    maxNodes="4"
else 
    maxNodes=$1
fi

: > critical.txt # empty the file

for ((i = 0; i < maxNodes; i++));
do 
    go run node/node.go $i $maxNodes > "log$i.txt" 2>&1 & 
done
