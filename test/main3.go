// package main

// import (
// 	"bytes"
// 	"encoding/hex"
// 	"errors"
// 	"flag"
// 	"fmt"
// 	"net/http"
// 	"net/url"
// 	"os"
// 	"strings"
// 	"sync"
// 	"time"

// 	"github.com/edgelesssys/ego/attestation"
// 	"github.com/edgelesssys/ego/eclient"
// )

// type Transaction struct {
// 	Key        int    `json:"Key"`
// 	NewVal     int    `json:"NewVal"`
// 	ClientName string `json:"ClientName"`
// }

// type SetValue struct {
// 	Key    int
// 	NewVal int
// }

// type ResponsesRuntime struct {
// 	err      error
// 	key      int
// 	duration time.Duration
// }

// // var endpoints = []string{"http://localhost:8087", "http://localhost:8087", "http://localhost:8087"}

// func main() {
// 	// setval := &SetValue{2, 1}

// 	uniqueID, _ := hex.DecodeString("6ae31df74b9a9d4ada73f9a15f260a3faf03ad40ae08f43ec5297cd338446c26")

// 	verifyReport := func(report attestation.Report) error {
// 		if !bytes.Equal(report.UniqueID, uniqueID) {
// 			return errors.New("invalid UniqueID")
// 		}
// 		return nil
// 	}
// 	tlsConfig := eclient.CreateAttestationClientTLSConfig(verifyReport)
// 	client := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}
// 	//client := &http.Client{}

// 	wg := &sync.WaitGroup{}
// 	//waitResponses := &sync.WaitGroup{}
// 	// var mu sync.Mutex
// 	//c := make(chan ResponsesRuntime)
// 	wg.Add(1)

// 	const orderingURL = "https://localhost:8086/Add"
// 	//const stopURL = "https://localhost:8086/Stop"
// 	//var storeResponse []string

// 	flag.Parse()
// 	args := flag.Args()
// 	username := args[0]
// 	var key, value int
// 	key = 0

// 	go func() {
// 		// Create a new ticker that ticks every 1000 milliseconds
// 		ticker := time.NewTicker(250 * time.Microsecond)

// 		// Create a timer that will stop the ticker after 1 second
// 		timer := time.NewTimer(1 * time.Second)

// 		for {
// 			select {
// 			case <-timer.C:
// 				// Stop the ticker
// 				ticker.Stop()

// 				// Signal that the WaitGroup is done
// 				wg.Done()

// 				return
// 			case <-ticker.C:
// 				wg.Add(1)
// 				value += 2
// 				key++
// 				go sendToRuntime(key, value, orderingURL, wg, client, username)
// 				// fmt.Println(statusUpdate())
// 			}
// 		}
// 	}()

// 	// go func() {
// 	// 	for res := range c {
// 	// 		fmt.Printf("%v, duration: %v\n", res.vals, res.duration.Microseconds())
// 	// 		storeResponse = append(storeResponse, strconv.FormatInt(res.duration.Microseconds(), 10))
// 	// 		waitResponses.Done()
// 	// 	}
// 	// }()
// 	// Wait for all goroutines to finish
// 	wg.Wait()

// 	fmt.Println("finished calling endpoint(s)...")

// 	// wait for all responses to be added to the list...

// 	// req, err := http.NewRequest("GET", stopURL, nil)

// 	// if err != nil {
// 	// 	fmt.Println(err)
// 	// }
// 	// _, err = client.Do(req)

// 	// if err != nil {
// 	// 	fmt.Println(err)
// 	// }
// 	// fmt.Println("called to stop...")

// 	//err := storeDataInFile(&storeResponse)
// 	// if err != nil {
// 	// 	fmt.Printf("Error: %v\n", err)
// 	// }

// 	// time.Sleep(1 * time.Second)
// 	// ticker.Stop()
// 	// done <- true
// }

// func storeDataInFile(data *[]string) error {
// 	os.Remove("storeResponseInFile.txt")
// 	err := os.WriteFile("storeResponseInFile.txt", []byte(toString(data)), 0o777)
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println("Success writing to the file!")
// 	return nil
// }

// func toString(data *[]string) string {
// 	return strings.Join([]string(*data), ",")
// }

// // func decrementCounter(wg *sync.WaitGroup) {
// // 	defer (*wg).Done()
// // }

// func sendToRuntime(key int, value int, endpoint string, wg *sync.WaitGroup, runtime *http.Client, username string) {
// 	defer wg.Done()
// 	// fmt.Println(value)

// 	// t := Transaction{
// 	// 	ClientName: "robin",
// 	// 	Key:        key,
// 	// 	NewVal:     value,
// 	// }
// 	// q := url.Values{}
// 	// body := map[string]int{"Key": setvalues.Key, "NewVal": setvalues.NewVal, "OldVal": setvalues.OldVal}
// 	// q.Add("client", nameClient)
// 	// jsonBody, err := json.Marshal(t)
// 	// if err != nil {
// 	// 	c <- ResponsesRuntime{endpoint, "", err, SetValue{}, 0}
// 	// 	return
// 	// }
// 	q := &url.Values{}
// 	q.Add("username", username)
// 	q.Add("key", fmt.Sprint(key))
// 	q.Add("value", fmt.Sprint(value))

// 	req, err := http.NewRequest("GET", endpoint, nil)
// 	if err != nil {
// 		//c <- ResponsesRuntime{endpoint, "", err, SetValue{}, 0}
// 		return
// 	}
// 	// //req.Header.Add("Content-Type", "application/json")
// 	req.URL.RawQuery = q.Encode()
// 	_, err = runtime.Do(req)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	// defer res.Body.Close()
// 	// // responseData, err := io.ReadAll(res.Body)
// 	// // if err != nil {
// 	// // 	c <- ResponsesRuntime{endpoint, "", err, SetValue{}, 0}
// 	// // 	panic("Error posting")
// 	// // }
// 	// // fmt.Println(string(responseData))
// 	// var duration time.Duration = time.Since(tm)
// 	// c <- ResponsesRuntime{endpoint, res.Status, nil, SetValue{Key: key, NewVal: value}, duration}
// }