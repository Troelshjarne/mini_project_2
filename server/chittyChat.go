package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	chatpackage "github.com/Troelshjarne/mini_project_2/chat"

	"google.golang.org/grpc"
)

var serverlamtime = 1

// contains all unique message id's
var msgIds []int
var idMutex sync.Mutex

type Server struct {
	chatpackage.UnimplementedCommunicationServer

	//Map to store channel pointers. These are clients connecting to the service.
	channel map[string][]chan *chatpackage.ChatMessage
}

// helper method -> checks if msg already exists
func containsID(id int) bool {
	for _, val := range msgIds {
		if val == id {
			return true
		}
	}
	return id == 0
}

// helper method -> add id.
func addId(id int) {
	if len(msgIds) > 20 {
		msgIds = msgIds[1:]
	}
	msgIds = append(msgIds, id)
}

// close channel when client leaves.
func removeChannel(s *Server, ch chan *chatpackage.ChatMessage) {
	for key, val := range s.channel {
		for j := 0; j < len(val); j++ {
			if val[j] == ch {
				s.channel[key] = append(val[:j], val[j+1:]...)
				break
			}
		}
	}
}

// sends messages to all connected clients.
func ServerBroadcast(s *Server, messageBody string) {
	go func() {
		for _, streams := range s.channel {
			msg := chatpackage.ChatMessage{
				Channel: &chatpackage.Channel{
					Name:      "all",
					SendersID: "server"},
				Message:       messageBody,
				ParticipantID: "server",
				LamTime:       int32(serverlamtime) - 1,
			}
			for _, msgChan := range streams {
				msgChan <- &msg
			}
		}
	}()
}

func (s *Server) Broadcast(ch *chatpackage.Channel, msgStream chatpackage.Communication_BroadcastServer) error {

	msgChannel := make(chan *chatpackage.ChatMessage)

	//Joining information is stored in the channel map.
	s.channel[ch.Name] = append(s.channel[ch.Name], msgChannel)
	body := fmt.Sprintf("Participant %v joined the chat at timestamp %d", ch.SendersID, serverlamtime)
	log.Println(body)
	serverlamtime++
	ServerBroadcast(s, body)

	// keeping the stream open
	for {

		select {
		//If the channel is closed, this case is chosen.
		case <-msgStream.Context().Done():
			body := fmt.Sprintf("%v left the chat at timestamp %d", ch.SendersID, serverlamtime)
			log.Print(body)
			// Let all know, that user left.
			ServerBroadcast(s, body)
			// Event has passed; increment lamtime.
			serverlamtime++

			// Remove obsolete channel.
			removeChannel(s, msgChannel)
			close(msgChannel)

			// Close loop.
			return nil
		//If the server is running, messages are recieved through this channel.
		case msg := <-msgChannel:
			var inc bool = false
			fmt.Printf("Recieved message: %v at timestamp %v \n", msg, serverlamtime)
			idMutex.Lock()
			if !containsID(int(msg.Id)) {
				// Increment after recieving from client.
				serverlamtime++
				addId(int(msg.Id))
				inc = true
			}
			idMutex.Unlock()
			msgStream.Send(msg)

			if inc {
				// Increment after sending message
				serverlamtime++
			}
		}
	}

}

func (s *Server) Publish(msgStream chatpackage.Communication_PublishServer) error {

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
		fmt.Println("Channel Name: " + msg.Channel.Name)
		msg.LamTime = int32(serverlamtime)
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
	grpcServer := grpc.NewServer(options...)

	chatpackage.RegisterCommunicationServer(grpcServer, &Server{
		channel: make(map[string][]chan *chatpackage.ChatMessage),
	})
	grpcServer.Serve(list)
}
