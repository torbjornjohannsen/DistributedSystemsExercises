### How to start the system

In a bash shell, simply run the run.sh script (make sure to make it executable first). 
By default it will start 3 nodes and 3 clients.
It will then write their logs to respective .txt files, named as nlog0, nlog1 &c for nodes, and clog0, clog1 &c for clients.
The nodes live for 100 seconds, and have an implemented 2% chance to fail each second.
This simulate and demonstrate our handling of a server node failing.

Note that the nodes all run in the background, so the script will immediately return. Waiting a minute should guarantee all nodes are finished.

You can provide an optional argument to the script which tells it how many nodes to start. For instance

    ./run.sh 5 10

would launch 5 nodes, 10 clients, and produce 5 nlog files and 10 clog files.