if [ -z "$1" ]; then 
    maxNodes="3"
else 
    maxNodes=$1
fi

if [ -z "$2" ]; then 
    maxClients="3"
else 
    maxClients=$2
fi


for ((i = 0; i < maxNodes; i++));
do 
    go run node/node.go $i $maxNodes > "nlog$i.txt" 2>&1 & 
done

sleep 3

for ((i = 0; i < maxClients; i++));
do 
    go run client/client.go $i $maxNodes > "clog$i.txt" 2>&1 & 
done
