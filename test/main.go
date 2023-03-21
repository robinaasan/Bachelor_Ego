package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Transaction struct {
	Key        int    `json:"Key"`
	NewVal     int    `json:"NewVal"`
	OldVal     int    `json:"OldVal"`
	ClientName string `json:"ClientName"`
}

type SetValue struct {
	Key    int
	NewVal int
	OldVal int
}

type ResponsesRuntime struct {
	response string
	endpoint string
	err      error
	vals     SetValue
}

const orderingURL = "http://localhost:8087"

// var endpoints = []string{"http://localhost:8087", "http://localhost:8087", "http://localhost:8087"}

func main() {
	setvals := []SetValue{{2, 3, 4}, {3, 4, 5}, {6, 7, 8}, {9, 10, 11}, {12, 13, 14}}

	cl := &http.Client{}
	var wg sync.WaitGroup
	c := make(chan ResponsesRuntime)
	for _, setval := range setvals {
		wg.Add(1)
		go sendToOrdering(orderingURL, setval, &wg, cl, c)
	}
	go func() {
		wg.Wait()
		close(c)
	}()

	for r := range c {
		// if r.err != nil {

		// 	s := fmt.Sprintf("Error: endpoint: %s got: %v\n", r.endpoint, r.err)
		// 	fmt.Printf("%v", s)
		// } else {
		// 	fmt.Println(r.response + "\n")
		// }

		// if r.err != nil {
		// 	fmt.Printf("Error requesting %s: %v\n", r.endpoint, r.err)
		// 	continue
		// }
		fmt.Printf("%+v, %v\n", r.response, time.Now().Nanosecond())
	}
}

func sendToOrdering(endpoint string, setvalues SetValue, wg *sync.WaitGroup, runtime *http.Client, c chan ResponsesRuntime) {
	defer (*wg).Done()
	t := Transaction{
		ClientName: "robin",
		Key:        setvalues.Key,
		NewVal:     setvalues.NewVal,
		OldVal:     setvalues.OldVal,
	}
	// q := url.Values{}
	// body := map[string]int{"Key": setvalues.Key, "NewVal": setvalues.NewVal, "OldVal": setvalues.OldVal}
	// q.Add("client", nameClient)
	jsonBody, err := json.Marshal(t)
	if err != nil {
		c <- ResponsesRuntime{endpoint, "", err, SetValue{}}
		return
	}
	req, err := http.NewRequest("POST", orderingURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		c <- ResponsesRuntime{endpoint, "", err, SetValue{}}
		return
	}
	req.Header.Add("Content-Type", "application/json")
	// req.URL.RawQuery = q.Encode()
	res, err := runtime.Do(req)
	if err != nil {
		c <- ResponsesRuntime{endpoint, "", err, SetValue{}}
		return
	}
	defer res.Body.Close()
	// responseData, err := io.ReadAll(res.Body)
	if err != nil {
		c <- ResponsesRuntime{endpoint, "", err, SetValue{}}
		return
	}
	// fmt.Println(string(responseData))
	c <- ResponsesRuntime{endpoint, res.Status, nil, setvalues}
}
