package runtimeclients

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Runtimeclient struct {
	Conn *websocket.Conn        // Websocket connection for each runtime
	Send chan SendBackToRuntime // Message channel for each client
}

// Struct for getting the transaction from the runtimes
type TransactionContent struct {
	Key        int    `json:"Key"`
	NewVal     int    `json:"NewVal"`
	OldVal     int    `json:"OldVal"`
	ClientName string `json:"ClientName"`
}

type BlockFromTransactions struct {
	TransactionContentSlice []TransactionContent
	BroadcastToRuntimes     bool
	Runtimeclient           *Runtimeclient
}

type SendBackToRuntime struct {
	TransactionContentSlice []TransactionContent `json:"TransactionContentSlice"`
}

// goroutine to handle sending messages to a single client, this only sends the created blocks
func (rc *Runtimeclient) WritePump() {
	for {
		select {
		case message := <-rc.Send:

			// Send the message and wait for acknowledgement
			dataToSend, err := json.Marshal(message)
			if err != nil {
				panic("Error marshalling data to send to runtime")
			}
			
			err = rc.Conn.WriteMessage(websocket.TextMessage, dataToSend)
			if err != nil {
				fmt.Println("Error writing to runtime", err)
				continue
			}
		}
	}
}

// goroutine to handle receiving messages from a single runtime
func (rc *Runtimeclient) ReadPump(blockSize int, allTransactions *[]TransactionContent, mu *sync.Mutex, createdblockFromTransactions chan BlockFromTransactions) {
	var count int
	for {
		_, message, err := rc.Conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading from runtime", err)
			rc.Conn.Close()
			return
		}
		// the transaction will include some ACK from a message from earlier
		newTransAction := &TransactionContent{}
		err = json.Unmarshal(message, newTransAction)
		if err != nil {
			log.Println(err)
			fmt.Println("Error reading from runtime", err)
			continue
		}
		// send the transaction
		mu.Lock()
		count++
		(*allTransactions) = append((*allTransactions), *newTransAction)
		if count >= blockSize {
			count = 0
			// allTransactionBytes, err := json.Marshal(allTransactions)
			if err != nil {
				log.Println(err)
				return
			}

			createdblockFromTransactions <- BlockFromTransactions{TransactionContentSlice: *allTransactions, BroadcastToRuntimes: true, Runtimeclient: rc}
			(*allTransactions) = []TransactionContent{}
		} else {
			createdblockFromTransactions <- BlockFromTransactions{TransactionContentSlice: []TransactionContent{}, BroadcastToRuntimes: false, Runtimeclient: rc}
		}
		mu.Unlock()
	}
}

// send a message to all connected runtimez
func BroadcastMessage(message *SendBackToRuntime, allruntimeclients []Runtimeclient, mu *sync.Mutex) {
	mu.Lock()
	defer mu.Unlock()
	for _, client := range allruntimeclients {
		select {
		case client.Send <- *message:
			// fmt.Println("Callback:", string(message))
		default:
			// TODO: handle error (runtmime disconnected)
			log.Println("Check when this runs")
		}
	}
}
