package channel

import (
	"fmt"
	"strconv"

	"github.com/adrianbrad/chat/message"
	"github.com/adrianbrad/chat/model"
	"github.com/adrianbrad/chat/repository"
)

type usersRepoMock struct{}

func NewUsersRepoMock() repository.Repository {
	return &usersRepoMock{}
}

func (r usersRepoMock) CheckIfExists(id int) bool {
	return true
}

func (r usersRepoMock) GetOne(id int) (interface{}, error) {
	return nil, nil
}

func (r usersRepoMock) GetAll() []interface{} {
	return nil
}

func (r usersRepoMock) GetAllWhere(column string, value int, limit int) []interface{} {
	return nil
}

func (r usersRepoMock) Create(interface{}) (int, error) {
	return 0, nil
}

type usersChannelsRepoMock struct {
	users    map[int]bool
	channels map[int]bool
	messages map[int][]string
}

func NewUSersChannelsRepoMock() repository.UsersChannelsRepository {
	return &usersChannelsRepoMock{
		users: map[int]bool{
			1: true,
			2: true,
			3: true,
			4: true,
			5: true,
		},
		channels: map[int]bool{
			1: true,
		},
	}
}

func (r usersChannelsRepoMock) AddOrUpdateUserToChannel(userID, channelID int) error {
	_, userExists := r.users[userID]
	_, channelExists := r.channels[channelID]
	if !userExists || !channelExists {
		return fmt.Errorf("User or Channel does not exist")
	}
	return nil
}

type messagesRepoMock struct {
	messages map[int][]model.Message
}

func NewMessagesRepoMock(nrOfMessages, nrOfRooms int) repository.Repository {
	messages := make(map[int][]model.Message)
	for j := 1; j <= nrOfRooms; j++ {
		for i := 1; i <= nrOfMessages; i++ {
			messages[j] = append(messages[j], model.Message{
				UserID:  1,
				Content: fmt.Sprintf("Room %d message %d", j, i),
				RoomIDs: []int{j},
			})
		}
	}
	return &messagesRepoMock{messages}
}

func (r messagesRepoMock) CheckIfExists(id int) bool {
	return true
}

func (r messagesRepoMock) GetOne(id int) (interface{}, error) {
	return nil, nil
}

func (r messagesRepoMock) GetAll() []interface{} {
	return nil
}

func (r messagesRepoMock) GetAllWhere(column string, roomid int, limit int) (history []interface{}) {
	return append(history, r.messages[roomid][len(r.messages[roomid])-1-limit:len(r.messages[roomid])-1])
}

func (r messagesRepoMock) Create(interface{}) (int, error) {
	return 0, nil
}

type clientMock struct {
	userID              int
	join                chan ClientJoinsRooms
	leave               chan ClientRooms
	channelMessageQueue chan ClientMessage
	forwardMessage      chan *message.BroadcastedMessage
	nextMessageToRead   *message.ReceivedMessage
	messages            []*message.BroadcastedMessage
}

func NewClientMock(userID int, join chan ClientJoinsRooms, leave chan ClientRooms, channelMessageQueue chan ClientMessage) *clientMock {
	c := &clientMock{
		userID:              userID,
		join:                join,
		leave:               leave,
		channelMessageQueue: channelMessageQueue,
		forwardMessage:      make(chan *message.BroadcastedMessage),
	}
	go c.Write()
	return c
}

func (c *clientMock) Read() {
	switch c.nextMessageToRead.Action {
	case "join":
		historyLimit := 30
		givenHistoryLimit, err := strconv.Atoi(c.nextMessageToRead.Content)
		if err == nil {
			historyLimit = givenHistoryLimit
		}
		c.join <- ClientJoinsRooms{
			ClientRooms: ClientRooms{
				Client: c,
				Rooms:  c.nextMessageToRead.RoomIDs,
			},
			HistoryLimit: historyLimit}
	case "leave":
		c.leave <- ClientRooms{
			Client: c,
			Rooms:  c.nextMessageToRead.RoomIDs,
		}
	case "message":
		c.channelMessageQueue <- ClientMessage{c, c.nextMessageToRead}
	}
}

func (c *clientMock) Write() {
	for msg := range c.ForwardMessage() {
		c.messages = append(c.messages, msg)
	}
}

func (c clientMock) ForwardMessage() chan *message.BroadcastedMessage {
	return c.forwardMessage
}

func (c clientMock) GetUserID() int {
	return c.userID
}

func (c clientMock) Close() {

}

func (c *clientMock) setNextMessageToRead(rm *message.ReceivedMessage) {
	c.nextMessageToRead = rm
}

func (c *clientMock) setNextMessageToReadAndRead(rm *message.ReceivedMessage) {
	c.nextMessageToRead = rm
	c.Read()
}

func (c clientMock) joinRoomsMessage(historyLimit int, roomIDs ...int) *message.ReceivedMessage {
	return &message.ReceivedMessage{
		UserID:  c.userID,
		Action:  "join",
		RoomIDs: roomIDs,
		Content: strconv.Itoa(historyLimit),
	}
}

func (c clientMock) leaveRoomsMessage(roomIDs ...int) *message.ReceivedMessage {
	return &message.ReceivedMessage{
		UserID:  c.userID,
		Action:  "leave",
		RoomIDs: roomIDs,
	}
}

func (c clientMock) sendMessageToRooms(content string, roomIDs ...int) *message.ReceivedMessage {
	return &message.ReceivedMessage{
		UserID:  c.userID,
		Action:  "message",
		RoomIDs: roomIDs,
		Content: content,
	}
}

func (c *clientMock) clearMessages() {
	c.messages = nil
}
