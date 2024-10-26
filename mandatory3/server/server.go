package main

import (
	pb "dissys/mandatory3/Chitt_chat"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
)

// stolen from example
var (
	tls        = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile   = flag.String("cert_file", "", "The TLS cert file")
	keyFile    = flag.String("key_file", "", "The TLS key file")
	jsonDBFile = flag.String("json_db_file", "", "A json file containing a list of features")
	port       = flag.Int("port", 50051, "The server port")
)

type ChatServer struct {
	pb.UnimplementedChittChatServer

	mu             sync.Mutex
	clientCounter  int
	lamportTime    int32
	msgToBroadcast chan *pb.Message
	channels       []chan *pb.Message
}

func (s *ChatServer) Chat(stream grpc.BidiStreamingServer[pb.Message, pb.Message]) error {
	broadcastChan := make(chan *pb.Message)
	s.mu.Lock()
	s.clientCounter++
	log.Println("New connection established")
	s.channels = append(s.channels, broadcastChan)

	if len(s.channels) <= 1 { // first time we get a connection launch the broadcaster thread
		go s.Broadcaster()
	}

	s.msgToBroadcast <- &pb.Message{
		Text:        fmt.Sprint("Participant ", s.clientCounter, " joined Chitty-Chat at Lamport time ", s.lamportTime),
		LamportTime: s.lamportTime,
		SenderId:    -1, // for server
		LastMessage: false}

	s.mu.Unlock()

	// goroutine to broadcast to this particular client
	go func() {
		for {
			msg := <-broadcastChan

			stream.Send(msg)
		}
	}()

	for {
		in, err := stream.Recv()
		if err != nil {
			return err
		}

		s.mu.Lock()
		s.lamportTime = max(s.lamportTime+1, in.LamportTime)
		s.mu.Unlock()

		if in.LastMessage { // graceful exit

			s.msgToBroadcast <- &pb.Message{
				Text:        fmt.Sprint("Participant ", in.SenderId, " left Chitty-Chat at Lamport time ", s.lamportTime),
				LamportTime: s.lamportTime,
				SenderId:    in.SenderId,
				LastMessage: in.LastMessage}

			return nil
		} else {

			s.msgToBroadcast <- &pb.Message{
				Text:        in.Text,
				LamportTime: s.lamportTime,
				SenderId:    in.SenderId,
				LastMessage: in.LastMessage}
		}
	}
}

func (s *ChatServer) Broadcaster() {
	log.Println("Broadcaster launched")

	for { // just relay every message we get to all the comms we have going on
		msg := <-s.msgToBroadcast
		for _, channel := range s.channels {
			channel <- msg
		}
	}
}

func newServer(broadcastChan chan *pb.Message) *ChatServer {
	return &ChatServer{
		lamportTime:    0,
		msgToBroadcast: broadcastChan,
		channels:       make([]chan *pb.Message, 0),
		clientCounter:  0}
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	broadcastChan := make(chan *pb.Message)
	grpcServer := grpc.NewServer()
	pb.RegisterChittChatServer(grpcServer, newServer(broadcastChan))
	grpcServer.Serve(lis)

}
