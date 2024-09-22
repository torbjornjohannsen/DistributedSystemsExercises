# Mandatory Activity 2, TCP/IP Handshake 

## a) What are the packages in your implementation? What data structure do you use to transmit data and meta-data?
We have created a data structure called Message, which has the following properties/meta-data:
-sequence_num (int):??
-acknowledgement (int): confirms that the message was received, by adding one to the sequential number
-syn(bool): a flag indicating when the client wants to connect with the server
-ack(bool): a flag indicating whether the message was received
-fin(bool): a flag indicating the end of the conncetion
-window_size(int): the size of how much data the client can send without getting an ackowledgment
-segmet_size(int): the size of the data segment
-data([rune]): a slice holding the message 

packages??

## b) Does your implementation use threads or processes? Why is it not realistic to use threads?
Our implementation uses goroutines. This way the send (waiting for an acknowledgments from the server) and the receive (handling messages coming from the server) funstions can work simultaneously.
?? why not realistic 

## c) In case the network changes the order in which messages are delivered, how would you handle message re-ordering?
The send function assigns a sequence number to the message, which is unique for each message. The receive function checks whether that number matches the expected acknowledgement number, if it does not, then the message is temporarily ignored until the missing ones are received.

## d) In case messages can be delayed or lost, how does your implementation handle message loss?
If the client does not receive the acknowledgement within some period of time, it assumes that the message is lost. Our implementation handles that by creating a fucntion called getMessageWithTimeout, which triggers a resend of the message.
(more details?)

## e) Why is the 3-way handshake important?
It is important because thanks to the 3 way handshake, both- the client and  the server are ready to communicate and update each other on their state. This ensures that the data is transmitted correctly between the client and the server.