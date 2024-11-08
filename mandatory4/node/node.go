package main

import (
	"context"
	pb "dissys/mandatory4/proto"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

// stolen from example
var (
	tls        = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile   = flag.String("cert_file", "", "The TLS cert file")
	keyFile    = flag.String("key_file", "", "The TLS key file")
	jsonDBFile = flag.String("json_db_file", "", "A json file containing a list of features")
	serverAddr = flag.String("addr", "localhost:50051", "The server address in the format of host:port")
)

type DMENode struct {
	pb.UnimplementedDMEServer

	token    *pb.Token
	hasToken bool
	mu       sync.Mutex
	id       int32
	thisPort int32
	nextPort int32
}

// server
func (node *DMENode) RecieveToken(ctx context.Context, token *pb.Token) (*emptypb.Empty, error) {
	log.Println("Recieved as ", node.id, " on port ", node.thisPort)
	node.mu.Lock()

	log.Println("Recieved Token \"", token.Value, "\" from ", token.PeerID)

	node.token = &pb.Token{Value: token.Value, PeerID: node.id} // set ID of the token to our ID
	node.hasToken = true

	node.mu.Unlock()

	go Dial(node, node.nextPort)

	return &emptypb.Empty{}, nil
}

// client
func (node *DMENode) SendToken(client pb.DMEClient) {
	if !node.hasToken {
		log.Fatalln("Tried to send token despite not having it")
	}
	context := context.Background()

	node.mu.Lock()

	_, err := client.RecieveToken(context, node.token)

	log.Println("Sent token: \"", node.token, "\" to ", node.nextPort)
	node.hasToken = false

	node.mu.Unlock()
	if err != nil {
		log.Fatalln("Failed to send due to ", err)
	}

}

func Dial(node *DMENode, port int32) {
	// create options array
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials())) //

	addr := "localhost:" + strconv.Itoa(int(node.nextPort))

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		log.Fatalln("Client failed to open connection: ", err)
	}
	defer conn.Close()

	client := pb.NewDMEClient(conn)

	node.SendToken(client)
}

func newNode(id int32, thisPort int32, nextPort int32) *DMENode {
	return &DMENode{hasToken: false, id: id, thisPort: thisPort, nextPort: nextPort}
}

func main() {
	// get port number, throws errors if anything is wrong
	id, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalln("Failed to convert port:", err)
	}
	flag.Parse()

	thisPort := int32(8080 + id)
	var nextPort int32
	if id == 3 {
		nextPort = 8080
	} else {
		nextPort = int32(thisPort) + 1
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", thisPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("qwe", id, ": ", thisPort)
	grpcServer := grpc.NewServer()
	node := newNode(int32(id), thisPort, nextPort)
	pb.RegisterDMEServer(grpcServer, node)

	if node.id == 0 {
		node.mu.Lock()
		node.token = &pb.Token{Value: "token Value 123", PeerID: node.id}
		node.hasToken = true
		node.mu.Unlock()

		time.Sleep(time.Duration(5 * time.Second))
		log.Println("Inital send \"", node.token.Value, "\" to ", node.nextPort)
		Dial(node, node.nextPort)

	}

	grpcServer.Serve(lis)

}
