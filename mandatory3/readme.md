## Running the program 

First navigate to this directory 

    DistributedSystemsExercises/mandatory3

Run the server and any clients in seperate terminals

Make sure you run the server before the clients, otherwise they will simply fail to connect and exit 

### Running the server

Use the command  

    go run server/server.go

To launch the server, which will then live untill you manually kill it, letting clients connect and leave as they wish 

### Running the client(s)

Use the command 

    go run client/client.go

To launch the client, which will try to connect to the server, die if it can't and if it does it will send a random number of messages(at most 50, at least 0) after which it will tell the server it is leaving and the process will exit. 

Note: The client might or might not get the message announcing it's departure, due to the recieving functionality being in a different thread, which might or might not get to print the message before the calling thread exits and kills it 

### Creating logs

If you want to create logfiles, simply add 

    2> logname.txt  

after the command to run.

I.e. for the server 

    go run server/server.go > serverlog.txt

The 

    2> logname.txt

part pipes stderr(the output the log package writes to by default) to the .txt file 
