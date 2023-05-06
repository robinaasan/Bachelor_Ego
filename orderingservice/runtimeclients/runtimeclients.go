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
	Send chan SendBackToRuntime // Message channel for each runtime
}

// Transaction data from the runtime
type TransactionContent struct {
	Key        int    `json:"Key"`
	NewVal     int    `json:"NewVal"`
	OldVal     int    `json:"OldVal"`
	ClientName string `json:"ClientName"` //Name/hash of the vendor
}

// Content from message from the runtime
type messageFromRuntime struct {
	TransactionContent `json:"TransactionContent"`
	MessageId          string `json:"MessageId"`
	ClientHash         string `json:"ClientHash"` //Name/hash of the vendor
}

type BlockFromTransactions struct {
	TransactionContentSlice []TransactionContent
	BroadcastToRuntimes     bool
	Runtimeclient           *Runtimeclient
	MessageId               string
	ClientHash              string
}

// struct for sending message back to a runtime
type SendBackToRuntime struct {
	TransactionContentSlice []TransactionContent `json:"TransactionContentSlice"`
	ACK                     bool                 `json:"ACK"` // The runtime who sent the last transaction should recieve a message back inlcuding an ACK
	MessageId               string               `json:"MessageId"`
	ClientHash              string               `json:"ClientHash"`
}

// goroutine to handle sending messages to a single client, this only sends the created blocks
func (rc *Runtimeclient) WritePump() {
	for {
		select {
		case message := <-rc.Send:

			dataToSendRuntime, err := json.Marshal(message)
			if err != nil {
				panic("Error marshalling data to send to runtime")
			}

			err = rc.Conn.WriteMessage(websocket.TextMessage, dataToSendRuntime)
			if err != nil {
				fmt.Println("Error writing to runtime", err)
				rc.Conn.Close()
				continue
			}
		}
	}
}

// goroutine to handle receiving messages from a single runtime
func (rc *Runtimeclient) ReadPump(blockSize int, allTransactions *[]TransactionContent, mu *sync.Mutex, createdBlockFromTransactions chan BlockFromTransactions) {
	var count int
	for {
		_, message, err := rc.Conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading from runtime", err)
			rc.Conn.Close()
			return
		}

		m := &messageFromRuntime{}
		err = json.Unmarshal(message, m)
		if err != nil {
			log.Println(err)
			fmt.Println("Error reading from runtime", err)
			continue
		}

		mu.Lock()
		count++
		(*allTransactions) = append((*allTransactions), m.TransactionContent)
		if count >= blockSize {
			count = 0
			if err != nil {
				log.Println(err)
				return
			}

			// createdBlockFromTransactions is a channel used in main to wait for created blocks
			createdBlockFromTransactions <- BlockFromTransactions{TransactionContentSlice: *allTransactions, BroadcastToRuntimes: true, Runtimeclient: rc, MessageId: m.MessageId, ClientHash: m.ClientHash}
			(*allTransactions) = []TransactionContent{}
		} else {
			// send empty slice as acknowledgement to the runtime who send the last message
			createdBlockFromTransactions <- BlockFromTransactions{TransactionContentSlice: []TransactionContent{}, BroadcastToRuntimes: false, Runtimeclient: rc, MessageId: m.MessageId, ClientHash: m.ClientHash}
		}
		mu.Unlock()
	}
}

// send a message to all connected runtimes
func BroadcastMessage(message *SendBackToRuntime, allruntimeclients []Runtimeclient, mu *sync.Mutex) {
	mu.Lock()
	defer mu.Unlock()
	for _, client := range allruntimeclients {
		select {
		case client.Send <- *message:
		default:
			// TODO: Error handling not implemented
			log.Println("No message recieved")
			return
		}
	}
}
