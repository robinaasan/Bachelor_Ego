package runtimeclients

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Runtimeclient struct {
	Conn  *websocket.Conn       // Websocket connection for each runtime
	Send  chan []byte           // Message channel for each client
	Ack   map[int]chan struct{} // Map of messages IDs to acknowledgement channel
	Queue []Message
}

type Message struct {
	ID      int
	Payload []byte
}

// Struct for getting the transaction from the runtimes
type TransactionContent struct {
	Key        int    `json:"Key"`
	NewVal     int    `json:"NewVal"`
	OldVal     int    `json:"OldVal"`
	ClientName string `json:"ClientName"`
}

type Transaction struct {
	TransactionContent `json:"TransactionContent"`
	ACK                string `json:"ACK"`
}

type BlockFromTransactions struct {
	TransactionContentSlice []TransactionContent `json:"TransactionContentSlice"`
}

// goroutine to handle sending messages to a single client, this only sends the created blocks
func (rc *Runtimeclient) WritePump() {
	for {
		select {
		case message := <-rc.Send:
			msg := Message{
				ID:      len(rc.Queue),
				Payload: message,
			}

			// Add the message to the queue
			rc.Queue = append(rc.Queue, msg)

			// Send the message and wait for acknowledgement
			ackCh := make(chan struct{})
			rc.Ack[msg.ID] = ackCh
			err := rc.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				fmt.Println("Error writing to runtime", err)
				delete(rc.Ack, msg.ID)
				continue
			}
			select {
			case <-ackCh:
				// Message acknowledged, remove from queue
				rc.Queue = rc.Queue[1:]
			case <-time.After(5 * time.Second):
				// Timeout waiting for acknowledgement, resend message
				fmt.Println("Timeout waiting for acknowledgement, resending message")
				delete(rc.Ack, msg.ID)
				rc.Send <- message
			}
		}
	}
}

// goroutine to handle receiving messages from a single runtime
func (rc *Runtimeclient) ReadPump(blockSize int, allTransactions *[]TransactionContent, mu *sync.Mutex, blockToTime chan BlockFromTransactions) {
	var count int
	for {
		_, message, err := rc.Conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading from runtime", err)
			rc.Conn.Close()
			return
		}
		// the transaction will include some ACK from a message from earlier
		newTransAction := &Transaction{}
		err = json.Unmarshal(message, newTransAction)
		if err != nil {
			log.Println(err)
			fmt.Println("Error reading from runtime", err)
			continue
		}

		// Send acknowledgement for the message
		id, err := strconv.Atoi(string(newTransAction.ACK))
		if err != nil {
			fmt.Println("Invalid message ID", err)
			continue
		}

		// sending ack with 0 means that the runtime is simply sending a transaction with no ack for previous message
		if id != 0 {
			ackCh, ok := rc.Ack[id]
			if !ok {
				fmt.Println("Unknown message ID", id)
				continue
			}
			close(ackCh)
			delete(rc.Ack, id)
		}

		// send the transaction
		mu.Lock()
		count++
		(*allTransactions) = append((*allTransactions), newTransAction.TransactionContent)
		// fmt.Printf("%+v\n", newTransAction) //TODO: remove
		if count >= blockSize {
			count = 0
			// allTransactionBytes, err := json.Marshal(allTransactions)
			if err != nil {
				log.Println(err)
				return
			}
			blockFromTransactions := BlockFromTransactions{TransactionContentSlice: *allTransactions}
			blockToTime <- blockFromTransactions
			// timerChan <- timeNow
			// createdBlock <- allTransactionBytes

			// wait for the blocks to be created and broadcasted
			(*allTransactions) = []TransactionContent{}
		}
		mu.Unlock()
	}
}

// send a message to all connected runtimez
func BroadcastMessage(message []byte, allruntimeclients []Runtimeclient, mu *sync.Mutex) {
	mu.Lock()
	defer mu.Unlock()
	for _, client := range allruntimeclients {
		select {
		case client.Send <- message:
			// fmt.Println("Callback:", string(message))
		default:
			// TODO: handle error (runtmime disconnected)
			log.Println("Check when this runs")
		}
	}
}
