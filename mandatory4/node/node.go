package main

import (
	"context"
	pb "dissys/mandatory4/proto"
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
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

	token         *pb.Token
	hasToken      bool
	mu            sync.Mutex
	accessCounter int32
	id            int32
	thisPort      int32
	nextPort      int32
}

func (node *DMENode) AccessCriticalSection() {
	f, err := os.OpenFile("critical.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)

	if err != nil {
		log.Fatalln("Failed to open critical file: ", err)
	}

	node.accessCounter++
	strToWrite := fmt.Sprintln("Cool string ", node.accessCounter, " from node ", node.id)

	f.WriteString(strToWrite)
	log.Println("Wrote \"", strToWrite, "\" to the critical file")
}

// server
func (node *DMENode) RecieveToken(ctx context.Context, token *pb.Token) (*emptypb.Empty, error) {
	log.Println("Recieved as ", node.id, " on port ", node.thisPort)
	node.mu.Lock()

	log.Println("Recieved Token \"", token.Value, "\" from ", token.PeerID)

	node.hasToken = true
	if rand.IntN(2) > 0 {
		node.AccessCriticalSection()
	}

	token.PeerID = node.id // set ID of the token to our ID
	token.MinAccessCounter = min(token.MinAccessCounter, node.accessCounter)
	node.token = token
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

	node.hasToken = false

	node.mu.Unlock()

	if err != nil {
		log.Fatalln("Failed to send due to ", err)
	}
	log.Println("Sent token: \"", node.token, "\" to ", node.nextPort)

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
	return &DMENode{hasToken: false, id: id, thisPort: thisPort, nextPort: nextPort, accessCounter: 0}
}

func serverKiller(server *grpc.Server, node *DMENode) {
	time.Sleep(100 * time.Second)
	server.GracefulStop()
	log.Println("Killing node ", node.id)
	os.Exit(0)
}

func main() {
	// get port number, throws errors if anything is wrong
	id, err := strconv.Atoi(os.Args[1])
	numNodes, err2 := strconv.Atoi(os.Args[2])
	if err != nil || err2 != nil {
		log.Fatalln("Failed to convert id or numNodes:", err)
	}
	flag.Parse()

	thisPort := int32(8080 + id)
	var nextPort int32
	if id == numNodes-1 {
		nextPort = 8080
	} else {
		nextPort = int32(thisPort) + 1
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", thisPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("Node:", id, " using port ", thisPort)
	grpcServer := grpc.NewServer()
	node := newNode(int32(id), thisPort, nextPort)
	pb.RegisterDMEServer(grpcServer, node)

	if node.id == 0 {
		node.mu.Lock()
		node.token = &pb.Token{Value: "token Value 123", PeerID: node.id, MinAccessCounter: 0}
		node.hasToken = true
		node.mu.Unlock()

		time.Sleep(time.Duration(5 * time.Second))
		log.Println("Inital send \"", node.token.Value, "\" to ", node.nextPort)
		go Dial(node, node.nextPort)

	}
	go serverKiller(grpcServer, node)
	grpcServer.Serve(lis)

}
