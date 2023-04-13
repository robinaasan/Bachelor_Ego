package runtimeclients

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Runtimeclient struct {
	Conn *websocket.Conn // Websocket connection for each runtime
	Send chan []byte     // Message channel for each client
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
	TimeStamp          int64 `json:"TimeStamp"`
}

type BlockFromTransactions struct {
	TransactionContentSlice []TransactionContent `json:"TransactionContentSlice"`
	TimeStamp               int64                `json:"TimeStamp"`
}

// goroutine to handle sending messages to a single client, this only sends the created blocks
func (rc *Runtimeclient) WritePump() {
	for {
		select {
		case message, ok := <-rc.Send:
			if !ok {
				// channel closed, client disconnected
				return
			}
			//(*timeEvaluation) = append((*timeEvaluation), time.Since(*timer))
			err := rc.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				fmt.Println("Error writing to runtime", err)
				continue
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
			log.Println(err)
			break
		}

		newTransAction := &Transaction{}
		err = json.Unmarshal(message, newTransAction)
		if err != nil {
			log.Println(err)
			fmt.Println("Error reading from runtime", err)
			continue
		}
		mu.Lock()
		count++
		(*allTransactions) = append((*allTransactions), newTransAction.TransactionContent)
		//fmt.Printf("%+v\n", newTransAction) //TODO: remove
		if count >= blockSize {
			count = 0
			//allTransactionBytes, err := json.Marshal(allTransactions)
			if err != nil {
				log.Println(err)
				return
			}
			blockFromTransactions := BlockFromTransactions{TransactionContentSlice: *allTransactions, TimeStamp: newTransAction.TimeStamp}
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
			//fmt.Println("Callback:", string(message))
		default:
			// TODO: handle error (runtmime disconnected)
			log.Println("Check when this runs")
		}
	}
}
