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

	mu             sync.Mutex //used whenever we modify resources that several threads might use
	clientCounter  int
	lamportTime    int32
	msgToBroadcast chan *pb.Message
	channels       []chan *pb.Message
}

func (s *ChatServer) getLamportTime(altTime int32) int32 {
	s.mu.Lock()
	s.lamportTime = max(s.lamportTime, altTime) + 1
	s.mu.Unlock()
	return s.lamportTime
}

func (s *ChatServer) Chat(stream grpc.BidiStreamingServer[pb.Message, pb.Message]) error {
	broadcastChan := make(chan *pb.Message)

	s.clientCounter++
	log.Println("New connection established with client ", s.clientCounter)

	stream.Send(&pb.Message{SenderId: int32(s.clientCounter), LamportTime: s.getLamportTime(0)}) // special message to assign ID

	s.mu.Lock()
	s.channels = append(s.channels, broadcastChan)

	if len(s.channels) <= 1 { // first time we get a connection launch the broadcaster thread
		go s.Broadcaster()
	}
	s.mu.Unlock()

	s.msgToBroadcast <- &pb.Message{
		Text:        fmt.Sprint("Participant ", s.clientCounter, " joined Chitty-Chat at Lamport time ", s.lamportTime),
		LamportTime: s.getLamportTime(0),
		SenderId:    -1, // for server
		LastMessage: false}

	// goroutine to broadcast to this particular client
	go func() {
		for {
			msg := <-broadcastChan
			//msg.LamportTime = s.getLamportTime(0)

			s.mu.Lock()
			s.lamportTime++
			stream.Send(&pb.Message{Text: msg.Text, LamportTime: s.lamportTime})
			log.Println("\t@", s.lamportTime)
			s.mu.Unlock()
		}
	}()

	for {
		in, err := stream.Recv()
		if err != nil {
			return err
		}

		var msgText string

		if in.LastMessage { // graceful exit
			lT := s.getLamportTime(in.LamportTime)
			msgText = fmt.Sprint("Participant ", in.SenderId, " left Chitty-Chat at Lamport time ", lT)

			s.msgToBroadcast <- &pb.Message{
				Text:        msgText,
				LamportTime: lT,
				SenderId:    in.SenderId,
				LastMessage: in.LastMessage}

			log.Println("Ended chat with participant ", in.SenderId)

			return nil
		} else {
			msgText = fmt.Sprint(in.SenderId, ": ", in.Text)

			s.msgToBroadcast <- &pb.Message{
				Text:        msgText,
				LamportTime: s.getLamportTime(in.LamportTime),
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
		log.Println("Broadcasted \"", msg.Text, "\"")
	}
}

func newServer(broadcastChan chan *pb.Message) *ChatServer {
	return &ChatServer{
		lamportTime:    1, // 1 because the setup stuff counts as an event
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
