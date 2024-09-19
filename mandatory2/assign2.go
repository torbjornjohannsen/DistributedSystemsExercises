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
	return fmt.Sprintf("Seq: %d\nAck: %d\nFlags:\tSyn\tAck\tFin\n \t%t\t%t\t%t\nWindow_Size %d\nSegment_size %d\n", msg.sequence_num, msg.acknowledgment, msg.syn, msg.ack, msg.fin, msg.window_size, msg.segment_size)
}

func client(channel chan Message) {
	msg := Message{rand.Int(), 0, true, false, false, 0, 0, nil}
	channel <- msg
	msg2 := <-channel
	if !(msg2.syn && !msg2.fin && msg2.ack && msg2.acknowledgment == msg.sequence_num+1) {
		fmt.Println("Wrong message")
		return
	}
	fmt.Printf("Client recieved: %s\n", msg2.print())
	msg3 := Message{msg2.acknowledgment, msg2.sequence_num + 1, false, true, false, 0, 0, nil}
	channel <- msg3

}

func server(channel chan Message) {
	msg := <-channel
	if !(msg.syn && !msg.fin && !msg.ack) {
		fmt.Println("Wrong message")
		return
	}
	fmt.Printf("Server Recieved: %s\n", msg.print())
	msg2 := Message{rand.Int(), msg.sequence_num + 1, true, true, false, 0, 0, nil}
	channel <- msg2
	msg3 := <-channel
	if !(!msg3.syn && !msg2.fin && msg2.ack && msg3.sequence_num == msg2.acknowledgment && msg3.acknowledgment == msg2.sequence_num+1) {
		fmt.Println("Wrong message")
		return
	}
	fmt.Printf("Server recieved: %s\n", msg.print())

}

func main() {
	channel := make(chan Message)
	go client(channel)
	go server(channel)

	time.Sleep(2000 * time.Millisecond)
}
