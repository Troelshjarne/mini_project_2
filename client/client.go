package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	chatpackage "github.com/Troelshjarne/mini_project_2/chat"

	"google.golang.org/grpc"
)

var channelName = flag.String("channel", "default", "Channel name for chatting")
var senderName = flag.String("sender", "default", "Senders name")
var tcpServer = flag.String("server", ":9080", "Tcp server")

func main() {

	flag.Parse()

	fmt.Println("=== Welcome to Chitty Chat - Beta 0.1.2 ===")
	var options []grpc.DialOption
	options = append(options, grpc.WithBlock(), grpc.WithInsecure())

	conn, err := grpc.Dial(*tcpServer, options...)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}

	defer conn.Close()

	ctx := context.Background()
	client := chatpackage.NewCommunicationClient(conn)

	go joinChannel(ctx, client)

	scanner := bufio.NewScanner(os.Stdin)
	// check for string length!!!!!
	for scanner.Scan() {
		go sendMessage(ctx, client, scanner.Text(), int32(0))
	}

}

func joinChannel(ctx context.Context, client chatpackage.CommunicationClient) {

	channel := chatpackage.Channel{Name: *channelName, SendersID: *senderName}
	stream, err := client.JoinChannel(ctx, &channel)
	if err != nil {
		log.Fatalf("Client join channel error! Throws %v", err)
	}

	//fmt.Printf("Joined channel: %v \n", *channelName)

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

func sendMessage(ctx context.Context, client chatpackage.CommunicationClient, message string, lamtime int32) {
	stream, err := client.SendMessage(ctx)
	if err != nil {
		log.Printf("Fail sending message! Got error: %v", err)
	}
	// include timestamp.
	msg := chatpackage.ChatMessage{
		Channel: &chatpackage.Channel{
			Name:      *channelName,
			SendersID: *senderName},
		Message:       message,
		ParticipantID: *senderName,
		LamTime:       lamtime,
	}
	stream.Send(&msg)

	ack, err := stream.CloseAndRecv()
	fmt.Printf("Message sent: %v \n", ack)
}
