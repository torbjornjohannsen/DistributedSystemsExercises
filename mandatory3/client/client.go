package main

import (
	"context"
	pb "dissys/mandatory3/Chitt_chat"
	"flag"
	"log"
	"math/rand/v2"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// yoinked from the example
var (
	tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile             = flag.String("ca_file", "", "The file containing the CA root cert file")
	serverAddr         = flag.String("addr", "localhost:50051", "The server address in the format of host:port")
	serverHostOverride = flag.String("server_host_override", "x.test.example.com", "The server name used to verify the hostname returned by the TLS handshake")
)

func runChat(client pb.ChittChatClient) {
	id := rand.Int()
	lamportTime := int32(0)
	var mu sync.Mutex

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()                  // run cancel after the execution of this function dies
	stream, err := client.Chat(ctx) // contact the server, recieve a stream
	if err != nil {
		log.Fatalln("Client failed due to", err)
	}

	// goroutine to recieve and log messages
	go func() {
		for {
			in, err := stream.Recv()
			if err != nil {
				log.Fatalln("Client failed due to ", err)
			}

			mu.Lock()
			lamportTime = max(lamportTime+1, in.LamportTime)
			mu.Unlock()

			log.Println("Client recieved: ", in.Text, " @ ", in.LamportTime)
		}
	}()

	numMessages := rand.IntN(50)
	for i := 0; i < numMessages; i++ {
		mu.Lock()
		lamportTime++
		mu.Unlock()

		msg := pb.Message{
			Text:        "Client ",
			LamportTime: lamportTime,
			SenderId:    int32(id),
			LastMessage: false}

		err := stream.Send(&msg)
		if err != nil {
			log.Fatalln("Client failed due to ", err)
		}
		time.Sleep(time.Duration(rand.IntN(2000)) * time.Millisecond)
	}

	quitMsg := pb.Message{
		Text:        "",
		LamportTime: lamportTime,
		SenderId:    int32(id),
		LastMessage: true,
	}
	stream.Send(&quitMsg)

	stream.CloseSend()
}

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.NewClient(*serverAddr, opts...)
	if err != nil {
		log.Fatalln("Client failed to open connection: ", err)
	}

	defer conn.Close()

	client := pb.NewChittChatClient(conn)

	runChat(client)
}
