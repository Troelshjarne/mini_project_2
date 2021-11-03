package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"

	chatpackage "github.com/Troelshjarne/mini_project_2/chat"

	"google.golang.org/grpc"
)

// commandline input stored in flags
var channelName = flag.String("channel", "default", "Channel name for chatting")
var senderName = flag.String("sender", "default", "Senders name")
var tcpServer = flag.String("server", ":9080", "Tcp server")

func main() {

	flag.Parse()

	fmt.Println("=== Welcome to Chitty Chat - Beta 0.1.2 ===")
	var options []grpc.DialOption
	options = append(options, grpc.WithBlock(), grpc.WithInsecure())
	//connect to server
	conn, err := grpc.Dial(*tcpServer, options...)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}

	defer conn.Close()

	ctx := context.Background()

	// client connection interface
	client := chatpackage.NewCommunicationClient(conn)

	go broadcast(ctx, client)

	//Logging

	LOG_FILE := "./log.txt"

	logFile, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer logFile.Close()

	mw := io.MultiWriter(os.Stdout, logFile)

	log.SetOutput(mw)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		if valid(text) {
			go publish(ctx, client, text, int32(0))
		} else {
			log.Printf("(%v): Invalid message length. Must be at least 1 character long, and at most 128 characters long.", *senderName)
		}
	}

}

// validates length og client message
func valid(input string) bool {
	return !(len(input) == 0 || len(input) > 128)
}

// rpc call -> client sends message to server
func broadcast(ctx context.Context, client chatpackage.CommunicationClient) {

	channel := chatpackage.Channel{Name: *channelName, SendersID: *senderName}
	stream, err := client.Broadcast(ctx, &channel)
	if err != nil {
		log.Fatalf("Client join channel error! Throws %v", err)
	}

	//Logging

	LOG_FILE := "./log.txt"

	logFile, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer logFile.Close()

	mw := io.MultiWriter(os.Stdout, logFile)

	log.SetOutput(mw)

	waitChannel := make(chan struct{})

	go func() {

		for {

			in, err := stream.Recv()
			if err == io.EOF {
				close(waitChannel)
				return
			}
			if err != nil {
				log.Fatalf("Failed to recieve message from channel joining. \n Got error: %v", err)
			}

			if *senderName != in.ParticipantID {

				log.Printf("Time: (%v) Message: (%v) -> %v \n", in.LamTime, in.ParticipantID, in.Message)
			}

		}
	}()

	<-waitChannel

}

// rpc call -> reads the messages broad by the server
func publish(ctx context.Context, client chatpackage.CommunicationClient, message string, lamtime int32) {

	stream, err := client.Publish(ctx)
	if err != nil {
		log.Printf("Failure sending message! Got error: %v", err)
	}

	/* generates a 'random' 64-bit int cryptographically secure random number in a range.
	   Gives a message an unique ID -> used as bugfix, since a message initally was broadcasted n (client) times.
	*/
	randId, _ := rand.Int(rand.Reader, big.NewInt(8192*8192*8192*8192))
	//Include timestamp.
	msg := chatpackage.ChatMessage{
		Channel: &chatpackage.Channel{
			Name:      *channelName,
			SendersID: *senderName},
		Message:       message,
		ParticipantID: *senderName,
		LamTime:       lamtime,
		Id:            randId.Int64(),
	}
	stream.Send(&msg)

	ack, err := stream.CloseAndRecv()
	fmt.Printf("Message sent: %v \n", ack)
}
