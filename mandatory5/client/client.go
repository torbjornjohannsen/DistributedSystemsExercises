package main

import (
	"context"
	pb "dissys/mandatory5/proto"
	"flag"
	"log"
	"math/rand/v2"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// stolen from example
var (
	tls        = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile   = flag.String("cert_file", "", "The TLS cert file")
	keyFile    = flag.String("key_file", "", "The TLS key file")
	jsonDBFile = flag.String("json_db_file", "", "A json file containing a list of features")
	serverAddr = flag.String("addr", "localhost:50051", "The server address in the format of host:port")
)

type AuctionClient struct {
	c pb.AuctionClient

	maxBid      int32
	maxBidID    int32
	auctionover bool
	context     context.Context
	ID          int
	serverIDs   []int
}

func (this *AuctionClient) SendBid() error {
	ourBid := this.maxBid + int32(rand.IntN(10)+1)

	bidAmount := &pb.Amount{Amount: ourBid, IsServer: false, ClientID: int32(this.ID)}

	log.Println("Send Bid of ", ourBid)
	acknowledgement, err := this.c.Bid(this.context, bidAmount)
	if err != nil {
		log.Println("Sendbid err: ", err)
		return err
	}

	if acknowledgement.Success {
		this.maxBid = ourBid
		log.Println("Successful bid")
	} else {
		log.Println("Unsuccessful bid with exception code", acknowledgement.Exception)
	}
	return nil
}

func (this *AuctionClient) GetResult() error {
	log.Println("Send result call")
	result, err := this.c.Result(this.context, &pb.IsServer{IsServer: false})
	if err != nil {
		log.Println("GetResult err: ", err)
		return err
	}

	log.Println("Got result {maxBid:", result.MaxBid, ", auctionOver: ", result.Auctionover, "}")
	this.maxBid = result.MaxBid
	this.auctionover = result.Auctionover
	this.maxBidID = result.ClientID
	return nil
}

func RunClient(id int, serverList []int) {
	auctionClient := AuctionClient{context: context.Background(), maxBid: 0, auctionover: false, ID: id, maxBidID: -1, serverIDs: serverList}

	conn := auctionClient.SetupConnection(auctionClient.serverIDs[0])
	log.Println("Intially using server", auctionClient.serverIDs[0])
	sId := 0
	goNextServer := false

	for {
		if auctionClient.auctionover {
			break
		}

		if auctionClient.maxBidID != int32(auctionClient.ID) {

			goNextServer = (auctionClient.SendBid() != nil)
		}

		goNextServer = goNextServer || (auctionClient.GetResult() != nil)

		if !goNextServer {
			time.Sleep(time.Second * 1)
			continue
		}

		if sId < len(auctionClient.serverIDs)-1 {
			log.Println("Server ", auctionClient.serverIDs[sId], " failed, moving on to server ", auctionClient.serverIDs[sId+1])
			conn.Close()
			sId++
			conn = auctionClient.SetupConnection(auctionClient.serverIDs[sId])
		} else if sId >= len(auctionClient.serverIDs)-1 {
			log.Fatalln("All servers failed")
		}

	}

	auctionClient.GetResult()

	if auctionClient.maxBidID == int32(auctionClient.ID) {
		log.Println("Won with a bid of ", auctionClient.maxBid)
	} else {
		log.Println("Lost bid to ", auctionClient.maxBidID)
	}
}

func (this *AuctionClient) SetupConnection(serverID int) *grpc.ClientConn {
	thisPort := 8080 + serverID

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials())) //

	addr := "localhost:" + strconv.Itoa(thisPort)

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		log.Fatalln("Client failed to open connection: ", err)
	}

	this.c = pb.NewAuctionClient(conn)
	return conn
}

func main() {
	id, err := strconv.Atoi(os.Args[1])
	serverAmt, err2 := strconv.Atoi(os.Args[2])
	if err != nil || err2 != nil {
		log.Fatalln("Failed to convert id or numNodes:", err)
	}

	serverList := make([]int, serverAmt)
	for i := 0; i < serverAmt; i++ {
		serverList[i] = i
	}

	// shuffle the list of server ID's
	// https://stackoverflow.com/questions/12264789/shuffle-array-in-go
	for i := range serverList {
		j := rand.IntN(i + 1)
		serverList[i], serverList[j] = serverList[j], serverList[i]
	}

	RunClient(id, serverList)
}
