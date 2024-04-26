package pkg

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

type messageUnit struct {
	ClientName        string
	MessageBody       string
	MessageUniqueCode int
	ClientUniqueCode  int
}

type messageHandle struct {
	MQue []messageUnit
	mu   sync.Mutex
}

// @todo global variable messageHandleObject error check
// var messageHandleObject = messageHandle{}

type ChatServer struct {
	ServicesServer
	ChatRooms map[int32]*messageHandle // -> chat room
}

type RoomNumber int32

// define ChatService
func (is *ChatServer) ChatService(csi Services_ChatServiceServer) error {
	roomNumber := csi.Context().Value(RoomNumber(10)).(int32)
	chatRoom := is.ChatRooms[roomNumber]

	// @todo rand.Intn(1e6) clientUniqueCode duplication check
	clientUniqueCode := rand.Intn(1e6)
	// clientUniqueCode := clientUniqueCode_
	errch := make(chan error)

	// receive messages - init a go routine
	go receiveFromStream(csi, chatRoom, clientUniqueCode, errch)

	// send messages - init a go routine
	go sendToStream(csi, chatRoom, clientUniqueCode, errch)

	return <-errch

}

// receive messages
func receiveFromStream(csi_ Services_ChatServiceServer, messageHandleObject_ *messageHandle, clientUniqueCode_ int, errch_ chan error) {

	//implement a loop
	for {
		mssg, err := csi_.Recv()
		if err != nil {
			log.Printf("Error in receiving message from client :: %v", err)
			errch_ <- err
		} else {

			messageHandleObject_.mu.Lock()

			messageHandleObject_.MQue = append(messageHandleObject_.MQue, messageUnit{
				ClientName:        mssg.Name,
				MessageBody:       mssg.Body,
				MessageUniqueCode: rand.Intn(1e8),
				ClientUniqueCode:  clientUniqueCode_,
			})

			log.Printf("%v", messageHandleObject_.MQue[len(messageHandleObject_.MQue)-1])

			messageHandleObject_.mu.Unlock()

		}
	}
}

// send message
func sendToStream(csi_ Services_ChatServiceServer, messageHandleObject_ *messageHandle, clientUniqueCode_ int, errch_ chan error) {

	//implement a loop
	for {

		//loop through messages in MQue
		for {
			// @todo time.Sleep efficiency check
			time.Sleep(500 * time.Millisecond)

			messageHandleObject_.mu.Lock()

			if len(messageHandleObject_.MQue) == 0 {
				messageHandleObject_.mu.Unlock()
				break
			}

			senderUniqueCode := messageHandleObject_.MQue[0].ClientUniqueCode
			senderName4Client := messageHandleObject_.MQue[0].ClientName
			message4Client := messageHandleObject_.MQue[0].MessageBody

			messageHandleObject_.mu.Unlock()

			//send message to designated client (do not send to the same client)
			if senderUniqueCode != clientUniqueCode_ {

				err := csi_.Send(&FromServer{Name: senderName4Client, Body: message4Client})

				if err != nil {
					errch_ <- err
				}

				messageHandleObject_.mu.Lock()

				if len(messageHandleObject_.MQue) > 1 {
					messageHandleObject_.MQue = messageHandleObject_.MQue[1:] // delete the message at index 0 after sending to receiver
				} else {
					messageHandleObject_.MQue = []messageUnit{}
				}

				messageHandleObject_.mu.Unlock()

			}

		}

		time.Sleep(100 * time.Millisecond)
	}
}
