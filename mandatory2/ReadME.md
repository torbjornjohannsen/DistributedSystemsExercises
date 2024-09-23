# Mandatory Activity 2, TCP/IP Handshake 

## a) What are the packages in your implementation? What data structure do you use to transmit data and meta-data?
We have created a data structure called Message to represent packages, which has the following properties/meta-data:
- sequence_num (int): a number that is used for the TLS handshake to ensure that the other party correctly parsed the last message, and to identify packets in data transfer
- acknowledgement (int): in the TLS handshake it confirms that the message was received, by adding one to the sequence number
- syn(bool): a flag used for the TLS handshake
- ack(bool): a flag used for the TLS handshake
- fin(bool): a flag indicating the end of the conncetion, ie that asking that it be shut down gently by the other party
- window_size(int): the size of how much data the client can send without getting an ackowledgment in the sliding window model
- segmet_size(int): the size of the data segment(ie how many bytes)
- data([rune]): the actual data we wish to transmit(null in the TLS handshake) 

## b) Does your implementation use threads or processes? Why is it not realistic to use threads?
Our implementation uses threads. This way the send (waiting for an acknowledgments from the server) and the receive (handling messages coming from the server) functions can work simultaneously.


It's not realistic because communication has 0 latency and packages never get lost. We use channels to communicate between the threads, which also ensures that packages are never out of order. You can't expect the same of the real internet, where a package has to be rerouted to tons of different computers, all of whom can fail or have delays. 

## c) In case the network changes the order in which messages are delivered, how would you handle message re-ordering?
We use the lost segment heuristic from the lecture where the send function assigns a sequence number to the message, which is unique for each message. The receive function checks whether that number matches the expected number, if it does not, then the message is temporarily ignored until the missing one are received.


The server on the other hand will ship off the entirety of the sliding window, and if it reaches the end before it has acknowledgement from the client that the client got the first message, it will "restart" sending the window. 

## d) In case messages can be delayed or lost, how does your implementation handle message loss?
Our implementation handles that by creating a fucntion called getMessageWithTimeout, which triggers a resend of the message. In the TLS handshake, if any message times out, the party that doesn't recieve a message will restart the handshake, ie the server will go back to waiting for the first SYN message, and the client will send it. 


In data transfer, the client will keep asking for whatever data segment it is missing untill it gets it, and the server will restart the window as described above. 

## e) Why is the 3-way handshake important?
It is important because thanks to the 3 way handshake, both- the client and  the server know that the other party has correctly recieved and decoded their messages, due to the seq ack "dance". So basically you're sure that you're able to communicate.
