package main

import (
	"fmt"
	"io"
	"log"
	"net"

	chatpackage "github.com/Troelshjarne/mini_project_2/chat"

	"google.golang.org/grpc"
)

type Server struct {
	chatpackage.UnimplementedCommunicationServer

	//Map to store channel pointers. These are clients connecting to the service.
	channel map[string][]chan *chatpackage.ChatMessage
}

func (s *Server) JoinChannel(ch *chatpackage.Channel, msgStream chatpackage.Communication_JoinChannelServer) error {

	msgChannel := make(chan *chatpackage.ChatMessage)

	//Joining information is stored in the channel map.
	s.channel[ch.Name] = append(s.channel[ch.Name], msgChannel)

	// keeping the stream open
	for {

		select {
		//If the channel is closed, this case is chosen.
		case <-msgStream.Context().Done():
			log.Print(ch.SendersID + " " + "is gooone bby")
			//msgStream.Send(&msg)
			return nil
		//If the server is running, messages are recieved through this channel.
		case msg := <-msgChannel:
			fmt.Printf("Recieved message: %v at timestamp T \n", msg)
			msgStream.Send(msg)
		}
	}

}

func (s *Server) SendMessage(msgStream chatpackage.Communication_SendMessageServer) error {

	//Used for recieving messages
	msg, err := msgStream.Recv()

	if err == io.EOF {
		return nil
	}

	if err != nil {
		return err
	}

	//Used for sending acknowledgements, that an message has been sent.
	ack := chatpackage.MessageAck{Status: "Message sent..."}
	msgStream.SendAndClose(&ack)

	//Function to loop over the channel to send incoming messages
	go func() {
		streams := s.channel[msg.Channel.Name]
		for _, msgChan := range streams {
			msgChan <- msg
		}
	}()

	return nil

}

func main() {

	fmt.Println("=== Server starting up ===")
	list, err := net.Listen("tcp", ":9080")

	if err != nil {
		log.Fatalf("Failed to listen on port 9080: %v", err)
	}

	var options []grpc.ServerOption
	grpcServer := grpc.NewServer(options...) //NOTE <--- What does this do?

	chatpackage.RegisterCommunicationServer(grpcServer, &Server{
		channel: make(map[string][]chan *chatpackage.ChatMessage),
	})
	grpcServer.Serve(list)
}
