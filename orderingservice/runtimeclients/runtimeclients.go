package runtimeclients

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type Runtimeclient struct {
	Conn *websocket.Conn // Websocket connection for each runtime
	Send chan []byte     // Message channel for each client
}

// Struct for getting the transaction from the runtimes
type Transaction struct {
	Key        int    `json:"Key"`
	NewVal     int    `json:"NewVal"`
	OldVal     int    `json:"OldVal"`
	ClientName string `json:"ClientName"`
}

// goroutine to handle sending messages to a single client
func (rc *Runtimeclient) WritePump() {
	for {
		select {
		case message, ok := <-rc.Send:
			if !ok {
				// channel closed, client disconnected
				return
			}
			err := rc.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

// goroutine to handle receiving messages from a single runtime
func (rc *Runtimeclient) ReadPump(blockSize int, allTransactions *[]Transaction, createdBlock chan []byte, done chan bool) {
	var count int
	for {
		_, message, err := rc.Conn.ReadMessage()
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

		count++
		(*allTransactions) = append((*allTransactions), *newTransAction)
		fmt.Printf("%+v\n", newTransAction) //TODO: remove
		if count >= blockSize {
			count = 0
			allTransactionBytes, err := json.Marshal(allTransactions)
			if err != nil {
				log.Println(err)
				return
			}
			createdBlock <- allTransactionBytes
			// wait for the blocks to be created and broadcasted
			<-done 
			(*allTransactions) = []Transaction{}
		}
	}
}

// send a message to all connected runtimez
func BroadcastMessage(message []byte, allruntimeclients []Runtimeclient) {
	for _, client := range allruntimeclients {
		select {
		case client.Send <- message:
			fmt.Println(string(message))
		default:
			// TODO: handle error (runtmime disconnected)
			log.Println("Check when this runs")
		}
	}
}
