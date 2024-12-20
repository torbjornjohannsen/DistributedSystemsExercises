package main

import (
	"bytes"
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
	data           []rune
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

func tls_client(inChannel chan Message, outChannel chan Message, timeout int) bool {
	msg := Message{rand.Int(), 0, true, false, false, 0, 0, nil}
	inChannel <- msg
	// get and validate message 2, with timeout
	var msg2 Message

	if !(getMessageWithTimeout(&msg2, outChannel, timeout) &&
		msg2.syn && !msg2.fin && msg2.ack && msg2.acknowledgment == msg.sequence_num+1) {

		fmt.Println("Invalid message: %s", msg.print())
		return false
	}

	fmt.Printf("Client recieved: %s\n", msg2.print())
	msg3 := Message{msg2.acknowledgment, msg2.sequence_num + 1, false, true, false, 0, 0, nil}
	outChannel <- msg3
	return true
}

func tls_server(inChannel chan Message, outChannel chan Message, timeout int) bool {
	var msg Message
	if !(getMessageWithTimeout(&msg, inChannel, timeout) &&
		msg.syn && !msg.fin && !msg.ack) {

		fmt.Println("Invalid Message: %s", msg.print())
		return false
	}

	fmt.Printf("Server Recieved: %s\n", msg.print())
	msg2 := Message{rand.Int(), msg.sequence_num + 1, true, true, false, 0, 0, nil}
	outChannel <- msg2
	var msg3 Message
	if !(getMessageWithTimeout(&msg3, outChannel, timeout) &&
		!msg3.syn && !msg2.fin && msg2.ack && msg3.sequence_num == msg2.acknowledgment && msg3.acknowledgment == msg2.sequence_num+1) {

		fmt.Println("Wrong message")
		return false
	}

	fmt.Printf("Server Recieved: %s\nTLS handshake done\n", msg3.print())
	return true
}

func send(inChannel chan Message, outChannel chan Message, data []rune) {
	//
	for !tls_server(outChannel, inChannel, 1000) {
		time.Sleep(time.Duration(500) * time.Millisecond)
	}

	var requestMsg Message
	for !getMessageWithTimeout(&requestMsg, inChannel, 5000) {
		fmt.Println("No req msg")
	}

	fmt.Printf("Got request msg: %s\n", requestMsg.print())

	windowSize := requestMsg.window_size
	segmentSize := requestMsg.segment_size
	totalSegments := len(data) / segmentSize
	if len(data)%segmentSize > 0 {
		totalSegments++
	}

	indice := 0
	counter := 0
	segmentRecieved := make([]bool, totalSegments+2) // +2 to avoid off by one error from acknowledgements being incremented automatically

	go func(segArr []bool, msgCh chan Message) {
		for {
			var ackMsg Message

			ackMsg = <-msgCh
			segArr[ackMsg.acknowledgment-1] = true
		}
	}(segmentRecieved, inChannel)

	for !segmentRecieved[totalSegments-1] {
		if !segmentRecieved[max(0, counter-windowSize)] { // if we have fired off the whole window-size without acknowledgement
			counter = max(0, counter-windowSize) // reset to start of block
			indice = segmentSize * counter
		}

		block := data[indice : min(indice+segmentSize, len(data))]
		indice = min(indice+segmentSize, len(data))
		msgBlock := Message{counter, 0, true, false, false, windowSize, segmentSize, block}
		fmt.Printf("Server sent msg %d\n", counter)
		outChannel <- msgBlock
		counter++
	}

	finMsg := Message{counter, 0, false, false, true, 0, 0, nil}
	outChannel <- finMsg

	fmt.Println("Server finished sending off")
}

func receive(inChannel chan Message, outChannel chan Message, timeout int) {
	for !tls_client(inChannel, outChannel, 1000) {
		time.Sleep(time.Duration(500) * time.Millisecond)
	}

	// establish window and segment size
	requestMsg := Message{0, 0, false, false, false, 4, 10, nil}
	outChannel <- requestMsg

	var msg Message
	var ack = 0
	var data bytes.Buffer
	for getMessageWithTimeout(&msg, inChannel, timeout) && !msg.fin {

		if msg.sequence_num == ack {
			ack++
			data.WriteString(string(msg.data))
			fmt.Printf("Client received %d {%s}\n", msg.sequence_num, string(msg.data))
		}
		outChannel <- Message{msg.sequence_num, ack, msg.syn, msg.ack, msg.fin, msg.window_size, msg.segment_size, nil}
	}
	fmt.Println("Entire message:\n" + data.String())
}

func main() {
	inChannel := make(chan Message)
	outChannel := make(chan Message)
	go send(inChannel, outChannel, []rune("weqewqewqewq\nqweqweqweqw\nrejthetjerthreikthrwjithwjoithewjorhewoir\nrhjithwrtgweoigtewoitgegtewiogtewyiotgewitoew"))
	go receive(outChannel, inChannel, 1000)

	time.Sleep(2000 * time.Millisecond)
}
