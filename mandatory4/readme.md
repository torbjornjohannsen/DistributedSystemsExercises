### How to start the system

In a bash shell, simply run the run.sh script(Make sure to make it executable first). By default it will start 4 nodes and write their logs to respective .txt files, named as log0.txt, log1.txt ect. The nodes live for ~30 seconds. 

Note that the nodes all run in the background, so the script will immediately return. Waiting a minute should guarantee all nodes are finished. 

You can provide an optional argument to the script which tells it how many nodes to start. For instance 

    ./run.sh 10

would launch 10 nodes, and produce 10 logfiles. 

In addition to this the critical section in the form of a critical.txt file will also be created(or overwritten if it already exists.)
