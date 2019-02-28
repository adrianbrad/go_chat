package channel

import (
	"log"

	"github.com/adrianbrad/chat/message"
	"github.com/adrianbrad/chat/model"
	"github.com/gorilla/websocket"
)

//Client is a Websocket client opened by a user, it also holds information about that user
type Client interface {
	Read()
	Write()
	ForwardMessage() chan *message.BroadcastedMessage
	GetUserID() int
	CloseSocket()
}

type client struct {
	//socket is the websocket connection for this client
	socket *websocket.Conn
	//send is the buffered channel on which messages are queued ready to be forwarded to the user browser
	forwardMessage chan *message.BroadcastedMessage
	//channel is the channel this client is chatting in, used to broadcast messages to everyone else in the channel
	channel Channel
	//userData is used for storing information about the user
	user model.User
}

//We process the message here. This is the first place they reach
func (client *client) Read() {
	defer client.socket.Close()

	for {
		var receivedMessage *message.ReceivedMessage
		err := client.socket.ReadJSON(&receivedMessage)
		log.Println(err)
		//if reading from socket fails the for loop is broken and the socket is closed
		if err != nil {
			return
		}

		switch receivedMessage.Action {
		case "join":
			client.channel.JoinRoom() <- client
			client.ForwardMessage() <- client.sendHistory(receivedMessage)
		case "leave":
			client.channel.LeaveRoom() <- client
		case "message":
			messageToBeBroadcasted := client.processMessage(receivedMessage)
			client.channel.MessageQueue() <- messageToBeBroadcasted
		}
	}
}

func (client *client) Write() {
	defer client.socket.Close()

	for msg := range client.ForwardMessage() {
		err := client.socket.WriteJSON(msg)
		//if writing to socket fails the for loop is brocken and the socket is closed
		if err != nil {
			return
		}
	}
}

func (client client) processMessage(rm *message.ReceivedMessage) *message.BroadcastedMessage {
	//TODO return a message to be broadcasted over the specified rooms
	return nil
}

func (client client) sendHistory(rm *message.ReceivedMessage) *message.BroadcastedMessage {
	//TODO get history from the channel for the room he asked to join based on the amount of messages given
	return nil
}

func (client *client) ForwardMessage() chan *message.BroadcastedMessage {
	return client.forwardMessage
}

func (client *client) GetUserID() int {
	return client.user.ID
}

func (client *client) CloseSocket() {
	client.socket.Close()
}
