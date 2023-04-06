package runtimeclients

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Runtimeclient struct {
	Conn *websocket.Conn
	Send chan []byte
	mu   *sync.Mutex
}

// name below should be replaces by som hash later
type Transaction struct {
	Key        int    `json:"Key"`
	NewVal     int    `json:"NewVal"`
	OldVal     int    `json:"OldVal"`
	ClientName string `json:"ClientName"`
}

// goroutine to handle sending messages to a single client
func (cl *Runtimeclient) WritePump() {
	for {
		select {
		case message, ok := <-cl.Send:
			if !ok {
				// channel closed, client disconnected
				return
			}
			err := cl.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				// handle error
			}
		}
	}
}

// goroutine to handle receiving messages from a single client
func (cl *Runtimeclient) ReadPump(count int, allTransactions *[]Transaction, createdBlock chan []byte) {
	for {
		_, message, err := cl.Conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
		newTransAction := &Transaction{}
		err = json.Unmarshal(message, newTransAction)
		if err != nil {
			log.Println(err)
			continue
		}
		//cl.mu.Lock()
		count++
		if count < 5 {

			(*allTransactions) = append((*allTransactions), *newTransAction)
			fmt.Printf("%+v\n", newTransAction)
			//fmt.Printf("%+v", newTransAction)
		} else {
			count = 0
			allTransactionBytes, err := json.Marshal(allTransactions)
			if err != nil {
				//TODO:handle error
				return
			}
			createdBlock <- allTransactionBytes

			//allTransactions = nil
		}
		//cl.mu.Unlock()
		//err = conn.WriteMessage(websocket.TextMessage, []byte("ACK"))
		cl.Send <- []byte("ACK")
	}
}

// send a message to all connected clients
func BroadcastMessage(message []byte, allruntimeclients []Runtimeclient) {
	for _, client := range allruntimeclients {
		select {
		case client.Send <- message:
		default:
			// handle error (client disconnected)
		}
	}
}
