#!/bin/bash

: > critical.txt # empty the file

for i in {0..3}
do 
    (go run node/node.go $i > "log$i.txt" 2>&1) & 
done
