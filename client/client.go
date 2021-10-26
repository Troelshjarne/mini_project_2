package main

import (
	"context"
	"fmt"
	"log"

	t "time"

	chatpackage "github.com/Troelshjarne/mini_project_2/chat"

	"google.golang.org/grpc"
)

func main() {
	// Creat a virtual RPC Client Connection on port  9080 WithInsecure (because  of http)
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(":9080", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %s", err)
	} else {
		fmt.Print("we are live!")
	}

	// Defer means: When this function returns, call this method (meaing, one main is done, close connection)
	defer conn.Close()

	//  Create new Client from generated gRPC code from proto
	c := chatpackage.NewPublishClient(conn)

	for {
		SendGetMsgRequest(c)
		t.Sleep(5 * t.Second)
	}

}

func SendGetMsgRequest(c chatpackage.PublishClient) {

	messageExample := chatpackage.ChatMessageRequest{}

	response, err := c.GetMessage(context.Background(), &messageExample)
	if err != nil {
		log.Fatalf("Error when calling GetMessage: %s", err)
	}

	fmt.Printf("The current message is: %s\n", response.Message)

}
