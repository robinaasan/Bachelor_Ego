package main

import (
	"bytes"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/edgelesssys/ego/attestation"
	"github.com/edgelesssys/ego/eclient"
)

type Transaction struct {
	Key        int    `json:"Key"`
	NewVal     int    `json:"NewVal"`
	ClientName string `json:"ClientName"`
}

type SetValue struct {
	Key    int
	NewVal int
}

type ResponsesRuntime struct {
	response string
	endpoint string
	err      error
	vals     SetValue
	duration time.Duration
}

// var endpoints = []string{"http://localhost:8087", "http://localhost:8087", "http://localhost:8087"}

func main() {
	// setval := &SetValue{2, 1}

	uniqueID, _ := hex.DecodeString("4fb6dfaa42cb68d4f07e67f59b2eb6ad39c4ed50cf4c7d6ffea0b8075800483a")

	verifyReport := func(report attestation.Report) error {
		if !bytes.Equal(report.UniqueID, uniqueID) {
			return errors.New("invalid UniqueID")
		}
		return nil
	}
	tlsConfig := eclient.CreateAttestationClientTLSConfig(verifyReport)
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}
	// client := &http.Client{}

	wg := &sync.WaitGroup{}
	waitResponses := &sync.WaitGroup{}
	// var mu sync.Mutex
	c := make(chan ResponsesRuntime)
	wg.Add(1)

	const orderingURL = "https://localhost:8086/Add"
	var storeResponse []string

	flag.Parse()
	args := flag.Args()
	username := args[0]

	go func() {
		// Create a new ticker that ticks every 1000 milliseconds
		ticker := time.NewTicker(1000 * time.Microsecond)

		// Create a timer that will stop the ticker after 1 second
		timer := time.NewTimer(1 * time.Second)

		var key, value int
		key = 1

		for {
			select {
			case <-timer.C:
				// Stop the ticker
				ticker.Stop()

				// Signal that the WaitGroup is done
				wg.Done()

				return
			case <-ticker.C:
				// mu.Lock() // This will make sure no request is sent twice but dont need this for tesing
				wg.Add(1)
				waitResponses.Add(1)
				value += 2
				key++
				go sendToRuntime(key, value, orderingURL, wg, client, c, time.Now(), username)
				// fmt.Println(statusUpdate())
			}
		}
	}()

	go func() {
		for res := range c {
			fmt.Printf("%v, duration: %v\n", res.vals, res.duration.Microseconds())
			storeResponse = append(storeResponse, strconv.FormatInt(res.duration.Microseconds(), 10))
			waitResponses.Done()
		}
	}()
	// Wait for all goroutines to finish
	wg.Wait()
	fmt.Println("finished calling endpoint(s)...")

	// wait for all responses to be added to the list...
	waitResponses.Wait()
	err := storeDataInFile(&storeResponse)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	// time.Sleep(1 * time.Second)
	// ticker.Stop()
	// done <- true
}

func storeDataInFile(data *[]string) error {
	f, err := os.OpenFile("./storeResponseInFile.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o777)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	for _, times := range *data {
		if _, err := f.WriteString(times + ", "); err != nil {
			panic(err)
		}
	}

	// timeDiff := endTime.Sub(time.UnixMicro(blockFromTransactions.TimeStamp)).Microseconds()

	// os.Remove("storeResponseInFile.txt")
	// err := os.WriteFile("storeResponseInFile.txt", []byte(toString(data)), 0o777)
	// if err != nil {
	// 	return err
	// }
	fmt.Println("Success writing to the file!")
	return nil
}

func toString(data *[]string) string {
	return strings.Join([]string(*data), ",")
}

// func decrementCounter(wg *sync.WaitGroup) {
// 	defer (*wg).Done()
// }

func sendToRuntime(key int, value int, endpoint string, wg *sync.WaitGroup, runtime *http.Client, c chan ResponsesRuntime, tm time.Time, username string) {
	defer wg.Done()
	// fmt.Println(value)

	// t := Transaction{
	// 	ClientName: "robin",
	// 	Key:        key,
	// 	NewVal:     value,
	// }
	// q := url.Values{}
	// body := map[string]int{"Key": setvalues.Key, "NewVal": setvalues.NewVal, "OldVal": setvalues.OldVal}
	// q.Add("client", nameClient)
	// jsonBody, err := json.Marshal(t)
	// if err != nil {
	// 	c <- ResponsesRuntime{endpoint, "", err, SetValue{}, 0}
	// 	return
	// }
	q := &url.Values{}
	q.Add("username", username)
	q.Add("key", fmt.Sprint(key))
	q.Add("value", fmt.Sprint(value))

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		c <- ResponsesRuntime{endpoint, "", err, SetValue{}, 0}
		return
	}
	// req.Header.Add("Content-Type", "application/json")
	req.URL.RawQuery = q.Encode()
	res, err := runtime.Do(req)
	if err != nil {
		c <- ResponsesRuntime{endpoint, "", err, SetValue{}, 0}
		return
	}
	defer res.Body.Close()
	// responseData, err := io.ReadAll(res.Body)
	// if err != nil {
	// 	c <- ResponsesRuntime{endpoint, "", err, SetValue{}, 0}
	// 	panic("Error posting")
	// }
	// fmt.Println(string(responseData))
	var duration time.Duration = time.Since(tm)
	c <- ResponsesRuntime{endpoint, res.Status, nil, SetValue{Key: key, NewVal: value}, duration}
}
