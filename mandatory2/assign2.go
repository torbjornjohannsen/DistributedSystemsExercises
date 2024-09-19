package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Message struct {
	sequence_num   int
	acknowledgment int
	syn            bool
	ack            bool
	fin            bool
	window_size    int
	segment_size   int
	data           *rune
}

func (msg *Message) print() string {
	return fmt.Sprintf("Seq: %d\nAck: %d\nFlags:\tSyn\tAck\tFin\n \t%t\t%t\t%t\nWindow_Size %d\nSegment_size %d\n",
		msg.sequence_num, msg.acknowledgment, msg.syn, msg.ack, msg.fin, msg.window_size, msg.segment_size)
}

func getMessageWithTimeout(msg *Message, channel chan Message, timeout int) bool {
	select {
	case *msg = <-channel:
		return true
	case <-time.After(time.Duration(timeout) * time.Millisecond):
		fmt.Println("Channel timed out")
		return false
	}
}

func tls_client(channel chan Message, timeout int) bool {
	msg := Message{rand.Int(), 0, true, false, false, 0, 0, nil}
	channel <- msg
	// get and validate message 2, with timeout
	var msg2 Message

	if !(getMessageWithTimeout(&msg2, channel, timeout) &&
		msg2.syn && !msg2.fin && msg2.ack && msg2.acknowledgment == msg.sequence_num+1) {

		fmt.Println("Invalid message: %s", msg.print())
		return false
	}

	fmt.Printf("Client recieved: %s\n", msg2.print())
	msg3 := Message{msg2.acknowledgment, msg2.sequence_num + 1, false, true, false, 0, 0, nil}
	channel <- msg3
	return true
}

func tls_server(channel chan Message, timeout int) bool {
	var msg Message
	if !(getMessageWithTimeout(&msg, channel, timeout) &&
		msg.syn && !msg.fin && !msg.ack) {

		fmt.Println("Invalid Message: %s", msg.print())
		return false
	}

	fmt.Printf("Server Recieved: %s\n", msg.print())
	msg2 := Message{rand.Int(), msg.sequence_num + 1, true, true, false, 0, 0, nil}
	channel <- msg2
	var msg3 Message
	if !(getMessageWithTimeout(&msg3, channel, timeout) &&
		!msg3.syn && !msg2.fin && msg2.ack && msg3.sequence_num == msg2.acknowledgment && msg3.acknowledgment == msg2.sequence_num+1) {

		fmt.Println("Wrong message")
		return false
	}

	fmt.Printf("Server Recieved: %s\nTLS handshake done", msg3.print())
	return true
}

func send(channel chan Message, data []rune) {
	//
	for !tls_server(channel, 1000) {
		time.Sleep(time.Duration(500) * time.Millisecond)
	}
}

func recieve(channel chan Message) {
	for !tls_client(channel, 1000) {
		time.Sleep(time.Duration(500) * time.Millisecond)
	}

	// establish window and segment size
	requestMsg := Message{0, 0, false, false, false, 4, 10, nil}
	channel <- requestMsg

}

func main() {
	channel := make(chan Message)
	go tls_client(channel, 1000)
	go tls_server(channel, 1000)

	time.Sleep(2000 * time.Millisecond)
}
