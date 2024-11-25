package main

import (
	"context"
	pb "dissys/mandatory5/proto"
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
)

type AuctionNode struct {
	pb.UnimplementedAuctionServer

	maxBid      int32
	maxBidID    int32
	auctionOver bool
	nodeClients []pb.AuctionClient
	mu          sync.Mutex
}

func (this *AuctionNode) Bid(ctx context.Context, bid *pb.Amount) (*pb.Acknowledgement, error) {
	log.Println("Recieved Bid call for ", bid.Amount, " (from server? ", bid.IsServer, ") from client", bid.ClientID)

	this.mu.Lock()
	if bid.Amount > this.maxBid {
		this.maxBid = bid.Amount
		this.maxBidID = bid.ClientID
	}
	this.mu.Unlock()

	if !bid.IsServer {
		bid.IsServer = true
		waitGroup := sync.WaitGroup{}
		for i := range this.nodeClients {
			waitGroup.Add(1)
			go func(node *AuctionNode, cI int, wg *sync.WaitGroup) {
				defer wg.Done()

				log.Println("Replicate bid call to #", cI)
				ack, err := node.nodeClients[cI].Bid(ctx, bid)
				if err != nil {
					log.Println("Propegating bid err: ", err)
					return
				}
				log.Println("Recieved acknowledgement #", cI)
				if !ack.Success {
					log.Println("Unsuccessful node to node bid with exception: ", ack.Exception)
				}
			}(this, i, &waitGroup)
		}

		waitGroup.Wait()
	}

	log.Println("Send acknowledgement")
	return &pb.Acknowledgement{Success: true, Exception: 0}, nil
}

func (this *AuctionNode) Result(ctx context.Context, isServer *pb.IsServer) (*pb.Outcome, error) {
	log.Println("Recieved Result call (from server? ", isServer.IsServer, ")")

	if !isServer.IsServer {
		waitGroup := sync.WaitGroup{}
		amServer := &pb.IsServer{IsServer: true}

		for i := range this.nodeClients {
			waitGroup.Add(1)
			go func(node *AuctionNode, cI int, wg *sync.WaitGroup, msg *pb.IsServer) {
				defer wg.Done()
				log.Println("Replicate Result call to ", cI)
				outcome, err := node.nodeClients[cI].Result(ctx, msg)
				if err != nil {
					log.Println("Client propegating bid err: ", err)
					return
				}

				log.Println("Recieved outcome #", cI, " {maxBid: ", outcome.MaxBid, ", auctionover: ", outcome.Auctionover, "}")

				node.mu.Lock()
				defer node.mu.Unlock()
				if outcome.MaxBid >= node.maxBid {
					node.maxBid = outcome.MaxBid
					node.maxBidID = outcome.ClientID
				}
				if outcome.Auctionover {
					node.auctionOver = true
				}
			}(this, i, &waitGroup, amServer)
		}

		waitGroup.Wait()
	}

	log.Println("Send outcome {maxBid: ", this.maxBid, ", auctionover: ", this.auctionOver, " clientID: ", this.maxBidID, "}")
	return &pb.Outcome{MaxBid: this.maxBid, Auctionover: this.auctionOver, ClientID: this.maxBidID}, nil
}

func (this *AuctionNode) AuctionTimer(timeInMs int) {
	time.Sleep(time.Duration(timeInMs) * time.Millisecond)
	this.mu.Lock()
	this.auctionOver = true
	log.Println("Auction set to over")
	this.mu.Unlock()
	go func() {
		time.Sleep(time.Second * 5)
		os.Exit(0)
	}()
}

func CreateClient(port int) (pb.AuctionClient, *grpc.ClientConn) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials())) //

	addr := "localhost:" + strconv.Itoa(port)

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		log.Fatalln("Client failed to open connection: ", err)
	}

	return pb.NewAuctionClient(conn), conn
}

func fail() {
	time.Sleep(3 * time.Second)
	for {
		time.Sleep(1 * time.Second)
		if rand.IntN(100) > 98 {
			log.Fatalln("Server died")
			os.Exit(1)
		}
	}
}

func main() {
	// get port number, throws errors if anything is wrong
	id, err := strconv.Atoi(os.Args[1])
	numNodes, err2 := strconv.Atoi(os.Args[2])
	if err != nil || err2 != nil {
		log.Fatalln("Failed to convert id or numNodes:", err)
	}

	thisPort := int32(8080 + id)

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", thisPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	cliPort := 8079
	// Create array of clients to send messages to all the other servers we know of
	clientArr := make([]pb.AuctionClient, numNodes-1)
	offset := 0
	for i := 0; i < numNodes; i++ {
		cliPort++
		log.Println("a: ", i, " of ", numNodes, " with port ", cliPort, " vs ", thisPort)
		if cliPort == int(thisPort) {
			offset = 1
			continue
		}
		var conn *grpc.ClientConn
		clientArr[i-offset], conn = CreateClient(cliPort)
		defer conn.Close()
	}

	grpcServer := grpc.NewServer()
	node := &AuctionNode{maxBid: 0, auctionOver: false, nodeClients: clientArr}
	pb.RegisterAuctionServer(grpcServer, node)

	go node.AuctionTimer(100000)
	go fail()
	grpcServer.Serve(lis)
}
